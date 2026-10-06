package discordfake

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newServer(t *testing.T) (*Fake, *httptest.Server) {
	t.Helper()
	f := New("123", "segredo", User{ID: "111", Username: "grimbold"})
	srv := httptest.NewServer(f.Handler())
	t.Cleanup(srv.Close)
	return f, srv
}

func noRedirect() *http.Client {
	return &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

// do faz a requisição, fecha o corpo e devolve só o status e os headers.
func do(t *testing.T, c *http.Client, method, target, contentType, body string) (int, http.Header) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, target, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	return resp.StatusCode, resp.Header
}

// CA-01.3: a página de autorização registra o escopo e o state pedidos.
func TestAuthorize_CA01_3_RecordsScopeAndState(t *testing.T) {
	f, srv := newServer(t)
	q := url.Values{"client_id": {"123"}, "response_type": {"code"}, "scope": {"identify"},
		"state": {"abc"}, "redirect_uri": {"http://localhost:3000/cb"}}
	status, _ := do(t, http.DefaultClient, http.MethodGet, srv.URL+"/oauth2/authorize?"+q.Encode(), "", "")
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	if got := f.LastAuthorize(); got.Scope != "identify" || got.State != "abc" {
		t.Fatalf("autorização registrada = %+v", got)
	}
}

// CA-04.1: "Cancelar" volta para o redirect_uri com error=access_denied e o state.
func TestAuthorize_CA04_1_DenyRedirectsWithError(t *testing.T) {
	_, srv := newServer(t)
	form := url.Values{"redirect_uri": {"http://localhost:3000/cb"}, "state": {"abc"}, "decision": {"deny"}}
	_, header := do(t, noRedirect(), http.MethodPost, srv.URL+"/oauth2/authorize", "application/x-www-form-urlencoded", form.Encode())
	loc, _ := url.Parse(header.Get("Location"))
	if loc.Query().Get("error") != "access_denied" || loc.Query().Get("state") != "abc" || loc.Query().Has("code") {
		t.Fatalf("Location = %s", loc)
	}
}

// RNF-04 (login-discord): o falso imita a autorização; "Autorizar" volta ao redirect_uri
// com o code e o state.
func TestAuthorize_RNF04_AllowRedirectsWithCode(t *testing.T) {
	_, srv := newServer(t)
	form := url.Values{"redirect_uri": {"http://localhost:3000/cb"}, "state": {"abc"}, "decision": {"allow"}}
	_, header := do(t, noRedirect(), http.MethodPost, srv.URL+"/oauth2/authorize", "application/x-www-form-urlencoded", form.Encode())
	loc, _ := url.Parse(header.Get("Location"))
	if loc.Query().Get("code") == "" || loc.Query().Get("state") != "abc" {
		t.Fatalf("Location = %s", loc)
	}
}

// D-11 (personagens): a página aceita entrar como outro usuário, com ID fixo pelo nome; sem
// nome, entra o usuário padrão.
func TestAuthorize_D11_ChooseAnotherUser(t *testing.T) {
	f, _ := newServer(t)
	if got := f.userFor(""); got != f.User {
		t.Errorf("sem nome = %+v", got)
	}
	if got := f.userFor(" grimbold "); got != f.User {
		t.Errorf("nome do padrão = %+v", got)
	}
	bia := f.userFor("bia")
	if bia.Username != "bia" || bia.ID == f.User.ID || bia != f.userFor("bia") {
		t.Errorf("bia = %+v", bia)
	}
	if strings.Trim(bia.ID, "0123456789") != "" {
		t.Errorf("ID do Discord precisa ser só dígitos: %q", bia.ID)
	}

	// O código emitido pela página carrega o usuário escolhido.
	f2, srv := newServer(t)
	form := url.Values{"redirect_uri": {"http://localhost:3000/cb"}, "state": {"abc"}, "decision": {"allow"}, "username": {"bia"}}
	_, header := do(t, noRedirect(), http.MethodPost, srv.URL+"/oauth2/authorize", "application/x-www-form-urlencoded", form.Encode())
	loc, _ := url.Parse(header.Get("Location"))
	code := loc.Query().Get("code")
	f2.mu.Lock()
	got := f2.codes[code].user
	f2.mu.Unlock()
	if got.Username != "bia" {
		t.Fatalf("o código ficou com %+v", got)
	}
}

// RNF-04 (login-discord): o endpoint de token recusa JSON, como o Discord de verdade.
func TestToken_RNF04_RejectsJSON(t *testing.T) {
	_, srv := newServer(t)
	status, _ := do(t, http.DefaultClient, http.MethodPost, srv.URL+"/api/oauth2/token", "application/json", `{}`)
	if status != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d", status)
	}
}
