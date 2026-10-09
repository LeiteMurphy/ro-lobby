package server

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
)

// Testes das rotas do banco de talentos com serviço falso, sem banco (spec
// banco-de-talentos, T-02 a T-04). O fluxo com o PostgreSQL fica em internal/talents.

type fakeTalents struct {
	list      []talents.Talent
	count     int
	gotFilter talents.CatalogFilter
	gotCount  talents.CountInput
	err       error
	got       *talents.Availability
	gotIn     talents.AvailabilityInput
	gotIDs    [2]string
}

func (f *fakeTalents) SetAvailability(_ context.Context, userID, characterID string, in talents.AvailabilityInput) (*talents.Availability, error) {
	f.gotIDs, f.gotIn = [2]string{userID, characterID}, in
	return f.got, f.err
}

func (f *fakeTalents) ForLobby(_ context.Context, userID, lobbyID string) ([]talents.Talent, error) {
	f.gotIDs = [2]string{userID, lobbyID}
	return f.list, f.err
}

func (f *fakeTalents) Catalog(_ context.Context, filter talents.CatalogFilter) ([]talents.Talent, error) {
	f.gotFilter = filter
	return f.list, f.err
}

func (f *fakeTalents) Count(_ context.Context, userID string, in talents.CountInput) (int, error) {
	f.gotIDs[0], f.gotCount = userID, in
	return f.count, f.err
}

var fogo = talents.Talent{
	CharacterID: charID, Nick: "Fogo", ClassID: "arquimago", Level: 200, Role: "dps", Portrait: "retrato-2",
	Days: []int{3}, Start: "18:00", End: "00:00", AnyInstance: true, Instances: []talents.Instance{},
	DiscordUsername: "caio", Removed: true,
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
		"removed": true, "blocked": false,
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

// CA-04.1 / CA-04.2 / RN-11 / RN-12 / RNF-05: os filtros chegam ao serviço; sem sessão o
// nome no Discord não sai, com sessão sai.
func TestListTalents_CA04_1_CA04_2(t *testing.T) {
	f := &fakeTalents{list: []talents.Talent{fogo}}
	rec, got := callList(t, newTalentsServer(f), "/talents?instanceId=sonho-sombrio&role=dps&day=6&time=01:00")
	day := 6
	want := talents.CatalogFilter{InstanceID: "sonho-sombrio", Role: "dps", Day: &day, Time: "01:00"}
	if rec.Code != http.StatusOK || !reflect.DeepEqual(f.gotFilter, want) {
		t.Fatalf("status %d, filtro %+v", rec.Code, f.gotFilter)
	}
	if got[0]["discordUsername"] != "caio" {
		t.Errorf("com sessão = %v", got[0])
	}
	rec, anon := call(t, newTalentsServer(f), http.MethodGet, "/talents", "", "")
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "caio") || strings.Contains(rec.Body.String(), "discordUsername") {
		t.Errorf("sem sessão = %d %s %v", rec.Code, rec.Body.String(), anon)
	}
}

// RN-11: filtro inválido responde 422.
func TestListTalents_RN11_Invalid(t *testing.T) {
	invalid := &talents.ValidationError{Fields: []talents.FieldError{{Field: talents.FieldTime, Code: talents.CodeInvalid}}}
	if rec, _ := call(t, newTalentsServer(&fakeTalents{err: invalid}), http.MethodGet, "/talents?time=01:15", "", ""); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status %d", rec.Code)
	}
}

// CA-03.1 (API) / RN-10: a contagem exige sessão e repassa o formulário.
func TestCountTalents_CA03_1(t *testing.T) {
	q := "/talents/count?instanceId=templo-do-demonio-rei&startsAt=2026-10-07T23:00:00Z&minLevel=160&tank=1&support=2&dps=3&characterId=" + charID
	if rec, _ := call(t, newTalentsServer(&fakeTalents{}), http.MethodGet, q, "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("sem sessão: %d", rec.Code)
	}
	f := &fakeTalents{count: 3}
	rec, got := call(t, newTalentsServer(f), http.MethodGet, q, "token-valido", "")
	want := talents.CountInput{
		InstanceID: "templo-do-demonio-rei", StartsAt: time.Date(2026, 10, 7, 23, 0, 0, 0, time.UTC), MinLevel: 160,
		Tank: 1, Support: 2, Dps: 3, CharacterID: charID,
	}
	if rec.Code != http.StatusOK || got["count"] != float64(3) || !reflect.DeepEqual(f.gotCount, want) || f.gotIDs[0] != userID {
		t.Errorf("status %d, corpo %v, entrada %+v", rec.Code, got, f.gotCount)
	}
}
