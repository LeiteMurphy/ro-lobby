// Package server monta as rotas HTTP da API a partir da interface gerada do contrato.
package server

import (
	"net/http"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/health"
)

// handlers junta as rotas de cada área numa implementação da interface gerada.
type handlers struct {
	*health.Handler
	authHandler
	charactersHandler
}

var _ api.StrictServerInterface = handlers{}

// New devolve o handler HTTP com todas as rotas do openapi.yaml.
func New(db health.Pinger, authenticator Authenticator, chars CharacterService) http.Handler {
	strict := api.NewStrictHandler(handlers{
		Handler:           health.New(db),
		authHandler:       authHandler{auth: authenticator},
		charactersHandler: charactersHandler{auth: authenticator, chars: chars},
	}, nil)
	return withSessionToken(api.HandlerFromMux(strict, http.NewServeMux()))
}
