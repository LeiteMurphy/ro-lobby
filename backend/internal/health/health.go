// Package health implementa o GET /healthz (RN-06, RN-07).
package health

import (
	"context"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
)

// PingTimeout é o prazo para o banco responder ao ping (RN-06, RN-07).
const PingTimeout = 2 * time.Second

// Pinger é o que o health check precisa do banco. O *pgxpool.Pool satisfaz.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Handler responde o /healthz conforme o contrato gerado do openapi.yaml.
type Handler struct {
	db      Pinger
	timeout time.Duration
}

var _ api.StrictServerInterface = (*Handler)(nil)

// New cria o handler com o prazo padrão de 2 segundos.
func New(db Pinger) *Handler {
	return &Handler{db: db, timeout: PingTimeout}
}

// GetHealthz responde 200 quando o banco responde ao ping dentro do prazo e 503
// quando falha ou demora mais que isso.
func (h *Handler) GetHealthz(ctx context.Context, _ api.GetHealthzRequestObject) (api.GetHealthzResponseObject, error) {
	if h.ping(ctx) != nil {
		return api.GetHealthz503JSONResponse{
			Status:   api.HealthStatusDegraded,
			Database: api.HealthDatabaseUnavailable,
		}, nil
	}
	return api.GetHealthz200JSONResponse{
		Status:   api.HealthStatusOk,
		Database: api.HealthDatabaseOk,
	}, nil
}

// ping aplica o prazo mesmo que o Pinger ignore o cancelamento do contexto, para a
// requisição nunca ficar presa esperando o banco.
func (h *Handler) ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- h.db.Ping(ctx) }()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func  demoLint( ) {}
