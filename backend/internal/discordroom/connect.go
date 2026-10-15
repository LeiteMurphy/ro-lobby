package discordroom

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discord"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordbot"
)

// Conexão ao servidor oficial (spec discord-bot, US-01, RN-01 a RN-04, D-03).

// OAuth troca o código da conexão e entrega o token só para incluir no servidor; o
// *discord.Client satisfaz.
type OAuth interface {
	JoinFromCode(ctx context.Context, code, redirectURI, expectedID string,
		join func(ctx context.Context, userID, accessToken string) error) (discord.Profile, error)
}

var (
	// ErrInvalidCode: o Discord recusou o código, ou o token não serve para entrar no
	// servidor (sem o escopo guilds.join).
	ErrInvalidCode = errors.New("discordroom: código da conexão recusado")
)

// CodeWrongAccount: a conta autorizada não é a do login (RN-03).
const CodeWrongAccount = "wrong_account"

// Connection é o estado da conexão do Usuário ao servidor oficial.
type Connection struct {
	Enabled bool
	// ConnectedAt fica zerado se o Usuário nunca se conectou.
	ConnectedAt time.Time
}

// Connection devolve se a sala está ligada e quando o Usuário se conectou (RN-01, RN-22).
func (s *Service) Connection(ctx context.Context, userID string) (Connection, error) {
	out := Connection{Enabled: s.Enabled()}
	uid, err := parseUUID(userID)
	if err != nil {
		return out, ErrNotFound
	}
	row, err := s.queries.GetDiscordConnection(ctx, uid)
	if isNoRows(err) {
		return out, ErrNotFound
	}
	if err != nil {
		return out, fmt.Errorf("discordroom: conexão: %w", err)
	}
	if row.DiscordConnectedAt.Valid {
		out.ConnectedAt = row.DiscordConnectedAt.Time.UTC()
	}
	return out, nil
}

// Connect inclui o Usuário no servidor oficial com o código da conexão e grava a data
// (RN-01 a RN-04). O token do Discord só passa pela memória (RN-02).
func (s *Service) Connect(ctx context.Context, userID, code, redirectURI string) (Connection, error) {
	if !s.Enabled() || s.OAuth == nil {
		return Connection{}, ErrDisabled
	}
	uid, err := parseUUID(userID)
	if err != nil {
		return Connection{}, ErrNotFound
	}
	row, err := s.queries.GetDiscordConnection(ctx, uid)
	if isNoRows(err) {
		return Connection{}, ErrNotFound
	}
	if err != nil {
		return Connection{}, fmt.Errorf("discordroom: conexão: %w", err)
	}
	_, err = s.OAuth.JoinFromCode(ctx, code, redirectURI, row.DiscordID, s.bot.AddMember)
	switch {
	case errors.Is(err, discord.ErrWrongAccount):
		return Connection{}, rule(CodeWrongAccount)
	case errors.Is(err, discord.ErrInvalidCode), errors.Is(err, discordbot.ErrInvalidToken):
		return Connection{}, ErrInvalidCode
	case errors.Is(err, discord.ErrUnavailable), errors.Is(err, discordbot.ErrUnavailable):
		return Connection{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	case err != nil:
		return Connection{}, fmt.Errorf("discordroom: conectar: %w", err)
	}
	now := s.Now().UTC()
	var at pgtype.Timestamptz
	err = s.inTx(ctx, func(q *db.Queries) error {
		var err error
		if at, err = q.SetDiscordConnected(ctx, db.SetDiscordConnectedParams{UserID: uid, Now: now}); err != nil {
			return err
		}
		return onConnected(ctx, q, uid, row.DiscordID, now) // RN-13
	})
	if err != nil {
		return Connection{}, fmt.Errorf("discordroom: gravar conexão: %w", err)
	}
	return Connection{Enabled: true, ConnectedAt: at.Time.UTC()}, nil
}
