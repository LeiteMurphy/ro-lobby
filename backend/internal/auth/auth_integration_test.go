//go:build integration

package auth

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discord"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordfake"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/migrate"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

const redirect = "http://localhost:3000/auth/discord/callback"

type env struct {
	svc   *Service
	fake  *discordfake.Fake
	pool  *pgxpool.Pool
	clock time.Time
}

func setup(t *testing.T) *env {
	t.Helper()
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
	pool, err := database.Open(t.Context(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	fake := discordfake.New("123", "segredo", discordfake.User{ID: "111", Username: "grimbold", GlobalName: "Grimbold"})
	srv := httptest.NewServer(fake.Handler())
	t.Cleanup(srv.Close)

	e := &env{fake: fake, pool: pool, clock: time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)}
	e.svc = NewService(db.New(pool), discord.New(srv.URL+"/api", "123", "segredo", discord.WithTimeout(500*time.Millisecond)))
	e.svc.Now = func() time.Time { return e.clock }
	return e
}

func (e *env) login(t *testing.T, user discordfake.User) (string, User) {
	t.Helper()
	token, u, err := e.svc.Login(t.Context(), e.fake.IssueCode(user, redirect), redirect)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	return token, u
}

func (e *env) count(t *testing.T, table string) int {
	t.Helper()
	var n int
	if err := e.pool.QueryRow(t.Context(), "SELECT count(*) FROM "+table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// CA-01.1: o primeiro login cria o Usuário e a Sessão.
func TestLogin_CA01_1_FirstLoginCreatesUserAndSession(t *testing.T) {
	e := setup(t)
	token, u := e.login(t, e.fake.User)
	if u.Username != "grimbold" || u.GlobalName != "Grimbold" || u.ID == "" {
		t.Fatalf("usuário = %+v", u)
	}
	if token == "" || e.count(t, "users") != 1 || e.count(t, "sessions") != 1 {
		t.Fatalf("token vazio ou contagens erradas")
	}
	got, err := e.svc.Authenticate(t.Context(), token)
	if err != nil || got != u {
		t.Fatalf("Authenticate = %+v, %v", got, err)
	}
}

// CA-01.2: o login seguinte mantém o mesmo Usuário e atualiza o nome de exibição.
func TestLogin_CA01_2_NextLoginUpdatesSameUser(t *testing.T) {
	e := setup(t)
	_, first := e.login(t, discordfake.User{ID: "111", Username: "grimbold", GlobalName: "Grimbold"})
	e.clock = e.clock.Add(72 * time.Hour)
	_, second := e.login(t, discordfake.User{ID: "111", Username: "grimbold", GlobalName: "Grimbold, o Sábio"})

	if first.ID != second.ID || e.count(t, "users") != 1 {
		t.Fatalf("o segundo login criou outro Usuário: %v / %v", first.ID, second.ID)
	}
	if second.GlobalName != "Grimbold, o Sábio" {
		t.Errorf("nome de exibição = %q", second.GlobalName)
	}
}

// CA-01.4 / RN-07: o banco guarda o hash do token, não o token.
func TestLogin_CA01_4_StoresOnlyTokenHash(t *testing.T) {
	e := setup(t)
	token, _ := e.login(t, e.fake.User)
	var stored []byte
	if err := e.pool.QueryRow(t.Context(), "SELECT token_hash FROM sessions").Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stored, HashToken(token)) || bytes.Equal(stored, []byte(token)) {
		t.Fatal("o banco não guarda o hash SHA-256 do token")
	}
}

// CA-04.4 / RN-13: código recusado não cria Usuário nem Sessão.
func TestLogin_CA04_4_InvalidCodeCreatesNothing(t *testing.T) {
	e := setup(t)
	_, _, err := e.svc.Login(t.Context(), "nao-existe", redirect)
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("err = %v, esperado ErrInvalidCode", err)
	}
	if e.count(t, "users") != 0 || e.count(t, "sessions") != 0 {
		t.Fatal("código recusado criou registros")
	}
}

// CA-04.5 / RN-13: Discord lento não cria Usuário nem Sessão.
func TestLogin_CA04_5_SlowDiscordCreatesNothing(t *testing.T) {
	e := setup(t)
	e.fake.Delay = 2 * time.Second
	_, _, err := e.svc.Login(t.Context(), e.fake.IssueCode(e.fake.User, redirect), redirect)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, esperado ErrUnavailable", err)
	}
	if e.count(t, "users") != 0 || e.count(t, "sessions") != 0 {
		t.Fatal("Discord lento criou registros")
	}
}

// CA-06.3 / RN-09: usada há 29 dias, continua válida e passa a vencer 30 dias depois.
func TestAuthenticate_CA06_3_SlidingExpiry(t *testing.T) {
	e := setup(t)
	token, _ := e.login(t, e.fake.User)

	e.clock = e.clock.Add(29 * 24 * time.Hour)
	if _, err := e.svc.Authenticate(t.Context(), token); err != nil {
		t.Fatalf("29 dias: %v", err)
	}
	e.clock = e.clock.Add(29 * 24 * time.Hour) // 58 dias do login, 29 do último uso
	if _, err := e.svc.Authenticate(t.Context(), token); err != nil {
		t.Fatalf("29 dias depois do último uso: %v", err)
	}
}

// CA-06.4 / RN-09: usada há 31 dias, vencida.
func TestAuthenticate_CA06_4_ExpiredAfter30Days(t *testing.T) {
	e := setup(t)
	token, _ := e.login(t, e.fake.User)
	e.clock = e.clock.Add(31 * 24 * time.Hour)
	if _, err := e.svc.Authenticate(t.Context(), token); !errors.Is(err, ErrNoSession) {
		t.Fatalf("err = %v, esperado ErrNoSession", err)
	}
}

// RN-09: o último uso é gravado no máximo uma vez por hora.
func TestAuthenticate_RN09_TouchAtMostHourly(t *testing.T) {
	e := setup(t)
	token, _ := e.login(t, e.fake.User)
	lastUsed := func() time.Time {
		var ts time.Time
		if err := e.pool.QueryRow(t.Context(), "SELECT last_used_at FROM sessions").Scan(&ts); err != nil {
			t.Fatal(err)
		}
		return ts.UTC()
	}
	start := e.clock

	e.clock = start.Add(30 * time.Minute)
	if _, err := e.svc.Authenticate(t.Context(), token); err != nil {
		t.Fatal(err)
	}
	if !lastUsed().Equal(start) {
		t.Fatalf("gravou o uso com 30 min: %v", lastUsed())
	}
	e.clock = start.Add(90 * time.Minute)
	if _, err := e.svc.Authenticate(t.Context(), token); err != nil {
		t.Fatal(err)
	}
	if !lastUsed().Equal(start.Add(90 * time.Minute)) {
		t.Fatalf("não gravou o uso com 90 min: %v", lastUsed())
	}
}

// RN-10 / CA-06.2: token ausente, malformado ou desconhecido vale como visitante.
func TestAuthenticate_CA06_2_UnknownOrMalformedToken(t *testing.T) {
	e := setup(t)
	unknown, err := newToken()
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"", "nao-e-um-token", unknown} {
		if _, err := e.svc.Authenticate(t.Context(), token); !errors.Is(err, ErrNoSession) {
			t.Errorf("token %q: err = %v, esperado ErrNoSession", token, err)
		}
	}
}

// CA-03.2 / RN-11: sair num navegador não derruba o outro.
func TestLogout_CA03_2_OtherSessionsSurvive(t *testing.T) {
	e := setup(t)
	browser1, _ := e.login(t, e.fake.User)
	browser2, _ := e.login(t, e.fake.User)

	if err := e.svc.Logout(t.Context(), browser1); err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.Authenticate(t.Context(), browser1); !errors.Is(err, ErrNoSession) {
		t.Errorf("sessão encerrada continua: %v", err)
	}
	if _, err := e.svc.Authenticate(t.Context(), browser2); err != nil {
		t.Errorf("a outra sessão caiu: %v", err)
	}
	if err := e.svc.Logout(t.Context(), browser1); err != nil {
		t.Errorf("sair duas vezes deveria ser inofensivo: %v", err)
	}
}
