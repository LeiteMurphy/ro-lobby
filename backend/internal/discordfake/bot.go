package discordfake

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Rotas do bot (spec discord-bot, T-02, RNF-03): incluir no servidor, criar e apagar
// canais, mexer nas permissões e postar mensagens. O Discord falso guarda o que o bot fez,
// para os testes conferirem, e tem chaves para simular falhas.

// Overwrite é uma permissão de canal, como o Discord manda.
type Overwrite struct {
	ID    string `json:"id"`
	Type  int    `json:"type"` // 0 = cargo, 1 = membro
	Allow string `json:"allow"`
	Deny  string `json:"deny"`
}

// Channel é um canal criado pelo bot.
type Channel struct {
	ID         string               `json:"id"`
	Type       int                  `json:"type"` // 0 = texto, 2 = voz, 4 = categoria
	Name       string               `json:"name"`
	ParentID   string               `json:"parent_id,omitempty"`
	Overwrites map[string]Overwrite `json:"-"`
}

// Message é uma mensagem postada pelo bot.
type Message struct {
	ChannelID       string
	Content         string
	AllowedMentions AllowedMentions
}

// AllowedMentions é o controle de menções que o bot manda (D-06).
type AllowedMentions struct {
	Parse []string `json:"parse"`
	Users []string `json:"users"`
}

// BotState é o servidor do lado do Discord falso.
type BotState struct {
	members  map[string]bool
	channels map[string]*Channel
	messages []Message
	nextID   int64
}

// BotFailures simula o Discord fora do ar ou no limite (RN-20, RN-21, RNF-02).
type BotFailures struct {
	// Down faz toda rota do bot responder 500.
	Down bool
	// GuildFull faz a criação de canal responder o código 30013.
	GuildFull bool
	// RateLimitNext faz as próximas N chamadas responderem 429 com RetryAfter.
	RateLimitNext int
	RetryAfter    time.Duration
	// FailCreateAfter, se > 0, faz a criação de canal falhar depois desse número de canais
	// criados com sucesso (para a compensação de CA-02.7).
	FailCreateAfter int
}

func (f *Fake) botState() *BotState {
	if f.bot == nil {
		f.bot = &BotState{members: map[string]bool{}, channels: map[string]*Channel{}, nextID: 900000000000000000}
	}
	return f.bot
}

// SetBotFailures troca as falhas simuladas.
func (f *Fake) SetBotFailures(b BotFailures) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failures = b
}

// Members devolve os IDs incluídos no servidor.
func (f *Fake) Members() map[string]bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := map[string]bool{}
	for id := range f.botState().members {
		out[id] = true
	}
	return out
}

// AddMemberDirect põe alguém no servidor sem OAuth, para preparar testes.
func (f *Fake) AddMemberDirect(userID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.botState().members[userID] = true
}

// Channel devolve uma cópia do canal, ou false se ele não existe.
func (f *Fake) Channel(id string) (Channel, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.botState().channels[id]
	if !ok {
		return Channel{}, false
	}
	cp := *c
	cp.Overwrites = map[string]Overwrite{}
	for k, v := range c.Overwrites {
		cp.Overwrites[k] = v
	}
	return cp, true
}

// ChannelCount é quantos canais existem no servidor.
func (f *Fake) ChannelCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.botState().channels)
}

// DeleteChannelDirect apaga um canal como se alguém tivesse apagado à mão (CA-05.5).
func (f *Fake) DeleteChannelDirect(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.botState().channels, id)
}

// Messages devolve as mensagens postadas no canal, na ordem.
func (f *Fake) Messages(channelID string) []Message {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []Message
	for _, m := range f.botState().messages {
		if m.ChannelID == channelID {
			out = append(out, m)
		}
	}
	return out
}

// BotCalls é quantas chamadas o bot fez.
func (f *Fake) BotCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.botCalls
}

func (f *Fake) registerBot(mux *http.ServeMux) {
	mux.HandleFunc("PUT /api/guilds/{guild}/members/{user}", f.botRoute(f.addMember))
	mux.HandleFunc("POST /api/guilds/{guild}/channels", f.botRoute(f.createChannel))
	mux.HandleFunc("PUT /api/channels/{channel}/permissions/{target}", f.botRoute(f.putOverwrite))
	mux.HandleFunc("DELETE /api/channels/{channel}/permissions/{target}", f.botRoute(f.deleteOverwrite))
	mux.HandleFunc("POST /api/channels/{channel}/messages", f.botRoute(f.postMessage))
	mux.HandleFunc("DELETE /api/channels/{channel}", f.botRoute(f.deleteChannel))
}

// botRoute confere o token do bot, conta a chamada e aplica as falhas simuladas.
func (f *Fake) botRoute(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if f.BotToken == "" || r.Header.Get("Authorization") != "Bot "+f.BotToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"message": "401: Unauthorized", "code": 0})
			return
		}
		f.mu.Lock()
		f.botCalls++
		fail := f.failures
		if f.failures.RateLimitNext > 0 {
			f.failures.RateLimitNext--
		}
		f.mu.Unlock()
		switch {
		case fail.Down:
			writeJSON(w, http.StatusInternalServerError, map[string]any{"message": "500: Internal Server Error", "code": 0})
			return
		case fail.RateLimitNext > 0:
			writeJSON(w, http.StatusTooManyRequests, map[string]any{
				"message": "You are being rate limited.", "retry_after": fail.RetryAfter.Seconds(), "global": false,
			})
			return
		}
		next(w, r)
	}
}

func (f *Fake) addMember(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccessToken string `json:"access_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	userID := r.PathValue("user")
	f.mu.Lock()
	defer f.mu.Unlock()
	user, ok := f.tokens[body.AccessToken]
	if !ok || user.ID != userID {
		writeJSON(w, http.StatusForbidden, map[string]any{"message": "Invalid OAuth2 access token", "code": 50025})
		return
	}
	st := f.botState()
	if st.members[userID] {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	st.members[userID] = true
	writeJSON(w, http.StatusCreated, map[string]any{"user": map[string]string{"id": userID}})
}

func (f *Fake) createChannel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name       string      `json:"name"`
		Type       int         `json:"type"`
		ParentID   string      `json:"parent_id"`
		Overwrites []Overwrite `json:"permission_overwrites"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": "Invalid Form Body", "code": 50035})
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	st := f.botState()
	if f.failures.GuildFull {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": "Maximum number of guild channels reached (500)", "code": 30013})
		return
	}
	if f.failures.FailCreateAfter > 0 && len(st.channels) >= f.failures.FailCreateAfter {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": "500: Internal Server Error", "code": 0})
		return
	}
	if body.ParentID != "" {
		if p, ok := st.channels[body.ParentID]; !ok || p.Type != 4 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"message": "Invalid Form Body", "code": 50035})
			return
		}
	}
	st.nextID++
	c := &Channel{ID: strconv.FormatInt(st.nextID, 10), Type: body.Type, Name: body.Name, ParentID: body.ParentID, Overwrites: map[string]Overwrite{}}
	for _, o := range body.Overwrites {
		c.Overwrites[o.ID] = o
	}
	st.channels[c.ID] = c
	writeJSON(w, http.StatusCreated, c)
}

func (f *Fake) channelOr404(w http.ResponseWriter, id string) (*Channel, bool) {
	c, ok := f.botState().channels[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"message": "Unknown Channel", "code": 10003})
	}
	return c, ok
}

func (f *Fake) putOverwrite(w http.ResponseWriter, r *http.Request) {
	var o Overwrite
	_ = json.NewDecoder(r.Body).Decode(&o)
	o.ID = r.PathValue("target")
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.channelOr404(w, r.PathValue("channel"))
	if !ok {
		return
	}
	c.Overwrites[o.ID] = o
	w.WriteHeader(http.StatusNoContent)
}

func (f *Fake) deleteOverwrite(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.channelOr404(w, r.PathValue("channel"))
	if !ok {
		return
	}
	delete(c.Overwrites, r.PathValue("target"))
	w.WriteHeader(http.StatusNoContent)
}

func (f *Fake) postMessage(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content         string          `json:"content"`
		AllowedMentions AllowedMentions `json:"allowed_mentions"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.channelOr404(w, r.PathValue("channel")); !ok {
		return
	}
	st := f.botState()
	st.messages = append(st.messages, Message{ChannelID: r.PathValue("channel"), Content: body.Content, AllowedMentions: body.AllowedMentions})
	st.nextID++
	writeJSON(w, http.StatusOK, map[string]string{"id": strconv.FormatInt(st.nextID, 10)})
}

func (f *Fake) deleteChannel(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	st := f.botState()
	id := r.PathValue("channel")
	c, ok := f.channelOr404(w, id)
	if !ok {
		return
	}
	delete(st.channels, id)
	// Apagar uma categoria deixa os canais dela soltos, como no Discord.
	for _, child := range st.channels {
		if child.ParentID == id {
			child.ParentID = ""
		}
	}
	writeJSON(w, http.StatusOK, c)
}

// HasMention diz se o texto menciona o usuário no formato do Discord.
func HasMention(content, userID string) bool {
	return strings.Contains(content, "<@"+userID+">")
}
