package database

import "testing"

// RN-08 / CA-02.5: a conexão pede o fuso UTC. A leitura real do fuso da sessão
// está no teste de integração.
func TestPoolConfig_RN08_ForcaUTC(t *testing.T) {
	cfg, err := PoolConfig("postgres://u:p@localhost:5432/db?timezone=America/Sao_Paulo")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got := cfg.ConnConfig.RuntimeParams["timezone"]; got != "UTC" {
		t.Fatalf("timezone = %q, esperado UTC", got)
	}
}

func TestPoolConfig_RejeitaURLInvalida(t *testing.T) {
	if _, err := PoolConfig("::não é url::"); err == nil {
		t.Fatal("esperado erro")
	}
}
