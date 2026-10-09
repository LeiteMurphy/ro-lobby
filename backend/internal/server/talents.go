package server

import (
	"context"
	"errors"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
)

// talentsHandler serve o banco de talentos (spec banco-de-talentos). As rotas chegam nas
// tasks T-02 a T-04.
type talentsHandler struct{}

var errNotImplemented = errors.New("banco de talentos: rota ainda não implementada")

func (talentsHandler) SetAvailability(context.Context, api.SetAvailabilityRequestObject) (api.SetAvailabilityResponseObject, error) {
	return nil, errNotImplemented
}

func (talentsHandler) ListTalents(context.Context, api.ListTalentsRequestObject) (api.ListTalentsResponseObject, error) {
	return nil, errNotImplemented
}

func (talentsHandler) CountTalents(context.Context, api.CountTalentsRequestObject) (api.CountTalentsResponseObject, error) {
	return nil, errNotImplemented
}

func (talentsHandler) ListLobbyTalents(context.Context, api.ListLobbyTalentsRequestObject) (api.ListLobbyTalentsResponseObject, error) {
	return nil, errNotImplemented
}
