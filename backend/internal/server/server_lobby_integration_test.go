//go:build integration

package server

import (
	"fmt"
	"net/http"
	"testing"
)

// Lobby no servidor do dono e grupo no mesmo servidor, pelas rotas, com o PostgreSQL real
// (spec servidor, T-03).

func (f flow) lobbyServer(t *testing.T, path, token string) string {
	t.Helper()
	rec, got := call(t, f.h, http.MethodGet, path, token, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("detalhe: %d %v", rec.Code, got)
	}
	return got["server"].(string)
}

func wantWrongServer(t *testing.T, what string, rec int, got map[string]any) {
	t.Helper()
	if rec != http.StatusConflict || got["code"] != "wrong_server" {
		t.Errorf("%s: %d %v, quer 409 wrong_server", what, rec, got)
	}
}

// CA-02.1 / RN-06: o lobby fica no servidor do personagem do dono, e a lista da Home traz
// o servidor.
func TestLobbyServer_CA02_1_InheritsOwnerServer(t *testing.T) {
	f := newFlow(t)
	ana := f.session(t, "1", "Ana")
	brasa := f.character(t, ana, `{"nick":"Brasa","classId":"arcebispo","level":200,"role":"support","server":"nidhogg"}`)
	path, day := f.lobby(t, ana, brasa)
	if got := f.lobbyServer(t, path, ana); got != "nidhogg" {
		t.Errorf("detalhe: servidor %q", got)
	}
	code, list := callListAs(t, f.h, "/lobbies?from="+day+"&to="+day, "")
	if code != http.StatusOK || len(list) != 1 || list[0]["server"] != "nidhogg" {
		t.Errorf("lista: %d %v", code, list)
	}
}

// CA-02.3 / RN-07: o servidor não muda na edição, na troca do personagem do dono nem
// quando o personagem do dono é excluído depois do início.
func TestLobbyServer_CA02_3_NeverChanges(t *testing.T) {
	f := newFlow(t)
	ana := f.session(t, "1", "Ana")
	host := f.character(t, ana, `{"nick":"Anfitria","classId":"arcebispo","level":200,"role":"support","server":"nidhogg"}`)
	other := f.character(t, ana, `{"nick":"Outra","classId":"arcebispo","level":200,"role":"support","server":"nidhogg"}`)
	path, _ := f.lobby(t, ana, host)
	_, lobby := call(t, f.h, http.MethodGet, path, ana, "")

	upd := fmt.Sprintf(`{"startsAt":%q,"slots":{"tank":1,"support":2,"dps":5},"minLevel":170}`, lobby["startsAt"])
	if rec, got := call(t, f.h, http.MethodPut, path, ana, upd); rec.Code != http.StatusOK || got["server"] != "nidhogg" {
		t.Errorf("editar: %d %v", rec.Code, got)
	}
	if rec, got := call(t, f.h, http.MethodPut, path+"/owner-character", ana, fmt.Sprintf(`{"characterId":%q}`, other)); rec.Code != http.StatusOK {
		t.Fatalf("trocar o personagem do dono: %d %v", rec.Code, got)
	}
	if got := f.lobbyServer(t, path, ana); got != "nidhogg" {
		t.Errorf("depois da troca: %q", got)
	}

	if _, err := f.pool.Exec(t.Context(), `UPDATE lobbies SET starts_at = now() - interval '1 hour'`); err != nil {
		t.Fatal(err)
	}
	if rec, got := call(t, f.h, http.MethodDelete, "/characters/"+other, ana, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("excluir depois do início: %d %v", rec.Code, got)
	}
	if got := f.lobbyServer(t, path, ana); got != "nidhogg" {
		t.Errorf("depois da exclusão: %q", got)
	}
}

// CA-03.1 / CA-03.2 / RN-09: só personagem do servidor do lobby se candidata.
func TestLobbyServer_CA03_1_CA03_2_Apply(t *testing.T) {
	f := newFlow(t)
	ana, bia := f.session(t, "1", "Ana"), f.session(t, "2", "Bia")
	host := f.character(t, ana, `{"nick":"Anfitria","classId":"arcebispo","level":200,"role":"support","server":"freya"}`)
	nid := f.character(t, bia, `{"nick":"Brasa","classId":"guardiao-real","level":200,"role":"tank","server":"nidhogg"}`)
	fre := f.character(t, bia, `{"nick":"Fogo","classId":"arquimago","level":200,"role":"dps","server":"freya"}`)
	path, _ := f.lobby(t, ana, host)

	rec, got := call(t, f.h, http.MethodPost, path+"/applications", bia, fmt.Sprintf(`{"characterId":%q}`, nid))
	wantWrongServer(t, "candidatar de Nidhogg", rec.Code, got)
	_, detail := call(t, f.h, http.MethodGet, path, ana, "")
	if detail["pendingCount"] != float64(0) {
		t.Errorf("candidatura criada: %v", detail["pendingCount"])
	}
	if rec, got := call(t, f.h, http.MethodPost, path+"/applications", bia, fmt.Sprintf(`{"characterId":%q}`, fre)); rec.Code != http.StatusCreated || got["status"] != "pending" {
		t.Errorf("candidatar de Freya: %d %v", rec.Code, got)
	}
}

// CA-03.3 / RN-10: o dono não troca para um personagem de outro servidor.
func TestLobbyServer_CA03_3_OwnerSwap(t *testing.T) {
	f := newFlow(t)
	ana := f.session(t, "1", "Ana")
	host := f.character(t, ana, `{"nick":"Anfitria","classId":"arcebispo","level":200,"role":"support","server":"freya"}`)
	nid := f.character(t, ana, `{"nick":"Outra","classId":"arcebispo","level":200,"role":"support","server":"nidhogg"}`)
	path, _ := f.lobby(t, ana, host)

	rec, got := call(t, f.h, http.MethodPut, path+"/owner-character", ana, fmt.Sprintf(`{"characterId":%q}`, nid))
	wantWrongServer(t, "troca do dono", rec.Code, got)
	_, detail := call(t, f.h, http.MethodGet, path, ana, "")
	if detail["owner"].(map[string]any)["characterId"] != host {
		t.Errorf("personagem do dono mudou: %v", detail["owner"])
	}
}

// CA-03.4 / CA-03.5 / RN-04, RN-10: o membro não pede troca para outro servidor; para o
// mesmo servidor, o pedido e o aceite seguem as regras de sempre.
func TestLobbyServer_CA03_4_CA03_5_MemberSwap(t *testing.T) {
	f := newFlow(t)
	ana, bia := f.session(t, "1", "Ana"), f.session(t, "2", "Bia")
	host := f.character(t, ana, `{"nick":"Anfitria","classId":"arcebispo","level":200,"role":"support","server":"freya"}`)
	fogo := f.character(t, bia, `{"nick":"Fogo","classId":"arquimago","level":200,"role":"dps","server":"freya"}`)
	nid := f.character(t, bia, `{"nick":"Brasa","classId":"guardiao-real","level":200,"role":"tank","server":"nidhogg"}`)
	fre := f.character(t, bia, `{"nick":"Lamina","classId":"sicario","level":200,"role":"dps","server":"freya"}`)
	path, _ := f.lobby(t, ana, host)

	_, app := call(t, f.h, http.MethodPost, path+"/applications", bia, fmt.Sprintf(`{"characterId":%q}`, fogo))
	appPath := "/applications/" + app["id"].(string)
	if rec, got := call(t, f.h, http.MethodPost, appPath+"/accept", ana, ""); rec.Code != http.StatusOK {
		t.Fatalf("aceitar: %d %v", rec.Code, got)
	}

	rec, got := call(t, f.h, http.MethodPost, appPath+"/swap-requests", bia, fmt.Sprintf(`{"characterId":%q,"reason":"quero ir de tank hoje"}`, nid))
	wantWrongServer(t, "pedido de troca", rec.Code, got)
	_, detail := call(t, f.h, http.MethodGet, path, ana, "")
	if swaps, _ := detail["swapRequests"].([]any); len(swaps) != 0 {
		t.Errorf("pedido criado: %v", swaps)
	}

	rec, swap := call(t, f.h, http.MethodPost, appPath+"/swap-requests", bia, fmt.Sprintf(`{"characterId":%q,"reason":"quero ir de lâmina hoje"}`, fre))
	if rec.Code != http.StatusCreated {
		t.Fatalf("pedido do mesmo servidor: %d %v", rec.Code, swap)
	}
	// RN-04: com o pedido pendente, o personagem novo não muda de servidor.
	if rec, got := call(t, f.h, http.MethodPut, "/characters/"+fre, bia, `{"nick":"Lamina","classId":"sicario","level":200,"role":"dps","server":"nidhogg"}`); rec.Code != http.StatusConflict {
		t.Errorf("mudar o servidor do personagem do pedido: %d %v", rec.Code, got)
	}
	if rec, got := call(t, f.h, http.MethodPost, "/swap-requests/"+swap["id"].(string)+"/accept", ana, ""); rec.Code != http.StatusOK {
		t.Errorf("aceitar a troca: %d %v", rec.Code, got)
	}
}

func callListAs(t *testing.T, h http.Handler, path, token string) (int, []map[string]any) {
	t.Helper()
	rec, _ := call(t, h, http.MethodGet, path, token, "")
	var got []map[string]any
	decode(t, rec.Body.Bytes(), &got)
	return rec.Code, got
}
