package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Testes dos limites do servidor (spec endurecimento, RN-13 e RN-14). Os serviços são nil:
// se a requisição chegasse à rota, o teste quebraria, o que prova que nada é gravado.

func okPinger() fakePinger { return func(context.Context) error { return nil } }

func bigBody(n int) string { return `{"nick":"` + strings.Repeat("a", n) + `"}` }

// CA-05.1: corpo de 65 KiB com tamanho declarado → 413, sem chegar à rota.
func TestLimitBody_CA05_1_DeclaredTooLarge(t *testing.T) {
	h := New(okPinger(), nil, nil, nil, nil, nil)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/characters", strings.NewReader(bigBody(65<<10)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, esperado 413", rec.Code)
	}
}

// CA-05.1: sem tamanho declarado (chunked), a leitura para no limite e também dá 413.
func TestLimitBody_CA05_1_ChunkedTooLarge(t *testing.T) {
	h := New(okPinger(), nil, nil, nil, nil, nil)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/characters",
		io.NopCloser(strings.NewReader(bigBody(65<<10))))
	req.ContentLength = -1
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, esperado 413", rec.Code)
	}
}

// RN-13: um corpo dentro do limite passa pelo filtro e chega à leitura normal.
func TestLimitBody_RN13_WithinLimitPasses(t *testing.T) {
	var got int
	h := limitBody(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("leitura dentro do limite falhou: %v", err)
		}
		got = len(b)
	}))
	body := strings.Repeat("a", MaxBodyBytes)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(body))
	h.ServeHTTP(httptest.NewRecorder(), req)
	if got != MaxBodyBytes {
		t.Fatalf("leu %d bytes, esperado %d", got, MaxBodyBytes)
	}
}

// CA-05.6: os quatro prazos do servidor.
func TestNewHTTPServer_CA05_6_Timeouts(t *testing.T) {
	srv := NewHTTPServer(":8080", http.NotFoundHandler())
	checks := []struct {
		name      string
		got, want time.Duration
	}{
		{"ReadHeaderTimeout", srv.ReadHeaderTimeout, 5 * time.Second},
		{"ReadTimeout", srv.ReadTimeout, 15 * time.Second},
		{"WriteTimeout", srv.WriteTimeout, 30 * time.Second},
		{"IdleTimeout", srv.IdleTimeout, 120 * time.Second},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %s, esperado %s", c.name, c.got, c.want)
		}
	}
	if srv.Addr != ":8080" {
		t.Errorf("Addr = %q", srv.Addr)
	}
}
