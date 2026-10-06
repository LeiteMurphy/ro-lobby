package server

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/auth"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/catalog"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/characters"
)

// CharacterService é o que as rotas precisam do serviço de personagens; o
// *characters.Service satisfaz.
type CharacterService interface {
	List(ctx context.Context, userID string) ([]characters.Character, error)
	Create(ctx context.Context, userID string, in characters.Input) (characters.Character, error)
	Update(ctx context.Context, userID, id string, in characters.Input) (characters.Character, error)
	Delete(ctx context.Context, userID, id string) error
	SetMain(ctx context.Context, userID, id string) error
}

type charactersHandler struct {
	auth  Authenticator
	chars CharacterService
}

var (
	noSession = api.Error{Error: api.ErrorErrorNoSession}
	notFound  = api.Error{Error: api.ErrorErrorNotFound}
)

// currentUser devolve o id do Usuário da sessão, ou ok=false para responder 401 (RN-01).
func (h charactersHandler) currentUser(ctx context.Context) (string, bool, error) {
	user, err := h.auth.Authenticate(ctx, sessionToken(ctx))
	if errors.Is(err, auth.ErrNoSession) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return user.ID, true, nil
}

// ListClasses serve o catálogo de classes, sem sessão (RN-20, CA-07.1).
func (charactersHandler) ListClasses(context.Context, api.ListClassesRequestObject) (api.ListClassesResponseObject, error) {
	classes := catalog.Classes()
	body := make(api.ListClasses200JSONResponse, len(classes))
	for i, c := range classes {
		body[i] = api.Class{Id: c.ID, Name: c.Name, Plural: c.Plural, Tier: api.ClassTier(c.Tier), Family: c.Family}
	}
	return body, nil
}

// ListCharacters devolve os personagens do Usuário da sessão (RN-01, RN-17).
func (h charactersHandler) ListCharacters(ctx context.Context, _ api.ListCharactersRequestObject) (api.ListCharactersResponseObject, error) {
	userID, ok, err := h.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.ListCharacters401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	list, err := h.chars.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	body := make(api.ListCharacters200JSONResponse, len(list))
	for i, c := range list {
		if body[i], err = toAPICharacter(c); err != nil {
			return nil, err
		}
	}
	return body, nil
}

// CreateCharacter cadastra um personagem (RN-04 a RN-12).
func (h charactersHandler) CreateCharacter(ctx context.Context, req api.CreateCharacterRequestObject) (api.CreateCharacterResponseObject, error) {
	userID, ok, err := h.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.CreateCharacter401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	created, err := h.chars.Create(ctx, userID, toInput(req.Body))
	var invalid *characters.ValidationError
	switch {
	case errors.As(err, &invalid):
		return api.CreateCharacter422JSONResponse{InvalidJSONResponse: toAPIValidation(invalid)}, nil
	case errors.Is(err, characters.ErrLimitReached):
		return api.CreateCharacter409JSONResponse{Error: api.ErrorErrorCharacterLimit}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPICharacter(created)
	return api.CreateCharacter201JSONResponse(body), err
}

// UpdateCharacter grava todos os campos do personagem do Usuário (RN-02, RN-15).
func (h charactersHandler) UpdateCharacter(ctx context.Context, req api.UpdateCharacterRequestObject) (api.UpdateCharacterResponseObject, error) {
	userID, ok, err := h.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.UpdateCharacter401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	updated, err := h.chars.Update(ctx, userID, req.Id, toInput(req.Body))
	var invalid *characters.ValidationError
	switch {
	case errors.As(err, &invalid):
		return api.UpdateCharacter422JSONResponse{InvalidJSONResponse: toAPIValidation(invalid)}, nil
	case errors.Is(err, characters.ErrNotFound):
		return api.UpdateCharacter404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse(notFound)}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPICharacter(updated)
	return api.UpdateCharacter200JSONResponse(body), err
}

// DeleteCharacter exclui o personagem do Usuário (RN-02, RN-14).
func (h charactersHandler) DeleteCharacter(ctx context.Context, req api.DeleteCharacterRequestObject) (api.DeleteCharacterResponseObject, error) {
	userID, ok, err := h.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.DeleteCharacter401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	err = h.chars.Delete(ctx, userID, req.Id)
	switch {
	case errors.Is(err, characters.ErrNotFound):
		return api.DeleteCharacter404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse(notFound)}, nil
	case err != nil:
		return nil, err
	}
	return api.DeleteCharacter204Response{}, nil
}

// SetMainCharacter torna o personagem o principal do Usuário (RN-02, RN-13).
func (h charactersHandler) SetMainCharacter(ctx context.Context, req api.SetMainCharacterRequestObject) (api.SetMainCharacterResponseObject, error) {
	userID, ok, err := h.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return api.SetMainCharacter401JSONResponse{NoSessionJSONResponse: api.NoSessionJSONResponse(noSession)}, nil
	}
	err = h.chars.SetMain(ctx, userID, req.Id)
	switch {
	case errors.Is(err, characters.ErrNotFound):
		return api.SetMainCharacter404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse(notFound)}, nil
	case err != nil:
		return nil, err
	}
	return api.SetMainCharacter204Response{}, nil
}

func toInput(body *api.CharacterInput) characters.Input {
	if body == nil {
		return characters.Input{}
	}
	in := characters.Input{Nick: body.Nick, ClassID: body.ClassId, Level: body.Level, Role: string(body.Role)}
	if body.Portrait != nil {
		in.Portrait = string(*body.Portrait)
	}
	if body.Link != nil {
		in.Link = *body.Link
	}
	return in
}

func toAPICharacter(c characters.Character) (api.Character, error) {
	id, err := uuid.Parse(c.ID)
	if err != nil {
		return api.Character{}, err
	}
	body := api.Character{
		Id:        id,
		Nick:      c.Nick,
		ClassId:   c.ClassID,
		Level:     c.Level,
		Role:      api.Role(c.Role),
		Portrait:  api.Portrait(c.Portrait),
		IsMain:    c.IsMain,
		CreatedAt: c.CreatedAt,
	}
	if c.Link != "" {
		body.Link = &c.Link
	}
	return body, nil
}

func toAPIValidation(e *characters.ValidationError) api.InvalidJSONResponse {
	fields := make([]api.FieldError, len(e.Fields))
	for i, f := range e.Fields {
		fields[i] = api.FieldError{Field: api.FieldErrorField(f.Field), Code: api.FieldErrorCode(f.Code)}
	}
	return api.InvalidJSONResponse{Error: api.ValidationErrorErrorValidation, Fields: fields}
}
