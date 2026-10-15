package discordroom

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
)

// Avisar no Discord (spec discord-bot, US-08, RN-26 a RN-28).

// AlertCooldown é o intervalo mínimo entre avisos do mesmo lobby (RN-27).
const AlertCooldown = 10 * time.Minute

// CooldownError: o último aviso foi há menos de AlertCooldown (RN-27).
type CooldownError struct {
	RetryAfter time.Duration
}

func (e *CooldownError) Error() string {
	return fmt.Sprintf("discordroom: aviso em intervalo, falta %s", e.RetryAfter)
}

// Alert posta "<dono> está chamando o grupo" na sala aberta, mencionando os membros
// conectados sem o dono (RN-26). Só o dono avisa, com intervalo mínimo de 10 minutos
// (RN-27); um aviso que falha no Discord não conta para o intervalo nem é refeito (RN-28).
func (s *Service) Alert(ctx context.Context, ownerID, lobbyID string) error {
	if !s.Enabled() {
		return ErrDisabled
	}
	lid, err := parseUUID(lobbyID)
	if err != nil {
		return ErrNotFound
	}
	now := s.Now().UTC()
	err = s.inTx(ctx, func(q *db.Queries) error {
		room, err := q.LockRoomForAlert(ctx, lid)
		if isNoRows(err) {
			owner, err := q.GetLobbyOwner(ctx, lid)
			switch {
			case isNoRows(err):
				return ErrNotFound
			case err != nil:
				return err
			case owner.String() != ownerID:
				return rule(CodeNotOwner)
			}
			return rule(CodeNoRoom)
		}
		if err != nil {
			return err
		}
		if room.OwnerID.String() != ownerID {
			return rule(CodeNotOwner) // RN-28
		}
		if room.Status != StatusOpen {
			return rule(CodeNoRoom) // RN-26
		}
		if room.LastAlertAt.Valid {
			if next := room.LastAlertAt.Time.Add(AlertCooldown); now.Before(next) {
				return &CooldownError{RetryAfter: next.Sub(now)} // RN-27: sem chamar o Discord
			}
		}
		members, err := q.RoomMembers(ctx, lid)
		if err != nil {
			return err
		}
		owner, ids := "", []string{}
		for _, m := range members {
			if m.IsOwner {
				owner = nickOf(m)
			} else if m.Connected {
				ids = append(ids, m.DiscordID)
			}
		}
		msg := alertMessage(owner, ids, s.lobbyURL(lobbyID))
		if err := s.bot.Post(ctx, room.TextChannelID, msg.Content, msg.Mentions); err != nil {
			return fmt.Errorf("%w: %w", ErrUnavailable, err) // RN-28: a transação desfaz tudo
		}
		return q.SetLastAlert(ctx, db.SetLastAlertParams{LobbyID: lid, Now: now})
	})
	var re *RuleError
	var ce *CooldownError
	if err == nil || errors.As(err, &re) || errors.As(err, &ce) || errors.Is(err, ErrNotFound) || errors.Is(err, ErrUnavailable) {
		return err
	}
	return fmt.Errorf("discordroom: avisar: %w", err)
}
