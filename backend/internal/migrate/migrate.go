// Package migrate aplica e reverte as migrações do goose (ADR-04, RN-09). É o mesmo
// código usado pelo comando cmd/migrate e pelos testes de integração.
package migrate

import (
	"database/sql"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/migrations"
)

// OpenDB abre um *sql.DB sobre o pgx, com a sessão em UTC (RN-08).
func OpenDB(databaseURL string) (*sql.DB, error) {
	cfg, err := database.PoolConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	return stdlib.OpenDB(*cfg.ConnConfig), nil
}

// NewProvider devolve o provider do goose com as migrações embutidas.
func NewProvider(db *sql.DB) (*goose.Provider, error) {
	return goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
}
