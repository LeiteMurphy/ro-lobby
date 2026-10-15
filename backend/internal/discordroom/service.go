// Package discordroom guarda as regras da sala do grupo no Discord (spec discord-bot,
// design seção 3, ADR-09): abrir a sala, a fila de efeitos (outbox) e o trabalhador que a
// executa, o lembrete e a remoção.
package discordroom

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordbot"
)

// Bot é o que a sala precisa do Discord; o *discordbot.Client satisfaz.
type Bot interface {
	CreateRoom(ctx context.Context, name string, allow []string) (discordbot.Room, error)
	Allow(ctx context.Context, channelID, userID string) error
	Revoke(ctx context.Context, channelID, userID string) error
	Post(ctx context.Context, channelID, content string, mentions []string) error
	DeleteChannel(ctx context.Context, channelID string) error
	AddMember(ctx context.Context, userID, accessToken string) error
	GuildID() string
}

// Estados da sala (RN-09).
const (
	StatusOpen    = "open"
	StatusClosing = "closing"
	StatusDeleted = "deleted"
)

// Tipos de job da fila (D-02).
const (
	JobSyncAccess = "sync_access"
	JobPost       = "post"
)

// Service é a sala do grupo. Com bot nulo, a feature está desligada (RN-22).
type Service struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	bot     Bot
	log     *slog.Logger
	// OAuth troca o código da conexão ao servidor (RN-01); sem ele, conectar fica desligado.
	OAuth OAuth
	// Origin é a origem pública do web, para o link do lobby nas mensagens (RN-08).
	Origin string
	// Now é o relógio do serviço; os testes trocam por um relógio fixo.
	Now func() time.Time
}

func NewService(pool *pgxpool.Pool, bot Bot) *Service {
	return &Service{pool: pool, queries: db.New(pool), bot: bot, log: slog.Default(), Now: time.Now}
}

// Enabled diz se a sala no Discord está ligada (RN-22).
func (s *Service) Enabled() bool { return s.bot != nil }

// postPayload é o texto pronto de uma mensagem e quem ela menciona (D-06).
type postPayload struct {
	Content  string   `json:"content"`
	Mentions []string `json:"mentions"`
}

// enqueuePost põe uma mensagem na fila da sala, na transação de quem chama (D-02).
func enqueuePost(ctx context.Context, q *db.Queries, lobbyID pgtype.UUID, msg postPayload, now time.Time) error {
	if msg.Mentions == nil {
		msg.Mentions = []string{}
	}
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return q.EnqueueDiscordJob(ctx, db.EnqueueDiscordJobParams{LobbyID: lobbyID, Kind: JobPost, Payload: raw, Now: now})
}

// enqueueSync põe na fila a conferência do acesso de um Usuário à sala (D-02).
func enqueueSync(ctx context.Context, q *db.Queries, lobbyID, userID pgtype.UUID, now time.Time) error {
	return q.EnqueueDiscordJob(ctx, db.EnqueueDiscordJobParams{
		LobbyID: lobbyID, Kind: JobSyncAccess, UserID: userID, Payload: []byte("{}"), Now: now,
	})
}

func (s *Service) inTx(ctx context.Context, fn func(q *db.Queries) error) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error { return fn(s.queries.WithTx(tx)) })
}

func parseUUID(id string) (pgtype.UUID, error) {
	var u pgtype.UUID
	err := u.Scan(id)
	return u, err
}

var errNoRows = pgx.ErrNoRows

func isNoRows(err error) bool { return errors.Is(err, errNoRows) }
