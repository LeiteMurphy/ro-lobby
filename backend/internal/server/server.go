// Package server monta as rotas HTTP da API a partir da interface gerada do contrato.
package server

import (
	"net/http"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/health"
)

// New devolve o handler HTTP com todas as rotas do openapi.yaml.
func New(db health.Pinger) http.Handler {
	strict := api.NewStrictHandler(health.New(db), nil)
	return api.HandlerFromMux(strict, http.NewServeMux())
}
