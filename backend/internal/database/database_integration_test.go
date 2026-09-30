//go:build integration

package database

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

// CA-02.5: com o backend conectado, SHOW TimeZone devolve UTC, mesmo que o banco
// esteja configurado com outro fuso.
func TestConexao_CA02_5_FusoDaSessaoEUTC(t *testing.T) {
	url := testdb.New(t)
	ctx := context.Background()

	// Given: um banco cujo fuso padrão não é UTC.
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	var dbName string
	if err := conn.QueryRow(ctx, "SELECT current_database()").Scan(&dbName); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(ctx, "ALTER DATABASE "+pgx.Identifier{dbName}.Sanitize()+" SET timezone TO 'America/Sao_Paulo'"); err != nil {
		t.Fatal(err)
	}
	conn.Close(ctx)

	// When: o backend abre a conexão.
	pool, err := Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	// Then: o fuso da sessão é UTC.
	var show string
	if err := pool.QueryRow(ctx, "SHOW TimeZone").Scan(&show); err != nil {
		t.Fatal(err)
	}
	if show != "UTC" {
		t.Errorf("SHOW TimeZone = %q, esperado UTC", show)
	}
	got, err := db.New(pool).SessionTimeZone(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "UTC" {
		t.Errorf("SessionTimeZone = %q, esperado UTC", got)
	}
}
