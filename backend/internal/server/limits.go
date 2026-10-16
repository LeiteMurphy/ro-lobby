package server

import (
	"errors"
	"net/http"
	"time"
)

// Limites do servidor da API (spec endurecimento, RN-13 e RN-14).

// MaxBodyBytes é o maior corpo de requisição aceito (RN-13). Os formulários do RO Lobby
// cabem com folga: o maior é um lobby com observação, bem abaixo de 1 KiB.
const MaxBodyBytes = 64 << 10

// Prazos do servidor (RN-14).
const (
	ReadHeaderTimeout = 5 * time.Second
	ReadTimeout       = 15 * time.Second
	WriteTimeout      = 30 * time.Second
	IdleTimeout       = 120 * time.Second
)

// NewHTTPServer devolve o servidor da API com os prazos da RN-14.
func NewHTTPServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: ReadHeaderTimeout,
		ReadTimeout:       ReadTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       IdleTimeout,
	}
}

// limitBody recusa com 413 o corpo maior que MaxBodyBytes (RN-13). Com o tamanho
// declarado, recusa antes de ler; sem ele (chunked), a leitura para no limite e o erro vira
// 413 em requestError.
func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > MaxBodyBytes {
			tooLarge(w)
			return
		}
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		}
		next.ServeHTTP(w, r)
	})
}

// requestError responde a uma requisição que não pôde ser lida: 413 quando passou do
// limite, 400 no resto.
func requestError(w http.ResponseWriter, _ *http.Request, err error) {
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		tooLarge(w)
		return
	}
	http.Error(w, err.Error(), http.StatusBadRequest)
}

func tooLarge(w http.ResponseWriter) {
	http.Error(w, "corpo da requisição grande demais", http.StatusRequestEntityTooLarge)
}
