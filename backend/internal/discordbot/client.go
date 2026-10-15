// Package discordbot fala com o Discord como o bot do RO Lobby, só por REST (spec
// discord-bot, ADR-09, design D-01, D-04 e D-06). Cria e apaga a sala (uma categoria com
// um canal de texto e um de voz), libera e revoga o acesso, posta mensagens e inclui o
// jogador no servidor oficial. O token do bot nunca vai para o log (RNF-01).
package discordbot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL é a API oficial do Discord. Os testes apontam para o Discord falso.
const DefaultBaseURL = "https://discord.com/api"

// DefaultTimeout é o prazo de cada chamada (RNF-02).
const DefaultTimeout = 5 * time.Second

var (
	// ErrUnavailable: o Discord não respondeu no prazo, respondeu com erro ou seguiu no
	// limite de chamadas (RN-20, RN-21).
	ErrUnavailable = errors.New("discordbot: Discord indisponível")
	// ErrGuildFull: o servidor chegou ao limite de canais do Discord (RN-20).
	ErrGuildFull = errors.New("discordbot: servidor no limite de canais")
	// ErrInvalidToken: o token do jogador não serve para incluir no servidor (RN-03).
	ErrInvalidToken = errors.New("discordbot: token do jogador recusado")
)

// Códigos de erro do Discord que importam aqui.
const (
	codeUnknownChannel   = 10003
	codeMaxGuildChannels = 30013
)

// Tipos de canal do Discord.
const (
	channelText     = 0
	channelVoice    = 2
	channelCategory = 4
)

// Permissões do Discord (bits). O membro vê, lê o histórico, escreve, entra na voz e fala.
const (
	permView    = 1 << 10
	permSend    = 1 << 11
	permHistory = 1 << 16
	permConnect = 1 << 20
	permSpeak   = 1 << 21

	memberAllow = permView | permSend | permHistory | permConnect | permSpeak
)

// Room são os IDs da sala no Discord (D-04).
type Room struct {
	CategoryID string
	TextID     string
	VoiceID    string
}

// Channels devolve os três canais, da categoria para a voz.
func (r Room) Channels() []string { return []string{r.CategoryID, r.TextID, r.VoiceID} }

// Client é o bot do RO Lobby num servidor.
type Client struct {
	baseURL string
	token   string
	guildID string
	// botUserID é o usuário do bot, que precisa ver a sala que nega @everyone.
	botUserID string
	timeout   time.Duration
	http      *http.Client
	log       *slog.Logger
}

type Option func(*Client)

// WithTimeout troca o prazo de cada chamada (usado nos testes).
func WithTimeout(d time.Duration) Option { return func(c *Client) { c.timeout = d } }

// WithLogger troca o logger padrão (usado nos testes).
func WithLogger(l *slog.Logger) Option { return func(c *Client) { c.log = l } }

// New monta o bot. botUserID é o ID do usuário do bot, que para aplicativos do Discord é o
// mesmo do aplicativo (DISCORD_CLIENT_ID).
func New(baseURL, token, guildID, botUserID string, opts ...Option) *Client {
	c := &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		token:     token,
		guildID:   guildID,
		botUserID: botUserID,
		timeout:   DefaultTimeout,
		http:      &http.Client{},
		log:       slog.Default(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// GuildID é o servidor oficial.
func (c *Client) GuildID() string { return c.guildID }

type overwrite struct {
	ID    string `json:"id"`
	Type  int    `json:"type"` // 0 = cargo, 1 = membro
	Allow string `json:"allow"`
	Deny  string `json:"deny"`
}

func bits(n int64) string { return strconv.FormatInt(n, 10) }

// overwrites da sala (RN-07): @everyone (o cargo com o ID do servidor) não vê; o bot e os
// membros veem e falam.
func (c *Client) overwrites(allow []string) []overwrite {
	out := []overwrite{
		{ID: c.guildID, Type: 0, Allow: "0", Deny: bits(permView)},
		{ID: c.botUserID, Type: 1, Allow: bits(memberAllow), Deny: "0"},
	}
	for _, id := range allow {
		if id == c.botUserID {
			continue
		}
		out = append(out, overwrite{ID: id, Type: 1, Allow: bits(memberAllow), Deny: "0"})
	}
	return out
}

// CreateRoom cria a categoria, o canal de texto e o de voz com as permissões (RN-05,
// RN-07, D-04). Se uma criação falha, apaga o que já criou antes de devolver o erro
// (CA-02.7).
func (c *Client) CreateRoom(ctx context.Context, name string, allow []string) (Room, error) {
	ow := c.overwrites(allow)
	var created []string
	create := func(kind int, channelName, parent string) (string, error) {
		body := map[string]any{"name": channelName, "type": kind, "permission_overwrites": ow}
		if parent != "" {
			body["parent_id"] = parent
		}
		var out struct {
			ID string `json:"id"`
		}
		if err := c.do(ctx, http.MethodPost, "/guilds/"+c.guildID+"/channels", body, &out); err != nil {
			return "", err
		}
		created = append(created, out.ID)
		return out.ID, nil
	}
	room, err := func() (Room, error) {
		var r Room
		var err error
		if r.CategoryID, err = create(channelCategory, name, ""); err != nil {
			return r, err
		}
		if r.TextID, err = create(channelText, textName(name), r.CategoryID); err != nil {
			return r, err
		}
		if r.VoiceID, err = create(channelVoice, name, r.CategoryID); err != nil {
			return r, err
		}
		return r, nil
	}()
	if err != nil {
		c.undo(created)
		return Room{}, err
	}
	return room, nil
}

// undo apaga os canais de uma criação que falhou, com um prazo próprio: o pedido original
// pode já ter estourado o dele.
func (c *Client) undo(ids []string) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	for i := len(ids) - 1; i >= 0; i-- {
		if err := c.DeleteChannel(ctx, ids[i]); err != nil {
			c.log.Error("canal órfão no Discord depois de uma criação que falhou", "channel_id", ids[i], "err", err)
		}
	}
}

// textName é o nome do canal de texto no formato do Discord: minúsculo e com hífens.
func textName(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(name) {
		switch r {
		case ' ', '·', '-', '_':
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
				dash = true
			}
		default:
			b.WriteRune(r)
			dash = false
		}
	}
	return strings.TrimRight(b.String(), "-")
}

// Allow libera a sala para o membro (RN-10, RN-13).
func (c *Client) Allow(ctx context.Context, channelID, userID string) error {
	err := c.do(ctx, http.MethodPut, "/channels/"+channelID+"/permissions/"+userID,
		overwrite{ID: userID, Type: 1, Allow: bits(memberAllow), Deny: "0"}, nil)
	if errors.Is(err, errNotFound) {
		return nil // canal apagado à mão: nada a liberar
	}
	return err
}

// Revoke tira a permissão do membro (RN-11).
func (c *Client) Revoke(ctx context.Context, channelID, userID string) error {
	err := c.do(ctx, http.MethodDelete, "/channels/"+channelID+"/permissions/"+userID, nil, nil)
	if errors.Is(err, errNotFound) {
		return nil
	}
	return err
}

// Post manda uma mensagem que só menciona os IDs dados (D-06): nick, título e motivo, que
// vêm do usuário, não conseguem marcar @everyone nem cargos.
func (c *Client) Post(ctx context.Context, channelID, content string, mentions []string) error {
	if mentions == nil {
		mentions = []string{}
	}
	body := map[string]any{
		"content":          content,
		"allowed_mentions": map[string]any{"parse": []string{}, "users": mentions},
	}
	return c.do(ctx, http.MethodPost, "/channels/"+channelID+"/messages", body, nil)
}

// DeleteChannel apaga o canal; se ele já não existe, é sucesso (CA-05.5).
func (c *Client) DeleteChannel(ctx context.Context, channelID string) error {
	err := c.do(ctx, http.MethodDelete, "/channels/"+channelID, nil, nil)
	if errors.Is(err, errNotFound) {
		return nil
	}
	return err
}

// AddMember inclui o jogador no servidor oficial com o token dele, que tem o escopo
// guilds.join (RN-02). Já estar no servidor também é sucesso (RN-04). O token não é
// guardado nem vai para o log.
func (c *Client) AddMember(ctx context.Context, userID, accessToken string) error {
	err := c.do(ctx, http.MethodPut, "/guilds/"+c.guildID+"/members/"+userID,
		map[string]string{"access_token": accessToken}, nil)
	if errors.Is(err, errForbidden) {
		return ErrInvalidToken
	}
	return err
}

var (
	errNotFound  = errors.New("discordbot: não encontrado")
	errForbidden = errors.New("discordbot: recusado")
)

type apiError struct {
	Code       int     `json:"code"`
	RetryAfter float64 `json:"retry_after"`
}

// do faz a chamada com o token do bot. No 429, espera retry_after uma vez se couber no
// prazo; senão devolve ErrUnavailable (RNF-02).
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	for attempt := 0; ; attempt++ {
		status, payload, err := c.send(ctx, method, path, body)
		if err != nil {
			return fmt.Errorf("%w: %s %s: %w", ErrUnavailable, method, path, err)
		}
		var apiErr apiError
		if status >= 300 {
			_ = json.Unmarshal(payload, &apiErr)
		}
		switch {
		case status == http.StatusTooManyRequests:
			wait := time.Duration(apiErr.RetryAfter * float64(time.Second))
			deadline, _ := ctx.Deadline()
			if attempt > 0 || time.Until(deadline) <= wait {
				return fmt.Errorf("%w: limite de chamadas em %s %s", ErrUnavailable, method, path)
			}
			select {
			case <-time.After(wait):
				continue
			case <-ctx.Done():
				return fmt.Errorf("%w: %w", ErrUnavailable, ctx.Err())
			}
		case status < 300:
			if out != nil && len(payload) > 0 {
				if err := json.Unmarshal(payload, out); err != nil {
					return fmt.Errorf("%w: resposta inválida de %s %s", ErrUnavailable, method, path)
				}
			}
			return nil
		case apiErr.Code == codeMaxGuildChannels:
			return ErrGuildFull
		case status == http.StatusNotFound || apiErr.Code == codeUnknownChannel:
			return errNotFound
		case status == http.StatusForbidden:
			return errForbidden
		case status == http.StatusUnauthorized:
			c.log.Error("o Discord recusou o token do bot; confira DISCORD_BOT_TOKEN", "status", status)
			return fmt.Errorf("%w: token do bot recusado", ErrUnavailable)
		default:
			return fmt.Errorf("%w: %s %s respondeu %d", ErrUnavailable, method, path, status)
		}
	}
}

func (c *Client) send(ctx context.Context, method, path string, body any) (int, []byte, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", "Bot "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, payload, err
}
