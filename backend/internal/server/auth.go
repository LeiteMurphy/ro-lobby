package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/auth"
)

// Authenticator é o que as rotas precisam do serviço de autenticação; o *auth.Service
// satisfaz.
type Authenticator interface {
	Login(ctx context.Context, code, redirectURI string) (string, auth.User, error)
	Authenticate(ctx context.Context, token string) (auth.User, error)
	Logout(ctx context.Context, token string) error
}

type authHandler struct {
	auth Authenticator
}

type tokenKey struct{}

// withSessionToken lê "Authorization: Bearer <token>" e põe o token no contexto, para as
// rotas do contrato que exigem sessão (ADR-07).
func withSessionToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
			r = r.WithContext(context.WithValue(r.Context(), tokenKey{}, strings.TrimSpace(token)))
		}
		next.ServeHTTP(w, r)
	})
}

func sessionToken(ctx context.Context) string {
	token, _ := ctx.Value(tokenKey{}).(string)
	return token
}

// CreateSessionFromDiscord troca o código por uma Sessão (RN-04, RN-05, RN-07, RN-13).
func (h authHandler) CreateSessionFromDiscord(ctx context.Context, req api.CreateSessionFromDiscordRequestObject) (api.CreateSessionFromDiscordResponseObject, error) {
	if req.Body == nil || req.Body.Code == "" || req.Body.RedirectUri == "" {
		return api.CreateSessionFromDiscord400JSONResponse{Error: api.ErrorErrorInvalidCode}, nil
	}
	token, user, err := h.auth.Login(ctx, req.Body.Code, req.Body.RedirectUri)
	switch {
	case errors.Is(err, auth.ErrInvalidCode):
		return api.CreateSessionFromDiscord400JSONResponse{Error: api.ErrorErrorInvalidCode}, nil
	case errors.Is(err, auth.ErrUnavailable):
		return api.CreateSessionFromDiscord502JSONResponse{Error: api.ErrorErrorDiscordUnavailable}, nil
	case err != nil:
		return nil, err
	}
	body, err := toAPIUser(user)
	if err != nil {
		return nil, err
	}
	return api.CreateSessionFromDiscord201JSONResponse{SessionToken: token, User: body}, nil
}

// GetMe devolve o Usuário da Sessão, ou 401 (RN-10, RN-16).
func (h authHandler) GetMe(ctx context.Context, _ api.GetMeRequestObject) (api.GetMeResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, sessionToken(ctx))
	if errors.Is(err, auth.ErrNoSession) {
		return api.GetMe401JSONResponse{Error: api.ErrorErrorNoSession}, nil
	}
	if err != nil {
		return nil, err
	}
	body, err := toAPIUser(user)
	if err != nil {
		return nil, err
	}
	return api.GetMe200JSONResponse(body), nil
}

// DeleteSession encerra só a Sessão atual (RN-11).
func (h authHandler) DeleteSession(ctx context.Context, _ api.DeleteSessionRequestObject) (api.DeleteSessionResponseObject, error) {
	if err := h.auth.Logout(ctx, sessionToken(ctx)); err != nil {
		return nil, err
	}
	return api.DeleteSession204Response{}, nil
}

func toAPIUser(u auth.User) (api.User, error) {
	id, err := uuid.Parse(u.ID)
	if err != nil {
		return api.User{}, err
	}
	body := api.User{Id: id, Username: u.Username}
	if u.GlobalName != "" {
		body.GlobalName = &u.GlobalName
	}
	return body, nil
}
