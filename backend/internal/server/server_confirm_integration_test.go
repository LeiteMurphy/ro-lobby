//go:build integration

package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Conferir o servidor dos personagens migrados pelas rotas, com o PostgreSQL real (spec
// servidor, T-02).

// toConfirm marca o personagem como migrado, como a 00010 deixa os que já existiam (RN-15).
func (f flow) toConfirm(t *testing.T, characterID string) {
	t.Helper()
	if _, err := f.pool.Exec(t.Context(), `UPDATE characters SET server_to_confirm = true WHERE id = $1`, characterID); err != nil {
		t.Fatal(err)
	}
}

func (f flow) characters(t *testing.T, token string) (*httptest.ResponseRecorder, []map[string]any) {
	t.Helper()
	rec, _ := call(t, f.h, http.MethodGet, "/characters", token, "")
	var got []map[string]any
	decode(t, rec.Body.Bytes(), &got)
	return rec, got
}

func (f flow) confirmState(t *testing.T, token, characterID string) bool {
	t.Helper()
	rec, list := f.characters(t, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("listar: %d", rec.Code)
	}
	for _, c := range list {
		if c["id"] == characterID {
			return c["serverToConfirm"].(bool)
		}
	}
	t.Fatalf("personagem %s fora da lista", characterID)
	return false
}

// CA-06.4 / RN-15: personagem criado depois da migração nasce conferido.
func TestConfirmServer_CA06_4_NewIsConfirmed(t *testing.T) {
	f := newFlow(t)
	ana := f.session(t, "1", "Ana")
	brasa := f.character(t, ana, `{"nick":"Brasa","classId":"guardiao-real","level":172,"role":"tank","server":"nidhogg"}`)
	if f.confirmState(t, ana, brasa) {
		t.Error("personagem novo a conferir")
	}
}

// CA-06.2 / RN-16: "Está certo" tira só aquele personagem da conferência, sem mudar o
// servidor; confirmar de novo continua 204.
func TestConfirmServer_CA06_2_ConfirmsOnlyThatOne(t *testing.T) {
	f := newFlow(t)
	ana := f.session(t, "1", "Ana")
	brasa := f.character(t, ana, `{"nick":"Brasa","classId":"guardiao-real","level":172,"role":"tank","server":"freya"}`)
	lirien := f.character(t, ana, `{"nick":"Lirien","classId":"arcebispo","level":178,"role":"support","server":"freya"}`)
	f.toConfirm(t, brasa)
	f.toConfirm(t, lirien)

	for range 2 {
		if rec, got := call(t, f.h, http.MethodPost, "/characters/"+brasa+"/confirm-server", ana, ""); rec.Code != http.StatusNoContent {
			t.Fatalf("confirmar: %d %v", rec.Code, got)
		}
	}
	if f.confirmState(t, ana, brasa) || !f.confirmState(t, ana, lirien) {
		t.Errorf("Brasa a conferir %v, Lirien a conferir %v", f.confirmState(t, ana, brasa), f.confirmState(t, ana, lirien))
	}
	_, list := f.characters(t, ana)
	for _, c := range list {
		if c["server"] != "freya" {
			t.Errorf("%s mudou para %v", c["nick"], c["server"])
		}
	}
}

// CA-06.3 / RN-16: salvar a edição confirma, inclusive mudando o servidor.
func TestConfirmServer_CA06_3_EditConfirms(t *testing.T) {
	f := newFlow(t)
	ana := f.session(t, "1", "Ana")
	brasa := f.character(t, ana, `{"nick":"Brasa","classId":"guardiao-real","level":172,"role":"tank","server":"freya"}`)
	f.toConfirm(t, brasa)
	rec, got := call(t, f.h, http.MethodPut, "/characters/"+brasa, ana, `{"nick":"Brasa","classId":"guardiao-real","level":172,"role":"tank","server":"nidhogg"}`)
	if rec.Code != http.StatusOK || got["server"] != "nidhogg" || got["serverToConfirm"] != false {
		t.Errorf("editar: %d %v", rec.Code, got)
	}
}

// CA-06.6 / RN-16: personagem de outro Usuário responde 404 e continua a conferir.
func TestConfirmServer_CA06_6_OnlyOwner(t *testing.T) {
	f := newFlow(t)
	ana, bia := f.session(t, "1", "Ana"), f.session(t, "2", "Bia")
	brasa := f.character(t, ana, `{"nick":"Brasa","classId":"guardiao-real","level":172,"role":"tank","server":"freya"}`)
	f.toConfirm(t, brasa)
	if rec, got := call(t, f.h, http.MethodPost, "/characters/"+brasa+"/confirm-server", bia, ""); rec.Code != http.StatusNotFound {
		t.Errorf("de outro Usuário: %d %v", rec.Code, got)
	}
	if !f.confirmState(t, ana, brasa) {
		t.Error("o personagem saiu da conferência")
	}
}

// CA-06.5 / RN-16: personagem a conferir cria lobby, se candidata e entra no banco de
// talentos normalmente.
func TestConfirmServer_CA06_5_WorksNormally(t *testing.T) {
	f := newFlow(t)
	ana, bia := f.session(t, "1", "Ana"), f.session(t, "2", "Bia")
	host := f.character(t, ana, `{"nick":"Anfitria","classId":"arcebispo","level":200,"role":"support","server":"freya"}`)
	brasa := f.character(t, bia, `{"nick":"Brasa","classId":"guardiao-real","level":200,"role":"tank","server":"freya"}`)
	f.toConfirm(t, host)
	f.toConfirm(t, brasa)

	path, _ := f.lobby(t, ana, host)
	if rec, got := call(t, f.h, http.MethodPost, path+"/applications", bia, fmt.Sprintf(`{"characterId":%q}`, brasa)); rec.Code != http.StatusCreated {
		t.Errorf("candidatar: %d %v", rec.Code, got)
	}
	body := `{"enabled":true,"days":[0,1,2,3,4,5,6],"start":"18:00","end":"23:00","anyInstance":true}`
	if rec, got := call(t, f.h, http.MethodPut, "/characters/"+brasa+"/availability", bia, body); rec.Code != http.StatusOK {
		t.Errorf("banco de talentos: %d %v", rec.Code, got)
	}
}
