package server

import (
	"context"
	"errors"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
)

// TalentService é o que as rotas precisam do banco de talentos; o *talents.Service
// satisfaz.
type TalentService interface {
	SetAvailability(ctx context.Context, userID, characterID string, in talents.AvailabilityInput) (*talents.Availability, error)
}

// talentsHandler serve o banco de talentos (spec banco-de-talentos).
type talentsHandler struct {
	session charactersHandler
	talents TalentService
}

var errNotImplemented = errors.New("banco de talentos: rota ainda não implementada")

// SetAvailability liga, edita ou desliga o personagem no banco (RN-01 a RN-05).
func (h talentsHandler) SetAvailability(ctx context.Context, req api.SetAvailabilityRequestObject) (api.SetAvailabilityResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.SetAvailability401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	in := talents.AvailabilityInput{}
	if b := req.Body; b != nil {
		in = talents.AvailabilityInput{
			Enabled: b.Enabled, Days: valueOf(b.Days), Start: valueOf(b.Start), End: valueOf(b.End),
			AnyInstance: valueOf(b.AnyInstance), InstanceIDs: valueOf(b.InstanceIds),
		}
	}
	got, err := h.talents.SetAvailability(ctx, userID, req.Id, in)
	var invalid *talents.ValidationError
	switch {
	case errors.As(err, &invalid):
		return api.SetAvailability422JSONResponse{InvalidJSONResponse: toAPITalentValidation(invalid)}, nil
	case errors.Is(err, talents.ErrNotFound):
		return api.SetAvailability404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse(notFound)}, nil
	case err != nil:
		return nil, err
	}
	body := api.SetAvailability200JSONResponse{}
	if got != nil {
		a := toAPIAvailability(*got)
		body.Availability = &a
	}
	return body, nil
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

func toAPIAvailability(a talents.Availability) api.Availability {
	return api.Availability{
		Enabled: a.Enabled, Days: a.Days, Start: a.Start, End: a.End,
		AnyInstance: a.AnyInstance, InstanceIds: a.InstanceIDs,
	}
}

func toAPITalentValidation(e *talents.ValidationError) api.InvalidJSONResponse {
	fields := make([]api.FieldError, len(e.Fields))
	for i, f := range e.Fields {
		fields[i] = api.FieldError{Field: api.FieldErrorField(f.Field), Code: api.FieldErrorCode(f.Code)}
	}
	return api.InvalidJSONResponse{Error: api.ValidationErrorErrorValidation, Fields: fields}
}

// valueOf devolve o valor do campo opcional, ou o zero do tipo.
func valueOf[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
