package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/auth"
)

// Testes das rotas de autenticação com um Authenticator falso, sem banco (RN-16).

const userID = "6f1c2b8e-3a4d-4e5f-9a1b-2c3d4e5f6a7b"

type fakeAuth struct {
	loginErr  error
	loggedOut string
	gotCode   string
	gotURI    string
}

func (f *fakeAuth) Login(_ context.Context, code, redirectURI string) (string, auth.User, error) {
	f.gotCode, f.gotURI = code, redirectURI
	if f.loginErr != nil {
		return "", auth.User{}, f.loginErr
	}
	return "token-novo", auth.User{ID: userID, Username: "grimbold", GlobalName: "Grimbold"}, nil
}

func (f *fakeAuth) Authenticate(_ context.Context, token string) (auth.User, error) {
	if token != "token-valido" {
		return auth.User{}, auth.ErrNoSession
	}
	return auth.User{ID: userID, Username: "mirai.exe"}, nil
}

func (f *fakeAuth) Logout(_ context.Context, token string) error {
	f.loggedOut = token
	return nil
}

func call(t *testing.T, h http.Handler, method, path, token, body string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var parsed map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &parsed)
	return rec, parsed
}

func newAuthServer(f *fakeAuth) http.Handler {
	return New(fakePinger(func(context.Context) error { return nil }), f)
}

// CA-01.1 / RN-16: código aceito → 201 com o token e o usuário.
func TestAuthDiscord_CA01_1_Created(t *testing.T) {
	f := &fakeAuth{}
	rec, body := call(t, newAuthServer(f), http.MethodPost, "/auth/discord", "",
		`{"code":"abc","redirectUri":"http://localhost:3000/auth/discord/callback"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, corpo %s", rec.Code, rec.Body)
	}
	user, _ := body["user"].(map[string]any)
	if body["sessionToken"] != "token-novo" || user["id"] != userID || user["globalName"] != "Grimbold" {
		t.Fatalf("corpo = %v", body)
	}
	if f.gotCode != "abc" || f.gotURI != "http://localhost:3000/auth/discord/callback" {
		t.Errorf("repassou code=%q redirectUri=%q", f.gotCode, f.gotURI)
	}
}

// CA-04.4: código recusado ou ausente → 400 invalid_code.
func TestAuthDiscord_CA04_4_InvalidCode(t *testing.T) {
	for name, body := range map[string]string{
		"recusado": `{"code":"x","redirectUri":"http://localhost:3000/cb"}`,
		"sem code": `{"code":"","redirectUri":"http://localhost:3000/cb"}`,
	} {
		rec, parsed := call(t, newAuthServer(&fakeAuth{loginErr: auth.ErrInvalidCode}), http.MethodPost, "/auth/discord", "", body)
		if rec.Code != http.StatusBadRequest || parsed["error"] != "invalid_code" {
			t.Errorf("%s: status = %d, corpo = %v", name, rec.Code, parsed)
		}
	}
}

// CA-04.5: Discord fora do ar → 502 discord_unavailable.
func TestAuthDiscord_CA04_5_DiscordUnavailable(t *testing.T) {
	rec, parsed := call(t, newAuthServer(&fakeAuth{loginErr: auth.ErrUnavailable}), http.MethodPost, "/auth/discord", "",
		`{"code":"x","redirectUri":"http://localhost:3000/cb"}`)
	if rec.Code != http.StatusBadGateway || parsed["error"] != "discord_unavailable" {
		t.Fatalf("status = %d, corpo = %v", rec.Code, parsed)
	}
}

func TestAuthDiscord_UnexpectedErrorIs500(t *testing.T) {
	rec, _ := call(t, newAuthServer(&fakeAuth{loginErr: errors.New("banco caiu")}), http.MethodPost, "/auth/discord", "",
		`{"code":"x","redirectUri":"http://localhost:3000/cb"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

// CA-06.1: sessão válida → 200 com o usuário; sem nome de exibição, globalName é nulo.
func TestMe_CA06_1_ValidSession(t *testing.T) {
	rec, body := call(t, newAuthServer(&fakeAuth{}), http.MethodGet, "/me", "token-valido", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if body["id"] != userID || body["username"] != "mirai.exe" || body["globalName"] != nil {
		t.Fatalf("corpo = %v", body)
	}
	if _, present := body["globalName"]; !present {
		t.Error("globalName deveria vir como null, não ausente")
	}
}

// CA-06.2: sem sessão, ou com token desconhecido → 401 no_session.
func TestMe_CA06_2_NoSession(t *testing.T) {
	for _, token := range []string{"", "token-desconhecido"} {
		rec, body := call(t, newAuthServer(&fakeAuth{}), http.MethodGet, "/me", token, "")
		if rec.Code != http.StatusUnauthorized || body["error"] != "no_session" {
			t.Errorf("token %q: status = %d, corpo = %v", token, rec.Code, body)
		}
	}
}

// CA-03.1 / RN-11: sair encerra a sessão do token enviado → 204.
func TestDeleteSession_CA03_1_LogsOutThisToken(t *testing.T) {
	f := &fakeAuth{}
	rec, _ := call(t, newAuthServer(f), http.MethodDelete, "/session", "token-valido", "")
	if rec.Code != http.StatusNoContent || f.loggedOut != "token-valido" {
		t.Fatalf("status = %d, encerrou %q", rec.Code, f.loggedOut)
	}
}

// RN-16: métodos fora do contrato → 405.
func TestAuthRoutes_RN16_MethodNotAllowed(t *testing.T) {
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/auth/discord"}, {http.MethodPost, "/me"}, {http.MethodGet, "/session"},
	} {
		rec, _ := call(t, newAuthServer(&fakeAuth{}), tc.method, tc.path, "", "")
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s: status = %d, esperado 405", tc.method, tc.path, rec.Code)
		}
	}
}
