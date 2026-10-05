// Package discord troca o código do OAuth2 do Discord pelo perfil do usuário
// (spec login-discord, RN-01, RN-04, RN-13, RN-17).
package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL é a API oficial do Discord. Os testes apontam para um Discord falso.
const DefaultBaseURL = "https://discord.com/api"

// DefaultTimeout é o prazo da troca do código, somando as duas chamadas (RN-13).
const DefaultTimeout = 5 * time.Second

var (
	// ErrInvalidCode: o Discord recusou o código (cancelado, vencido, repetido ou de
	// outro redirect_uri).
	ErrInvalidCode = errors.New("discord: código recusado")
	// ErrUnavailable: o Discord não respondeu no prazo ou respondeu com erro.
	ErrUnavailable = errors.New("discord: indisponível")
)

// Profile é o que o escopo identify devolve e o RO Lobby guarda (RN-06). Os tokens do
// Discord não saem deste pacote (RN-04).
type Profile struct {
	ID         string
	Username   string
	GlobalName string
}

type Client struct {
	baseURL      string
	clientID     string
	clientSecret string
	timeout      time.Duration
	http         *http.Client
}

type Option func(*Client)

// WithTimeout troca o prazo padrão de 5 segundos (usado nos testes).
func WithTimeout(d time.Duration) Option { return func(c *Client) { c.timeout = d } }

func New(baseURL, clientID, clientSecret string, opts ...Option) *Client {
	c := &Client{
		baseURL:      strings.TrimRight(baseURL, "/"),
		clientID:     clientID,
		clientSecret: clientSecret,
		timeout:      DefaultTimeout,
		http:         &http.Client{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// ProfileFromCode troca o código pelo token, lê GET /users/@me e descarta o token.
func (c *Client) ProfileFromCode(ctx context.Context, code, redirectURI string) (Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	token, err := c.exchange(ctx, code, redirectURI)
	if err != nil {
		return Profile{}, err
	}
	return c.me(ctx, token)
}

func (c *Client) exchange(ctx context.Context, code, redirectURI string) (string, error) {
	form := url.Values{
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	// O Discord só aceita formulário neste endpoint.
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized:
		return "", ErrInvalidCode
	case resp.StatusCode != http.StatusOK:
		return "", fmt.Errorf("%w: token respondeu %d", ErrUnavailable, resp.StatusCode)
	}

	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.AccessToken == "" {
		return "", fmt.Errorf("%w: resposta de token inválida", ErrUnavailable)
	}
	return body.AccessToken, nil
}

func (c *Client) me(ctx context.Context, token string) (Profile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/users/@me", nil)
	if err != nil {
		return Profile{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return Profile{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return Profile{}, fmt.Errorf("%w: /users/@me respondeu %d", ErrUnavailable, resp.StatusCode)
	}

	var body struct {
		ID         string  `json:"id"`
		Username   string  `json:"username"`
		GlobalName *string `json:"global_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.ID == "" || body.Username == "" {
		return Profile{}, fmt.Errorf("%w: perfil inválido", ErrUnavailable)
	}
	p := Profile{ID: body.ID, Username: body.Username}
	if body.GlobalName != nil {
		p.GlobalName = *body.GlobalName
	}
	return p, nil
}
