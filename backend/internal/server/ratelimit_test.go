package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/ratelimit"
)

// Testes do limite de requisições (spec endurecimento, RN-15 e RN-16). A rota final é um
// handler que só conta; o relógio é falso.

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time { return c.t }

func testLimits(c *fakeClock) Limits {
	return Limits{
		General: ratelimit.NewWithClock(300, time.Minute, c.now),
		Login:   ratelimit.NewWithClock(10, time.Minute, c.now),
		Actions: ratelimit.NewWithClock(30, time.Minute, c.now),
	}
}

type counting struct{ n int }

func (c *counting) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	c.n++
	w.WriteHeader(http.StatusNoContent)
}

func send(h http.Handler, method, path, ip, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, path, nil)
	req.RemoteAddr = "172.18.0.5:41000" // o container do web
	if ip != "" {
		req.Header.Set(ClientIPHeader, ip)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func newLimited() (http.Handler, *counting, *fakeClock) {
	c := &fakeClock{t: time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)}
	next := &counting{}
	return rateLimit(testLimits(c), next), next, c
}

// CA-05.2: 30 ações da mesma sessão passam; a 31ª é 429 com Retry-After e não chega à rota.
func TestRateLimit_CA05_2_ActionsPerSession(t *testing.T) {
	h, next, _ := newLimited()
	for i := range 30 {
		if rec := send(h, http.MethodPost, "/lobbies", "203.0.113.1", "sessao-a"); rec.Code != http.StatusNoContent {
			t.Fatalf("ação %d: status %d", i+1, rec.Code)
		}
	}
	rec := send(h, http.MethodPost, "/lobbies", "203.0.113.1", "sessao-a")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("31ª ação: status %d, esperado 429", rec.Code)
	}
	if got := rec.Header().Get("Retry-After"); got != "60" {
		t.Fatalf("Retry-After = %q, esperado 60", got)
	}
	if rec.Body.String() != `{"error":"rate_limited"}` {
		t.Fatalf("corpo = %q", rec.Body.String())
	}
	if next.n != 30 {
		t.Fatalf("a rota recebeu %d requisições, esperado 30", next.n)
	}
}

// CA-05.2: outra sessão no mesmo minuto não é afetada; leituras não contam como ação.
func TestRateLimit_CA05_2_OtherSessionAndReadsUnaffected(t *testing.T) {
	h, _, _ := newLimited()
	for range 31 {
		send(h, http.MethodPost, "/lobbies", "203.0.113.1", "sessao-a")
	}
	if rec := send(h, http.MethodPost, "/lobbies", "203.0.113.1", "sessao-b"); rec.Code != http.StatusNoContent {
		t.Fatalf("outra sessão: status %d", rec.Code)
	}
	if rec := send(h, http.MethodGet, "/lobbies", "203.0.113.1", "sessao-a"); rec.Code != http.StatusNoContent {
		t.Fatalf("leitura da sessão bloqueada: status %d", rec.Code)
	}
}

// CA-05.3: 10 trocas de código de login por IP; a 11ª é 429 e outro IP entra.
func TestRateLimit_CA05_3_LoginPerIP(t *testing.T) {
	h, _, _ := newLimited()
	for i := range 10 {
		if rec := send(h, http.MethodPost, "/auth/discord", "203.0.113.7", ""); rec.Code != http.StatusNoContent {
			t.Fatalf("login %d: status %d", i+1, rec.Code)
		}
	}
	if rec := send(h, http.MethodPost, "/auth/discord", "203.0.113.7", ""); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("11º login: status %d, esperado 429", rec.Code)
	}
	if rec := send(h, http.MethodPost, "/auth/discord", "198.51.100.2", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("outro IP: status %d", rec.Code)
	}
}

// CA-05.4: depois do Retry-After, a ação volta a ser aceita.
func TestRateLimit_CA05_4_ReleasedAfterWindow(t *testing.T) {
	h, _, c := newLimited()
	for range 30 {
		send(h, http.MethodPost, "/lobbies", "203.0.113.1", "sessao-a")
	}
	rec := send(h, http.MethodPost, "/lobbies", "203.0.113.1", "sessao-a")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status %d, esperado 429", rec.Code)
	}
	c.t = c.t.Add(60 * time.Second)
	if rec := send(h, http.MethodPost, "/lobbies", "203.0.113.1", "sessao-a"); rec.Code != http.StatusNoContent {
		t.Fatalf("depois da janela: status %d", rec.Code)
	}
}

// CA-05.5 / RN-16: cada IP repassado pelo web conta separado, mesmo vindo do mesmo container.
func TestRateLimit_CA05_5_ForwardedIPIsTheKey(t *testing.T) {
	h, _, _ := newLimited()
	for range 300 {
		send(h, http.MethodGet, "/lobbies", "203.0.113.1", "")
	}
	if rec := send(h, http.MethodGet, "/lobbies", "203.0.113.1", ""); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("301ª do mesmo IP: status %d, esperado 429", rec.Code)
	}
	if rec := send(h, http.MethodGet, "/lobbies", "203.0.113.2", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("outro IP repassado: status %d", rec.Code)
	}
}

// RN-16: sem IP repassado (ou com valor inválido), vale o endereço da conexão.
func TestClientIP_RN16_FallbackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.RemoteAddr = "172.18.0.5:41000"
	if got := clientIP(req); got != "172.18.0.5" {
		t.Fatalf("sem cabeçalho: %q", got)
	}
	req.Header.Set(ClientIPHeader, "não é ip")
	if got := clientIP(req); got != "172.18.0.5" {
		t.Fatalf("cabeçalho inválido: %q", got)
	}
	req.Header.Set(ClientIPHeader, "2001:db8::1")
	if got := clientIP(req); got != "2001:db8::1" {
		t.Fatalf("IPv6: %q", got)
	}
}

// RN-15: o /healthz do healthcheck nunca conta nem é barrado.
func TestRateLimit_RN15_HealthzExempt(t *testing.T) {
	h, _, _ := newLimited()
	for range 400 {
		if rec := send(h, http.MethodGet, "/healthz", "", ""); rec.Code != http.StatusNoContent {
			t.Fatalf("healthz barrado: status %d", rec.Code)
		}
	}
}

// RN-15: o logout não conta como ação, para nunca prender alguém logado.
func TestRateLimit_RN15_LogoutNotAnAction(t *testing.T) {
	h, _, _ := newLimited()
	for range 30 {
		send(h, http.MethodPost, "/lobbies", "203.0.113.1", "sessao-a")
	}
	if rec := send(h, http.MethodDelete, "/session", "203.0.113.1", "sessao-a"); rec.Code != http.StatusNoContent {
		t.Fatalf("logout barrado: status %d", rec.Code)
	}
}

// RN-15: New sem WithRateLimits não limita (testes de integração); com ela, limita.
func TestNew_RN15_RateLimitsOption(t *testing.T) {
	plain := New(okPinger(), nil, nil, nil, nil, nil)
	limited := New(okPinger(), nil, nil, nil, nil, nil, WithRateLimits(Limits{
		General: ratelimit.New(1, time.Minute),
		Login:   ratelimit.New(1, time.Minute),
		Actions: ratelimit.New(1, time.Minute),
	}))
	for range 2 {
		send(plain, http.MethodGet, "/classes", "203.0.113.9", "")
		send(limited, http.MethodGet, "/classes", "203.0.113.9", "")
	}
	if rec := send(plain, http.MethodGet, "/classes", "203.0.113.9", ""); rec.Code == http.StatusTooManyRequests {
		t.Fatal("New sem opção limitou")
	}
	if rec := send(limited, http.MethodGet, "/classes", "203.0.113.9", ""); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("New com opção não limitou: %d", rec.Code)
	}
}
