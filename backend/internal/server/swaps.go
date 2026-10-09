package server

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/applications"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// Rotas da Parte 2 da candidatura: sair, remover e trocas de personagem (US-05 a US-08,
// design P2.5).

var swapRequestNotFound = api.SwapRequestNotFoundJSONResponse(api.Error{Error: api.ErrorErrorNotFound})

// LeaveLobby tira o membro da sessão do grupo (RN-14, RN-24).
func (h applicationsHandler) LeaveLobby(ctx context.Context, req api.LeaveLobbyRequestObject) (api.LeaveLobbyResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.LeaveLobby401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	left, err := h.apps.Leave(ctx, userID, req.Id)
	if rule, ok := ruleError(err); ok {
		return api.LeaveLobby409JSONResponse{ApplicationRuleJSONResponse: rule}, nil
	}
	switch {
	case errors.Is(err, applications.ErrNotFound):
		return api.LeaveLobby404JSONResponse{ApplicationNotFoundJSONResponse: applicationNotFound}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPIApplication(left)
	return api.LeaveLobby200JSONResponse(body), err
}

// RemoveMember tira um membro do grupo, se quem pede é o dono (RN-07, RN-09, RN-15).
func (h applicationsHandler) RemoveMember(ctx context.Context, req api.RemoveMemberRequestObject) (api.RemoveMemberResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.RemoveMember401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	reason, block := "", false
	if b := req.Body; b != nil {
		reason, block = b.Reason, b.Block != nil && *b.Block
	}
	removed, err := h.apps.Remove(ctx, userID, req.Id, reason, block)
	var invalid *lobbies.ValidationError
	if rule, ok := ruleError(err); ok {
		return api.RemoveMember409JSONResponse{ApplicationRuleJSONResponse: rule}, nil
	}
	switch {
	case errors.As(err, &invalid):
		return api.RemoveMember422JSONResponse{InvalidJSONResponse: toAPILobbyValidation(invalid)}, nil
	case errors.Is(err, applications.ErrNotFound):
		return api.RemoveMember404JSONResponse{ApplicationNotFoundJSONResponse: applicationNotFound}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPIApplication(removed)
	return api.RemoveMember200JSONResponse(body), err
}

// SwapOwnerCharacter troca o personagem do dono no lobby e devolve o detalhe (RN-11,
// RN-19, RN-36).
func (h applicationsHandler) SwapOwnerCharacter(ctx context.Context, req api.SwapOwnerCharacterRequestObject) (api.SwapOwnerCharacterResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.SwapOwnerCharacter401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	characterID := ""
	if req.Body != nil {
		characterID = req.Body.CharacterId
	}
	err = h.apps.SwapOwnerCharacter(ctx, userID, req.Id, characterID)
	var invalid *lobbies.ValidationError
	if rule, ok := ruleError(err); ok {
		return api.SwapOwnerCharacter409JSONResponse{ApplicationRuleJSONResponse: rule}, nil
	}
	switch {
	case errors.As(err, &invalid):
		return api.SwapOwnerCharacter422JSONResponse{InvalidJSONResponse: toAPILobbyValidation(invalid)}, nil
	case errors.Is(err, applications.ErrNotFound):
		return api.SwapOwnerCharacter404JSONResponse{LobbyNotFoundJSONResponse: api.LobbyNotFoundJSONResponse(lobbyNotFound)}, nil
	case err != nil:
		return nil, err
	}
	detail, err := h.lobbies.Detail(ctx, req.Id, userID)
	if err != nil {
		return nil, err
	}
	body, err := toAPIDetail(detail)
	return api.SwapOwnerCharacter200JSONResponse(body), err
}

// UnblockInLobby desbloqueia o jogador no lobby aberto do dono (RN-15, CA-06.9).
func (h applicationsHandler) UnblockInLobby(ctx context.Context, req api.UnblockInLobbyRequestObject) (api.UnblockInLobbyResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.UnblockInLobby401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	characterID := ""
	if req.Body != nil {
		characterID = req.Body.CharacterId
	}
	err = h.apps.Unblock(ctx, userID, req.Id, characterID)
	if rule, ok := ruleError(err); ok && rule.Code == applications.CodeNotOpen {
		return api.UnblockInLobby409JSONResponse{LobbyNotOpenJSONResponse: api.LobbyNotOpenJSONResponse(lobbyNotOpen)}, nil
	}
	switch {
	case errors.Is(err, applications.ErrNotFound):
		return api.UnblockInLobby404JSONResponse{LobbyNotFoundJSONResponse: api.LobbyNotFoundJSONResponse(lobbyNotFound)}, nil
	case err != nil:
		return nil, err
	}
	return api.UnblockInLobby204Response{}, nil
}

// RequestSwap abre o pedido de troca do membro da sessão (RN-20, RN-21, RN-36).
func (h applicationsHandler) RequestSwap(ctx context.Context, req api.RequestSwapRequestObject) (api.RequestSwapResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.RequestSwap401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	in := applications.SwapInput{}
	if b := req.Body; b != nil {
		in = applications.SwapInput{CharacterID: b.CharacterId, Reason: b.Reason}
	}
	created, err := h.apps.RequestSwap(ctx, userID, req.Id, in)
	var invalid *lobbies.ValidationError
	if rule, ok := ruleError(err); ok {
		return api.RequestSwap409JSONResponse{ApplicationRuleJSONResponse: rule}, nil
	}
	switch {
	case errors.As(err, &invalid):
		return api.RequestSwap422JSONResponse{InvalidJSONResponse: toAPILobbyValidation(invalid)}, nil
	case errors.Is(err, applications.ErrNotFound):
		return api.RequestSwap404JSONResponse{ApplicationNotFoundJSONResponse: applicationNotFound}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPISwapRequest(created)
	return api.RequestSwap201JSONResponse(body), err
}

// AcceptSwapRequest aceita o pedido de troca, se quem pede é o dono (RN-22, RN-23).
func (h applicationsHandler) AcceptSwapRequest(ctx context.Context, req api.AcceptSwapRequestRequestObject) (api.AcceptSwapRequestResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.AcceptSwapRequest401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	accepted, err := h.apps.AcceptSwap(ctx, userID, req.Id)
	if rule, ok := ruleError(err); ok {
		return api.AcceptSwapRequest409JSONResponse{ApplicationRuleJSONResponse: rule}, nil
	}
	switch {
	case errors.Is(err, applications.ErrNotFound):
		return api.AcceptSwapRequest404JSONResponse{SwapRequestNotFoundJSONResponse: swapRequestNotFound}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPISwapRequest(accepted)
	return api.AcceptSwapRequest200JSONResponse(body), err
}

// RejectSwapRequest recusa o pedido de troca com justificativa, se quem pede é o dono
// (RN-09, RN-22).
func (h applicationsHandler) RejectSwapRequest(ctx context.Context, req api.RejectSwapRequestRequestObject) (api.RejectSwapRequestResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.RejectSwapRequest401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	reason := ""
	if req.Body != nil {
		reason = req.Body.Reason
	}
	rejected, err := h.apps.RejectSwap(ctx, userID, req.Id, reason)
	var invalid *lobbies.ValidationError
	if rule, ok := ruleError(err); ok {
		return api.RejectSwapRequest409JSONResponse{ApplicationRuleJSONResponse: rule}, nil
	}
	switch {
	case errors.As(err, &invalid):
		return api.RejectSwapRequest422JSONResponse{InvalidJSONResponse: toAPILobbyValidation(invalid)}, nil
	case errors.Is(err, applications.ErrNotFound):
		return api.RejectSwapRequest404JSONResponse{SwapRequestNotFoundJSONResponse: swapRequestNotFound}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPISwapRequest(rejected)
	return api.RejectSwapRequest200JSONResponse(body), err
}

// WithdrawSwapRequest retira o próprio pedido de troca pendente (RN-27).
func (h applicationsHandler) WithdrawSwapRequest(ctx context.Context, req api.WithdrawSwapRequestRequestObject) (api.WithdrawSwapRequestResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.WithdrawSwapRequest401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	withdrawn, err := h.apps.WithdrawSwap(ctx, userID, req.Id)
	if rule, ok := ruleError(err); ok {
		return api.WithdrawSwapRequest409JSONResponse{ApplicationRuleJSONResponse: rule}, nil
	}
	switch {
	case errors.Is(err, applications.ErrNotFound):
		return api.WithdrawSwapRequest404JSONResponse{SwapRequestNotFoundJSONResponse: swapRequestNotFound}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPISwapRequest(withdrawn)
	return api.WithdrawSwapRequest200JSONResponse(body), err
}

func toAPISwapRequest(r lobbies.SwapRequest) (api.SwapRequest, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return api.SwapRequest{}, err
	}
	appID, err := uuid.Parse(r.ApplicationID)
	if err != nil {
		return api.SwapRequest{}, err
	}
	from, err := optionalUUID(r.FromCharacterID)
	if err != nil {
		return api.SwapRequest{}, err
	}
	to, err := optionalUUID(r.ToCharacterID)
	if err != nil {
		return api.SwapRequest{}, err
	}
	return api.SwapRequest{
		Id:              id,
		ApplicationId:   appID,
		FromCharacterId: from,
		ToCharacterId:   to,
		ToRole:          api.Role(r.ToRole),
		Reason:          r.Reason,
		Status:          api.SwapRequestStatus(r.Status),
		DecisionReason:  optional(r.DecisionReason),
		CreatedAt:       r.CreatedAt,
		DecidedAt:       optionalTime(r.DecidedAt),
	}, nil
}

func toAPISwapCharacter(c lobbies.SwapCharacter) (api.SwapCharacter, error) {
	id, err := optionalUUID(c.CharacterID)
	if err != nil {
		return api.SwapCharacter{}, err
	}
	out := api.SwapCharacter{
		CharacterId: id,
		Nick:        optional(c.Nick),
		ClassId:     optional(c.ClassID),
		Level:       optional(c.Level),
		Role:        api.Role(c.Role),
	}
	if c.Portrait != "" {
		portrait := api.SwapCharacterPortrait(c.Portrait)
		out.Portrait = &portrait
	}
	return out, nil
}

func toAPILobbySwapRequest(r lobbies.LobbySwapRequest) (api.LobbySwapRequest, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return api.LobbySwapRequest{}, err
	}
	appID, err := uuid.Parse(r.ApplicationID)
	if err != nil {
		return api.LobbySwapRequest{}, err
	}
	userID, err := uuid.Parse(r.UserID)
	if err != nil {
		return api.LobbySwapRequest{}, err
	}
	from, err := toAPISwapCharacter(r.From)
	if err != nil {
		return api.LobbySwapRequest{}, err
	}
	to, err := toAPISwapCharacter(r.To)
	if err != nil {
		return api.LobbySwapRequest{}, err
	}
	return api.LobbySwapRequest{
		Id:            id,
		ApplicationId: appID,
		UserId:        userID,
		DiscordName:   optional(r.DiscordName),
		From:          from,
		To:            to,
		Reason:        r.Reason,
		CreatedAt:     r.CreatedAt,
	}, nil
}
