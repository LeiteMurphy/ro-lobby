//go:build integration

package migrate

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/pressly/goose/v3"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

func setup(t *testing.T) (*sql.DB, *goose.Provider) {
	t.Helper()
	db, err := OpenDB(testdb.New(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	provider, err := NewProvider(db)
	if err != nil {
		t.Fatal(err)
	}
	return db, provider
}

func lastVersion(t *testing.T, p *goose.Provider) int64 {
	t.Helper()
	sources := p.ListSources()
	if len(sources) == 0 {
		t.Fatal("nenhuma migração embutida")
	}
	return sources[len(sources)-1].Version
}

func projectTables(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	err := db.QueryRowContext(t.Context(), `
		SELECT count(*) FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
		  AND table_name <> 'goose_db_version'`).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// CA-03.1: aplicar todas as migrações num banco vazio termina sem erro, e a versão
// registrada é a da última migração.
func TestMigrations_CA03_1_UpOnEmptyDatabase(t *testing.T) {
	db, p := setup(t)
	ctx := context.Background()

	if n := projectTables(t, db); n != 0 {
		t.Fatalf("banco de teste deveria começar vazio, tem %d tabelas", n)
	}
	if _, err := p.Up(ctx); err != nil {
		t.Fatalf("up: %v", err)
	}
	got, err := p.GetDBVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if want := lastVersion(t, p); got != want {
		t.Fatalf("versão = %d, esperado %d", got, want)
	}
}

// CA-03.2: reverter todas as migrações termina sem erro, e o banco volta a não ter
// tabelas do projeto.
func TestMigrations_CA03_2_DownToZero(t *testing.T) {
	db, p := setup(t)
	ctx := context.Background()

	if _, err := p.Up(ctx); err != nil {
		t.Fatalf("up: %v", err)
	}
	if _, err := p.DownTo(ctx, 0); err != nil {
		t.Fatalf("down até 0: %v", err)
	}
	got, err := p.GetDBVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Fatalf("versão = %d, esperado 0", got)
	}
	if n := projectTables(t, db); n != 0 {
		t.Fatalf("sobraram %d tabelas do projeto", n)
	}
	// A linha de base desfaz o fuso do banco (D-04).
	var setting sql.NullString
	err = db.QueryRowContext(t.Context(), `
		SELECT s.setconfig::text FROM pg_db_role_setting s
		JOIN pg_database d ON d.oid = s.setdatabase
		WHERE d.datname = current_database() AND s.setrole = 0`).Scan(&setting)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatal(err)
	}
	if setting.Valid && setting.String != "" && setting.String != "{}" {
		t.Errorf("configuração do banco não foi desfeita: %s", setting.String)
	}
}
