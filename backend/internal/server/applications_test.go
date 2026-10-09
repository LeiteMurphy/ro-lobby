package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/applications"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// Testes das rotas de candidatura e do detalhe conforme quem olha, com serviços falsos,
// sem banco (spec candidatura-lobby, T-04). O fluxo com o PostgreSQL fica em
// applications_integration_test.go.

const appID = "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"

var pendingApp = applications.Application{
	ID: appID, LobbyID: lobbyID, UserID: userID, CharacterID: charID, Role: "support",
	Message: "tenho buff de ASPD", Status: applications.StatusPending,
	CreatedAt: time.Date(2026, 10, 6, 20, 0, 0, 0, time.UTC),
}

type fakeApps struct {
	err       error
	called    bool
	gotUser   string
	gotID     string
	gotIn     applications.ApplyInput
	gotReason string
	gotBlock  bool
	gotChar   string
	gotSwap   applications.SwapInput
	mine      []applications.Mine
}

func (f *fakeApps) Apply(_ context.Context, userID, lobbyID string, in applications.ApplyInput) (applications.Application, error) {
	f.called, f.gotUser, f.gotID, f.gotIn = true, userID, lobbyID, in
	return pendingApp, f.err
}

func (f *fakeApps) Accept(_ context.Context, ownerID, id string) (applications.Application, error) {
	f.called, f.gotUser, f.gotID = true, ownerID, id
	a := pendingApp
	a.Status = applications.StatusAccepted
	return a, f.err
}

func (f *fakeApps) Reject(_ context.Context, ownerID, id, reason string) (applications.Application, error) {
	f.called, f.gotUser, f.gotID, f.gotReason = true, ownerID, id, reason
	a := pendingApp
	a.Status, a.Reason = applications.StatusRejected, reason
	a.DecidedAt = time.Date(2026, 10, 6, 21, 0, 0, 0, time.UTC)
	return a, f.err
}

func (f *fakeApps) Withdraw(_ context.Context, userID, id string) (applications.Application, error) {
	f.called, f.gotUser, f.gotID = true, userID, id
	a := pendingApp
	a.Status = applications.StatusWithdrawn
	return a, f.err
}

func (f *fakeApps) ListMine(_ context.Context, userID string) ([]applications.Mine, error) {
	f.called, f.gotUser = true, userID
	return f.mine, f.err
}

func newAppsServer(f *fakeApps) http.Handler {
	return New(fakePinger(func(context.Context) error { return nil }), &fakeAuth{}, &fakeChars{}, &fakeLobbies{}, f, nil)
}

var applicationRoutes = []struct{ method, path, body string }{
	{http.MethodPost, "/lobbies/" + lobbyID + "/applications", `{"characterId":"` + charID + `"}`},
	{http.MethodPost, "/applications/" + appID + "/accept", ""},
	{http.MethodPost, "/applications/" + appID + "/reject", `{"reason":"já temos suporte"}`},
	{http.MethodPost, "/applications/" + appID + "/withdraw", ""},
	{http.MethodGet, "/me/applications", ""},
	{http.MethodPost, "/applications/" + appID + "/leave", ""},
	{http.MethodPost, "/applications/" + appID + "/remove", `{"reason":"mudamos o horário da run"}`},
	{http.MethodPut, "/lobbies/" + lobbyID + "/owner-character", `{"characterId":"` + charID + `"}`},
	{http.MethodPost, "/applications/" + appID + "/swap-requests", `{"characterId":"` + charID + `","reason":"ninguém apareceu de tank"}`},
	{http.MethodPost, "/swap-requests/" + swapID + "/accept", ""},
	{http.MethodPost, "/swap-requests/" + swapID + "/reject", `{"reason":"já achamos um tank"}`},
	{http.MethodPost, "/swap-requests/" + swapID + "/withdraw", ""},
}

// RNF-02: toda rota de candidatura exige sessão; o serviço nem é chamado.
func TestApplications_RNF02_RequireSession(t *testing.T) {
	for _, r := range applicationRoutes {
		for _, token := range []string{"", "token-vencido"} {
			f := &fakeApps{}
			rec, body := call(t, newAppsServer(f), r.method, r.path, token, r.body)
			if rec.Code != http.StatusUnauthorized || body["error"] != "no_session" || f.called {
				t.Errorf("%s %s (%q): status %d, corpo %v, chamou %v", r.method, r.path, token, rec.Code, body, f.called)
			}
		}
	}
}

// CA-01.1 / CA-01.2: candidatar repassa personagem e mensagem do Usuário da sessão.
func TestApply_CA01_1_Created(t *testing.T) {
	f := &fakeApps{}
	rec, body := call(t, newAppsServer(f), http.MethodPost, "/lobbies/"+lobbyID+"/applications", "token-valido",
		`{"characterId":"`+charID+`","message":"tenho buff de ASPD"}`)
	want := map[string]any{
		"id": appID, "lobbyId": lobbyID, "characterId": charID, "role": "support",
		"message": "tenho buff de ASPD", "status": "pending", "reason": nil, "blocked": false,
		"createdAt": "2026-10-06T20:00:00Z", "decidedAt": nil,
	}
	if rec.Code != http.StatusCreated || !reflect.DeepEqual(body, want) {
		t.Errorf("status %d, corpo %v", rec.Code, body)
	}
	if f.gotUser != userID || f.gotID != lobbyID || f.gotIn != (applications.ApplyInput{CharacterID: charID, Message: "tenho buff de ASPD"}) {
		t.Errorf("repassou %s %s %+v", f.gotUser, f.gotID, f.gotIn)
	}
}

// D-07: cada regra vira 409 com o código; campos viram 422; lobby inexistente, 404.
func TestApply_D07_Errors(t *testing.T) {
	for _, code := range []string{
		applications.CodeNotOpen, applications.CodeOwnLobby, applications.CodeAlreadyActive,
		applications.CodeRoleFull, applications.CodeRejectedBefore, applications.CodeBelowMinLevel,
	} {
		rec, body := call(t, newAppsServer(&fakeApps{err: &applications.RuleError{Code: code}}), http.MethodPost,
			"/lobbies/"+lobbyID+"/applications", "token-valido", `{"characterId":"`+charID+`"}`)
		if rec.Code != http.StatusConflict || body["error"] != "application_rule" || body["code"] != code {
			t.Errorf("%s: status %d, corpo %v", code, rec.Code, body)
		}
	}
	invalid := &lobbies.ValidationError{Fields: []lobbies.FieldError{{Field: "message", Code: "too_long"}}}
	rec, body := call(t, newAppsServer(&fakeApps{err: invalid}), http.MethodPost,
		"/lobbies/"+lobbyID+"/applications", "token-valido", `{"characterId":"`+charID+`"}`)
	if rec.Code != http.StatusUnprocessableEntity || body["error"] != "validation" {
		t.Errorf("422: status %d, corpo %v", rec.Code, body)
	}
	rec, body = call(t, newAppsServer(&fakeApps{err: applications.ErrNotFound}), http.MethodPost,
		"/lobbies/x/applications", "token-valido", `{"characterId":"`+charID+`"}`)
	if rec.Code != http.StatusNotFound || body["error"] != "not_found" {
		t.Errorf("404: status %d, corpo %v", rec.Code, body)
	}
}

// CA-02.1, CA-02.2, CA-03.2: aceitar, recusar e retirar repassam quem pede e a
// candidatura.
func TestDecide_CA02_1_CA02_2_CA03_2(t *testing.T) {
	for _, c := range []struct {
		path, body, status string
	}{
		{"/applications/" + appID + "/accept", "", "accepted"},
		{"/applications/" + appID + "/reject", `{"reason":"já temos suporte"}`, "rejected"},
		{"/applications/" + appID + "/withdraw", "", "withdrawn"},
	} {
		f := &fakeApps{}
		rec, body := call(t, newAppsServer(f), http.MethodPost, c.path, "token-valido", c.body)
		if rec.Code != http.StatusOK || body["status"] != c.status || f.gotUser != userID || f.gotID != appID {
			t.Errorf("%s: status %d, corpo %v", c.path, rec.Code, body)
		}
		if c.status == "rejected" && (f.gotReason != "já temos suporte" || body["reason"] != "já temos suporte" || body["decidedAt"] != "2026-10-06T21:00:00Z") {
			t.Errorf("recusa: %q, corpo %v", f.gotReason, body)
		}
	}
}

// CA-02.3, CA-02.6, CA-02.7, CA-03.4: 422 da justificativa, 409 com o código e 404.
func TestDecide_D07_Errors(t *testing.T) {
	routes := []struct{ path, body string }{
		{"/applications/" + appID + "/accept", ""},
		{"/applications/" + appID + "/reject", `{"reason":"já temos suporte"}`},
		{"/applications/" + appID + "/withdraw", ""},
	}
	for _, r := range routes {
		for _, code := range []string{applications.CodeNotOwner, applications.CodeNotPending, applications.CodeNotYours, applications.CodeRoleFull, applications.CodeScheduleConflict} {
			rec, body := call(t, newAppsServer(&fakeApps{err: &applications.RuleError{Code: code}}), http.MethodPost, r.path, "token-valido", r.body)
			if rec.Code != http.StatusConflict || body["code"] != code {
				t.Errorf("%s %s: status %d, corpo %v", r.path, code, rec.Code, body)
			}
		}
		rec, body := call(t, newAppsServer(&fakeApps{err: applications.ErrNotFound}), http.MethodPost, r.path, "token-valido", r.body)
		if rec.Code != http.StatusNotFound || body["error"] != "not_found" {
			t.Errorf("%s 404: status %d, corpo %v", r.path, rec.Code, body)
		}
	}
	invalid := &lobbies.ValidationError{Fields: []lobbies.FieldError{{Field: "reason", Code: "required"}}}
	rec, body := call(t, newAppsServer(&fakeApps{err: invalid}), http.MethodPost, "/applications/"+appID+"/reject", "token-valido", `{"reason":""}`)
	if rec.Code != http.StatusUnprocessableEntity || body["error"] != "validation" {
		t.Errorf("422: status %d, corpo %v", rec.Code, body)
	}
}

// CA-03.1 / RN-33: minhas candidaturas com lobby, personagem e estado; personagem
// excluído vem nulo.
func TestListMine_CA03_1(t *testing.T) {
	rejected := pendingApp
	rejected.Status, rejected.Reason = applications.StatusRejected, "já temos suporte"
	gone := pendingApp
	gone.CharacterID = ""
	f := &fakeApps{mine: []applications.Mine{
		{Application: rejected, InstanceName: "Templo do Demônio Rei", StartsAt: temple.StartsAt, LobbyStatus: "open",
			Nick: "Lirien", ClassID: "arcebispo", Level: 178, Portrait: "retrato-2"},
		{Application: gone, InstanceName: "Templo do Demônio Rei", StartsAt: temple.StartsAt, LobbyStatus: "started"},
	}}
	rec, body := callList(t, newAppsServer(f), "/me/applications")
	if rec.Code != http.StatusOK || len(body) != 2 || f.gotUser != userID {
		t.Fatalf("status %d, corpo %v", rec.Code, body)
	}
	first := body[0]
	if first["application"].(map[string]any)["reason"] != "já temos suporte" ||
		!reflect.DeepEqual(first["lobby"], map[string]any{"instanceName": "Templo do Demônio Rei", "startsAt": "2026-10-07T23:00:00Z", "status": "open"}) ||
		!reflect.DeepEqual(first["character"], map[string]any{"nick": "Lirien", "classId": "arcebispo", "level": float64(178), "portrait": "retrato-2"}) {
		t.Errorf("primeira = %v", first)
	}
	if body[1]["character"] != nil || body[1]["application"].(map[string]any)["characterId"] != nil {
		t.Errorf("personagem excluído = %v", body[1])
	}
}

// D-06 / RN-28 / RN-32: o detalhe recebe quem olha (vazio sem sessão) e devolve membros,
// pendentes só quando o serviço manda e a candidatura de quem olha.
func TestGetLobby_D06_Viewer(t *testing.T) {
	member := lobbies.Participant{
		ApplicationID: appID, UserID: userID, CharacterID: charID, Nick: "Brasa", ClassID: "guardiao-real",
		Level: 200, Portrait: "retrato-3", Link: "https://ragnaplace.com/brasa", Role: "tank",
		CreatedAt: time.Date(2026, 10, 6, 20, 0, 0, 0, time.UTC),
	}
	visitor := lobbies.Detail{Lobby: temple, Members: []lobbies.Participant{member}}
	f := &fakeLobbies{detail: &visitor}
	rec, body := call(t, newLobbiesServer(f), http.MethodGet, "/lobbies/"+lobbyID, "", "")
	if rec.Code != http.StatusOK || f.gotUser != "" {
		t.Fatalf("visitante: status %d, quem olha %q", rec.Code, f.gotUser)
	}
	wantMember := map[string]any{
		"applicationId": appID, "userId": userID, "discordName": nil, "characterId": charID, "nick": "Brasa",
		"classId": "guardiao-real", "level": float64(200), "portrait": "retrato-3", "link": "https://ragnaplace.com/brasa",
		"role": "tank", "message": nil, "createdAt": "2026-10-06T20:00:00Z",
	}
	if !reflect.DeepEqual(body["members"], []any{wantMember}) {
		t.Errorf("membros = %v", body["members"])
	}
	if _, ok := body["pending"]; ok {
		t.Errorf("visitante não recebe pendentes: %v", body["pending"])
	}
	if _, ok := body["myApplication"]; ok {
		t.Errorf("visitante não tem candidatura: %v", body["myApplication"])
	}

	owner := visitor
	member.DiscordName = "Bia"
	candidate := member
	candidate.Message = "tenho buff de ASPD"
	owner.Members = []lobbies.Participant{member}
	owner.Pending = []lobbies.Participant{candidate}
	owner.MyApplication = &lobbies.ViewerApplication{ID: appID, Role: "tank", Status: "rejected", Reason: "já temos tank", CreatedAt: member.CreatedAt}
	f = &fakeLobbies{detail: &owner}
	_, body = call(t, newLobbiesServer(f), http.MethodGet, "/lobbies/"+lobbyID, "token-valido", "")
	if f.gotUser != userID {
		t.Errorf("quem olha = %q", f.gotUser)
	}
	pending, _ := body["pending"].([]any)
	if len(pending) != 1 || pending[0].(map[string]any)["message"] != "tenho buff de ASPD" || pending[0].(map[string]any)["discordName"] != "Bia" {
		t.Errorf("pendentes = %v", body["pending"])
	}
	mine, _ := body["myApplication"].(map[string]any)
	if mine["status"] != "rejected" || mine["reason"] != "já temos tank" || mine["characterId"] != nil {
		t.Errorf("minha candidatura = %v", mine)
	}
}

func callList(t *testing.T, h http.Handler, path string) (*httptest.ResponseRecorder, []map[string]any) {
	t.Helper()
	rec, _ := call(t, h, http.MethodGet, path, "token-valido", "")
	var got []map[string]any
	decode(t, rec.Body.Bytes(), &got)
	return rec, got
}
