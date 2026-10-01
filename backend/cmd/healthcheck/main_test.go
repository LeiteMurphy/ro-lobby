package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// RN-02 (home-local): o healthcheck da API só passa com o /healthz em 200, e é ele que
// segura o web até a API estar pronta.
func TestCheck_RN02_PassesOnlyOn200(t *testing.T) {
	for _, tc := range []struct {
		status  int
		wantErr bool
	}{
		{http.StatusOK, false},
		{http.StatusServiceUnavailable, true},
		{http.StatusInternalServerError, true},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tc.status)
		}))
		err := check(t.Context(), srv.URL+"/healthz")
		srv.Close()
		if (err != nil) != tc.wantErr {
			t.Errorf("status %d: err = %v, esperado erro = %v", tc.status, err, tc.wantErr)
		}
	}
}

func TestCheck_RN02_FailsWhenAPIIsDown(t *testing.T) {
	if err := check(t.Context(), "http://127.0.0.1:1/healthz"); err == nil {
		t.Fatal("esperado erro com a API fora do ar")
	}
}
