package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/api"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/catalog"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/characters"
)

// Testes das rotas de personagem e do catálogo com serviços falsos, sem banco (spec
// personagens, T-04). O fluxo com o PostgreSQL fica em characters_integration_test.go.

const charID = "0b7e7f0e-8d4c-4f34-9a55-3f9f4c1a2b3c"

type fakeChars struct {
	err     error
	gotUser string
	gotID   string
	gotIn   characters.Input
}

var brasa = characters.Character{
	ID: charID, Nick: "Brasa", ClassID: "guardiao-real", Level: 172, Role: "tank",
	Portrait: "retrato-1", IsMain: true, CreatedAt: time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC),
}

func (f *fakeChars) List(_ context.Context, userID string) ([]characters.Character, error) {
	f.gotUser = userID
	return []characters.Character{brasa}, f.err
}

func (f *fakeChars) Create(_ context.Context, userID string, in characters.Input) (characters.Character, error) {
	f.gotUser, f.gotIn = userID, in
	return brasa, f.err
}

func (f *fakeChars) Update(_ context.Context, userID, id string, in characters.Input) (characters.Character, error) {
	f.gotUser, f.gotID, f.gotIn = userID, id, in
	return brasa, f.err
}

func (f *fakeChars) Delete(_ context.Context, userID, id string) error {
	f.gotUser, f.gotID = userID, id
	return f.err
}

func (f *fakeChars) SetMain(_ context.Context, userID, id string) error {
	f.gotUser, f.gotID = userID, id
	return f.err
}

func newCharsServer(f *fakeChars) http.Handler {
	return New(fakePinger(func(context.Context) error { return nil }), &fakeAuth{}, f, nil)
}

const brasaJSON = `{"nick":"Brasa","classId":"guardiao-real","level":172,"role":"tank"}`

var characterRoutes = []struct{ method, path, body string }{
	{http.MethodGet, "/characters", ""},
	{http.MethodPost, "/characters", brasaJSON},
	{http.MethodPut, "/characters/" + charID, brasaJSON},
	{http.MethodDelete, "/characters/" + charID, ""},
	{http.MethodPut, "/characters/" + charID + "/main", ""},
}

// CA-07.2 / RN-01: sem sessão (ou com sessão inválida), toda rota de personagem responde
// 401 e o serviço nem é chamado.
func TestCharacters_CA07_2_RequireSession(t *testing.T) {
	for _, token := range []string{"", "token-desconhecido"} {
		for _, r := range characterRoutes {
			f := &fakeChars{}
			rec, body := call(t, newCharsServer(f), r.method, r.path, token, r.body)
			if rec.Code != http.StatusUnauthorized || body["error"] != "no_session" {
				t.Errorf("%s %s (token %q): status %d, corpo %v", r.method, r.path, token, rec.Code, body)
			}
			if f.gotUser != "" {
				t.Errorf("%s %s: chamou o serviço sem sessão", r.method, r.path)
			}
		}
	}
}

// CA-07.1 / RN-20: GET /classes devolve as 82 classes sem sessão.
func TestClasses_CA07_1_PublicCatalog(t *testing.T) {
	rec, _ := call(t, newCharsServer(&fakeChars{}), http.MethodGet, "/classes", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var got []map[string]string
	decode(t, rec.Body.Bytes(), &got)
	if len(got) != 82 {
		t.Fatalf("%d classes, quer 82", len(got))
	}
	want := map[string]string{"id": "guardiao-real", "name": "Guardião Real", "plural": "Guardiões Reais", "tier": "terceira", "family": "Espadachim"}
	if !slices.ContainsFunc(got, func(c map[string]string) bool { return reflect.DeepEqual(c, want) }) {
		t.Errorf("Guardião Real não veio como %v", want)
	}
}

// RN-01 / RN-17: a lista é do Usuário da sessão, no formato do contrato.
func TestCharacters_RN01_ListOfSessionUser(t *testing.T) {
	f := &fakeChars{}
	rec, _ := call(t, newCharsServer(f), http.MethodGet, "/characters", "token-valido", "")
	if rec.Code != http.StatusOK || f.gotUser != userID {
		t.Fatalf("status %d, usuário %q", rec.Code, f.gotUser)
	}
	var got []map[string]any
	decode(t, rec.Body.Bytes(), &got)
	want := map[string]any{
		"id": charID, "nick": "Brasa", "classId": "guardiao-real", "level": float64(172), "role": "tank",
		"portrait": "retrato-1", "link": nil, "isMain": true, "createdAt": "2026-10-06T12:00:00Z",
	}
	if len(got) != 1 || !reflect.DeepEqual(got[0], want) {
		t.Errorf("corpo = %v", got)
	}
}

// CA-02.1 / CA-02.2: criar responde 201 e repassa os campos, com retrato e link.
func TestCharacters_CA02_1_Create(t *testing.T) {
	f := &fakeChars{}
	rec, body := call(t, newCharsServer(f), http.MethodPost, "/characters", "token-valido",
		`{"nick":"Lirien","classId":"arcebispo","level":178,"role":"support","portrait":"retrato-3","link":"https://exemplo.com"}`)
	if rec.Code != http.StatusCreated || body["nick"] != "Brasa" {
		t.Fatalf("status %d, corpo %v", rec.Code, body)
	}
	want := characters.Input{Nick: "Lirien", ClassID: "arcebispo", Level: 178, Role: "support", Portrait: "retrato-3", Link: "https://exemplo.com"}
	if f.gotIn != want {
		t.Errorf("entrada = %+v", f.gotIn)
	}
}

// CA-02.4 / RN-19 / D-07: erros de campo voltam 422 com campo e código.
func TestCharacters_CA02_4_ValidationIs422(t *testing.T) {
	f := &fakeChars{err: &characters.ValidationError{Fields: []characters.FieldError{
		{Field: "classId", Code: "invalid"}, {Field: "nick", Code: "taken"},
	}}}
	for _, r := range characterRoutes[1:3] {
		rec, body := call(t, newCharsServer(f), r.method, r.path, "token-valido", r.body)
		want := map[string]any{"error": "validation", "fields": []any{
			map[string]any{"field": "classId", "code": "invalid"},
			map[string]any{"field": "nick", "code": "taken"},
		}}
		if rec.Code != http.StatusUnprocessableEntity || !reflect.DeepEqual(body, want) {
			t.Errorf("%s %s: status %d, corpo %v", r.method, r.path, rec.Code, body)
		}
	}
}

// CA-02.8 / RN-11: o limite responde 409 character_limit.
func TestCharacters_CA02_8_LimitIs409(t *testing.T) {
	f := &fakeChars{err: characters.ErrLimitReached}
	rec, body := call(t, newCharsServer(f), http.MethodPost, "/characters", "token-valido", brasaJSON)
	if rec.Code != http.StatusConflict || body["error"] != "character_limit" {
		t.Errorf("status %d, corpo %v", rec.Code, body)
	}
}

// CA-03.3 / CA-04.4 / RN-02: personagem de outro Usuário responde 404 em editar, excluir
// e marcar principal.
func TestCharacters_CA03_3_CA04_4_NotFoundIs404(t *testing.T) {
	for _, r := range characterRoutes[2:] {
		f := &fakeChars{err: characters.ErrNotFound}
		rec, body := call(t, newCharsServer(f), r.method, r.path, "token-valido", r.body)
		if rec.Code != http.StatusNotFound || body["error"] != "not_found" {
			t.Errorf("%s %s: status %d, corpo %v", r.method, r.path, rec.Code, body)
		}
		if f.gotID != charID || f.gotUser != userID {
			t.Errorf("%s %s: id %q, usuário %q", r.method, r.path, f.gotID, f.gotUser)
		}
	}
}

// RN-13 / RN-14: excluir e marcar principal respondem 204.
func TestCharacters_RN13_RN14_DeleteAndSetMainAre204(t *testing.T) {
	for _, r := range characterRoutes[3:] {
		rec, _ := call(t, newCharsServer(&fakeChars{}), r.method, r.path, "token-valido", "")
		if rec.Code != http.StatusNoContent {
			t.Errorf("%s %s: status %d", r.method, r.path, rec.Code)
		}
	}
}

// RN-21: erro inesperado do serviço vira 500, sem detalhes no corpo.
func TestCharacters_RN21_UnexpectedErrorIs500(t *testing.T) {
	f := &fakeChars{err: errors.New("banco caiu")}
	rec, _ := call(t, newCharsServer(f), http.MethodGet, "/characters", "token-valido", "")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "banco caiu") {
		t.Errorf("o corpo vazou o erro interno: %q", rec.Body)
	}
}

// D-04 / RN-08 / RN-10: os enums do contrato são os mesmos do catálogo e do serviço.
func TestContractEnums_D04_MatchCatalog(t *testing.T) {
	portraits := []string{string(api.PortraitRetrato1), string(api.PortraitRetrato2), string(api.PortraitRetrato3), string(api.PortraitRetrato4)}
	if !slices.Equal(portraits, catalog.Portraits()) {
		t.Errorf("retratos do contrato %v, do catálogo %v", portraits, catalog.Portraits())
	}
	roles := []string{string(api.RoleTank), string(api.RoleSupport), string(api.RoleDps)}
	if !slices.Equal(roles, characters.Roles) {
		t.Errorf("funções do contrato %v, do serviço %v", roles, characters.Roles)
	}
	for _, c := range catalog.Classes() {
		if !api.ClassTier(c.Tier).Valid() {
			t.Errorf("tier %q de %s fora do contrato", c.Tier, c.ID)
		}
	}
}

func decode(t *testing.T, data []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("corpo não é JSON: %v (%s)", err, data)
	}
}
