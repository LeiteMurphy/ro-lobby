//go:build integration

package server

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordfake"
)

// A sala acompanha o grupo, pelas rotas (spec discord-bot, T-06).

// openRoom abre a sala do lobby com a dona e roda o trabalhador; devolve os canais.
func (f roomFlow) openRoom(t *testing.T, path, owner string) (category, text, voice string) {
	t.Helper()
	if rec, got := call(t, f.h, http.MethodPost, path+"/discord-room", owner, ""); rec.Code != http.StatusCreated {
		t.Fatalf("abrir: %d %v", rec.Code, got)
	}
	f.tick(t)
	_, category, text, voice = f.roomRow(t)
	return
}

func (f roomFlow) hasAccess(t *testing.T, channels []string, discordID string) bool {
	t.Helper()
	in := 0
	for _, id := range channels {
		ch, _ := f.fake.Channel(id)
		if _, ok := ch.Overwrites[discordID]; ok {
			in++
		}
	}
	if in != 0 && in != len(channels) {
		t.Fatalf("acesso parcial de %s: %d de %d canais", discordID, in, len(channels))
	}
	return in == len(channels)
}

func lastMessage(t *testing.T, f roomFlow, text string) discordfake.Message {
	t.Helper()
	msgs := f.fake.Messages(text)
	if len(msgs) == 0 {
		t.Fatal("sem mensagens")
	}
	return msgs[len(msgs)-1]
}

// apply candidata o personagem e devolve o caminho da candidatura.
func (f roomFlow) apply(t *testing.T, path, token, characterID string) string {
	t.Helper()
	rec, app := call(t, f.h, http.MethodPost, path+"/applications", token, fmt.Sprintf(`{"characterId":%q}`, characterID))
	if rec.Code != http.StatusCreated {
		t.Fatalf("candidatar: %d %v", rec.Code, app)
	}
	return "/applications/" + app["id"].(string)
}

// CA-04.1 / CA-04.2 / RN-10: o aceito entra na sala se estiver conectado; a sala é avisada,
// com menção só para quem está conectado.
func TestDiscordSync_CA04_1_CA04_2_Accept(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	f.connect(t, "301")
	cat, text, voice := f.openRoom(t, path, ana)
	channels := []string{cat, text, voice}

	duda := f.session(t, "304", "Duda")
	edu := f.session(t, "305", "Edu")
	f.connect(t, "304")
	cura := f.character(t, duda, `{"nick":"Cura","classId":"arcebispo","level":200,"role":"support","server":"freya"}`)
	lamina := f.character(t, edu, `{"nick":"Lamina","classId":"sicario","level":200,"role":"dps","server":"freya"}`)

	if rec, _ := call(t, f.h, http.MethodPost, f.apply(t, path, duda, cura)+"/accept", ana, ""); rec.Code != http.StatusOK {
		t.Fatal("aceitar Duda")
	}
	f.tick(t)
	if !f.hasAccess(t, channels, "304") {
		t.Error("CA-04.1: Duda sem acesso")
	}
	m := lastMessage(t, f, text)
	if !strings.Contains(m.Content, "**Cura** entrou no grupo") || !discordfake.HasMention(m.Content, "304") {
		t.Errorf("CA-04.1: mensagem = %q", m.Content)
	}

	if rec, _ := call(t, f.h, http.MethodPost, f.apply(t, path, edu, lamina)+"/accept", ana, ""); rec.Code != http.StatusOK {
		t.Fatal("aceitar Edu")
	}
	f.tick(t)
	if f.hasAccess(t, channels, "305") {
		t.Error("CA-04.2: Edu, não conectado, com acesso")
	}
	m = lastMessage(t, f, text)
	if !strings.Contains(m.Content, "**Lamina** entrou no grupo") || discordfake.HasMention(m.Content, "305") || len(m.AllowedMentions.Users) != 0 {
		t.Errorf("CA-04.2: mensagem = %q", m.Content)
	}
}

// CA-04.3 / RN-11: quem sai ou é removido perde o acesso, sem o motivo no Discord.
func TestDiscordSync_CA04_3_LeaveAndRemove(t *testing.T) {
	f := newRoomFlow(t, true)
	ana, bia, caio := f.session(t, "301", "Ana"), f.session(t, "302", "Bia"), f.session(t, "303", "Caio")
	for _, d := range []string{"301", "302", "303"} {
		f.connect(t, d)
	}
	host := f.character(t, ana, `{"nick":"Lirien","classId":"arcebispo","level":200,"role":"support","server":"freya"}`)
	brasa := f.character(t, bia, `{"nick":"Brasa","classId":"guardiao-real","level":200,"role":"tank","server":"freya"}`)
	fogo := f.character(t, caio, `{"nick":"Fogo","classId":"arquimago","level":200,"role":"dps","server":"freya"}`)
	path, _ := f.lobby(t, ana, host)
	biaApp, caioApp := f.apply(t, path, bia, brasa), f.apply(t, path, caio, fogo)
	call(t, f.h, http.MethodPost, biaApp+"/accept", ana, "")
	call(t, f.h, http.MethodPost, caioApp+"/accept", ana, "")
	cat, text, voice := f.openRoom(t, path, ana)
	channels := []string{cat, text, voice}
	if !f.hasAccess(t, channels, "302") || !f.hasAccess(t, channels, "303") {
		t.Fatal("membros sem acesso na abertura")
	}

	if rec, got := call(t, f.h, http.MethodPost, biaApp+"/leave", bia, ""); rec.Code != http.StatusOK {
		t.Fatalf("sair: %d %v", rec.Code, got)
	}
	f.tick(t)
	if f.hasAccess(t, channels, "302") || !strings.Contains(lastMessage(t, f, text).Content, "**Brasa** saiu do grupo") {
		t.Errorf("saída: acesso %v, mensagem %q", f.hasAccess(t, channels, "302"), lastMessage(t, f, text).Content)
	}

	if rec, got := call(t, f.h, http.MethodPost, caioApp+"/remove", ana, `{"reason":"Mudamos o horário da run"}`); rec.Code != http.StatusOK {
		t.Fatalf("remover: %d %v", rec.Code, got)
	}
	f.tick(t)
	m := lastMessage(t, f, text)
	if f.hasAccess(t, channels, "303") || !strings.Contains(m.Content, "**Fogo** saiu do grupo") || strings.Contains(m.Content, "horário da run") {
		t.Errorf("remoção: mensagem %q", m.Content)
	}
}

// CA-04.4 / RN-12: a troca de personagem não muda o acesso nem gera mensagem.
func TestDiscordSync_CA04_4_SwapDoesNothing(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	f.connect(t, "301")
	other := f.character(t, ana, `{"nick":"Outra","classId":"arcebispo","level":200,"role":"support","server":"freya"}`)
	_, text, _ := f.openRoom(t, path, ana)
	before := len(f.fake.Messages(text))
	if rec, got := call(t, f.h, http.MethodPut, path+"/owner-character", ana, fmt.Sprintf(`{"characterId":%q}`, other)); rec.Code != http.StatusOK {
		t.Fatalf("troca: %d %v", rec.Code, got)
	}
	f.tick(t)
	var jobs int
	if err := f.pool.QueryRow(t.Context(), `SELECT count(*) FROM discord_jobs WHERE created_at > now() - interval '1 minute' AND kind = 'sync_access'`).Scan(&jobs); err != nil {
		t.Fatal(err)
	}
	if n := len(f.fake.Messages(text)); n != before || jobs != 0 {
		t.Errorf("mensagens novas %d, jobs de acesso %d", n-before, jobs)
	}
}

// CA-04.5 / RN-13: quem se conecta depois ganha as salas abertas dos seus lobbies e é
// mencionado; a sala apagada não é tocada.
func TestDiscordSync_CA04_5_LateConnection(t *testing.T) {
	f := newRoomFlow(t, true)
	ana, bia := f.session(t, "301", "Ana"), f.session(t, "302", "Bia")
	f.connect(t, "301")
	paths := make([]string, 3)
	for i := range paths {
		host := f.character(t, ana, fmt.Sprintf(`{"nick":"Anf%d","classId":"arcebispo","level":200,"role":"support","server":"freya"}`, i))
		brasa := f.character(t, bia, fmt.Sprintf(`{"nick":"Bra%d","classId":"guardiao-real","level":200,"role":"tank","server":"freya"}`, i))
		tomorrow := time.Now().In(time.UTC).AddDate(0, 0, 1+i)
		start := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 23, 0, 0, 0, time.UTC).Format(time.RFC3339)
		rec, lobby := call(t, f.h, http.MethodPost, "/lobbies", ana,
			fmt.Sprintf(`{"instanceId":"templo-do-demonio-rei","startsAt":%q,"slots":{"tank":1,"support":2,"dps":3},"minLevel":160,"characterId":%q}`, start, host))
		if rec.Code != http.StatusCreated {
			t.Fatalf("lobby %d: %d %v", i, rec.Code, lobby)
		}
		paths[i] = "/lobbies/" + lobby["id"].(string)
		call(t, f.h, http.MethodPost, f.apply(t, paths[i], bia, brasa)+"/accept", ana, "")
		if rec, _ := call(t, f.h, http.MethodPost, paths[i]+"/discord-room", ana, ""); rec.Code != http.StatusCreated {
			t.Fatalf("abrir sala %d", i)
		}
	}
	f.tick(t)
	texts := map[string][]string{}
	rows, err := f.pool.Query(t.Context(), `SELECT lobby_id::text, category_id, text_channel_id, voice_channel_id FROM discord_rooms`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var id, c, x, v string
		_ = rows.Scan(&id, &c, &x, &v)
		texts[id] = []string{c, x, v}
	}
	rows.Close()
	deleted := strings.TrimPrefix(paths[2], "/lobbies/")
	if _, err := f.pool.Exec(t.Context(), `UPDATE discord_rooms SET status = 'deleted', deleted_at = now() WHERE lobby_id = $1`, deleted); err != nil {
		t.Fatal(err)
	}
	deletedBefore := len(f.fake.Messages(texts[deleted][1]))

	if rec, got := f.connectWith(t, bia, "302"); rec.Code != http.StatusOK {
		t.Fatalf("conectar: %d %v", rec.Code, got)
	}
	f.tick(t)
	for _, p := range paths[:2] {
		ch := texts[strings.TrimPrefix(p, "/lobbies/")]
		m := lastMessage(t, f, ch[1])
		if !f.hasAccess(t, ch, "302") || !strings.Contains(m.Content, "entrou na sala") || !discordfake.HasMention(m.Content, "302") {
			t.Errorf("%s: acesso %v, mensagem %q", p, f.hasAccess(t, ch, "302"), m.Content)
		}
	}
	if n := len(f.fake.Messages(texts[deleted][1])); n != deletedBefore || f.hasAccess(t, texts[deleted], "302") {
		t.Error("a sala apagada foi tocada")
	}
}

// CA-04.6 / RN-21: com o Discord fora, o aceite funciona; o acesso vem quando ele volta.
func TestDiscordSync_CA04_6_DiscordDown(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	cat, text, voice := f.openRoom(t, path, ana)
	duda := f.session(t, "304", "Duda")
	f.connect(t, "304")
	cura := f.character(t, duda, `{"nick":"Cura","classId":"arcebispo","level":200,"role":"support","server":"freya"}`)

	f.fake.SetBotFailures(discordfake.BotFailures{Down: true})
	if rec, got := call(t, f.h, http.MethodPost, f.apply(t, path, duda, cura)+"/accept", ana, ""); rec.Code != http.StatusOK || got["status"] != "accepted" {
		t.Fatalf("aceite com o Discord fora: %d %v", rec.Code, got)
	}
	f.tick(t)
	f.fake.SetBotFailures(discordfake.BotFailures{})
	later := time.Now().Add(time.Hour)
	f.rooms.Now = func() time.Time { return later }
	f.tick(t)
	if !f.hasAccess(t, []string{cat, text, voice}, "304") || !strings.Contains(lastMessage(t, f, text).Content, "**Cura** entrou no grupo") {
		t.Error("o efeito não foi refeito quando o Discord voltou")
	}
}

// CA-05.3 / RN-17: o cancelamento avisa com o motivo e deixa a sala encerrando.
func TestDiscordSync_CA05_3_Cancel(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	f.connect(t, "301")
	f.connect(t, "302")
	_, text, _ := f.openRoom(t, path, ana)
	if rec, got := call(t, f.h, http.MethodPost, path+"/cancel", ana, `{"reason":"Faltou o tank, fica para amanhã"}`); rec.Code != http.StatusOK {
		t.Fatalf("cancelar: %d %v", rec.Code, got)
	}
	f.tick(t)
	m := lastMessage(t, f, text)
	if !strings.Contains(m.Content, "cancelado") || !strings.Contains(m.Content, "Faltou o tank, fica para amanhã") ||
		!discordfake.HasMention(m.Content, "301") || !discordfake.HasMention(m.Content, "302") {
		t.Errorf("mensagem = %q", m.Content)
	}
	var status string
	var closing *time.Time
	if err := f.pool.QueryRow(t.Context(), `SELECT status, closing_at FROM discord_rooms`).Scan(&status, &closing); err != nil || status != "closing" || closing == nil {
		t.Errorf("sala = %s, %v, %v", status, closing, err)
	}
}

// CA-06.4 / RN-15: o horário novo vai para a sala, e o lembrete segue o horário novo.
func TestDiscordSync_CA06_4_StartChanged(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	f.connect(t, "301")
	_, text, _ := f.openRoom(t, path, ana)
	_, lobby := call(t, f.h, http.MethodGet, path, ana, "")
	oldStart, _ := time.Parse(time.RFC3339, lobby["startsAt"].(string))
	newStart := oldStart.Add(time.Hour)
	upd := fmt.Sprintf(`{"startsAt":%q,"slots":{"tank":1,"support":2,"dps":3},"minLevel":160}`, newStart.Format(time.RFC3339))
	if rec, got := call(t, f.h, http.MethodPut, path, ana, upd); rec.Code != http.StatusOK {
		t.Fatalf("editar: %d %v", rec.Code, got)
	}
	f.tick(t)
	m := lastMessage(t, f, text)
	if !strings.Contains(m.Content, "Horário novo") || !strings.Contains(m.Content, "21:00") || !discordfake.HasMention(m.Content, "301") {
		t.Errorf("mensagem = %q", m.Content)
	}
	before := len(f.fake.Messages(text))
	f.rooms.Now = func() time.Time { return oldStart.Add(-15 * time.Minute) }
	f.tick(t)
	if n := len(f.fake.Messages(text)); n != before {
		t.Error("lembrete no horário antigo")
	}
	f.rooms.Now = func() time.Time { return newStart.Add(-15 * time.Minute) }
	f.tick(t)
	if !strings.Contains(lastMessage(t, f, text).Content, "Lembrete") {
		t.Error("sem lembrete no horário novo")
	}
	f.rooms.Now = func() time.Time { return oldStart.Add(6 * time.Hour) }
	f.tick(t)
	if status, _, _, _ := f.roomRow(t); status != "open" {
		t.Errorf("apagada no horário antigo: %s", status)
	}
	f.rooms.Now = func() time.Time { return newStart.Add(6 * time.Hour) }
	f.tick(t)
	if status, _, _, _ := f.roomRow(t); status != "deleted" {
		t.Errorf("não apagada no horário novo: %s", status)
	}
}
