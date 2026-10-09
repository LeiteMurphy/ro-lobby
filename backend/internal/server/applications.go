package server

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/applications"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// ApplicationService é o que as rotas precisam do serviço de candidaturas; o
// *applications.Service satisfaz.
type ApplicationService interface {
	Apply(ctx context.Context, userID, lobbyID string, in applications.ApplyInput) (applications.Application, error)
	Accept(ctx context.Context, ownerID, applicationID string) (applications.Application, error)
	Reject(ctx context.Context, ownerID, applicationID, reason string) (applications.Application, error)
	Withdraw(ctx context.Context, userID, applicationID string) (applications.Application, error)
	ListMine(ctx context.Context, userID string) ([]applications.Mine, error)
	Leave(ctx context.Context, userID, applicationID string) (applications.Application, error)
	Remove(ctx context.Context, ownerID, applicationID, reason string, block bool) (applications.Application, error)
	SwapOwnerCharacter(ctx context.Context, ownerID, lobbyID, characterID string) error
	RequestSwap(ctx context.Context, userID, applicationID string, in applications.SwapInput) (applications.SwapRequest, error)
	AcceptSwap(ctx context.Context, ownerID, swapID string) (applications.SwapRequest, error)
	RejectSwap(ctx context.Context, ownerID, swapID, reason string) (applications.SwapRequest, error)
	WithdrawSwap(ctx context.Context, userID, swapID string) (applications.SwapRequest, error)
	Unblock(ctx context.Context, ownerID, lobbyID, characterID string) error
}

type applicationsHandler struct {
	session charactersHandler // reaproveita o currentUser (RN-04 da personagens)
	apps    ApplicationService
	// lobbies monta o detalhe que a troca de personagem do dono devolve.
	lobbies LobbyService
}

var applicationNotFound = api.ApplicationNotFoundJSONResponse(api.Error{Error: api.ErrorErrorNotFound})

func ruleError(err error) (api.ApplicationRuleJSONResponse, bool) {
	var re *applications.RuleError
	if !errors.As(err, &re) {
		return api.ApplicationRuleJSONResponse{}, false
	}
	return api.ApplicationRuleJSONResponse{
		Error: api.ApplicationRuleErrorErrorApplicationRule,
		Code:  api.ApplicationRuleErrorCode(re.Code),
	}, true
}

// ApplyToLobby cria a candidatura do Usuário da sessão (RN-01 a RN-07, RN-30, D-07).
func (h applicationsHandler) ApplyToLobby(ctx context.Context, req api.ApplyToLobbyRequestObject) (api.ApplyToLobbyResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.ApplyToLobby401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	in := applications.ApplyInput{}
	if b := req.Body; b != nil {
		in = applications.ApplyInput{CharacterID: b.CharacterId, Message: deref(b.Message)}
	}
	created, err := h.apps.Apply(ctx, userID, req.Id, in)
	var invalid *lobbies.ValidationError
	if rule, ok := ruleError(err); ok {
		return api.ApplyToLobby409JSONResponse{ApplicationRuleJSONResponse: rule}, nil
	}
	switch {
	case errors.As(err, &invalid):
		return api.ApplyToLobby422JSONResponse{InvalidJSONResponse: toAPILobbyValidation(invalid)}, nil
	case errors.Is(err, applications.ErrNotFound):
		return api.ApplyToLobby404JSONResponse{LobbyNotFoundJSONResponse: api.LobbyNotFoundJSONResponse(lobbyNotFound)}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPIApplication(created)
	return api.ApplyToLobby201JSONResponse(body), err
}

// AcceptApplication aceita a candidatura, se quem pede é o dono (RN-08, RN-10 a RN-12).
func (h applicationsHandler) AcceptApplication(ctx context.Context, req api.AcceptApplicationRequestObject) (api.AcceptApplicationResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.AcceptApplication401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	accepted, err := h.apps.Accept(ctx, userID, req.Id)
	if rule, ok := ruleError(err); ok {
		return api.AcceptApplication409JSONResponse{ApplicationRuleJSONResponse: rule}, nil
	}
	switch {
	case errors.Is(err, applications.ErrNotFound):
		return api.AcceptApplication404JSONResponse{ApplicationNotFoundJSONResponse: applicationNotFound}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPIApplication(accepted)
	return api.AcceptApplication200JSONResponse(body), err
}

// RejectApplication recusa a candidatura com justificativa, se quem pede é o dono
// (RN-08, RN-09).
func (h applicationsHandler) RejectApplication(ctx context.Context, req api.RejectApplicationRequestObject) (api.RejectApplicationResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.RejectApplication401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	reason := ""
	if req.Body != nil {
		reason = req.Body.Reason
	}
	rejected, err := h.apps.Reject(ctx, userID, req.Id, reason)
	var invalid *lobbies.ValidationError
	if rule, ok := ruleError(err); ok {
		return api.RejectApplication409JSONResponse{ApplicationRuleJSONResponse: rule}, nil
	}
	switch {
	case errors.As(err, &invalid):
		return api.RejectApplication422JSONResponse{InvalidJSONResponse: toAPILobbyValidation(invalid)}, nil
	case errors.Is(err, applications.ErrNotFound):
		return api.RejectApplication404JSONResponse{ApplicationNotFoundJSONResponse: applicationNotFound}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPIApplication(rejected)
	return api.RejectApplication200JSONResponse(body), err
}

// WithdrawApplication retira a própria candidatura pendente (RN-13).
func (h applicationsHandler) WithdrawApplication(ctx context.Context, req api.WithdrawApplicationRequestObject) (api.WithdrawApplicationResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.WithdrawApplication401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	withdrawn, err := h.apps.Withdraw(ctx, userID, req.Id)
	if rule, ok := ruleError(err); ok {
		return api.WithdrawApplication409JSONResponse{ApplicationRuleJSONResponse: rule}, nil
	}
	switch {
	case errors.Is(err, applications.ErrNotFound):
		return api.WithdrawApplication404JSONResponse{ApplicationNotFoundJSONResponse: applicationNotFound}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPIApplication(withdrawn)
	return api.WithdrawApplication200JSONResponse(body), err
}

// ListMyApplications devolve as candidaturas do Usuário da sessão (RN-33, CA-03.1).
func (h applicationsHandler) ListMyApplications(ctx context.Context, _ api.ListMyApplicationsRequestObject) (api.ListMyApplicationsResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.ListMyApplications401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	list, err := h.apps.ListMine(ctx, userID)
	if err != nil {
		return nil, err
	}
	body := make(api.ListMyApplications200JSONResponse, len(list))
	for i, m := range list {
		app, err := toAPIApplication(m.Application)
		if err != nil {
			return nil, err
		}
		body[i].Application = app
		body[i].Lobby.InstanceName = m.InstanceName
		body[i].Lobby.StartsAt = m.StartsAt
		body[i].Lobby.Status = api.LobbyStatus(m.LobbyStatus)
		if m.Nick != "" {
			body[i].Character = &struct {
				ClassId  string                             `json:"classId"`
				Level    int                                `json:"level"`
				Nick     string                             `json:"nick"`
				Portrait api.MyApplicationCharacterPortrait `json:"portrait"`
			}{ClassId: m.ClassID, Level: m.Level, Nick: m.Nick, Portrait: api.MyApplicationCharacterPortrait(m.Portrait)}
		}
	}
	return body, nil
}

func toAPIApplication(a applications.Application) (api.Application, error) {
	id, err := uuid.Parse(a.ID)
	if err != nil {
		return api.Application{}, err
	}
	lobbyID, err := uuid.Parse(a.LobbyID)
	if err != nil {
		return api.Application{}, err
	}
	characterID, err := optionalUUID(a.CharacterID)
	if err != nil {
		return api.Application{}, err
	}
	return api.Application{
		Id:          id,
		LobbyId:     lobbyID,
		CharacterId: characterID,
		Role:        api.Role(a.Role),
		Message:     optional(a.Message),
		Status:      api.ApplicationStatus(a.Status),
		Reason:      optional(a.Reason),
		Blocked:     a.Blocked,
		CreatedAt:   a.CreatedAt,
		DecidedAt:   optionalTime(a.DecidedAt),
	}, nil
}

func toAPIParticipant(p lobbies.Participant) (api.LobbyParticipant, error) {
	appID, err := uuid.Parse(p.ApplicationID)
	if err != nil {
		return api.LobbyParticipant{}, err
	}
	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		return api.LobbyParticipant{}, err
	}
	characterID, err := optionalUUID(p.CharacterID)
	if err != nil {
		return api.LobbyParticipant{}, err
	}
	out := api.LobbyParticipant{
		ApplicationId: appID,
		UserId:        userID,
		DiscordName:   optional(p.DiscordName),
		CharacterId:   characterID,
		Nick:          optional(p.Nick),
		ClassId:       optional(p.ClassID),
		Level:         optional(p.Level),
		Link:          optional(p.Link),
		Role:          api.Role(p.Role),
		Message:       optional(p.Message),
		CreatedAt:     p.CreatedAt,
	}
	if p.Portrait != "" {
		portrait := api.LobbyParticipantPortrait(p.Portrait)
		out.Portrait = &portrait
	}
	return out, nil
}

func toAPIViewerApplication(v *lobbies.ViewerApplication) (*api.ViewerApplication, error) {
	if v == nil {
		return nil, nil
	}
	id, err := uuid.Parse(v.ID)
	if err != nil {
		return nil, err
	}
	characterID, err := optionalUUID(v.CharacterID)
	if err != nil {
		return nil, err
	}
	out := &api.ViewerApplication{
		Id:          id,
		CharacterId: characterID,
		Role:        api.Role(v.Role),
		Message:     optional(v.Message),
		Status:      api.ApplicationStatus(v.Status),
		Reason:      optional(v.Reason),
		Blocked:     v.Blocked,
		CreatedAt:   v.CreatedAt,
		DecidedAt:   optionalTime(v.DecidedAt),
	}
	if v.SwapRequest != nil {
		swap, err := toAPISwapRequest(*v.SwapRequest)
		if err != nil {
			return nil, err
		}
		out.SwapRequest = &swap
	}
	return out, nil
}

// toAPIDetail monta o lobby do detalhe com o que quem olha pode ver (D-06).
func toAPIDetail(d lobbies.Detail) (api.Lobby, error) {
	body, err := toAPILobby(d.Lobby)
	if err != nil {
		return api.Lobby{}, err
	}
	members := make([]api.LobbyParticipant, len(d.Members))
	for i, m := range d.Members {
		if members[i], err = toAPIParticipant(m); err != nil {
			return api.Lobby{}, err
		}
	}
	body.Members = &members
	if d.Pending != nil {
		pending := make([]api.LobbyParticipant, len(d.Pending))
		for i, p := range d.Pending {
			if pending[i], err = toAPIParticipant(p); err != nil {
				return api.Lobby{}, err
			}
		}
		body.Pending = &pending
	}
	if d.SwapRequests != nil {
		swaps := make([]api.LobbySwapRequest, len(d.SwapRequests))
		for i, r := range d.SwapRequests {
			if swaps[i], err = toAPILobbySwapRequest(r); err != nil {
				return api.Lobby{}, err
			}
		}
		body.SwapRequests = &swaps
	}
	body.MyApplication, err = toAPIViewerApplication(d.MyApplication)
	return body, err
}

func optionalUUID(s string) (*uuid.UUID, error) {
	if s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func optionalTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
