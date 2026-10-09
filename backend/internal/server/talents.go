package server

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
)

// TalentService é o que as rotas precisam do banco de talentos; o *talents.Service
// satisfaz.
type TalentService interface {
	SetAvailability(ctx context.Context, userID, characterID string, in talents.AvailabilityInput) (*talents.Availability, error)
	ForLobby(ctx context.Context, userID, lobbyID string) ([]talents.Talent, error)
	Catalog(ctx context.Context, f talents.CatalogFilter) ([]talents.Talent, error)
	Count(ctx context.Context, userID string, in talents.CountInput) (int, error)
}

// talentsHandler serve o banco de talentos (spec banco-de-talentos).
type talentsHandler struct {
	session charactersHandler
	talents TalentService
}

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

// ListTalents serve o catálogo (RN-11). A sessão é opcional; o nome no Discord só sai
// com ela (RN-12, D-05).
func (h talentsHandler) ListTalents(ctx context.Context, req api.ListTalentsRequestObject) (api.ListTalentsResponseObject, error) {
	_, withDiscord, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	p := req.Params
	f := talents.CatalogFilter{InstanceID: valueOf(p.InstanceId), Day: p.Day, Time: valueOf(p.Time)}
	if p.Role != nil {
		f.Role = string(*p.Role)
	}
	list, err := h.talents.Catalog(ctx, f)
	var invalid *talents.ValidationError
	if errors.As(err, &invalid) {
		return api.ListTalents422JSONResponse{InvalidJSONResponse: toAPITalentValidation(invalid)}, nil
	}
	if err != nil {
		return nil, err
	}
	body, err := toAPITalents(list, withDiscord)
	return api.ListTalents200JSONResponse(body), err
}

// CountTalents conta a afinidade do lobby em criação (RN-10).
func (h talentsHandler) CountTalents(ctx context.Context, req api.CountTalentsRequestObject) (api.CountTalentsResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.CountTalents401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	p := req.Params
	n, err := h.talents.Count(ctx, userID, talents.CountInput{
		InstanceID: p.InstanceId, StartsAt: p.StartsAt, MinLevel: p.MinLevel,
		Tank: p.Tank, Support: p.Support, Dps: p.Dps, CharacterID: p.CharacterId,
	})
	var invalid *talents.ValidationError
	if errors.As(err, &invalid) {
		return api.CountTalents422JSONResponse{InvalidJSONResponse: toAPITalentValidation(invalid)}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.CountTalents200JSONResponse{Count: n}, nil
}

// ListLobbyTalents lista quem tem afinidade com o lobby aberto do dono (RN-06 a RN-09).
// O dono tem sessão, então o nome no Discord vai junto (RN-12).
func (h talentsHandler) ListLobbyTalents(ctx context.Context, req api.ListLobbyTalentsRequestObject) (api.ListLobbyTalentsResponseObject, error) {
	userID, ok, err := h.session.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.ListLobbyTalents401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	list, err := h.talents.ForLobby(ctx, userID, req.Id)
	switch {
	case errors.Is(err, talents.ErrNotFound):
		return api.ListLobbyTalents404JSONResponse{LobbyNotFoundJSONResponse: api.LobbyNotFoundJSONResponse(lobbyNotFound)}, nil
	case errors.Is(err, talents.ErrNotOpen):
		return api.ListLobbyTalents409JSONResponse{LobbyNotOpenJSONResponse: api.LobbyNotOpenJSONResponse(lobbyNotOpen)}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPITalents(list, true)
	return api.ListLobbyTalents200JSONResponse(body), err
}

// toAPITalents converte a lista; sem sessão, o nome no Discord fica de fora (D-05).
func toAPITalents(list []talents.Talent, withDiscord bool) ([]api.Talent, error) {
	out := make([]api.Talent, len(list))
	for i, t := range list {
		id, err := uuid.Parse(t.CharacterID)
		if err != nil {
			return nil, err
		}
		instances := make([]api.TalentInstance, len(t.Instances))
		for j, in := range t.Instances {
			instances[j] = api.TalentInstance{Id: in.ID, Name: in.Name}
		}
		out[i] = api.Talent{
			CharacterId: id, Nick: t.Nick, ClassId: t.ClassID, Level: t.Level, Role: api.Role(t.Role),
			Portrait: api.Portrait(t.Portrait), Days: t.Days, Start: t.Start, End: t.End,
			AnyInstance: t.AnyInstance, Instances: instances,
		}
		if t.Link != "" {
			link := t.Link
			out[i].Link = &link
		}
		if withDiscord {
			name := t.DiscordUsername
			out[i].DiscordUsername = &name
		}
	}
	return out, nil
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
