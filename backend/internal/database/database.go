// Package database abre a conexão com o PostgreSQL.
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig interpreta a URL e força o fuso UTC em toda sessão (RN-08).
func PoolConfig(databaseURL string) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("DATABASE_URL inválida: %w", err)
	}
	cfg.ConnConfig.RuntimeParams["timezone"] = "UTC"
	return cfg, nil
}

// Open cria o pool de conexões. Ele não conecta na hora: a primeira conexão acontece
// no primeiro uso, então a API sobe mesmo com o banco fora do ar e o /healthz
// consegue responder 503 (RN-07).
func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := PoolConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}
