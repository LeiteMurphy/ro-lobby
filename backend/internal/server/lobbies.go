package server

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/catalog"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// LobbyService é o que as rotas precisam do serviço de lobbies; o *lobbies.Service
// satisfaz.
type LobbyService interface {
	List(ctx context.Context, from, to string) ([]lobbies.Lobby, error)
	Detail(ctx context.Context, id, viewerID string) (lobbies.Detail, error)
	Create(ctx context.Context, userID string, in lobbies.Input) (lobbies.Lobby, error)
	Update(ctx context.Context, userID, id string, in lobbies.UpdateInput) (lobbies.Lobby, error)
	Cancel(ctx context.Context, userID, id, reason string) (lobbies.Lobby, error)
}

type lobbiesHandler struct {
	session charactersHandler // reaproveita o currentUser (RN-04)
	lobbies LobbyService
}

var (
	lobbyNotFound = api.Error{Error: api.ErrorErrorNotFound}
	lobbyNotOpen  = api.Error{Error: api.ErrorErrorLobbyNotOpen}
)

// ListInstances serve o catálogo de instâncias, sem sessão (RN-01, RN-02, CA-06.1).
func (lobbiesHandler) ListInstances(context.Context, api.ListInstancesRequestObject) (api.ListInstancesResponseObject, error) {
	list := catalog.Instances()
	body := make(api.ListInstances200JSONResponse, len(list))
	for i, in := range list {
		body[i] = api.Instance{Id: in.ID, Name: in.Name, Level: in.Level, Reset: api.InstanceReset(in.Reset)}
	}
	return body, nil
}

// ListLobbies devolve os lobbies abertos do intervalo, sem sessão (RN-13, RN-14, RN-22).
func (h lobbiesHandler) ListLobbies(ctx context.Context, req api.ListLobbiesRequestObject) (api.ListLobbiesResponseObject, error) {
	list, err := h.lobbies.List(ctx, req.Params.From.Format(time.DateOnly), req.Params.To.Format(time.DateOnly))
	if errors.Is(err, lobbies.ErrInvalidRange) {
		return api.ListLobbies400JSONResponse{Error: api.ErrorErrorInvalidRange}, nil
	}
	if err != nil {
		return nil, err
	}
	body := make(api.ListLobbies200JSONResponse, len(list))
	for i, l := range list {
		if body[i], err = toAPILobby(l); err != nil {
			return nil, err
		}
	}
	return body, nil
}

// GetLobby devolve o lobby em qualquer estado (RN-15), com o que quem olha pode ver; a
// sessão é opcional (D-06 da candidatura).
func (h lobbiesHandler) GetLobby(ctx context.Context, req api.GetLobbyRequestObject) (api.GetLobbyResponseObject, error) {
	viewerID, _, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	l, err := h.lobbies.Detail(ctx, req.Id, viewerID)
	if errors.Is(err, lobbies.ErrNotFound) {
		return api.GetLobby404JSONResponse{LobbyNotFoundJSONResponse: api.LobbyNotFoundJSONResponse(lobbyNotFound)}, nil
	}
	if err != nil {
		return nil, err
	}
	body, err := toAPIDetail(l)
	return api.GetLobby200JSONResponse(body), err
}

// CreateLobby cria o lobby do Usuário da sessão (RN-04 a RN-11).
func (h lobbiesHandler) CreateLobby(ctx context.Context, req api.CreateLobbyRequestObject) (api.CreateLobbyResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.CreateLobby401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	in := lobbies.Input{}
	if b := req.Body; b != nil {
		in = lobbies.Input{
			InstanceID:  b.InstanceId,
			StartsAt:    b.StartsAt,
			Slots:       fromAPISlots(b.Slots),
			MinLevel:    b.MinLevel,
			CharacterID: b.CharacterId,
			Note:        deref(b.Note),
		}
	}
	created, err := h.lobbies.Create(ctx, userID, in)
	var invalid *lobbies.ValidationError
	switch {
	case errors.As(err, &invalid):
		return api.CreateLobby422JSONResponse{InvalidJSONResponse: toAPILobbyValidation(invalid)}, nil
	case errors.Is(err, lobbies.ErrLimitReached):
		return api.CreateLobby409JSONResponse{Error: api.ErrorErrorLobbyLimit}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPILobby(created)
	return api.CreateLobby201JSONResponse(body), err
}

// UpdateLobby edita o lobby aberto do dono (RN-17, RN-18, RN-20).
func (h lobbiesHandler) UpdateLobby(ctx context.Context, req api.UpdateLobbyRequestObject) (api.UpdateLobbyResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.UpdateLobby401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	in := lobbies.UpdateInput{}
	if b := req.Body; b != nil {
		in = lobbies.UpdateInput{StartsAt: b.StartsAt, Slots: fromAPISlots(b.Slots), MinLevel: b.MinLevel, Note: deref(b.Note)}
	}
	updated, err := h.lobbies.Update(ctx, userID, req.Id, in)
	var invalid *lobbies.ValidationError
	switch {
	case errors.As(err, &invalid):
		return api.UpdateLobby422JSONResponse{InvalidJSONResponse: toAPILobbyValidation(invalid)}, nil
	case errors.Is(err, lobbies.ErrNotFound):
		return api.UpdateLobby404JSONResponse{LobbyNotFoundJSONResponse: api.LobbyNotFoundJSONResponse(lobbyNotFound)}, nil
	case errors.Is(err, lobbies.ErrNotOpen):
		return api.UpdateLobby409JSONResponse{LobbyNotOpenJSONResponse: api.LobbyNotOpenJSONResponse(lobbyNotOpen)}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPILobby(updated)
	return api.UpdateLobby200JSONResponse(body), err
}

// CancelLobby cancela o lobby aberto do dono, com motivo (RN-19, RN-20).
func (h lobbiesHandler) CancelLobby(ctx context.Context, req api.CancelLobbyRequestObject) (api.CancelLobbyResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.CancelLobby401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	reason := ""
	if req.Body != nil {
		reason = req.Body.Reason
	}
	cancelled, err := h.lobbies.Cancel(ctx, userID, req.Id, reason)
	var invalid *lobbies.ValidationError
	switch {
	case errors.As(err, &invalid):
		return api.CancelLobby422JSONResponse{InvalidJSONResponse: toAPILobbyValidation(invalid)}, nil
	case errors.Is(err, lobbies.ErrNotFound):
		return api.CancelLobby404JSONResponse{LobbyNotFoundJSONResponse: api.LobbyNotFoundJSONResponse(lobbyNotFound)}, nil
	case errors.Is(err, lobbies.ErrNotOpen):
		return api.CancelLobby409JSONResponse{LobbyNotOpenJSONResponse: api.LobbyNotOpenJSONResponse(lobbyNotOpen)}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPILobby(cancelled)
	return api.CancelLobby200JSONResponse(body), err
}

func fromAPISlots(s api.Slots) lobbies.Slots {
	return lobbies.Slots{Tank: s.Tank, Support: s.Support, Dps: s.Dps}
}

func toAPISlots(s lobbies.Slots) api.Slots {
	return api.Slots{Tank: s.Tank, Support: s.Support, Dps: s.Dps}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func optional[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

func toAPILobby(l lobbies.Lobby) (api.Lobby, error) {
	id, err := uuid.Parse(l.ID)
	if err != nil {
		return api.Lobby{}, err
	}
	ownerID, err := uuid.Parse(l.Owner.UserID)
	if err != nil {
		return api.Lobby{}, err
	}
	body := api.Lobby{
		Id: id,
		Instance: api.LobbyInstance{
			Id:    l.InstanceID,
			Name:  l.InstanceName,
			Level: l.InstanceLevel,
		},
		StartsAt:     l.StartsAt,
		Status:       api.LobbyStatus(l.Status),
		Slots:        toAPISlots(l.Slots),
		Occupied:     toAPISlots(l.Occupied),
		PendingCount: l.PendingCount,
		MinLevel:     l.MinLevel,
		Note:         optional(l.Note),
		CancelReason: optional(l.CancelReason),
		CreatedAt:    l.CreatedAt,
		Owner: api.LobbyOwner{
			UserId:      ownerID,
			DiscordName: optional(l.Owner.DiscordName),
			Role:        api.Role(l.Owner.Role),
			Nick:        optional(l.Owner.Nick),
			ClassId:     optional(l.Owner.ClassID),
			Level:       optional(l.Owner.Level),
			Link:        optional(l.Owner.Link),
		},
	}
	if l.InstanceReset != "" {
		reset := api.LobbyInstanceReset(l.InstanceReset)
		body.Instance.Reset = &reset
	}
	if l.Owner.Portrait != "" {
		portrait := api.LobbyOwnerPortrait(l.Owner.Portrait)
		body.Owner.Portrait = &portrait
	}
	if l.Owner.CharacterID != "" {
		cid, err := uuid.Parse(l.Owner.CharacterID)
		if err != nil {
			return api.Lobby{}, err
		}
		body.Owner.CharacterId = &cid
	}
	return body, nil
}

func toAPILobbyValidation(e *lobbies.ValidationError) api.InvalidJSONResponse {
	fields := make([]api.FieldError, len(e.Fields))
	for i, f := range e.Fields {
		fields[i] = api.FieldError{Field: api.FieldErrorField(f.Field), Code: api.FieldErrorCode(f.Code)}
	}
	return api.InvalidJSONResponse{Error: api.ValidationErrorErrorValidation, Fields: fields}
}
