// Package auth cria o Usuário a partir do Discord e cuida da Sessão (spec login-discord,
// ADR-07, RN-04 a RN-11).
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discord"
)

const (
	// SessionTTL: a Sessão vale 30 dias a partir do último uso (RN-09).
	SessionTTL = 30 * 24 * time.Hour
	// TouchInterval: o último uso só é gravado se o registrado tiver mais que isso (RN-09).
	TouchInterval = time.Hour
	tokenBytes    = 32
)

var (
	// ErrInvalidCode e ErrUnavailable vêm do Discord (RN-13).
	ErrInvalidCode = discord.ErrInvalidCode
	ErrUnavailable = discord.ErrUnavailable
	// ErrNoSession: token ausente, malformado, desconhecido, vencido ou de Usuário apagado
	// (RN-10).
	ErrNoSession = errors.New("auth: sem sessão válida")
)

// ProfileFetcher é o que o serviço precisa do Discord; o *discord.Client satisfaz.
type ProfileFetcher interface {
	ProfileFromCode(ctx context.Context, code, redirectURI string) (discord.Profile, error)
}

type User struct {
	ID         string
	Username   string
	GlobalName string
}

type Service struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	discord ProfileFetcher
	// Now é o relógio do serviço; os testes trocam por um relógio fixo (D-06).
	Now func() time.Time
}

func NewService(pool *pgxpool.Pool, fetcher ProfileFetcher) *Service {
	return &Service{pool: pool, queries: db.New(pool), discord: fetcher, Now: time.Now}
}

// Login troca o código pelo perfil do Discord, cria ou atualiza o Usuário (RN-05) e abre
// uma Sessão (RN-07). Se o Discord falhar, nada é criado (RN-13).
func (s *Service) Login(ctx context.Context, code, redirectURI string) (string, User, error) {
	profile, err := s.discord.ProfileFromCode(ctx, code, redirectURI)
	if err != nil {
		return "", User{}, err
	}

	token, err := newToken()
	if err != nil {
		return "", User{}, err
	}

	// AJ-02 (RN-13): o Usuário e a Sessão são gravados juntos; se a Sessão falhar, o
	// primeiro login não deixa um Usuário sem sessão no banco.
	now := s.Now().UTC()
	var row db.User
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.queries.WithTx(tx)
		var err error
		row, err = q.UpsertUserByDiscordID(ctx, db.UpsertUserByDiscordIDParams{
			DiscordID:  profile.ID,
			Username:   profile.Username,
			GlobalName: pgtype.Text{String: profile.GlobalName, Valid: profile.GlobalName != ""},
			Now:        now,
		})
		if err != nil {
			return fmt.Errorf("auth: salvar usuário: %w", err)
		}
		if err := q.CreateSession(ctx, db.CreateSessionParams{TokenHash: HashToken(token), UserID: row.ID, Now: now}); err != nil {
			return fmt.Errorf("auth: criar sessão: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", User{}, err
	}
	return token, toUser(row), nil
}

// Authenticate devolve o Usuário da Sessão e renova o último uso, gravando no máximo uma
// vez por hora (RN-09).
func (s *Service) Authenticate(ctx context.Context, token string) (User, error) {
	if !wellFormed(token) {
		return User{}, ErrNoSession
	}
	now := s.Now().UTC()
	hash := HashToken(token)

	row, err := s.queries.GetActiveSession(ctx, db.GetActiveSessionParams{TokenHash: hash, ValidAfter: now.Add(-SessionTTL)})
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNoSession
	}
	if err != nil {
		return User{}, fmt.Errorf("auth: buscar sessão: %w", err)
	}
	if now.Sub(row.LastUsedAt) > TouchInterval {
		if _, err := s.queries.TouchSession(ctx, db.TouchSessionParams{TokenHash: hash, Now: now, StaleBefore: now.Add(-TouchInterval)}); err != nil {
			return User{}, fmt.Errorf("auth: renovar sessão: %w", err)
		}
	}
	return toUser(row.User), nil
}

// Logout apaga só esta Sessão (RN-11). Token inválido ou já apagado não é erro.
func (s *Service) Logout(ctx context.Context, token string) error {
	if !wellFormed(token) {
		return nil
	}
	return s.queries.DeleteSession(ctx, HashToken(token))
}

// HashToken é o que o banco guarda no lugar do token (RN-07).
func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func newToken() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: gerar token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func wellFormed(token string) bool {
	b, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil && len(b) == tokenBytes
}

func toUser(u db.User) User {
	return User{ID: u.ID.String(), Username: u.Username, GlobalName: u.GlobalName.String}
}
