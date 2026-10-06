//go:build integration

package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

// CA-02.1 com um PostgreSQL real: banco respondendo → 200.
func TestHealthzIntegration_CA02_1_DatabaseAvailable(t *testing.T) {
	pool, err := database.Open(context.Background(), testdb.New(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	rec, _ := get(t, New(pool, nil), http.MethodGet)
	assertBody(t, rec, http.StatusOK, map[string]string{"status": "ok", "database": "ok"})
}

// CA-02.2 com o driver real: banco inacessível → 503 em até 3 s.
func TestHealthzIntegration_CA02_2_DatabaseDown(t *testing.T) {
	// Porta 1 no loopback: nada escuta ali, então a conexão é recusada.
	pool, err := database.Open(context.Background(), "postgres://ro_lobby:x@127.0.0.1:1/ro_lobby?sslmode=disable&connect_timeout=5")
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	rec, took := get(t, New(pool, nil), http.MethodGet)
	assertBody(t, rec, http.StatusServiceUnavailable, map[string]string{"status": "degraded", "database": "unavailable"})
	if took > 3*time.Second {
		t.Errorf("resposta levou %s, esperado até 3s", took)
	}
}
