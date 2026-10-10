//go:build integration

package talents_test

import (
	"slices"
	"testing"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
)

// Banco de talentos no lobby sem instância (spec lobby-sem-instancia, RN-07).

// CA-04.1 / RN-07: o lobby sem instância traz quem só marcou outra instância; as outras
// condições continuam valendo.
func TestForLobby_CA04_1_AnyInstance(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "ana", "Lirien:arcebispo:178:support")
	l, err := e.lobbies.Create(t.Context(), ana.userID, lobbies.Input{
		AnyInstance: true, Title: "Caça ao MVP", StartsAt: at(1, 20, 0), MinLevel: 1,
		Slots: lobbies.Slots{Tank: 1, Support: 1, Dps: 3}, CharacterID: ana.chars["Lirien"],
	})
	if err != nil {
		t.Fatal(err)
	}
	p := e.user(t, "pedro", "Sonho:arquimago:200:dps", "Quinta:arquimago:200:dps")
	e.put(t, p, "Sonho", talents.AvailabilityInput{Days: []int{wednesday}, Start: "19:00", End: "23:00", InstanceIDs: []string{"sonho-sombrio"}})
	e.put(t, p, "Quinta", talents.AvailabilityInput{Days: []int{4}, Start: "19:00", End: "23:00", InstanceIDs: []string{"sonho-sombrio"}})
	got, err := e.svc.ForLobby(t.Context(), ana.userID, l.ID)
	if err != nil || !slices.Equal(nicks(got), []string{"Sonho"}) {
		t.Errorf("sem instância = %v, %v", nicks(got), err)
	}
	n, err := e.svc.Count(t.Context(), ana.userID, talents.CountInput{
		AnyInstance: true, StartsAt: at(1, 20, 0), MinLevel: 1, Tank: 1, Support: 2, Dps: 3,
		CharacterID: ana.chars["Lirien"],
	})
	if err != nil || n != 1 {
		t.Errorf("contagem = %d, %v", n, err)
	}
}
