package server

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/ratelimit"
)

// Limite de requisições (spec endurecimento, RN-15 e RN-16).

// ClientIPHeader é o cabeçalho em que o web repassa o IP de quem acessa (RN-16). A API só
// é alcançada pela rede interna do Compose (RN-07), então o valor vem sempre do web.
const ClientIPHeader = "X-Client-IP"

// Limits são os três limites da RN-15.
type Limits struct {
	// General conta qualquer rota por IP.
	General *ratelimit.Limiter
	// Login conta a troca do código do login por IP.
	Login *ratelimit.Limiter
	// Actions conta as ações (requisições que mudam algo) por sessão.
	Actions *ratelimit.Limiter
}

// DefaultLimits devolve os limites da RN-15: 300, 10 e 30 por minuto.
func DefaultLimits() Limits {
	return Limits{
		General: ratelimit.New(300, time.Minute),
		Login:   ratelimit.New(10, time.Minute),
		Actions: ratelimit.New(30, time.Minute),
	}
}

// WithRateLimits liga o limite de requisições. Sem ela (testes de integração), não há
// limite; a API em produção sempre liga (cmd/api).
func WithRateLimits(l Limits) Option { return func(o *options) { o.limits = &l } }

func rateLimit(l Limits, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		ip := clientIP(r)
		if ok, retry := l.General.Allow(ip); !ok {
			tooManyRequests(w, retry)
			return
		}
		if r.Method == http.MethodPost && r.URL.Path == "/auth/discord" {
			if ok, retry := l.Login.Allow(ip); !ok {
				tooManyRequests(w, retry)
				return
			}
		}
		if key, ok := actionKey(r); ok {
			if ok, retry := l.Actions.Allow(key); !ok {
				tooManyRequests(w, retry)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP é o IP repassado pelo web ou, sem ele, o endereço da conexão (RN-16).
func clientIP(r *http.Request) string {
	if ip := net.ParseIP(strings.TrimSpace(r.Header.Get(ClientIPHeader))); ip != nil {
		return ip.String()
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// actionKey identifica a sessão de uma ação: requisição que muda algo, com sessão, fora do
// login e do logout. A chave é o hash do token, para o token não ficar em memória.
func actionKey(r *http.Request) (string, bool) {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return "", false
	}
	if r.URL.Path == "/auth/discord" || r.URL.Path == "/session" {
		return "", false
	}
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		return "", false
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:]), true
}

// tooManyRequests responde 429 com Retry-After em segundos, arredondado para cima.
func tooManyRequests(w http.ResponseWriter, retry time.Duration) {
	seconds := int((retry + time.Second - 1) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = w.Write([]byte(`{"error":"rate_limited"}`))
}
