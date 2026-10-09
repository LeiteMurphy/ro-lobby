package server

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
)

// Testes das rotas do banco de talentos com serviço falso, sem banco (spec
// banco-de-talentos, T-02 a T-04). O fluxo com o PostgreSQL fica em internal/talents.

type fakeTalents struct {
	list   []talents.Talent
	err    error
	got    *talents.Availability
	gotIn  talents.AvailabilityInput
	gotIDs [2]string
}

func (f *fakeTalents) SetAvailability(_ context.Context, userID, characterID string, in talents.AvailabilityInput) (*talents.Availability, error) {
	f.gotIDs, f.gotIn = [2]string{userID, characterID}, in
	return f.got, f.err
}

func (f *fakeTalents) ForLobby(_ context.Context, userID, lobbyID string) ([]talents.Talent, error) {
	f.gotIDs = [2]string{userID, lobbyID}
	return f.list, f.err
}

var fogo = talents.Talent{
	CharacterID: charID, Nick: "Fogo", ClassID: "arquimago", Level: 200, Role: "dps", Portrait: "retrato-2",
	Days: []int{3}, Start: "18:00", End: "00:00", AnyInstance: true, Instances: []talents.Instance{},
	DiscordUsername: "caio",
}

func newTalentsServer(f *fakeTalents) http.Handler {
	return New(fakePinger(func(context.Context) error { return nil }), &fakeAuth{}, &fakeChars{}, nil, nil, f)
}

// CA-01.1 / RN-01: a rota repassa o pedido ao serviço e devolve a disponibilidade.
func TestSetAvailability_CA01_1_PassesInput(t *testing.T) {
	f := &fakeTalents{got: &talents.Availability{
		Enabled: true, Days: []int{1, 2}, Start: "19:00", End: "23:00", InstanceIDs: []string{"sonho-sombrio"},
	}}
	body := `{"enabled":true,"days":[1,2],"start":"19:00","end":"23:00","anyInstance":false,"instanceIds":["sonho-sombrio"]}`
	rec, got := call(t, newTalentsServer(f), http.MethodPut, "/characters/"+charID+"/availability", "token-valido", body)
	if rec.Code != http.StatusOK || f.gotIDs != [2]string{userID, charID} {
		t.Fatalf("status %d, ids %v", rec.Code, f.gotIDs)
	}
	wantIn := talents.AvailabilityInput{Enabled: true, Days: []int{1, 2}, Start: "19:00", End: "23:00", InstanceIDs: []string{"sonho-sombrio"}}
	if !reflect.DeepEqual(f.gotIn, wantIn) {
		t.Errorf("entrada = %+v", f.gotIn)
	}
	a, _ := got["availability"].(map[string]any)
	if a["start"] != "19:00" || a["enabled"] != true {
		t.Errorf("corpo = %v", got)
	}
}

// RN-01: desligar sem nunca ter ligado devolve availability null.
func TestSetAvailability_RN01_NullWhenNeverEnabled(t *testing.T) {
	rec, got := call(t, newTalentsServer(&fakeTalents{}), http.MethodPut, "/characters/"+charID+"/availability", "token-valido", `{"enabled":false}`)
	if v, ok := got["availability"]; rec.Code != http.StatusOK || !ok || v != nil {
		t.Errorf("status %d, corpo %v", rec.Code, got)
	}
}

// CA-01.3 / CA-01.5 / RN-05: sem sessão 401, de outro Usuário 404, inválido 422.
func TestSetAvailability_CA01_3_CA01_5_Errors(t *testing.T) {
	path := "/characters/" + charID + "/availability"
	if rec, _ := call(t, newTalentsServer(&fakeTalents{}), http.MethodPut, path, "", `{"enabled":false}`); rec.Code != http.StatusUnauthorized {
		t.Errorf("sem sessão: %d", rec.Code)
	}
	if rec, _ := call(t, newTalentsServer(&fakeTalents{err: talents.ErrNotFound}), http.MethodPut, path, "token-valido", `{"enabled":false}`); rec.Code != http.StatusNotFound {
		t.Errorf("de outro: %d", rec.Code)
	}
	invalid := &talents.ValidationError{Fields: []talents.FieldError{{Field: talents.FieldEnd, Code: talents.CodeSameAsStart}}}
	rec, got := call(t, newTalentsServer(&fakeTalents{err: invalid}), http.MethodPut, path, "token-valido", `{"enabled":true}`)
	fields, _ := got["fields"].([]any)
	if rec.Code != http.StatusUnprocessableEntity || len(fields) != 1 ||
		!reflect.DeepEqual(fields[0], map[string]any{"field": "end", "code": "same_as_start"}) {
		t.Errorf("inválido: %d %v", rec.Code, got)
	}
}

// CA-02.1 / RN-12: a lista do dono traz o personagem com o nome no Discord.
func TestListLobbyTalents_CA02_1(t *testing.T) {
	f := &fakeTalents{list: []talents.Talent{fogo}}
	rec, got := callList(t, newTalentsServer(f), "/lobbies/"+lobbyID+"/talents")
	if rec.Code != http.StatusOK || f.gotIDs != [2]string{userID, lobbyID} || len(got) != 1 {
		t.Fatalf("status %d, ids %v, corpo %v", rec.Code, f.gotIDs, got)
	}
	want := map[string]any{
		"characterId": charID, "nick": "Fogo", "classId": "arquimago", "level": float64(200), "role": "dps",
		"portrait": "retrato-2", "link": nil, "days": []any{float64(3)}, "start": "18:00", "end": "00:00",
		"anyInstance": true, "instances": []any{}, "discordUsername": "caio",
	}
	if !reflect.DeepEqual(got[0], want) {
		t.Errorf("corpo = %v", got[0])
	}
}

// CA-02.4 (API) / RN-09: sem sessão 401, de outro 404, encerrado 409.
func TestListLobbyTalents_CA02_4_Errors(t *testing.T) {
	path := "/lobbies/" + lobbyID + "/talents"
	if rec, _ := call(t, newTalentsServer(&fakeTalents{}), http.MethodGet, path, "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("sem sessão: %d", rec.Code)
	}
	if rec, _ := call(t, newTalentsServer(&fakeTalents{err: talents.ErrNotFound}), http.MethodGet, path, "token-valido", ""); rec.Code != http.StatusNotFound {
		t.Errorf("de outro: %d", rec.Code)
	}
	if rec, _ := call(t, newTalentsServer(&fakeTalents{err: talents.ErrNotOpen}), http.MethodGet, path, "token-valido", ""); rec.Code != http.StatusConflict {
		t.Errorf("encerrado: %d", rec.Code)
	}
}
