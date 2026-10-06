//go:build integration

package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/auth"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discord"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordfake"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/migrate"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

// Fluxo completo pela API, com o PostgreSQL real e o Discord falso: entrar, ver o
// usuário, sair e ser recusado depois (CA-01.1, CA-06.1, CA-06.2, CA-03.1).
func TestAuthFlowIntegration_CA06_1_LoginMeLogout(t *testing.T) {
	url := testdb.New(t)
	sqlDB, err := migrate.OpenDB(url)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := migrate.NewProvider(sqlDB)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(t.Context()); err != nil {
		t.Fatal(err)
	}
	_ = sqlDB.Close()
	pool, err := database.Open(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	fake := discordfake.New("123", "segredo", discordfake.User{ID: "111", Username: "grimbold", GlobalName: "Grimbold"})
	discordSrv := httptest.NewServer(fake.Handler())
	defer discordSrv.Close()
	svc := auth.NewService(pool, discord.New(discordSrv.URL+"/api", "123", "segredo"))
	h := New(pool, svc, nil, nil)

	const redirect = "http://localhost:3000/auth/discord/callback"
	code := fake.IssueCode(fake.User, redirect)
	rec, body := call(t, h, http.MethodPost, "/auth/discord", "", `{"code":"`+code+`","redirectUri":"`+redirect+`"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("login: status = %d, corpo %s", rec.Code, rec.Body)
	}
	token, _ := body["sessionToken"].(string)

	rec, me := call(t, h, http.MethodGet, "/me", token, "")
	if rec.Code != http.StatusOK || me["username"] != "grimbold" || me["globalName"] != "Grimbold" {
		t.Fatalf("me: status = %d, corpo = %v", rec.Code, me)
	}

	rec, _ = call(t, h, http.MethodDelete, "/session", token, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout: status = %d", rec.Code)
	}
	rec, _ = call(t, h, http.MethodGet, "/me", token, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("me depois de sair: status = %d, esperado 401", rec.Code)
	}
}
