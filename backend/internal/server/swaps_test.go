package server

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/applications"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// Testes das rotas da Parte 2 da candidatura (sair, remover e trocas), com serviços
// falsos, sem banco (spec candidatura-lobby, T-13). O fluxo com o PostgreSQL fica em
// swaps_integration_test.go.

const (
	swapID  = "1c2d3e4f-5a6b-4c7d-8e9f-0a1b2c3d4e5f"
	toChar  = "2d3e4f5a-6b7c-4d8e-9f0a-1b2c3d4e5f6a"
	fromJSN = "2026-10-06T21:00:00Z"
)

var pendingSwap = applications.SwapRequest{
	ID: swapID, ApplicationID: appID, FromCharacterID: charID, ToCharacterID: toChar, ToRole: "tank",
	Reason: "ninguém apareceu de tank", Status: applications.StatusPending,
	CreatedAt: time.Date(2026, 10, 6, 21, 0, 0, 0, time.UTC),
}

func (f *fakeApps) Leave(_ context.Context, userID, id string) (applications.Application, error) {
	f.called, f.gotUser, f.gotID = true, userID, id
	a := pendingApp
	a.Status = applications.StatusLeft
	return a, f.err
}

func (f *fakeApps) Remove(_ context.Context, ownerID, id, reason string, block bool) (applications.Application, error) {
	f.called, f.gotUser, f.gotID, f.gotReason, f.gotBlock = true, ownerID, id, reason, block
	a := pendingApp
	a.Status, a.Reason, a.Blocked = applications.StatusRemoved, reason, block
	return a, f.err
}

func (f *fakeApps) SwapOwnerCharacter(_ context.Context, ownerID, lobbyID, characterID string) error {
	f.called, f.gotUser, f.gotID, f.gotChar = true, ownerID, lobbyID, characterID
	return f.err
}

func (f *fakeApps) RequestSwap(_ context.Context, userID, id string, in applications.SwapInput) (applications.SwapRequest, error) {
	f.called, f.gotUser, f.gotID, f.gotSwap = true, userID, id, in
	return pendingSwap, f.err
}

func (f *fakeApps) AcceptSwap(_ context.Context, ownerID, id string) (applications.SwapRequest, error) {
	f.called, f.gotUser, f.gotID = true, ownerID, id
	r := pendingSwap
	r.Status = applications.StatusAccepted
	return r, f.err
}

func (f *fakeApps) RejectSwap(_ context.Context, ownerID, id, reason string) (applications.SwapRequest, error) {
	f.called, f.gotUser, f.gotID, f.gotReason = true, ownerID, id, reason
	r := pendingSwap
	r.Status, r.DecisionReason = applications.StatusRejected, reason
	r.DecidedAt = time.Date(2026, 10, 6, 22, 0, 0, 0, time.UTC)
	return r, f.err
}

func (f *fakeApps) WithdrawSwap(_ context.Context, userID, id string) (applications.SwapRequest, error) {
	f.called, f.gotUser, f.gotID = true, userID, id
	r := pendingSwap
	r.Status = applications.StatusWithdrawn
	return r, f.err
}

// CA-05.1 / CA-06.1 / CA-06.6: sair e remover repassam quem pede, a candidatura, a
// justificativa e o bloqueio.
func TestLeaveRemove_CA05_1_CA06_1(t *testing.T) {
	f := &fakeApps{}
	rec, body := call(t, newAppsServer(f), http.MethodPost, "/applications/"+appID+"/leave", "token-valido", "")
	if rec.Code != http.StatusOK || body["status"] != "left" || f.gotUser != userID || f.gotID != appID {
		t.Errorf("sair: %d %v", rec.Code, body)
	}
	for _, c := range []struct {
		body  string
		block bool
	}{
		{`{"reason":"mudamos o horário da run"}`, false},
		{`{"reason":"mudamos o horário da run","block":true}`, true},
	} {
		f := &fakeApps{}
		rec, body := call(t, newAppsServer(f), http.MethodPost, "/applications/"+appID+"/remove", "token-valido", c.body)
		if rec.Code != http.StatusOK || body["status"] != "removed" || body["reason"] != "mudamos o horário da run" ||
			body["blocked"] != c.block || f.gotReason != "mudamos o horário da run" || f.gotBlock != c.block {
			t.Errorf("remover %s: %d %v", c.body, rec.Code, body)
		}
	}
}

// CA-07.1: a troca do dono repassa o personagem e devolve o detalhe de quem pede.
func TestOwnerSwap_CA07_1_ReturnsDetail(t *testing.T) {
	f := &fakeApps{}
	l := &fakeLobbies{}
	h := New(fakePinger(func(context.Context) error { return nil }), &fakeAuth{}, &fakeChars{}, l, f, nil)
	rec, body := call(t, h, http.MethodPut, "/lobbies/"+lobbyID+"/owner-character", "token-valido", `{"characterId":"`+charID+`"}`)
	if rec.Code != http.StatusOK || body["id"] != lobbyID || f.gotUser != userID || f.gotID != lobbyID || f.gotChar != charID {
		t.Errorf("troca: %d %v, repassou %s %s %s", rec.Code, body, f.gotUser, f.gotID, f.gotChar)
	}
	if l.gotID != lobbyID || l.gotUser != userID {
		t.Errorf("detalhe de %s para %s", l.gotID, l.gotUser)
	}
}

// CA-08.1, CA-08.5, CA-08.8, CA-08.11: pedir, aceitar, recusar e retirar repassam quem
// pede e o pedido.
func TestSwapRequests_CA08_1_CA08_5_CA08_8_CA08_11(t *testing.T) {
	f := &fakeApps{}
	rec, body := call(t, newAppsServer(f), http.MethodPost, "/applications/"+appID+"/swap-requests", "token-valido",
		`{"characterId":"`+toChar+`","reason":"ninguém apareceu de tank"}`)
	want := map[string]any{
		"id": swapID, "applicationId": appID, "fromCharacterId": charID, "toCharacterId": toChar, "toRole": "tank",
		"reason": "ninguém apareceu de tank", "status": "pending", "decisionReason": nil,
		"createdAt": fromJSN, "decidedAt": nil,
	}
	if rec.Code != http.StatusCreated || !reflect.DeepEqual(body, want) {
		t.Errorf("pedir: %d %v", rec.Code, body)
	}
	if f.gotUser != userID || f.gotID != appID || f.gotSwap != (applications.SwapInput{CharacterID: toChar, Reason: "ninguém apareceu de tank"}) {
		t.Errorf("repassou %s %s %+v", f.gotUser, f.gotID, f.gotSwap)
	}
	for _, c := range []struct{ path, body, status string }{
		{"/swap-requests/" + swapID + "/accept", "", "accepted"},
		{"/swap-requests/" + swapID + "/reject", `{"reason":"já achamos um tank"}`, "rejected"},
		{"/swap-requests/" + swapID + "/withdraw", "", "withdrawn"},
	} {
		f := &fakeApps{}
		rec, body := call(t, newAppsServer(f), http.MethodPost, c.path, "token-valido", c.body)
		if rec.Code != http.StatusOK || body["status"] != c.status || f.gotUser != userID || f.gotID != swapID {
			t.Errorf("%s: %d %v", c.path, rec.Code, body)
		}
		if c.status == "rejected" && (f.gotReason != "já achamos um tank" || body["decisionReason"] != "já achamos um tank") {
			t.Errorf("recusa: %q, %v", f.gotReason, body)
		}
	}
}

// D-13: cada regra vira 409 com o código; campos, 422; candidatura, lobby ou pedido
// inexistente, 404.
func TestPart2_D13_Errors(t *testing.T) {
	routes := []struct {
		method, path, body string
		validation         bool
	}{
		{http.MethodPost, "/applications/" + appID + "/leave", "", false},
		{http.MethodPost, "/applications/" + appID + "/remove", `{"reason":"mudamos o horário da run"}`, true},
		{http.MethodPut, "/lobbies/" + lobbyID + "/owner-character", `{"characterId":"` + charID + `"}`, true},
		{http.MethodPost, "/applications/" + appID + "/swap-requests", `{"characterId":"` + toChar + `","reason":"ninguém apareceu de tank"}`, true},
		{http.MethodPost, "/swap-requests/" + swapID + "/accept", "", false},
		{http.MethodPost, "/swap-requests/" + swapID + "/reject", `{"reason":"já achamos um tank"}`, true},
		{http.MethodPost, "/swap-requests/" + swapID + "/withdraw", "", false},
	}
	for _, r := range routes {
		for _, code := range []string{
			applications.CodeNotOpen, applications.CodeNotOwner, applications.CodeNotYours, applications.CodeNotMember,
			applications.CodeSwapPending, applications.CodeNotPending, applications.CodeRoleFull,
			applications.CodeBelowMinLevel, applications.CodeScheduleConflict,
		} {
			rec, body := call(t, newAppsServer(&fakeApps{err: &applications.RuleError{Code: code}}), r.method, r.path, "token-valido", r.body)
			if rec.Code != http.StatusConflict || body["error"] != "application_rule" || body["code"] != code {
				t.Errorf("%s %s: %d %v", r.path, code, rec.Code, body)
			}
		}
		rec, body := call(t, newAppsServer(&fakeApps{err: applications.ErrNotFound}), r.method, r.path, "token-valido", r.body)
		if rec.Code != http.StatusNotFound || body["error"] != "not_found" {
			t.Errorf("%s 404: %d %v", r.path, rec.Code, body)
		}
		if r.validation {
			invalid := &lobbies.ValidationError{Fields: []lobbies.FieldError{{Field: "reason", Code: "required"}}}
			rec, body := call(t, newAppsServer(&fakeApps{err: invalid}), r.method, r.path, "token-valido", r.body)
			if rec.Code != http.StatusUnprocessableEntity || body["error"] != "validation" {
				t.Errorf("%s 422: %d %v", r.path, rec.Code, body)
			}
		}
	}
	// CA-06.6: candidatar bloqueado é 409 blocked.
	rec, body := call(t, newAppsServer(&fakeApps{err: &applications.RuleError{Code: applications.CodeBlocked}}), http.MethodPost,
		"/lobbies/"+lobbyID+"/applications", "token-valido", `{"characterId":"`+charID+`"}`)
	if rec.Code != http.StatusConflict || body["code"] != "blocked" {
		t.Errorf("bloqueado: %d %v", rec.Code, body)
	}
}

// D-12: o detalhe leva os pedidos de troca do dono e, na candidatura de quem olha, o
// bloqueio e o pedido mais recente.
func TestGetLobby_D12_SwapRequestsAndBlocked(t *testing.T) {
	owner := lobbies.Detail{Lobby: temple, Members: []lobbies.Participant{}, Pending: []lobbies.Participant{},
		SwapRequests: []lobbies.LobbySwapRequest{{
			ID: swapID, ApplicationID: appID, UserID: userID, DiscordName: "Bia",
			From:   lobbies.SwapCharacter{CharacterID: charID, Nick: "Fogo", ClassID: "arquimago", Level: 200, Portrait: "retrato-1", Role: "dps"},
			To:     lobbies.SwapCharacter{Nick: "Brasa", ClassID: "guardiao-real", Level: 200, Role: "tank"},
			Reason: "ninguém apareceu de tank", CreatedAt: pendingSwap.CreatedAt,
		}},
	}
	_, body := call(t, newLobbiesServer(&fakeLobbies{detail: &owner}), http.MethodGet, "/lobbies/"+lobbyID, "token-valido", "")
	want := []any{map[string]any{
		"id": swapID, "applicationId": appID, "userId": userID, "discordName": "Bia",
		"from":   map[string]any{"characterId": charID, "nick": "Fogo", "classId": "arquimago", "level": float64(200), "portrait": "retrato-1", "role": "dps"},
		"to":     map[string]any{"characterId": nil, "nick": "Brasa", "classId": "guardiao-real", "level": float64(200), "portrait": nil, "role": "tank"},
		"reason": "ninguém apareceu de tank", "createdAt": fromJSN,
	}}
	if !reflect.DeepEqual(body["swapRequests"], want) {
		t.Errorf("pedidos = %v", body["swapRequests"])
	}

	swap := pendingSwap
	member := lobbies.Detail{Lobby: temple, Members: []lobbies.Participant{},
		MyApplication: &lobbies.ViewerApplication{ID: appID, Role: "dps", Status: "accepted", CreatedAt: pendingSwap.CreatedAt, SwapRequest: &swap}}
	_, body = call(t, newLobbiesServer(&fakeLobbies{detail: &member}), http.MethodGet, "/lobbies/"+lobbyID, "token-valido", "")
	mine := body["myApplication"].(map[string]any)
	if body["swapRequests"] != nil || mine["blocked"] != false || mine["swapRequest"].(map[string]any)["status"] != "pending" {
		t.Errorf("membro: %v", body)
	}

	blocked := lobbies.Detail{Lobby: temple, Members: []lobbies.Participant{},
		MyApplication: &lobbies.ViewerApplication{ID: appID, Role: "dps", Status: "removed", Reason: "mudamos o horário da run", Blocked: true, CreatedAt: pendingSwap.CreatedAt}}
	_, body = call(t, newLobbiesServer(&fakeLobbies{detail: &blocked}), http.MethodGet, "/lobbies/"+lobbyID, "token-valido", "")
	mine = body["myApplication"].(map[string]any)
	if mine["blocked"] != true || mine["swapRequest"] != nil || mine["reason"] != "mudamos o horário da run" {
		t.Errorf("bloqueado: %v", mine)
	}
}
