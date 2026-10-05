// Package discordfake é um Discord falso para testes (spec login-discord, D-05, RNF-04).
// Ele imita a página de autorização, a troca do código e o GET /users/@me, sem rede.
// Não entra na imagem da API.
package discordfake

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type User struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	GlobalName string `json:"global_name,omitempty"`
}

type grant struct {
	user        User
	redirectURI string
}

// AuthorizeRequest guarda o que chegou na última autorização, para os testes conferirem
// o escopo e o state (CA-01.3).
type AuthorizeRequest struct {
	ClientID     string
	RedirectURI  string
	ResponseType string
	Scope        string
	State        string
}

type Fake struct {
	ClientID     string
	ClientSecret string
	// User é quem "autoriza" na página de autorização.
	User User
	// Delay atrasa as respostas da API, para simular o Discord lento (CA-04.5).
	Delay time.Duration

	mu            sync.Mutex
	codes         map[string]grant
	tokens        map[string]User
	lastAuthorize AuthorizeRequest
}

func New(clientID, clientSecret string, user User) *Fake {
	return &Fake{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		User:         user,
		codes:        map[string]grant{},
		tokens:       map[string]User{},
	}
}

// IssueCode gera um código como se o usuário tivesse autorizado com esse redirect_uri.
func (f *Fake) IssueCode(user User, redirectURI string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	code := randomHex()
	f.codes[code] = grant{user: user, redirectURI: redirectURI}
	return code
}

func (f *Fake) LastAuthorize() AuthorizeRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastAuthorize
}

// Handler atende /oauth2/authorize (página) e /api/oauth2/token e /api/users/@me (API).
func (f *Fake) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /oauth2/authorize", f.authorizePage)
	mux.HandleFunc("POST /oauth2/authorize", f.authorizeDecision)
	mux.HandleFunc("POST /api/oauth2/token", f.token)
	mux.HandleFunc("GET /api/users/@me", f.me)
	return mux
}

var page = template.Must(template.New("authorize").Parse(`<!doctype html>
<html lang="pt-BR"><meta charset="utf-8"><title>Discord falso</title>
<body style="font-family:system-ui;background:#1e1f22;color:#f2f3f5;padding:40px">
<h1>Discord falso</h1>
<p>{{.User.Username}} vai autorizar o aplicativo {{.ClientID}} com o escopo <code>{{.Scope}}</code>.</p>
<form method="post">
<input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
<input type="hidden" name="state" value="{{.State}}">
<button name="decision" value="allow">Autorizar</button>
<button name="decision" value="deny">Cancelar</button>
</form></body></html>`))

func (f *Fake) authorizePage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := AuthorizeRequest{
		ClientID:     q.Get("client_id"),
		RedirectURI:  q.Get("redirect_uri"),
		ResponseType: q.Get("response_type"),
		Scope:        q.Get("scope"),
		State:        q.Get("state"),
	}
	f.mu.Lock()
	f.lastAuthorize = req
	f.mu.Unlock()
	if req.ClientID != f.ClientID || req.ResponseType != "code" || req.RedirectURI == "" {
		http.Error(w, "pedido de autorização inválido", http.StatusBadRequest)
		return
	}
	_ = page.Execute(w, map[string]any{
		"User": f.User, "ClientID": req.ClientID, "Scope": req.Scope,
		"RedirectURI": req.RedirectURI, "State": req.State,
	})
}

func (f *Fake) authorizeDecision(w http.ResponseWriter, r *http.Request) {
	redirectURI, state := r.FormValue("redirect_uri"), r.FormValue("state")
	target, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "redirect_uri inválido", http.StatusBadRequest)
		return
	}
	q := target.Query()
	if r.FormValue("decision") == "allow" {
		q.Set("code", f.IssueCode(f.User, redirectURI))
	} else {
		q.Set("error", "access_denied")
	}
	q.Set("state", state)
	target.RawQuery = q.Encode()
	http.Redirect(w, r, target.String(), http.StatusFound)
}

func (f *Fake) token(w http.ResponseWriter, r *http.Request) {
	time.Sleep(f.Delay)
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		writeError(w, http.StatusUnsupportedMediaType, "invalid_request")
		return
	}
	if r.FormValue("client_id") != f.ClientID || r.FormValue("client_secret") != f.ClientSecret {
		writeError(w, http.StatusUnauthorized, "invalid_client")
		return
	}
	f.mu.Lock()
	g, ok := f.codes[r.FormValue("code")]
	delete(f.codes, r.FormValue("code")) // código de uso único
	f.mu.Unlock()
	if !ok || r.FormValue("grant_type") != "authorization_code" || g.redirectURI != r.FormValue("redirect_uri") {
		writeError(w, http.StatusBadRequest, "invalid_grant")
		return
	}
	token := randomHex()
	f.mu.Lock()
	f.tokens[token] = g.user
	f.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": token, "token_type": "Bearer", "expires_in": 604800,
		"refresh_token": randomHex(), "scope": "identify",
	})
}

func (f *Fake) me(w http.ResponseWriter, r *http.Request) {
	time.Sleep(f.Delay)
	const prefix = "Bearer "
	auth := r.Header.Get("Authorization")
	f.mu.Lock()
	user, ok := f.tokens[auth[min(len(prefix), len(auth)):]]
	f.mu.Unlock()
	if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix || !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	body := map[string]any{"id": user.ID, "username": user.Username, "global_name": nil, "avatar": "abc123"}
	if user.GlobalName != "" {
		body["global_name"] = user.GlobalName
	}
	writeJSON(w, http.StatusOK, body)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code, "error_description": fmt.Sprintf("discordfake: %s", code)})
}

func randomHex() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
