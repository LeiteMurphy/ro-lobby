//go:build integration

package server

import (
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordfake"
)

// Avisar no Discord, pelas rotas (spec discord-bot, T-08).

func (f roomFlow) at(t time.Time) { f.rooms.Now = func() time.Time { return t } }

func alertCount(f roomFlow, text string) int {
	n := 0
	for _, m := range f.fake.Messages(text) {
		if strings.Contains(m.Content, "está chamando o grupo") {
			n++
		}
	}
	return n
}

// CA-08.1 / RN-26: o aviso menciona os membros conectados, sem o dono, com o link.
func TestDiscordAlert_CA08_1_MentionsWithoutOwner(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	f.connect(t, "301")
	f.connect(t, "302")
	f.connect(t, "303")
	_, text, _ := f.openRoom(t, path, ana)
	if rec, got := call(t, f.h, http.MethodPost, path+"/discord-alert", ana, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("avisar: %d %v", rec.Code, got)
	}
	m := lastMessage(t, f, text)
	if !strings.Contains(m.Content, "**Lirien** está chamando o grupo") || !strings.Contains(m.Content, "https://rolobby.com.br"+path) {
		t.Errorf("mensagem = %q", m.Content)
	}
	if strings.Join(m.AllowedMentions.Users, ",") != "302,303" || discordfake.HasMention(m.Content, "301") {
		t.Errorf("menções = %v", m.AllowedMentions.Users)
	}
}

// CA-08.2 / CA-08.3 / RN-27: intervalo de 10 minutos, sem chamar o Discord; o dono vê
// quando pode avisar de novo, o membro não.
func TestDiscordAlert_CA08_2_CA08_3_Cooldown(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, bia, _ := f.group(t)
	_, text, _ := f.openRoom(t, path, ana)
	t0 := time.Now().Truncate(time.Second)
	f.at(t0)
	if rec, _ := call(t, f.h, http.MethodPost, path+"/discord-alert", ana, ""); rec.Code != http.StatusNoContent {
		t.Fatal("primeiro aviso")
	}
	f.at(t0.Add(4 * time.Minute))
	calls := f.fake.BotCalls()
	rec, got := call(t, f.h, http.MethodPost, path+"/discord-alert", ana, "")
	if rec.Code != http.StatusTooManyRequests || got["error"] != "cooldown" || got["retryAfterSeconds"] != float64(360) {
		t.Errorf("CA-08.2: %d %v", rec.Code, got)
	}
	if f.fake.BotCalls() != calls {
		t.Error("CA-08.2: chamou o Discord no intervalo")
	}
	_, detail := call(t, f.h, http.MethodGet, path, ana, "")
	room := detail["discord"].(map[string]any)["room"].(map[string]any)
	if room["alertAvailableAt"] != t0.Add(10*time.Minute).UTC().Format(time.RFC3339) {
		t.Errorf("CA-08.2: alertAvailableAt = %v", room["alertAvailableAt"])
	}
	_, member := call(t, f.h, http.MethodGet, path, bia, "")
	if member["discord"].(map[string]any)["room"].(map[string]any)["alertAvailableAt"] != nil {
		t.Error("CA-08.2: o membro vê o intervalo do dono")
	}
	f.at(t0.Add(10 * time.Minute))
	if rec, _ := call(t, f.h, http.MethodPost, path+"/discord-alert", ana, ""); rec.Code != http.StatusNoContent {
		t.Errorf("CA-08.3: aviso depois de 10 min: %d", rec.Code)
	}
	if alertCount(f, text) != 2 {
		t.Errorf("avisos = %d", alertCount(f, text))
	}
}

// CA-08.4 / CA-08.5 / RN-26 / RN-28: só o dono, e só com a sala aberta.
func TestDiscordAlert_CA08_4_CA08_5_OwnerAndOpenRoom(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, bia, _ := f.group(t)
	if rec, got := call(t, f.h, http.MethodPost, path+"/discord-alert", ana, ""); rec.Code != http.StatusConflict || got["code"] != "no_room" {
		t.Errorf("CA-08.5: sem sala: %d %v", rec.Code, got)
	}
	_, text, _ := f.openRoom(t, path, ana)
	if rec, got := call(t, f.h, http.MethodPost, path+"/discord-alert", bia, ""); rec.Code != http.StatusConflict || got["code"] != "not_owner" {
		t.Errorf("CA-08.4: membro: %d %v", rec.Code, got)
	}
	if _, err := f.pool.Exec(t.Context(), `UPDATE discord_rooms SET status = 'closing', closing_at = now()`); err != nil {
		t.Fatal(err)
	}
	if rec, got := call(t, f.h, http.MethodPost, path+"/discord-alert", ana, ""); rec.Code != http.StatusConflict || got["code"] != "no_room" {
		t.Errorf("CA-08.5: encerrando: %d %v", rec.Code, got)
	}
	if alertCount(f, text) != 0 {
		t.Error("postou aviso")
	}
}

// CA-08.6 / RN-28: a falha não conta para o intervalo e não é refeita.
func TestDiscordAlert_CA08_6_FailureDoesNotCount(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	_, text, _ := f.openRoom(t, path, ana)
	t0 := time.Now()
	f.at(t0)
	f.fake.SetBotFailures(discordfake.BotFailures{Down: true})
	if rec, got := call(t, f.h, http.MethodPost, path+"/discord-alert", ana, ""); rec.Code != http.StatusServiceUnavailable || got["error"] != "discord_unavailable" {
		t.Errorf("com o Discord fora: %d %v", rec.Code, got)
	}
	f.fake.SetBotFailures(discordfake.BotFailures{})
	f.tick(t)
	if alertCount(f, text) != 0 {
		t.Error("o aviso que falhou foi refeito")
	}
	f.at(t0.Add(time.Minute))
	if rec, _ := call(t, f.h, http.MethodPost, path+"/discord-alert", ana, ""); rec.Code != http.StatusNoContent {
		t.Errorf("depois da falha: %d", rec.Code)
	}
}

// CA-08.7 / RN-27: o lembrete não conta para o intervalo.
func TestDiscordAlert_CA08_7_ReminderDoesNotCount(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	_, text, _ := f.openRoom(t, path, ana)
	_, lobby := call(t, f.h, http.MethodGet, path, ana, "")
	start, _ := time.Parse(time.RFC3339, lobby["startsAt"].(string))
	f.at(start.Add(-15 * time.Minute))
	f.tick(t)
	if !strings.Contains(lastMessage(t, f, text).Content, "Lembrete") {
		t.Fatal("sem lembrete")
	}
	f.at(start.Add(-14 * time.Minute))
	if rec, _ := call(t, f.h, http.MethodPost, path+"/discord-alert", ana, ""); rec.Code != http.StatusNoContent {
		t.Errorf("aviso depois do lembrete: %d", rec.Code)
	}
}

// Caso de borda (RN-27): dois avisos ao mesmo tempo postam um só.
func TestDiscordAlert_RN27_Concurrent(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _, _ := f.group(t)
	_, text, _ := f.openRoom(t, path, ana)
	codes := make([]int, 2)
	var wg sync.WaitGroup
	for i := range codes {
		wg.Go(func() {
			rec, _ := call(t, f.h, http.MethodPost, path+"/discord-alert", ana, "")
			codes[i] = rec.Code
		})
	}
	wg.Wait()
	if alertCount(f, text) != 1 {
		t.Errorf("avisos = %d, respostas %v", alertCount(f, text), codes)
	}
}
