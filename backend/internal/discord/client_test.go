package discord

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordfake"
)

const redirect = "http://localhost:3000/auth/discord/callback"

func newFake(t *testing.T) (*discordfake.Fake, *httptest.Server) {
	t.Helper()
	fake := discordfake.New("123", "segredo", discordfake.User{ID: "111", Username: "grimbold", GlobalName: "Grimbold"})
	srv := httptest.NewServer(fake.Handler())
	t.Cleanup(srv.Close)
	return fake, srv
}

// RN-04 / CA-01.1: troca o código e lê o perfil; o token do Discord não sai do pacote.
func TestProfileFromCode_RN04_ValidCode(t *testing.T) {
	fake, srv := newFake(t)
	code := fake.IssueCode(fake.User, redirect)

	got, err := New(srv.URL+"/api", "123", "segredo").ProfileFromCode(t.Context(), code, redirect)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got != (Profile{ID: "111", Username: "grimbold", GlobalName: "Grimbold"}) {
		t.Fatalf("perfil = %+v", got)
	}
}

func TestProfileFromCode_RN06_NoGlobalName(t *testing.T) {
	fake, srv := newFake(t)
	code := fake.IssueCode(discordfake.User{ID: "222", Username: "mirai.exe"}, redirect)
	got, err := New(srv.URL+"/api", "123", "segredo").ProfileFromCode(t.Context(), code, redirect)
	if err != nil || got.GlobalName != "" || got.Username != "mirai.exe" {
		t.Fatalf("perfil = %+v, err = %v", got, err)
	}
}

// RN-13 / CA-04.4: código recusado pelo Discord.
func TestProfileFromCode_CA04_4_InvalidCode(t *testing.T) {
	_, srv := newFake(t)
	_, err := New(srv.URL+"/api", "123", "segredo").ProfileFromCode(t.Context(), "nao-existe", redirect)
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("err = %v, esperado ErrInvalidCode", err)
	}
}

// CA-04.6: o mesmo código não vale duas vezes.
func TestProfileFromCode_CA04_6_CodeIsSingleUse(t *testing.T) {
	fake, srv := newFake(t)
	code := fake.IssueCode(fake.User, redirect)
	c := New(srv.URL+"/api", "123", "segredo")
	if _, err := c.ProfileFromCode(t.Context(), code, redirect); err != nil {
		t.Fatal(err)
	}
	if _, err := c.ProfileFromCode(t.Context(), code, redirect); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("segundo uso: err = %v, esperado ErrInvalidCode", err)
	}
}

// RN-04: o redirect_uri enviado na troca precisa ser o mesmo da autorização.
func TestProfileFromCode_RN04_RedirectURIMustMatch(t *testing.T) {
	fake, srv := newFake(t)
	code := fake.IssueCode(fake.User, redirect)
	_, err := New(srv.URL+"/api", "123", "segredo").ProfileFromCode(t.Context(), code, "http://outro/callback")
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("err = %v, esperado ErrInvalidCode", err)
	}
}

// RN-03 / AJ-03: Client Secret errado também é recusado, e a API registra o motivo no log,
// sem o segredo.
func TestProfileFromCode_RN03_WrongSecret(t *testing.T) {
	fake, srv := newFake(t)
	code := fake.IssueCode(fake.User, redirect)
	var logs bytes.Buffer
	c := New(srv.URL+"/api", "123", "errado", WithLogger(slog.New(slog.NewTextHandler(&logs, nil))))
	_, err := c.ProfileFromCode(t.Context(), code, redirect)
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("err = %v, esperado ErrInvalidCode", err)
	}
	if !strings.Contains(logs.String(), "level=ERROR") || !strings.Contains(logs.String(), "DISCORD_CLIENT_SECRET") {
		t.Errorf("log = %q, esperado o aviso de Client Secret", logs.String())
	}
	if strings.Contains(logs.String(), "errado") {
		t.Errorf("o log vazou o Client Secret: %q", logs.String())
	}
}

// AJ-03: código recusado (400) não é problema de configuração, então não gera o aviso.
func TestProfileFromCode_RN03_InvalidCodeDoesNotLogSecretWarning(t *testing.T) {
	_, srv := newFake(t)
	var logs bytes.Buffer
	c := New(srv.URL+"/api", "123", "segredo", WithLogger(slog.New(slog.NewTextHandler(&logs, nil))))
	if _, err := c.ProfileFromCode(t.Context(), "nao-existe", redirect); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("err = %v", err)
	}
	if logs.Len() != 0 {
		t.Errorf("log inesperado: %q", logs.String())
	}
}

// RN-13 / CA-04.5: Discord lento vira indisponível dentro do prazo.
func TestProfileFromCode_CA04_5_SlowDiscord(t *testing.T) {
	fake, srv := newFake(t)
	fake.Delay = 2 * time.Second
	code := fake.IssueCode(fake.User, redirect)

	start := time.Now()
	_, err := New(srv.URL+"/api", "123", "segredo", WithTimeout(200*time.Millisecond)).ProfileFromCode(t.Context(), code, redirect)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, esperado ErrUnavailable", err)
	}
	if took := time.Since(start); took > time.Second {
		t.Errorf("levou %s, esperado perto de 200ms", took)
	}
}

func TestDefaultTimeout_RN13_IsFiveSeconds(t *testing.T) {
	if DefaultTimeout != 5*time.Second || New("x", "y", "z").timeout != 5*time.Second {
		t.Fatalf("prazo padrão = %s, esperado 5s", DefaultTimeout)
	}
}

// CA-04.5: Discord fora do ar.
func TestProfileFromCode_CA04_5_DiscordDown(t *testing.T) {
	_, err := New("http://127.0.0.1:1/api", "123", "segredo").ProfileFromCode(t.Context(), "x", redirect)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, esperado ErrUnavailable", err)
	}
}

// RN-13: erro 5xx do Discord também é indisponível.
func TestProfileFromCode_RN13_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)
	_, err := New(srv.URL, "123", "segredo").ProfileFromCode(t.Context(), "x", redirect)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, esperado ErrUnavailable", err)
	}
}
