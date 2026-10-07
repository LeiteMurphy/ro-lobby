//go:build integration

package server

import (
	"fmt"
	"net/http"
	"testing"
)

// Fluxo da Parte 2 pelas rotas, com o PostgreSQL real: dono, dois membros e um visitante
// (spec candidatura-lobby, T-13: CA-05.1, CA-06.1, CA-06.3, CA-06.6, CA-07.1, CA-07.5,
// CA-08.1, CA-08.5, CA-08.12, D-12, D-13).
func TestPart2FlowIntegration_CA05_1_CA06_1_CA07_5_CA08_12(t *testing.T) {
	f := newFlow(t)
	h := f.h
	ana, bia, caio, duda := f.session(t, "1", "Ana"), f.session(t, "2", "Bia"), f.session(t, "3", "Caio"), f.session(t, "4", "Duda")
	host := f.character(t, ana, `{"nick":"Anfitria","classId":"arcebispo","level":200,"role":"support"}`)
	hostDps := f.character(t, ana, `{"nick":"Lamina","classId":"sicario","level":200,"role":"dps"}`)
	fogo := f.character(t, bia, `{"nick":"Fogo","classId":"arquimago","level":200,"role":"dps"}`)
	brasa := f.character(t, bia, `{"nick":"Brasa","classId":"guardiao-real","level":200,"role":"tank"}`)
	cura := f.character(t, caio, `{"nick":"Cura","classId":"arcebispo","level":200,"role":"support"}`)
	path, _ := f.lobby(t, ana, host)

	member := func(token, characterID string) string {
		t.Helper()
		rec, app := call(t, h, http.MethodPost, path+"/applications", token, fmt.Sprintf(`{"characterId":%q}`, characterID))
		if rec.Code != http.StatusCreated {
			t.Fatalf("candidatar: %d %v", rec.Code, app)
		}
		appPath := "/applications/" + app["id"].(string)
		if rec, got := call(t, h, http.MethodPost, appPath+"/accept", ana, ""); rec.Code != http.StatusOK {
			t.Fatalf("aceitar: %d %v", rec.Code, got)
		}
		return appPath
	}
	biaApp, caioApp := member(bia, fogo), member(caio, cura)

	// CA-08.1: Bia pede a troca de Fogo (Dano) para Brasa (Tank) e continua como Dano.
	rec, swap := call(t, h, http.MethodPost, biaApp+"/swap-requests", bia, fmt.Sprintf(`{"characterId":%q,"reason":"ninguém apareceu de tank"}`, brasa))
	if rec.Code != http.StatusCreated || swap["status"] != "pending" || swap["toRole"] != "tank" || swap["fromCharacterId"] != fogo {
		t.Fatalf("pedir troca: %d %v", rec.Code, swap)
	}
	swapPath := "/swap-requests/" + swap["id"].(string)
	if rec, got := call(t, h, http.MethodPost, biaApp+"/swap-requests", bia, fmt.Sprintf(`{"characterId":%q,"reason":"ninguém apareceu de tank"}`, brasa)); rec.Code != http.StatusConflict || got["code"] != "swap_pending" {
		t.Errorf("segundo pedido: %d %v", rec.Code, got)
	}
	if rec, got := call(t, h, http.MethodPost, biaApp+"/swap-requests", bia, fmt.Sprintf(`{"characterId":%q,"reason":"curto"}`, brasa)); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("motivo curto: %d %v", rec.Code, got)
	}

	// D-12: os pedidos de troca só para o dono; o membro vê o próprio pedido.
	_, got := call(t, h, http.MethodGet, path, ana, "")
	swaps, _ := got["swapRequests"].([]any)
	if len(swaps) != 1 {
		t.Fatalf("pedidos do dono: %v", got["swapRequests"])
	}
	s := swaps[0].(map[string]any)
	if s["discordName"] != "Bia" || s["reason"] != "ninguém apareceu de tank" ||
		s["from"].(map[string]any)["nick"] != "Fogo" || s["to"].(map[string]any)["nick"] != "Brasa" || s["to"].(map[string]any)["role"] != "tank" {
		t.Errorf("pedido = %v", s)
	}
	for name, token := range map[string]string{"sem sessão": "", "Duda": duda, "Caio (membro)": caio, "Bia (membro)": bia} {
		_, got := call(t, h, http.MethodGet, path, token, "")
		if got["swapRequests"] != nil {
			t.Errorf("%s vê os pedidos: %v", name, got["swapRequests"])
		}
	}
	_, got = call(t, h, http.MethodGet, path, bia, "")
	mine := got["myApplication"].(map[string]any)
	if mine["swapRequest"].(map[string]any)["id"] != swap["id"] || mine["swapRequest"].(map[string]any)["status"] != "pending" || mine["blocked"] != false {
		t.Errorf("pedido do membro: %v", mine)
	}

	// CA-08.12 e RN-22: outro membro não retira nem decide o pedido de Bia.
	if rec, got := call(t, h, http.MethodPost, swapPath+"/withdraw", caio, ""); rec.Code != http.StatusConflict || got["code"] != "not_yours" {
		t.Errorf("retirar de outro: %d %v", rec.Code, got)
	}
	if rec, got := call(t, h, http.MethodPost, swapPath+"/accept", caio, ""); rec.Code != http.StatusConflict || got["code"] != "not_owner" {
		t.Errorf("aceitar sem ser dono: %d %v", rec.Code, got)
	}
	if rec, got := call(t, h, http.MethodPost, swapPath+"/reject", ana, `{"reason":""}`); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("recusa sem justificativa: %d %v", rec.Code, got)
	}

	// CA-08.5: o dono aceita; Bia passa a ocupar a vaga de Tank com Brasa.
	if rec, got := call(t, h, http.MethodPost, swapPath+"/accept", ana, ""); rec.Code != http.StatusOK || got["status"] != "accepted" {
		t.Fatalf("aceitar troca: %d %v", rec.Code, got)
	}
	_, got = call(t, h, http.MethodGet, path, ana, "")
	if occ := got["occupied"].(map[string]any); occ["tank"] != float64(1) || occ["dps"] != float64(0) || occ["support"] != float64(2) {
		t.Errorf("ocupantes depois da troca: %v", occ)
	}
	if rec, _ := call(t, h, http.MethodPost, "/swap-requests/00000000-0000-0000-0000-000000000000/accept", ana, ""); rec.Code != http.StatusNotFound {
		t.Errorf("pedido inexistente: %d", rec.Code)
	}

	// CA-07.5: membro não troca o personagem do dono; CA-07.1: o dono troca o dele e
	// recebe o detalhe.
	if rec, got := call(t, h, http.MethodPut, path+"/owner-character", caio, fmt.Sprintf(`{"characterId":%q}`, cura)); rec.Code != http.StatusConflict || got["code"] != "not_owner" {
		t.Errorf("troca do dono por membro: %d %v", rec.Code, got)
	}
	rec, got = call(t, h, http.MethodPut, path+"/owner-character", ana, fmt.Sprintf(`{"characterId":%q}`, hostDps))
	if rec.Code != http.StatusOK || got["owner"].(map[string]any)["characterId"] != hostDps || got["swapRequests"] == nil ||
		got["occupied"].(map[string]any)["dps"] != float64(1) || got["occupied"].(map[string]any)["support"] != float64(1) {
		t.Errorf("troca do dono: %d %v", rec.Code, got)
	}
	if rec, _ := call(t, h, http.MethodPut, "/lobbies/00000000-0000-0000-0000-000000000000/owner-character", ana, fmt.Sprintf(`{"characterId":%q}`, hostDps)); rec.Code != http.StatusNotFound {
		t.Errorf("lobby inexistente: %d", rec.Code)
	}

	// CA-05.1: Caio sai; sair de novo é 409 not_member.
	if rec, got := call(t, h, http.MethodPost, caioApp+"/leave", caio, ""); rec.Code != http.StatusOK || got["status"] != "left" {
		t.Errorf("sair: %d %v", rec.Code, got)
	}
	if rec, got := call(t, h, http.MethodPost, caioApp+"/leave", caio, ""); rec.Code != http.StatusConflict || got["code"] != "not_member" {
		t.Errorf("sair de novo: %d %v", rec.Code, got)
	}

	// CA-06.3: quem não é dono não remove; CA-06.1 e CA-06.6: o dono remove Bia com
	// bloqueio, e ela não se candidata de novo.
	if rec, got := call(t, h, http.MethodPost, biaApp+"/remove", duda, `{"reason":"mudamos o horário da run"}`); rec.Code != http.StatusConflict || got["code"] != "not_owner" {
		t.Errorf("remover sem ser dono: %d %v", rec.Code, got)
	}
	if rec, got := call(t, h, http.MethodPost, biaApp+"/remove", ana, `{"reason":"curta"}`); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("remover com justificativa curta: %d %v", rec.Code, got)
	}
	rec, got = call(t, h, http.MethodPost, biaApp+"/remove", ana, `{"reason":"mudamos o horário da run","block":true}`)
	if rec.Code != http.StatusOK || got["status"] != "removed" || got["blocked"] != true || got["reason"] != "mudamos o horário da run" {
		t.Errorf("remover: %d %v", rec.Code, got)
	}
	_, got = call(t, h, http.MethodGet, path, bia, "")
	if mine := got["myApplication"].(map[string]any); mine["status"] != "removed" || mine["blocked"] != true || mine["reason"] != "mudamos o horário da run" {
		t.Errorf("removida: %v", mine)
	}
	if rec, got := call(t, h, http.MethodPost, path+"/applications", bia, fmt.Sprintf(`{"characterId":%q}`, fogo)); rec.Code != http.StatusConflict || got["code"] != "blocked" {
		t.Errorf("candidatar bloqueada: %d %v", rec.Code, got)
	}
	rec, _ = call(t, h, http.MethodGet, "/me/applications", bia, "")
	var list []map[string]any
	decode(t, rec.Body.Bytes(), &list)
	if len(list) != 1 || list[0]["application"].(map[string]any)["status"] != "removed" || list[0]["application"].(map[string]any)["reason"] != "mudamos o horário da run" {
		t.Errorf("minhas de Bia: %v", list)
	}
}
