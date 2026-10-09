//go:build integration

package talents_test

import (
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
)

// Bordas da RN-04 e da RN-03 (spec banco-de-talentos, ciclo 2 da validação).

// rawAvailability grava direto no banco, sem a validação do catálogo do serviço, para
// simular uma instância que saiu do catálogo depois de escolhida.
func rawAvailability(t *testing.T, e *env, characterID string, ids []string) {
	t.Helper()
	var cid pgtype.UUID
	if err := cid.Scan(characterID); err != nil {
		t.Fatal(err)
	}
	if _, err := e.q.UpsertAvailability(t.Context(), db.UpsertAvailabilityParams{
		CharacterID: cid, Days: 127, StartMinute: 18 * 60, EndMinute: 0,
		InstanceIds: ids, Now: now,
	}); err != nil {
		t.Fatal(err)
	}
}

// Borda da RN-04: a instância que saiu do catálogo some da lista do personagem; quem não
// tem mais nenhuma sai da afinidade, mas continua no banco.
func TestRemovedInstance_RN04(t *testing.T) {
	e := setup(t)
	ana, lid := glast(t, e)
	p := e.user(t, "pedro", "Misto:arquimago:200:dps", "Orfao:arquimago:199:dps")
	rawAvailability(t, e, p.chars["Misto"], []string{"instancia-removida", "templo-do-demonio-rei"})
	rawAvailability(t, e, p.chars["Orfao"], []string{"instancia-removida"})

	all, err := e.svc.Catalog(t.Context(), talents.CatalogFilter{})
	if err != nil || !slices.Equal(nicks(all), []string{"Misto", "Orfao"}) {
		t.Fatalf("catálogo = %v, %v", nicks(all), err)
	}
	if len(all[0].Instances) != 1 || all[0].Instances[0].ID != "templo-do-demonio-rei" {
		t.Errorf("Misto = %+v", all[0].Instances)
	}
	if len(all[1].Instances) != 0 || all[1].AnyInstance {
		t.Errorf("Órfão = %+v", all[1])
	}
	got, err := e.svc.ForLobby(t.Context(), ana.userID, lid)
	if err != nil || !slices.Equal(nicks(got), []string{"Misto"}) {
		t.Errorf("afinidade = %v, %v", nicks(got), err)
	}
	list, err := e.chars.List(t.Context(), p.userID)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range list {
		if c.Availability == nil || slices.Contains(c.Availability.InstanceIDs, "instancia-removida") {
			t.Errorf("%s no perfil = %+v", c.Nick, c.Availability)
		}
	}
}

// RN-03: sábado 23:00–01:00 vira para domingo; um lobby no domingo à 00:30 tem afinidade
// e um no domingo às 23:30 não.
func TestMidnight_RN03_SaturdayToSunday(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "ana", "Lirien:arcebispo:178:support")
	p := e.user(t, "pedro", "Sabado:arquimago:200:dps")
	e.put(t, p, "Sabado", talents.AvailabilityInput{Days: []int{6}, Start: "23:00", End: "01:00", AnyInstance: true})
	// "Agora" é terça; domingo é daqui a 5 dias.
	for _, c := range []struct {
		hour, minute int
		want         []string
	}{{0, 30, []string{"Sabado"}}, {23, 30, []string{}}} {
		l, err := e.lobbies.Create(t.Context(), ana.userID, lobbies.Input{
			InstanceID: "templo-do-demonio-rei", StartsAt: at(5, c.hour, c.minute),
			Slots: lobbies.Slots{Support: 1, Dps: 3}, MinLevel: 160, CharacterID: ana.chars["Lirien"],
		})
		if err != nil {
			t.Fatal(err)
		}
		got, err := e.svc.ForLobby(t.Context(), ana.userID, l.ID)
		if err != nil || !slices.Equal(nicks(got), c.want) {
			t.Errorf("domingo %02d:%02d = %v, %v", c.hour, c.minute, nicks(got), err)
		}
	}
}
