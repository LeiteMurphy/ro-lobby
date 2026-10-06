//go:build integration

package migrate

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
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

// T-10 (lobbies): fresh apaga, recria e migra só bancos terminados em _e2e.
func TestFresh_T10_RecreatesOnlyE2EDatabases(t *testing.T) {
	ctx := t.Context()
	base := testdb.New(t) // um banco descartável; o fresh usa outro nome, com _e2e no fim
	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	u.Path += FreshSuffix
	target := u.String()
	t.Cleanup(func() {
		admin := *u
		admin.Path = "/postgres"
		conn, err := pgx.Connect(context.Background(), admin.String())
		if err != nil {
			t.Error(err)
			return
		}
		defer func() { _ = conn.Close(context.Background()) }()
		_, _ = conn.Exec(context.Background(), "DROP DATABASE IF EXISTS "+pgx.Identifier{strings.TrimPrefix(u.Path, "/")}.Sanitize()+" WITH (FORCE)")
	})

	if err := Fresh(ctx, base); !errors.Is(err, ErrNotFreshable) {
		t.Fatalf("banco sem _e2e: err = %v", err)
	}
	for range 2 { // a segunda vez apaga o que a primeira criou
		if err := Fresh(ctx, target); err != nil {
			t.Fatal(err)
		}
		db, err := OpenDB(target)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, "INSERT INTO users (discord_id, username, created_at, last_login_at) VALUES ('1', 'a', now(), now())"); err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.QueryRowContext(ctx, "SELECT count(*) FROM users").Scan(&n); err != nil || n != 1 {
			t.Errorf("usuários = %d (err %v): o banco não começou vazio", n, err)
		}
		_ = db.Close()
	}
}
