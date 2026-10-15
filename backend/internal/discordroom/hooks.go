package discordroom

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
)

// Ganchos da sala nas ações do grupo (spec discord-bot, US-04, RN-10 a RN-13, RN-15, RN-17,
// RN-21, D-02). Rodam dentro da transação de quem chama e só põem efeitos na fila: a ação
// do grupo nunca depende do Discord. Sem sala aberta, não fazem nada.

// openRoom diz se o lobby tem sala aberta.
func openRoom(ctx context.Context, q *db.Queries, lobbyID pgtype.UUID) (bool, error) {
	_, err := q.GetOpenDiscordRoom(ctx, lobbyID)
	if isNoRows(err) {
		return false, nil
	}
	return err == nil, err
}

// OnMemberJoined: quem virou membro ganha acesso, se estiver conectado, e é mencionado
// (RN-10).
func OnMemberJoined(ctx context.Context, q *db.Queries, lobbyID, userID pgtype.UUID, now time.Time) error {
	ok, err := openRoom(ctx, q, lobbyID)
	if err != nil || !ok {
		return err
	}
	conn, err := q.GetDiscordConnection(ctx, userID)
	if err != nil {
		return fmt.Errorf("discordroom: conexão do membro: %w", err)
	}
	nick, err := q.LastMemberNick(ctx, db.LastMemberNickParams{LobbyID: lobbyID, UserID: userID})
	if err != nil {
		return fmt.Errorf("discordroom: nick do membro: %w", err)
	}
	if err := enqueueSync(ctx, q, lobbyID, userID, now); err != nil {
		return err
	}
	return enqueuePost(ctx, q, lobbyID, joinedMessage(nickOrDeleted(nick), conn.DiscordID, conn.DiscordConnectedAt.Valid), now)
}

// OnMemberLeft: quem saiu ou foi removido perde o acesso, e a sala é avisada sem o motivo
// (RN-11).
func OnMemberLeft(ctx context.Context, q *db.Queries, lobbyID, userID pgtype.UUID, now time.Time) error {
	ok, err := openRoom(ctx, q, lobbyID)
	if err != nil || !ok {
		return err
	}
	nick, err := q.LastMemberNick(ctx, db.LastMemberNickParams{LobbyID: lobbyID, UserID: userID})
	if err != nil {
		return fmt.Errorf("discordroom: nick do membro: %w", err)
	}
	if err := enqueueSync(ctx, q, lobbyID, userID, now); err != nil {
		return err
	}
	return enqueuePost(ctx, q, lobbyID, leftMessage(nickOrDeleted(nick)), now)
}

// OnStartChanged: o horário novo vai para a sala, e o lembrete segue o horário novo
// (RN-15). A remoção já segue o horário do lobby (D-02).
func OnStartChanged(ctx context.Context, q *db.Queries, lobbyID pgtype.UUID, startsAt, now time.Time) error {
	ok, err := openRoom(ctx, q, lobbyID)
	if err != nil || !ok {
		return err
	}
	if err := q.ResetDiscordReminder(ctx, lobbyID); err != nil {
		return err
	}
	members, err := q.RoomMembers(ctx, lobbyID)
	if err != nil {
		return err
	}
	return enqueuePost(ctx, q, lobbyID, startChangedMessage(startsAt, connectedIDs(members)), now)
}

// OnCancelled: a sala recebe o motivo e fica encerrando; ela é apagada 1 hora depois
// (RN-17).
func OnCancelled(ctx context.Context, q *db.Queries, lobbyID pgtype.UUID, reason string, now time.Time) error {
	ok, err := openRoom(ctx, q, lobbyID)
	if err != nil || !ok {
		return err
	}
	members, err := q.RoomMembers(ctx, lobbyID)
	if err != nil {
		return err
	}
	if err := enqueuePost(ctx, q, lobbyID, cancelledMessage(reason, connectedIDs(members)), now); err != nil {
		return err
	}
	return q.CloseDiscordRoom(ctx, db.CloseDiscordRoomParams{LobbyID: lobbyID, Now: now})
}

// onConnected: quem se conecta ganha acesso às salas abertas dos lobbies em que é membro e
// é mencionado em cada uma (RN-13).
func onConnected(ctx context.Context, q *db.Queries, userID pgtype.UUID, discordID string, now time.Time) error {
	rooms, err := q.OpenRoomsOfMember(ctx, userID)
	if err != nil {
		return err
	}
	for _, lobbyID := range rooms {
		members, err := q.RoomMembers(ctx, lobbyID)
		if err != nil {
			return err
		}
		nick := "Personagem excluído"
		for _, m := range members {
			if m.UserID == userID {
				nick = nickOf(m)
			}
		}
		if err := enqueueSync(ctx, q, lobbyID, userID, now); err != nil {
			return err
		}
		if err := enqueuePost(ctx, q, lobbyID, connectedMessage(nick, discordID), now); err != nil {
			return err
		}
	}
	return nil
}

func nickOrDeleted(nick pgtype.Text) string {
	if nick.Valid {
		return nick.String
	}
	return "Personagem excluído"
}
