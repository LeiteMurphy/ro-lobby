//go:build integration

package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/applications"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/auth"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/characters"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordbot"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordfake"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordroom"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/migrate"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

// Sala do grupo no Discord pelas rotas, com o PostgreSQL real e o Discord falso (spec
// discord-bot, T-04).

const roomGuild = "1560117933705797643"

type roomFlow struct {
	flow
	fake  *discordfake.Fake
	rooms *discordroom.Service
}

func newRoomFlow(t *testing.T, enabled bool) roomFlow {
	t.Helper()
	url := testdb.New(t)
	sqlDB, err := migrate.OpenDB(url)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := migrate.NewProvider(sqlDB)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(t.Context()); err != nil {
		t.Fatal(err)
	}
	_ = sqlDB.Close()
	pool, err := database.Open(t.Context(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	fake := discordfake.New("id", "segredo", discordfake.User{ID: "1", Username: "u"})
	fake.BotToken = "token-do-bot"
	srv := httptest.NewServer(fake.Handler())
	t.Cleanup(srv.Close)
	var bot discordroom.Bot
	if enabled {
		bot = discordbot.New(srv.URL+"/api", "token-do-bot", roomGuild, "111", discordbot.WithTimeout(2*time.Second))
	}
	rooms := discordroom.NewService(pool, bot)
	rooms.Origin = "https://rolobby.com.br"
	h := New(pool, auth.NewService(pool, nil), characters.NewService(pool), lobbies.NewService(pool),
		applications.NewService(pool), talents.NewService(pool), WithRooms(rooms))
	return roomFlow{flow: flow{h: h, q: db.New(pool), pool: pool}, fake: fake, rooms: rooms}
}

// connect marca o Usuário com esse ID do Discord como conectado ao servidor (RN-01).
func (f roomFlow) connect(t *testing.T, discordID string) {
	t.Helper()
	if _, err := f.pool.Exec(t.Context(), `UPDATE users SET discord_connected_at = now() WHERE discord_id = $1`, discordID); err != nil {
		t.Fatal(err)
	}
}

func (f roomFlow) tick(t *testing.T) {
	t.Helper()
	if err := f.rooms.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
}

// member candidata e aceita o personagem.
func (f roomFlow) member(t *testing.T, path, owner, token, characterID string) {
	t.Helper()
	rec, app := call(t, f.h, http.MethodPost, path+"/applications", token, fmt.Sprintf(`{"characterId":%q}`, characterID))
	if rec.Code != http.StatusCreated {
		t.Fatalf("candidatar: %d %v", rec.Code, app)
	}
	if rec, got := call(t, f.h, http.MethodPost, "/applications/"+app["id"].(string)+"/accept", owner, ""); rec.Code != http.StatusOK {
		t.Fatalf("aceitar: %d %v", rec.Code, got)
	}
}

// group monta o lobby de amanhã às 20:00 com Ana (Suporte, dona), Bia (Tank) e Caio (Dano).
func (f roomFlow) group(t *testing.T) (path, ana, bia, caio string) {
	t.Helper()
	ana, bia, caio = f.session(t, "301", "Ana"), f.session(t, "302", "Bia"), f.session(t, "303", "Caio")
	host := f.character(t, ana, `{"nick":"Lirien","classId":"arcebispo","level":200,"role":"support","server":"freya"}`)
	brasa := f.character(t, bia, `{"nick":"Brasa","classId":"guardiao-real","level":200,"role":"tank","server":"freya"}`)
	fogo := f.character(t, caio, `{"nick":"Fogo","classId":"arquimago","level":200,"role":"dps","server":"freya"}`)
	path, _ = f.lobby(t, ana, host)
	f.member(t, path, ana, bia, brasa)
	f.member(t, path, ana, caio, fogo)
	return path, ana, bia, caio
}

func (f roomFlow) roomRow(t *testing.T) (status, category, text, voice string) {
	t.Helper()
	if err := f.pool.QueryRow(t.Context(), `SELECT status, category_id, text_channel_id, voice_channel_id FROM discord_rooms`).
		Scan(&status, &category, &text, &voice); err != nil {
		t.Fatal(err)
	}
	return
}

// CA-02.1 / CA-02.3 / RN-05 / RN-07: o dono abre a sala; só os conectados e o bot veem.
func TestDiscordRoom_CA02_1_CA02_3_OwnerOpens(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	f.connect(t, "301")
	f.connect(t, "302") // Caio (303) não está conectado

	rec, got := call(t, f.h, http.MethodPost, path+"/discord-room", ana, "")
	if rec.Code != http.StatusCreated || got["status"] != "open" || !strings.HasPrefix(got["url"].(string), "https://discord.com/channels/"+roomGuild+"/") {
		t.Fatalf("abrir: %d %v", rec.Code, got)
	}
	status, category, text, voice := f.roomRow(t)
	if status != "open" {
		t.Errorf("status = %s", status)
	}
	for _, id := range []string{category, text, voice} {
		ch, ok := f.fake.Channel(id)
		if !ok {
			t.Fatalf("canal %s não existe", id)
		}
		if _, ok := ch.Overwrites[roomGuild]; !ok {
			t.Errorf("%s: sem a regra de @everyone", ch.Name)
		}
		for _, d := range []string{"301", "302"} {
			if _, ok := ch.Overwrites[d]; !ok {
				t.Errorf("%s: %s sem acesso", ch.Name, d)
			}
		}
		if _, ok := ch.Overwrites["303"]; ok {
			t.Errorf("%s: não conectado com acesso", ch.Name)
		}
	}
	cat, _ := f.fake.Channel(category)
	if !strings.HasPrefix(cat.Name, "Templo do Demônio Rei · ") {
		t.Errorf("nome da categoria = %q", cat.Name)
	}
}

// CA-02.2 / RN-08: a mensagem de abertura traz título, servidor, horário, composição,
// menções só dos conectados, quem está fora e o link.
func TestDiscordRoom_CA02_2_OpeningMessage(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	f.connect(t, "301")
	f.connect(t, "302")
	if rec, got := call(t, f.h, http.MethodPost, path+"/discord-room", ana, ""); rec.Code != http.StatusCreated {
		t.Fatalf("abrir: %d %v", rec.Code, got)
	}
	f.tick(t)
	_, _, text, _ := f.roomRow(t)
	msgs := f.fake.Messages(text)
	if len(msgs) != 1 {
		t.Fatalf("mensagens = %d", len(msgs))
	}
	m := msgs[0]
	for _, want := range []string{
		"Sala aberta: Templo do Demônio Rei", "Servidor Freya", "20:00 (horário de Brasília)",
		"**Tank:** Brasa (Guardião Real) <@302>", "**Suporte:** Lirien (Arcebispo) <@301>",
		"**Dano:** Fogo (Arquimago)", "Ainda fora da sala: Fogo", "Lobby: https://rolobby.com.br" + path,
	} {
		if !strings.Contains(m.Content, want) {
			t.Errorf("mensagem sem %q:\n%s", want, m.Content)
		}
	}
	if discordfake.HasMention(m.Content, "303") {
		t.Error("menciona quem não está conectado")
	}
	if strings.Join(m.AllowedMentions.Users, ",") != "302,301" {
		t.Errorf("menções = %v", m.AllowedMentions.Users)
	}
}

// CA-02.4 / RN-06: só o dono abre.
func TestDiscordRoom_CA02_4_OnlyOwner(t *testing.T) {
	f := newRoomFlow(t, true)
	path, _, bia, _ := f.group(t)
	rec, got := call(t, f.h, http.MethodPost, path+"/discord-room", bia, "")
	if rec.Code != http.StatusConflict || got["code"] != "not_owner" {
		t.Errorf("membro: %d %v", rec.Code, got)
	}
	if f.fake.ChannelCount() != 0 {
		t.Error("criou canais")
	}
}

// CA-02.5 / RN-06: iniciado ou cancelado não abre.
func TestDiscordRoom_CA02_5_NotOpen(t *testing.T) {
	for name, sql := range map[string]string{
		"iniciado":  `UPDATE lobbies SET starts_at = now() - interval '1 hour'`,
		"cancelado": `UPDATE lobbies SET cancelled_at = now(), cancel_reason = 'Faltou o tank, fica para amanhã'`,
	} {
		t.Run(name, func(t *testing.T) {
			f := newRoomFlow(t, true)
			path, ana, _, _ := f.group(t)
			if _, err := f.pool.Exec(t.Context(), sql); err != nil {
				t.Fatal(err)
			}
			rec, got := call(t, f.h, http.MethodPost, path+"/discord-room", ana, "")
			if rec.Code != http.StatusConflict || got["code"] != "not_open" || f.fake.ChannelCount() != 0 {
				t.Errorf("%d %v, %d canais", rec.Code, got, f.fake.ChannelCount())
			}
		})
	}
}

// CA-02.6 / RN-09: com sala aberta, encerrando ou apagada, não abre outra.
func TestDiscordRoom_CA02_6_OneRoomPerLobby(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	if rec, _ := call(t, f.h, http.MethodPost, path+"/discord-room", ana, ""); rec.Code != http.StatusCreated {
		t.Fatal("primeira abertura falhou")
	}
	for _, sql := range []string{``, `UPDATE discord_rooms SET status = 'closing', closing_at = now()`,
		`UPDATE discord_rooms SET status = 'deleted', deleted_at = now()`} {
		if sql != "" {
			if _, err := f.pool.Exec(t.Context(), sql); err != nil {
				t.Fatal(err)
			}
		}
		rec, got := call(t, f.h, http.MethodPost, path+"/discord-room", ana, "")
		if rec.Code != http.StatusConflict || got["code"] != "room_exists" {
			t.Errorf("%q: %d %v", sql, rec.Code, got)
		}
	}
	if n := f.fake.ChannelCount(); n != 3 {
		t.Errorf("%d canais", n)
	}
}

// Caso de borda (RN-05): dois cliques ao mesmo tempo criam uma sala só.
func TestDiscordRoom_RN05_ConcurrentOpen(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	codes := make([]int, 2)
	var wg sync.WaitGroup
	for i := range codes {
		wg.Go(func() {
			rec, _ := call(t, f.h, http.MethodPost, path+"/discord-room", ana, "")
			codes[i] = rec.Code
		})
	}
	wg.Wait()
	created := 0
	for _, c := range codes {
		if c == http.StatusCreated {
			created++
		}
	}
	if created != 1 || f.fake.ChannelCount() != 3 {
		t.Errorf("respostas %v, %d canais", codes, f.fake.ChannelCount())
	}
}

// CA-02.7 / CA-02.8 / RN-20: Discord fora do ar ou cheio responde 503, sem sala nem canal.
func TestDiscordRoom_CA02_7_CA02_8_DiscordFails(t *testing.T) {
	for want, failure := range map[string]discordfake.BotFailures{
		"discord_unavailable": {FailCreateAfter: 1},
		"guild_full":          {GuildFull: true},
	} {
		t.Run(want, func(t *testing.T) {
			f := newRoomFlow(t, true)
			path, ana, _, _ := f.group(t)
			f.fake.SetBotFailures(failure)
			rec, got := call(t, f.h, http.MethodPost, path+"/discord-room", ana, "")
			if rec.Code != http.StatusServiceUnavailable || got["error"] != want {
				t.Errorf("%d %v", rec.Code, got)
			}
			var n int
			if err := f.pool.QueryRow(t.Context(), `SELECT count(*) FROM discord_rooms`).Scan(&n); err != nil || n != 0 {
				t.Errorf("salas gravadas: %d, %v", n, err)
			}
			if f.fake.ChannelCount() != 0 {
				t.Errorf("%d canais sobraram", f.fake.ChannelCount())
			}
		})
	}
}

// CA-03.1 / CA-03.2 / CA-03.4 / RN-18 / RN-19: o membro vê o link e quem está na sala;
// quem não é membro não recebe o bloco discord.
func TestDiscordRoom_CA03_1_CA03_4_DetailBlock(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, bia, caio := f.group(t)
	f.connect(t, "301")
	f.connect(t, "302")
	if rec, _ := call(t, f.h, http.MethodPost, path+"/discord-room", ana, ""); rec.Code != http.StatusCreated {
		t.Fatal("abrir falhou")
	}
	_, got := call(t, f.h, http.MethodGet, path, caio, "")
	d, ok := got["discord"].(map[string]any)
	if !ok {
		t.Fatalf("membro sem o bloco discord: %v", got)
	}
	room := d["room"].(map[string]any)
	if d["enabled"] != true || room["status"] != "open" || !strings.Contains(room["url"].(string), roomGuild) || d["viewerConnected"] != false {
		t.Errorf("bloco = %v", d)
	}
	connected := map[string]bool{}
	for _, m := range d["members"].([]any) {
		mm := m.(map[string]any)
		connected[mm["nick"].(string)] = mm["connected"].(bool)
	}
	if !connected["Lirien"] || !connected["Brasa"] || connected["Fogo"] {
		t.Errorf("membros = %v", connected)
	}
	_ = bia

	duda := f.session(t, "304", "Duda")
	pend := f.character(t, duda, `{"nick":"Cura","classId":"arcebispo","level":200,"role":"support","server":"freya"}`)
	if rec, _ := call(t, f.h, http.MethodPost, path+"/applications", duda, fmt.Sprintf(`{"characterId":%q}`, pend)); rec.Code != http.StatusCreated {
		t.Fatal("candidatar falhou")
	}
	for name, token := range map[string]string{"visitante": "", "pendente": duda, "outro": f.session(t, "305", "Edu")} {
		rec, got := call(t, f.h, http.MethodGet, path, token, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: %d", name, rec.Code)
		}
		if _, ok := got["discord"]; ok {
			t.Errorf("%s recebeu o bloco discord", name)
		}
		if raw := rec.Body.String(); strings.Contains(raw, "discord.com/channels") {
			t.Errorf("%s recebeu o link do canal", name)
		}
	}
}

// CA-03.3 / CA-03.5 / RN-18: sem sala, nulo; encerrando, sem link.
func TestDiscordRoom_CA03_3_CA03_5_NoRoomAndClosed(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	_, got := call(t, f.h, http.MethodGet, path, ana, "")
	if d := got["discord"].(map[string]any); d["room"] != nil || d["enabled"] != true {
		t.Errorf("sem sala: %v", d)
	}
	call(t, f.h, http.MethodPost, path+"/discord-room", ana, "")
	if _, err := f.pool.Exec(t.Context(), `UPDATE discord_rooms SET status = 'closing', closing_at = now()`); err != nil {
		t.Fatal(err)
	}
	_, got = call(t, f.h, http.MethodGet, path, ana, "")
	room := got["discord"].(map[string]any)["room"].(map[string]any)
	if room["status"] != "closing" || room["url"] != nil {
		t.Errorf("encerrando: %v", room)
	}
}

// CA-01.7 / RN-22: com a sala desligada, abrir responde discord_disabled e o bloco vem sem
// sala.
func TestDiscordRoom_RN22_Disabled(t *testing.T) {
	f := newRoomFlow(t, false)
	path, ana, _, _ := f.group(t)
	rec, got := call(t, f.h, http.MethodPost, path+"/discord-room", ana, "")
	if rec.Code != http.StatusNotFound || got["error"] != "discord_disabled" {
		t.Errorf("abrir: %d %v", rec.Code, got)
	}
	_, got = call(t, f.h, http.MethodGet, path, ana, "")
	if d := got["discord"].(map[string]any); d["enabled"] != false || d["room"] != nil {
		t.Errorf("bloco = %v", d)
	}
}
