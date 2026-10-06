package server

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/characters"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// Testes das rotas de lobbies e do catálogo de instâncias com serviços falsos, sem banco
// (spec lobbies, T-05). O fluxo com o PostgreSQL fica em lobbies_integration_test.go.

const lobbyID = "4b1c2d3e-5f60-4a7b-8c9d-0e1f2a3b4c5d"

var temple = lobbies.Lobby{
	ID: lobbyID, InstanceID: "templo-do-demonio-rei", InstanceName: "Templo do Demônio Rei",
	InstanceLevel: 160, InstanceReset: "daily", StartsAt: time.Date(2026, 10, 7, 23, 0, 0, 0, time.UTC),
	Status: lobbies.StatusOpen, Slots: lobbies.Slots{Tank: 1, Support: 2, Dps: 3},
	Occupied: lobbies.Slots{Support: 1}, MinLevel: 160,
	Owner: lobbies.Owner{UserID: userID, DiscordName: "Grimbold", CharacterID: charID, Nick: "Lirien",
		ClassID: "arcebispo", Level: 178, Portrait: "retrato-2", Role: "support"},
	CreatedAt: time.Date(2026, 10, 6, 19, 40, 0, 0, time.UTC),
}

type fakeLobbies struct {
	err      error
	gotUser  string
	gotID    string
	gotFrom  string
	gotTo    string
	gotIn    lobbies.Input
	gotUpd   lobbies.UpdateInput
	gotWhy   string
	listCall bool
}

func (f *fakeLobbies) List(_ context.Context, from, to string) ([]lobbies.Lobby, error) {
	f.listCall, f.gotFrom, f.gotTo = true, from, to
	return []lobbies.Lobby{temple}, f.err
}

func (f *fakeLobbies) Get(_ context.Context, id string) (lobbies.Lobby, error) {
	f.gotID = id
	return temple, f.err
}

func (f *fakeLobbies) Create(_ context.Context, userID string, in lobbies.Input) (lobbies.Lobby, error) {
	f.gotUser, f.gotIn = userID, in
	return temple, f.err
}

func (f *fakeLobbies) Update(_ context.Context, userID, id string, in lobbies.UpdateInput) (lobbies.Lobby, error) {
	f.gotUser, f.gotID, f.gotUpd = userID, id, in
	return temple, f.err
}

func (f *fakeLobbies) Cancel(_ context.Context, userID, id, reason string) (lobbies.Lobby, error) {
	f.gotUser, f.gotID, f.gotWhy = userID, id, reason
	return temple, f.err
}

func newLobbiesServer(f *fakeLobbies) http.Handler {
	return New(fakePinger(func(context.Context) error { return nil }), &fakeAuth{}, &fakeChars{}, f)
}

const templeJSON = `{"instanceId":"templo-do-demonio-rei","startsAt":"2026-10-07T23:00:00Z","slots":{"tank":1,"support":2,"dps":3},"minLevel":160,"characterId":"` + charID + `","note":"Chamar no Discord"}`

var lobbyWrites = []struct{ method, path, body string }{
	{http.MethodPost, "/lobbies", templeJSON},
	{http.MethodPut, "/lobbies/" + lobbyID, `{"startsAt":"2026-10-08T00:00:00Z","slots":{"tank":1,"support":2,"dps":5},"minLevel":170}`},
	{http.MethodPost, "/lobbies/" + lobbyID + "/cancel", `{"reason":"Metade do grupo não pode"}`},
}

// CA-06.3 / RN-04: criar, editar e cancelar exigem sessão; o serviço nem é chamado.
func TestLobbies_CA06_3_WritesRequireSession(t *testing.T) {
	for _, token := range []string{"", "token-desconhecido"} {
		for _, r := range lobbyWrites {
			f := &fakeLobbies{}
			rec, body := call(t, newLobbiesServer(f), r.method, r.path, token, r.body)
			if rec.Code != http.StatusUnauthorized || body["error"] != "no_session" || f.gotUser != "" {
				t.Errorf("%s %s (token %q): status %d, corpo %v", r.method, r.path, token, rec.Code, body)
			}
		}
	}
}

// CA-06.1 / RN-01 / RN-02: GET /instances sem sessão, na ordem da escolha.
func TestInstances_CA06_1_PublicCatalog(t *testing.T) {
	rec, _ := call(t, newLobbiesServer(&fakeLobbies{}), http.MethodGet, "/instances", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var got []map[string]any
	decode(t, rec.Body.Bytes(), &got)
	if len(got) != 51 {
		t.Fatalf("%d instâncias, quer 51", len(got))
	}
	want := map[string]any{"id": "torre-da-constelacao", "name": "Torre da Constelação", "level": float64(240), "reset": "three_days"}
	if !reflect.DeepEqual(got[0], want) {
		t.Errorf("primeira = %v", got[0])
	}
}

// RN-13 / RN-22: GET /lobbies sem sessão, com o intervalo de datas e o formato do contrato.
func TestLobbies_RN22_PublicList(t *testing.T) {
	f := &fakeLobbies{}
	rec, _ := call(t, newLobbiesServer(f), http.MethodGet, "/lobbies?from=2026-10-06&to=2026-10-19", "", "")
	if rec.Code != http.StatusOK || f.gotFrom != "2026-10-06" || f.gotTo != "2026-10-19" {
		t.Fatalf("status %d, intervalo %s..%s", rec.Code, f.gotFrom, f.gotTo)
	}
	var got []map[string]any
	decode(t, rec.Body.Bytes(), &got)
	want := map[string]any{
		"id":       lobbyID,
		"instance": map[string]any{"id": "templo-do-demonio-rei", "name": "Templo do Demônio Rei", "level": float64(160), "reset": "daily"},
		"startsAt": "2026-10-07T23:00:00Z", "status": "open",
		"slots":    map[string]any{"tank": float64(1), "support": float64(2), "dps": float64(3)},
		"occupied": map[string]any{"tank": float64(0), "support": float64(1), "dps": float64(0)},
		"minLevel": float64(160), "note": nil, "cancelReason": nil, "createdAt": "2026-10-06T19:40:00Z",
		"owner": map[string]any{"userId": userID, "discordName": "Grimbold", "characterId": charID, "nick": "Lirien",
			"classId": "arcebispo", "level": float64(178), "portrait": "retrato-2", "role": "support"},
	}
	if len(got) != 1 || !reflect.DeepEqual(got[0], want) {
		t.Errorf("corpo = %v", got)
	}
}

// Datas fora do formato ou intervalo invertido respondem 400.
func TestLobbies_RN22_InvalidRange(t *testing.T) {
	rec, _ := call(t, newLobbiesServer(&fakeLobbies{}), http.MethodGet, "/lobbies?from=ontem&to=2026-10-19", "", "")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("data malformada: status %d", rec.Code)
	}
	rec, body := call(t, newLobbiesServer(&fakeLobbies{err: lobbies.ErrInvalidRange}), http.MethodGet, "/lobbies?from=2026-10-19&to=2026-10-06", "", "")
	if rec.Code != http.StatusBadRequest || body["error"] != "invalid_range" {
		t.Errorf("intervalo invertido: status %d, corpo %v", rec.Code, body)
	}
}

// CA-03.2 / RN-15: lobby inexistente responde 404 sem sessão.
func TestGetLobby_CA03_2_NotFound(t *testing.T) {
	f := &fakeLobbies{err: lobbies.ErrNotFound}
	rec, body := call(t, newLobbiesServer(f), http.MethodGet, "/lobbies/x", "", "")
	if rec.Code != http.StatusNotFound || body["error"] != "not_found" || f.gotID != "x" {
		t.Errorf("status %d, corpo %v", rec.Code, body)
	}
}

// CA-01.1 / RN-04: criar repassa os campos do Usuário da sessão e responde 201.
func TestCreateLobby_CA01_1_Created(t *testing.T) {
	f := &fakeLobbies{}
	rec, body := call(t, newLobbiesServer(f), http.MethodPost, "/lobbies", "token-valido", templeJSON)
	if rec.Code != http.StatusCreated || body["id"] != lobbyID {
		t.Fatalf("status %d, corpo %v", rec.Code, body)
	}
	want := lobbies.Input{InstanceID: "templo-do-demonio-rei", StartsAt: time.Date(2026, 10, 7, 23, 0, 0, 0, time.UTC),
		Slots: lobbies.Slots{Tank: 1, Support: 2, Dps: 3}, MinLevel: 160, CharacterID: charID, Note: "Chamar no Discord"}
	if f.gotUser != userID || f.gotIn != want {
		t.Errorf("entrada = %+v (usuário %s)", f.gotIn, f.gotUser)
	}
}

// CA-01.7 / CA-01.12 / D-07: erros de campo respondem 422 com campo e código.
func TestCreateLobby_CA01_7_CA01_12_ValidationIs422(t *testing.T) {
	f := &fakeLobbies{err: &lobbies.ValidationError{Fields: []lobbies.FieldError{
		{Field: "instanceId", Code: "invalid"}, {Field: "characterId", Code: "invalid"},
	}}}
	for _, r := range lobbyWrites {
		rec, body := call(t, newLobbiesServer(f), r.method, r.path, "token-valido", r.body)
		want := map[string]any{"error": "validation", "fields": []any{
			map[string]any{"field": "instanceId", "code": "invalid"},
			map[string]any{"field": "characterId", "code": "invalid"},
		}}
		if rec.Code != http.StatusUnprocessableEntity || !reflect.DeepEqual(body, want) {
			t.Errorf("%s %s: status %d, corpo %v", r.method, r.path, rec.Code, body)
		}
	}
}

// CA-01.9 / RN-11: o limite responde 409 lobby_limit.
func TestCreateLobby_CA01_9_LimitIs409(t *testing.T) {
	rec, body := call(t, newLobbiesServer(&fakeLobbies{err: lobbies.ErrLimitReached}), http.MethodPost, "/lobbies", "token-valido", templeJSON)
	if rec.Code != http.StatusConflict || body["error"] != "lobby_limit" {
		t.Errorf("status %d, corpo %v", rec.Code, body)
	}
}

// CA-04.4 / CA-05.3 / RN-20 / D-08: lobby de outro é 404; iniciado ou cancelado é 409.
func TestLobbyWrites_CA04_4_CA05_3_NotFoundAndNotOpen(t *testing.T) {
	for _, r := range lobbyWrites[1:] {
		rec, body := call(t, newLobbiesServer(&fakeLobbies{err: lobbies.ErrNotFound}), r.method, r.path, "token-valido", r.body)
		if rec.Code != http.StatusNotFound || body["error"] != "not_found" {
			t.Errorf("%s %s de outro: status %d, corpo %v", r.method, r.path, rec.Code, body)
		}
		rec, body = call(t, newLobbiesServer(&fakeLobbies{err: lobbies.ErrNotOpen}), r.method, r.path, "token-valido", r.body)
		if rec.Code != http.StatusConflict || body["error"] != "lobby_not_open" {
			t.Errorf("%s %s iniciado: status %d, corpo %v", r.method, r.path, rec.Code, body)
		}
	}
}

// CA-04.1 / CA-05.1: editar e cancelar repassam os dados ao serviço.
func TestLobbyWrites_CA04_1_CA05_1_PassData(t *testing.T) {
	f := &fakeLobbies{}
	rec, _ := call(t, newLobbiesServer(f), lobbyWrites[1].method, lobbyWrites[1].path, "token-valido", lobbyWrites[1].body)
	want := lobbies.UpdateInput{StartsAt: time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC), Slots: lobbies.Slots{Tank: 1, Support: 2, Dps: 5}, MinLevel: 170}
	if rec.Code != http.StatusOK || f.gotID != lobbyID || f.gotUpd != want {
		t.Errorf("editar: status %d, %+v", rec.Code, f.gotUpd)
	}
	f = &fakeLobbies{}
	rec, _ = call(t, newLobbiesServer(f), lobbyWrites[2].method, lobbyWrites[2].path, "token-valido", lobbyWrites[2].body)
	if rec.Code != http.StatusOK || f.gotWhy != "Metade do grupo não pode" {
		t.Errorf("cancelar: status %d, motivo %q", rec.Code, f.gotWhy)
	}
}

// RN-21: erro inesperado do serviço de lobbies vira 500 sem detalhes (RN-21 da personagens).
func TestLobbies_RN21_UnexpectedErrorIs500(t *testing.T) {
	rec, _ := call(t, newLobbiesServer(&fakeLobbies{err: errors.New("banco caiu")}), http.MethodGet, "/lobbies/"+lobbyID, "", "")
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "banco caiu") {
		t.Errorf("status %d, corpo %q", rec.Code, rec.Body)
	}
}

// CA-06.4 / RN-21: excluir ou mudar a função do personagem dono de lobby aberto responde
// 409 character_in_open_lobby.
func TestCharacters_CA06_4_InOpenLobbyIs409(t *testing.T) {
	for _, r := range characterRoutes[2:4] {
		f := &fakeChars{err: characters.ErrInOpenLobby}
		rec, body := call(t, newCharsServer(f), r.method, r.path, "token-valido", r.body)
		if rec.Code != http.StatusConflict || body["error"] != "character_in_open_lobby" {
			t.Errorf("%s %s: status %d, corpo %v", r.method, r.path, rec.Code, body)
		}
	}
}
