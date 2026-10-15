//go:build integration

package server

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// Link do Discord do dono, pelas rotas (spec discord-bot, T-07).

const ownerLink = "https://discord.gg/abc123"

// linkedLobby cria o lobby de amanhã às 20:00 com o link do dono e Bia como membro.
func (f roomFlow) linkedLobby(t *testing.T) (path, ana, bia string) {
	t.Helper()
	ana, bia = f.session(t, "301", "Ana"), f.session(t, "302", "Bia")
	host := f.character(t, ana, `{"nick":"Lirien","classId":"arcebispo","level":200,"role":"support","server":"freya"}`)
	brasa := f.character(t, bia, `{"nick":"Brasa","classId":"guardiao-real","level":200,"role":"tank","server":"freya"}`)
	tomorrow := time.Now().In(lobbies.Location).AddDate(0, 0, 1)
	start := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 20, 0, 0, 0, lobbies.Location).UTC().Format(time.RFC3339)
	rec, lobby := call(t, f.h, http.MethodPost, "/lobbies", ana, fmt.Sprintf(
		`{"instanceId":"templo-do-demonio-rei","startsAt":%q,"slots":{"tank":1,"support":2,"dps":3},"minLevel":160,"characterId":%q,"discordLink":%q}`,
		start, host, ownerLink))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CA-07.1: criar com link: %d %v", rec.Code, lobby)
	}
	path = "/lobbies/" + lobby["id"].(string)
	f.member(t, path, ana, bia, brasa)
	return path, ana, bia
}

func ownerLinkOf(t *testing.T, got map[string]any) any {
	t.Helper()
	d, ok := got["discord"].(map[string]any)
	if !ok {
		return nil
	}
	return d["ownerLink"]
}

// CA-07.1 / CA-07.4 / RN-23: o link criado aparece para o dono e para o membro.
func TestOwnerLink_CA07_1_CA07_4_MembersSee(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, bia := f.linkedLobby(t)
	for name, token := range map[string]string{"dono": ana, "membro": bia} {
		_, got := call(t, f.h, http.MethodGet, path, token, "")
		if ownerLinkOf(t, got) != ownerLink {
			t.Errorf("%s: link = %v", name, ownerLinkOf(t, got))
		}
	}
}

// CA-07.2 / CA-07.3 / RN-23: link inválido recusa a edição e mantém o anterior; apagar
// tira o link.
func TestOwnerLink_CA07_2_CA07_3_EditAndClear(t *testing.T) {
	f := newRoomFlow(t, true)
	path, ana, _ := f.linkedLobby(t)
	_, lobby := call(t, f.h, http.MethodGet, path, ana, "")
	upd := func(link string) string {
		return fmt.Sprintf(`{"startsAt":%q,"slots":{"tank":1,"support":2,"dps":3},"minLevel":160,"discordLink":%q}`, lobby["startsAt"], link)
	}
	for _, bad := range []string{"https://exemplo.com/discord", "http://discord.gg/abc123"} {
		rec, got := call(t, f.h, http.MethodPut, path, ana, upd(bad))
		fields, _ := got["fields"].([]any)
		if rec.Code != http.StatusUnprocessableEntity || len(fields) != 1 || fields[0].(map[string]any)["field"] != "discordLink" {
			t.Errorf("CA-07.2: %q: %d %v", bad, rec.Code, got)
		}
	}
	_, got := call(t, f.h, http.MethodGet, path, ana, "")
	if ownerLinkOf(t, got) != ownerLink {
		t.Errorf("CA-07.2: link mudou: %v", ownerLinkOf(t, got))
	}
	if rec, got := call(t, f.h, http.MethodPut, path, ana, upd("")); rec.Code != http.StatusOK {
		t.Fatalf("CA-07.3: apagar: %d %v", rec.Code, got)
	}
	_, got = call(t, f.h, http.MethodGet, path, ana, "")
	if ownerLinkOf(t, got) != nil {
		t.Errorf("CA-07.3: link continua: %v", ownerLinkOf(t, got))
	}
}

// CA-07.5 / CA-07.6 / RN-19 / RN-25: quem não é membro não recebe o link, e ele não sai na
// lista da Home.
func TestOwnerLink_CA07_5_CA07_6_NotOutsideTheGroup(t *testing.T) {
	f := newRoomFlow(t, true)
	path, _, _ := f.linkedLobby(t)
	duda := f.session(t, "304", "Duda")
	cura := f.character(t, duda, `{"nick":"Cura","classId":"arcebispo","level":200,"role":"support","server":"freya"}`)
	f.apply(t, path, duda, cura)
	for name, token := range map[string]string{"visitante": "", "pendente": duda, "outro": f.session(t, "305", "Edu")} {
		rec, _ := call(t, f.h, http.MethodGet, path, token, "")
		if strings.Contains(rec.Body.String(), "discord.gg") {
			t.Errorf("CA-07.5: %s recebeu o link", name)
		}
	}
	day := time.Now().In(lobbies.Location).AddDate(0, 0, 1).Format(time.DateOnly)
	rec, _ := call(t, f.h, http.MethodGet, "/lobbies?from="+day+"&to="+day, "", "")
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "discord.gg") {
		t.Errorf("CA-07.6: a lista da Home traz o link: %d", rec.Code)
	}
}

// CA-07.7 / RN-24: com a sala desligada, o membro vê o link do anfitrião.
func TestOwnerLink_CA07_7_WithoutBot(t *testing.T) {
	f := newRoomFlow(t, false)
	path, _, bia := f.linkedLobby(t)
	_, got := call(t, f.h, http.MethodGet, path, bia, "")
	d := got["discord"].(map[string]any)
	if d["ownerLink"] != ownerLink || d["enabled"] != false || d["room"] != nil {
		t.Errorf("bloco = %v", d)
	}
}
