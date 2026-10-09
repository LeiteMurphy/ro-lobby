// Package characters guarda as regras dos personagens do Usuário (spec personagens,
// RN-01 a RN-15, design D-06, D-08 e D-09).
package characters

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/catalog"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
)

const (
	// MaxPerUser é o limite de personagens por Usuário (RN-11).
	MaxPerUser = 10
	// MaxNickLength e as faixas abaixo são as regras dos campos (RN-04, RN-07, RN-09).
	MaxNickLength = 24
	MinLevel      = 1
	MaxLevel      = 275
	MaxLinkLength = 300

	nickIndex = "characters_nick_key"
)

var (
	// ErrNotFound vale para personagem inexistente e para personagem de outro Usuário,
	// sem diferença (RN-02, D-08).
	ErrNotFound = errors.New("characters: personagem não encontrado")
	// ErrLimitReached: o Usuário já tem MaxPerUser personagens (RN-11).
	ErrLimitReached = errors.New("characters: limite de personagens atingido")
	// ErrInOpenLobby: o personagem é dono de um lobby aberto, ou tem candidatura pendente
	// ou aceita num lobby aberto, e não pode ser excluído nem mudar de nível ou função
	// (RN-21 da spec lobbies, RN-25 e RN-26 da candidatura-lobby).
	ErrInOpenLobby = errors.New("characters: personagem em lobby aberto")
)

// Campos e códigos de erro de validação (D-07). O web traduz para pt-BR.
const (
	FieldNick     = "nick"
	FieldClassID  = "classId"
	FieldLevel    = "level"
	FieldRole     = "role"
	FieldPortrait = "portrait"
	FieldLink     = "link"

	CodeRequired = "required"
	CodeTooLong  = "too_long"
	CodeInvalid  = "invalid"
	CodeTaken    = "taken"
)

// FieldError é um erro num campo do personagem.
type FieldError struct {
	Field string
	Code  string
}

// ValidationError junta os erros de campo de um pedido (RN-19).
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	parts := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		parts[i] = f.Field + ": " + f.Code
	}
	return "characters: dados inválidos (" + strings.Join(parts, ", ") + ")"
}

// Roles são as funções aceitas (RN-08).
var Roles = []string{"tank", "support", "dps"}

// Input é o que o Usuário informa ao criar ou editar. Portrait e Link vazios valem como
// "não informado".
type Input struct {
	Nick     string
	ClassID  string
	Level    int
	Role     string
	Portrait string
	Link     string
}

// Character é um personagem salvo.
type Character struct {
	ID        string
	Nick      string
	ClassID   string
	Level     int
	Role      string
	Portrait  string
	Link      string
	IsMain    bool
	CreatedAt time.Time
	// Availability é a disponibilidade no banco de talentos; nil se nunca entrou. Só a
	// lista do Usuário traz (spec banco-de-talentos, CA-01.4).
	Availability *talents.Availability
}

type Service struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	// Now é o relógio do serviço; os testes trocam por um relógio fixo.
	Now func() time.Time
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, queries: db.New(pool), Now: time.Now}
}

// List devolve os personagens do Usuário, o principal primeiro e os outros por ordem de
// cadastro (RN-01, RN-17).
func (s *Service) List(ctx context.Context, userID string) ([]Character, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("characters: id de usuário inválido: %w", err)
	}
	rows, err := s.queries.ListCharacters(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("characters: listar: %w", err)
	}
	avail, err := talents.ByUser(ctx, s.queries, uid)
	if err != nil {
		return nil, err
	}
	out := make([]Character, len(rows))
	for i, r := range rows {
		out[i] = toCharacter(r)
		if a, ok := avail[out[i].ID]; ok {
			out[i].Availability = &a
		}
	}
	return out, nil
}

// Create valida e cadastra um personagem. O primeiro do Usuário nasce principal (RN-12),
// e o limite de 10 vale mesmo com pedidos simultâneos (RN-11, D-06).
func (s *Service) Create(ctx context.Context, userID string, in Input) (Character, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return Character{}, fmt.Errorf("characters: id de usuário inválido: %w", err)
	}
	in, err = Validate(in)
	if err != nil {
		return Character{}, err
	}

	var created db.Character
	err = s.inTx(ctx, uid, func(q *db.Queries) error {
		count, err := q.CountCharacters(ctx, uid)
		if err != nil {
			return err
		}
		if count >= MaxPerUser {
			return ErrLimitReached
		}
		created, err = q.CreateCharacter(ctx, db.CreateCharacterParams{
			UserID:   uid,
			Nick:     in.Nick,
			ClassID:  in.ClassID,
			Level:    int16(in.Level), //nolint:gosec // Validate garante 1..275
			Role:     in.Role,
			Portrait: in.Portrait,
			Link:     optionalText(in.Link),
			IsMain:   count == 0,
			Now:      s.Now().UTC(),
		})
		return err
	})
	if err != nil {
		return Character{}, translate(err, "criar")
	}
	return toCharacter(created), nil
}

// Update valida e grava todos os campos do personagem do Usuário (RN-15).
func (s *Service) Update(ctx context.Context, userID, id string, in Input) (Character, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return Character{}, fmt.Errorf("characters: id de usuário inválido: %w", err)
	}
	cid, err := parseUUID(id)
	if err != nil {
		return Character{}, ErrNotFound
	}
	in, err = Validate(in)
	if err != nil {
		return Character{}, err
	}
	// Em transação com o Usuário travado, para a checagem da RN-25 não correr contra a
	// criação de um lobby nem contra uma candidatura (D-05 das specs lobbies e
	// candidatura-lobby).
	var updated db.Character
	err = s.inTx(ctx, uid, func(q *db.Queries) error {
		current, err := q.GetOwnCharacter(ctx, db.GetOwnCharacterParams{ID: cid, UserID: uid})
		if err != nil {
			return err // pgx.ErrNoRows vira ErrNotFound (RN-02)
		}
		if current.Role != in.Role || int(current.Level) != in.Level {
			if err := s.checkNotInOpenLobby(ctx, q, cid); err != nil {
				return err
			}
		}
		updated, err = q.UpdateCharacter(ctx, db.UpdateCharacterParams{
			ID:       cid,
			UserID:   uid,
			Nick:     in.Nick,
			ClassID:  in.ClassID,
			Level:    int16(in.Level), //nolint:gosec // Validate garante 1..275
			Role:     in.Role,
			Portrait: in.Portrait,
			Link:     optionalText(in.Link),
		})
		return err
	})
	if err != nil {
		return Character{}, translate(err, "editar")
	}
	return toCharacter(updated), nil
}

// checkNotInOpenLobby recusa mexer no personagem que está num lobby aberto, como dono,
// candidato pendente ou membro (RN-21 da lobbies, RN-25 e RN-26 da candidatura-lobby).
func (s *Service) checkNotInOpenLobby(ctx context.Context, q *db.Queries, cid pgtype.UUID) error {
	inLobby, err := q.CharacterInOpenLobby(ctx, db.CharacterInOpenLobbyParams{CharacterID: cid, Now: s.Now().UTC()})
	if err != nil {
		return err
	}
	if inLobby {
		return ErrInOpenLobby
	}
	return nil
}

// Delete exclui o personagem do Usuário. Se era o principal, o mais antigo que sobrar
// vira o principal (RN-14).
func (s *Service) Delete(ctx context.Context, userID, id string) error {
	uid, err := parseUUID(userID)
	if err != nil {
		return fmt.Errorf("characters: id de usuário inválido: %w", err)
	}
	cid, err := parseUUID(id)
	if err != nil {
		return ErrNotFound
	}
	err = s.inTx(ctx, uid, func(q *db.Queries) error {
		if _, err := q.GetOwnCharacter(ctx, db.GetOwnCharacterParams{ID: cid, UserID: uid}); err != nil {
			return err // pgx.ErrNoRows vira ErrNotFound (RN-02)
		}
		if err := s.checkNotInOpenLobby(ctx, q, cid); err != nil {
			return err
		}
		wasMain, err := q.DeleteCharacter(ctx, db.DeleteCharacterParams{ID: cid, UserID: uid})
		if err != nil {
			return err
		}
		if wasMain {
			return q.PromoteOldest(ctx, uid)
		}
		return nil
	})
	return translate(err, "excluir")
}

// SetMain torna o personagem o único principal do Usuário (RN-13).
func (s *Service) SetMain(ctx context.Context, userID, id string) error {
	uid, err := parseUUID(userID)
	if err != nil {
		return fmt.Errorf("characters: id de usuário inválido: %w", err)
	}
	cid, err := parseUUID(id)
	if err != nil {
		return ErrNotFound
	}
	err = s.inTx(ctx, uid, func(q *db.Queries) error {
		if err := q.ClearMain(ctx, uid); err != nil {
			return err
		}
		n, err := q.SetMain(ctx, db.SetMainParams{ID: cid, UserID: uid})
		if err != nil {
			return err
		}
		if n == 0 {
			return ErrNotFound // a transação desfaz o ClearMain
		}
		return nil
	})
	return translate(err, "marcar principal")
}

// Validate normaliza e confere os campos (RN-04 a RN-10). Devolve *ValidationError com
// todos os campos errados de uma vez.
func Validate(in Input) (Input, error) {
	var errs []FieldError
	add := func(field, code string) { errs = append(errs, FieldError{Field: field, Code: code}) }

	in.Nick = strings.TrimSpace(in.Nick)
	switch {
	case in.Nick == "":
		add(FieldNick, CodeRequired)
	case utf8.RuneCountInString(in.Nick) > MaxNickLength:
		add(FieldNick, CodeTooLong)
	case strings.IndexFunc(in.Nick, unicode.IsControl) >= 0:
		add(FieldNick, CodeInvalid)
	}

	if in.ClassID == "" {
		add(FieldClassID, CodeRequired)
	} else if _, ok := catalog.ClassByID(in.ClassID); !ok {
		add(FieldClassID, CodeInvalid)
	}

	if in.Level < MinLevel || in.Level > MaxLevel {
		add(FieldLevel, CodeInvalid)
	}

	switch {
	case in.Role == "":
		add(FieldRole, CodeRequired)
	case !isRole(in.Role):
		add(FieldRole, CodeInvalid)
	}

	if in.Portrait == "" {
		in.Portrait = catalog.DefaultPortrait
	} else if !catalog.HasPortrait(in.Portrait) {
		add(FieldPortrait, CodeInvalid)
	}

	in.Link = strings.TrimSpace(in.Link)
	if in.Link != "" {
		if utf8.RuneCountInString(in.Link) > MaxLinkLength {
			add(FieldLink, CodeTooLong)
		} else if !validLink(in.Link) {
			add(FieldLink, CodeInvalid)
		}
	}

	if len(errs) > 0 {
		return in, &ValidationError{Fields: errs}
	}
	return in, nil
}

// validLink aceita só https:// com host e sem usuário ou senha na URL (D-09).
func validLink(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	return u.Scheme == "https" && u.Host != "" && u.User == nil &&
		strings.HasPrefix(link, "https://") && strings.IndexFunc(link, unicode.IsSpace) < 0
}

func isRole(role string) bool {
	for _, r := range Roles {
		if r == role {
			return true
		}
	}
	return false
}

// inTx roda fn numa transação com o Usuário travado (D-06).
func (s *Service) inTx(ctx context.Context, uid pgtype.UUID, fn func(q *db.Queries) error) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.queries.WithTx(tx)
		if _, err := q.LockUser(ctx, uid); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
		return fn(q)
	})
}

// translate leva os erros do banco para os erros do serviço.
func translate(err error, action string) error {
	var pgErr *pgconn.PgError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pgx.ErrNoRows), errors.Is(err, ErrNotFound):
		return ErrNotFound
	case errors.Is(err, ErrLimitReached):
		return ErrLimitReached
	case errors.Is(err, ErrInOpenLobby):
		return ErrInOpenLobby
	case errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == nickIndex:
		return &ValidationError{Fields: []FieldError{{Field: FieldNick, Code: CodeTaken}}} // RN-05
	default:
		return fmt.Errorf("characters: %s: %w", action, err)
	}
}

func parseUUID(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(s)
	return id, err
}

func optionalText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func toCharacter(c db.Character) Character {
	return Character{
		ID:        c.ID.String(),
		Nick:      c.Nick,
		ClassID:   c.ClassID,
		Level:     int(c.Level),
		Role:      c.Role,
		Portrait:  c.Portrait,
		Link:      c.Link.String,
		IsMain:    c.IsMain,
		CreatedAt: c.CreatedAt,
	}
}
