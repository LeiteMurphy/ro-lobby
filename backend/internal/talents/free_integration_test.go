//go:build integration

package talents_test

import (
	"slices"
	"testing"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/applications"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
)

// Banco de talentos no grupo livre (spec grupo-livre, T-03, RN-13).

// CA-05.2 (API) / RN-13: grupo livre com vaga traz Tank, Suporte e Dano; cheio, ninguém.
func TestForLobby_CA05_2_Free(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "ana", "Lirien:arcebispo:178:support")
	l, err := e.lobbies.Create(t.Context(), ana.userID, lobbies.Input{
		InstanceID: "templo-do-demonio-rei", StartsAt: at(1, 20, 0), MinLevel: 160,
		CharacterID: ana.chars["Lirien"], Formation: lobbies.FormationFree, FreeSlots: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	p := e.user(t, "pedro", "Escudo:guardiao-real:190:tank", "Cura:arcebispo:180:support", "Fogo:arquimago:200:dps")
	for _, nick := range []string{"Escudo", "Cura", "Fogo"} {
		e.put(t, p, nick, every("18:00", "00:00"))
	}
	got, err := e.svc.ForLobby(t.Context(), ana.userID, l.ID)
	if err != nil || !slices.Equal(nicks(got), []string{"Fogo", "Escudo", "Cura"}) {
		t.Fatalf("com vaga = %v, %v", nicks(got), err)
	}

	bia := e.user(t, "bia", "Brasa:guardiao-real:172:tank")
	a, err := e.apps.Apply(t.Context(), bia.userID, l.ID, applications.ApplyInput{CharacterID: bia.chars["Brasa"]})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.apps.Accept(t.Context(), ana.userID, a.ID); err != nil {
		t.Fatal(err)
	}
	if got, err := e.svc.ForLobby(t.Context(), ana.userID, l.ID); err != nil || len(got) != 0 {
		t.Errorf("cheio = %v, %v", nicks(got), err)
	}
}

// CA-05.2 / RN-13: na criação de um grupo livre, as vagas livres são as do formulário menos a do
// dono, para qualquer função.
func TestCount_CA05_2_Free(t *testing.T) {
	e := setup(t)
	trio(t, e) // Brasa (Tank), Fogo (Dano), Cura (Suporte), quarta 19–23
	ana := e.user(t, "ana", "Lirien:arcebispo:178:support")
	in := talents.CountInput{
		InstanceID: "templo-do-demonio-rei", StartsAt: at(1, 20, 0), MinLevel: 160,
		CharacterID: ana.chars["Lirien"], Formation: lobbies.FormationFree, FreeSlots: 12,
	}
	if n, err := e.svc.Count(t.Context(), ana.userID, in); err != nil || n != 3 {
		t.Errorf("12 vagas = %d, %v", n, err)
	}
	in.FreeSlots = 1 // só a do dono
	if n, _ := e.svc.Count(t.Context(), ana.userID, in); n != 0 {
		t.Errorf("sem vaga livre = %d", n)
	}
}
