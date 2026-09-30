//go:build integration

// Package testdb cria um banco PostgreSQL vazio e descartável para cada teste de
// integração (RN-10). Ele parte da DATABASE_URL e apaga o banco no fim do teste.
package testdb

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

// New cria um banco vazio e devolve a URL de conexão dele.
func New(t *testing.T) string {
	t.Helper()

	adminURL := os.Getenv("DATABASE_URL")
	if adminURL == "" {
		t.Fatal("testes de integração precisam da DATABASE_URL de um PostgreSQL no ar")
	}

	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		t.Fatal(err)
	}
	name := "ro_lobby_test_" + hex.EncodeToString(buf)

	ctx := context.Background()
	admin, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		t.Fatalf("conectar em DATABASE_URL: %v", err)
	}
	defer func() { _ = admin.Close(ctx) }()

	if _, err := admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
		t.Fatalf("criar banco de teste: %v", err)
	}

	t.Cleanup(func() {
		conn, err := pgx.Connect(context.Background(), adminURL)
		if err != nil {
			t.Errorf("limpar banco de teste: %v", err)
			return
		}
		defer func() { _ = conn.Close(context.Background()) }()
		if _, err := conn.Exec(context.Background(), "DROP DATABASE IF EXISTS "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)"); err != nil {
			t.Errorf("apagar banco de teste: %v", err)
		}
	})

	u, err := url.Parse(adminURL)
	if err != nil {
		t.Fatalf("DATABASE_URL inválida: %v", err)
	}
	u.Path = "/" + name
	return u.String()
}
