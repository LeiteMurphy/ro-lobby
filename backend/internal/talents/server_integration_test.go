//go:build integration

package talents_test

import (
	"reflect"
	"testing"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
)

// Banco de talentos por servidor (spec servidor, T-04).

// CA-05.1 / CA-05.4 / RN-17, RN-19: num lobby em Freya, só o personagem de Freya tem
// afinidade, e cada item traz o servidor.
func TestServer_CA05_1_AffinityOnlySameServer(t *testing.T) {
	e := setup(t)
	ana, lid := glast(t, e) // Lirien, de Freya
	bia := e.user(t, "bia", "Fogo:arquimago:200:dps:freya")
	caio := e.user(t, "caio", "Chama:arquimago:200:dps:nidhogg")
	e.put(t, bia, "Fogo", every("18:00", "00:00"))
	e.put(t, caio, "Chama", every("18:00", "00:00"))

	list, err := e.svc.ForLobby(t.Context(), ana.userID, lid)
	if err != nil || !reflect.DeepEqual(nicks(list), []string{"Fogo"}) {
		t.Fatalf("afinidade = %v, %v", nicks(list), err)
	}
	if list[0].Server != "freya" {
		t.Errorf("servidor do item = %q", list[0].Server)
	}
}

// CA-05.2 / RN-17: a prévia da criação conta pelo servidor do personagem escolhido.
func TestServer_CA05_2_CountUsesOwnerServer(t *testing.T) {
	e := setup(t)
	bia := e.user(t, "bia", "Fogo:arquimago:200:dps:freya")
	caio := e.user(t, "caio", "Chama:arquimago:200:dps:nidhogg")
	e.put(t, bia, "Fogo", every("18:00", "00:00"))
	e.put(t, caio, "Chama", every("18:00", "00:00"))
	ana := e.user(t, "ana", "Lirien:arcebispo:178:support:nidhogg", "Brisa:arcebispo:178:support:freya")

	in := talents.CountInput{
		InstanceID: "templo-do-demonio-rei", StartsAt: at(1, 20, 0), MinLevel: 160,
		Tank: 1, Support: 2, Dps: 3, CharacterID: ana.chars["Lirien"],
	}
	if n, err := e.svc.Count(t.Context(), ana.userID, in); err != nil || n != 1 {
		t.Errorf("com Lirien (Nidhogg) = %d, %v", n, err)
	}
	in.CharacterID = ana.chars["Brisa"]
	if n, err := e.svc.Count(t.Context(), ana.userID, in); err != nil || n != 1 {
		t.Errorf("com Brisa (Freya) = %d, %v", n, err)
	}
	all, _ := e.svc.Catalog(t.Context(), talents.CatalogFilter{})
	if len(all) != 2 {
		t.Errorf("banco = %v", nicks(all))
	}
}

// CA-05.3 / CA-05.4 / RN-18, RN-19: o filtro de servidor do catálogo; código fora do
// catálogo é erro no campo.
func TestServer_CA05_3_CatalogFilter(t *testing.T) {
	e := setup(t)
	bia := e.user(t, "bia", "Fogo:arquimago:200:dps:freya", "Brasa:guardiao-real:172:tank:freya", "Cura:arcebispo:190:support:freya")
	caio := e.user(t, "caio", "Chama:arquimago:200:dps:nidhogg")
	for _, nick := range []string{"Fogo", "Brasa", "Cura"} {
		e.put(t, bia, nick, every("18:00", "00:00"))
	}
	e.put(t, caio, "Chama", every("18:00", "00:00"))

	list, err := e.svc.Catalog(t.Context(), talents.CatalogFilter{Server: "nidhogg"})
	if err != nil || !reflect.DeepEqual(nicks(list), []string{"Chama"}) || list[0].Server != "nidhogg" {
		t.Errorf("Nidhogg = %v, %v", list, err)
	}
	if list, _ := e.svc.Catalog(t.Context(), talents.CatalogFilter{Server: "freya"}); len(list) != 3 {
		t.Errorf("Freya = %v", nicks(list))
	}
	if list, _ := e.svc.Catalog(t.Context(), talents.CatalogFilter{}); len(list) != 4 {
		t.Errorf("Todos = %v", nicks(list))
	}
	_, err = e.svc.Catalog(t.Context(), talents.CatalogFilter{Server: "thor"})
	wantField(t, err, talents.FieldServer, talents.CodeInvalid)
}

// RN-17: o lobby em Nidhogg só enxerga Nidhogg, mesmo com o dono tendo personagem em Freya.
func TestServer_RN17_LobbyInNidhogg(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "ana", "Lirien:arcebispo:178:support:nidhogg")
	l, err := e.lobbies.Create(t.Context(), ana.userID, lobbies.Input{
		InstanceID: "templo-do-demonio-rei", StartsAt: at(1, 20, 0),
		Slots: lobbies.Slots{Tank: 1, Support: 1, Dps: 3}, MinLevel: 160, CharacterID: ana.chars["Lirien"],
	})
	if err != nil {
		t.Fatal(err)
	}
	bia := e.user(t, "bia", "Fogo:arquimago:200:dps:freya")
	caio := e.user(t, "caio", "Chama:arquimago:200:dps:nidhogg")
	e.put(t, bia, "Fogo", every("18:00", "00:00"))
	e.put(t, caio, "Chama", every("18:00", "00:00"))
	if list, err := e.svc.ForLobby(t.Context(), ana.userID, l.ID); err != nil || !reflect.DeepEqual(nicks(list), []string{"Chama"}) {
		t.Errorf("afinidade = %v, %v", nicks(list), err)
	}
}
