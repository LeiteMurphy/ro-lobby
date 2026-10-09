package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Testes unitários do /healthz: o banco é um Pinger falso, sem conexão real (RN-10,
// CA-06.7). O teste de integração com o PostgreSQL fica em healthz_integration_test.go.

type fakePinger func(ctx context.Context) error

func (f fakePinger) Ping(ctx context.Context) error { return f(ctx) }

func get(t *testing.T, h http.Handler, method string) (*httptest.ResponseRecorder, time.Duration) {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), method, "/healthz", nil)
	rec := httptest.NewRecorder()
	start := time.Now()
	h.ServeHTTP(rec, req)
	return rec, time.Since(start)
}

func assertBody(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, want map[string]string) {
	t.Helper()
	if rec.Code != wantStatus {
		t.Fatalf("status = %d, esperado %d", rec.Code, wantStatus)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, esperado application/json", ct)
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("corpo não é JSON: %v (%q)", err, rec.Body.String())
	}
	if len(got) != len(want) {
		t.Fatalf("corpo = %v, esperado %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, esperado %q", k, got[k], v)
		}
	}
}

// CA-02.1: banco respondendo → 200 com {"status":"ok","database":"ok"}.
func TestHealthz_CA02_1_DatabaseAvailable(t *testing.T) {
	h := New(fakePinger(func(context.Context) error { return nil }), nil, nil, nil, nil, nil)
	rec, _ := get(t, h, http.MethodGet)
	assertBody(t, rec, http.StatusOK, map[string]string{"status": "ok", "database": "ok"})
}

// CA-02.2: banco parado → 503 com {"status":"degraded","database":"unavailable"} em até 3 s.
func TestHealthz_CA02_2_DatabaseDown(t *testing.T) {
	h := New(fakePinger(func(context.Context) error { return errors.New("connection refused") }), nil, nil, nil, nil, nil)
	rec, took := get(t, h, http.MethodGet)
	assertBody(t, rec, http.StatusServiceUnavailable, map[string]string{"status": "degraded", "database": "unavailable"})
	if took > 3*time.Second {
		t.Errorf("resposta levou %s, esperado até 3s", took)
	}
}

// CA-02.3: banco que demora mais de 2 s para responder ao ping → 503.
func TestHealthz_CA02_3_SlowDatabase(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	// O Pinger ignora o contexto de propósito: o prazo tem que valer mesmo assim.
	h := New(fakePinger(func(context.Context) error { <-release; return nil }), nil, nil, nil, nil, nil)

	rec, took := get(t, h, http.MethodGet)
	assertBody(t, rec, http.StatusServiceUnavailable, map[string]string{"status": "degraded", "database": "unavailable"})
	if took < 2*time.Second || took > 3*time.Second {
		t.Errorf("resposta levou %s, esperado entre 2s e 3s", took)
	}
}

// CA-02.4: POST /healthz → 405.
func TestHealthz_CA02_4_MethodNotAllowed(t *testing.T) {
	called := false
	h := New(fakePinger(func(context.Context) error { called = true; return nil }), nil, nil, nil, nil, nil)
	rec, _ := get(t, h, http.MethodPost)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, esperado 405", rec.Code)
	}
	if called {
		t.Error("POST não deveria consultar o banco")
	}
}
