package server

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordroom"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// RoomService é o que as rotas precisam da sala do grupo no Discord (spec discord-bot); o
// *discordroom.Service satisfaz.
type RoomService interface {
	Enabled() bool
	Open(ctx context.Context, ownerID, lobbyID string) (discordroom.Room, error)
	View(ctx context.Context, lobbyID, viewerID string) (discordroom.View, error)
}

// Option ajusta o servidor.
type Option func(*options)

type options struct {
	rooms RoomService
}

// WithRooms liga as rotas da sala do grupo no Discord. Sem ela, a sala fica desligada
// (RN-22).
func WithRooms(r RoomService) Option { return func(o *options) { o.rooms = r } }

// noRooms é a sala desligada: sem bot, o bloco discord só traz os membros (RN-22).
type noRooms struct{}

func (noRooms) Enabled() bool { return false }
func (noRooms) Open(context.Context, string, string) (discordroom.Room, error) {
	return discordroom.Room{}, discordroom.ErrDisabled
}
func (noRooms) View(context.Context, string, string) (discordroom.View, error) {
	return discordroom.View{Members: []discordroom.Member{}}, nil
}

type discordHandler struct {
	session charactersHandler
	rooms   RoomService
}

// OpenDiscordRoom abre a sala do grupo (RN-05 a RN-09, RN-20).
func (h discordHandler) OpenDiscordRoom(ctx context.Context, req api.OpenDiscordRoomRequestObject) (api.OpenDiscordRoomResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.OpenDiscordRoom401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	room, err := h.rooms.Open(ctx, userID, req.Id)
	var re *discordroom.RuleError
	switch {
	case errors.As(err, &re):
		return api.OpenDiscordRoom409JSONResponse{Error: api.DiscordRuleErrorErrorDiscordRule, Code: api.DiscordRuleErrorCode(re.Code)}, nil
	case errors.Is(err, discordroom.ErrDisabled):
		return api.OpenDiscordRoom404JSONResponse{Error: api.DiscordErrorErrorDiscordDisabled}, nil
	case errors.Is(err, discordroom.ErrNotFound):
		return api.OpenDiscordRoom404JSONResponse{Error: api.DiscordErrorErrorNotFound}, nil
	case errors.Is(err, discordroom.ErrGuildFull):
		return api.OpenDiscordRoom503JSONResponse{Error: api.DiscordErrorErrorGuildFull}, nil
	case errors.Is(err, discordroom.ErrUnavailable):
		return api.OpenDiscordRoom503JSONResponse{Error: api.DiscordErrorErrorDiscordUnavailable}, nil
	case err != nil:
		return nil, err
	}
	url := room.URL()
	return api.OpenDiscordRoom201JSONResponse{Status: api.DiscordRoomStatus(room.Status), Url: &url}, nil
}

// viewerIsMember diz se quem olha é o dono ou um membro aceito (D-05).
func viewerIsMember(d lobbies.Detail, viewerID string) bool {
	if viewerID == "" {
		return false
	}
	if d.Owner.UserID == viewerID {
		return true
	}
	for _, m := range d.Members {
		if m.UserID == viewerID {
			return true
		}
	}
	return false
}

// toAPIDiscord converte o bloco discord do detalhe (RN-18).
func toAPIDiscord(v discordroom.View) (api.LobbyDiscord, error) {
	out := api.LobbyDiscord{Enabled: v.Enabled, Members: make([]api.DiscordMember, len(v.Members)), ViewerConnected: v.ViewerConnected}
	for i, m := range v.Members {
		id, err := uuid.Parse(m.UserID)
		if err != nil {
			return out, err
		}
		out.Members[i] = api.DiscordMember{UserId: id, Nick: m.Nick, Connected: m.Connected}
	}
	if v.Room != nil {
		room := api.DiscordRoom{Status: api.DiscordRoomStatus(v.Room.Status)}
		if v.Room.URL != "" {
			url := v.Room.URL
			room.Url = &url
		}
		out.Room = &room
	}
	return out, nil
}
