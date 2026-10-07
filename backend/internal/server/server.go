// Package server monta as rotas HTTP da API a partir da interface gerada do contrato.
package server

import (
	"log/slog"
	"net/http"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/health"
)

// handlers junta as rotas de cada área numa implementação da interface gerada.
type handlers struct {
	*health.Handler
	authHandler
	charactersHandler
	lobbiesHandler
	applicationsHandler
}

var _ api.StrictServerInterface = handlers{}

// New devolve o handler HTTP com todas as rotas do openapi.yaml.
func New(db health.Pinger, authenticator Authenticator, chars CharacterService, lobbySvc LobbyService, apps ApplicationService) http.Handler {
	session := charactersHandler{auth: authenticator, chars: chars}
	strict := api.NewStrictHandlerWithOptions(handlers{
		Handler:             health.New(db),
		authHandler:         authHandler{auth: authenticator},
		charactersHandler:   session,
		lobbiesHandler:      lobbiesHandler{session: session, lobbies: lobbySvc},
		applicationsHandler: applicationsHandler{session: session, apps: apps, lobbies: lobbySvc},
	}, nil, api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: internalError,
	})
	return withSessionToken(api.HandlerFromMux(strict, http.NewServeMux()))
}

// internalError registra o erro e responde 500 sem detalhes: o texto de um erro interno
// (do banco, por exemplo) não sai da API (RN-21 da spec personagens).
func internalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(r.Context(), "erro interno", "method", r.Method, "path", r.URL.Path, "err", err)
	http.Error(w, "erro interno", http.StatusInternalServerError)
}
