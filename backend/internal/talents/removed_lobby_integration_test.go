//go:build integration

package talents_test

import (
	"slices"
	"testing"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/applications"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
)

// Selo de removido no painel do dono (spec banco-de-talentos, RN-13).

// removedFromGlast põe Brasa (Tank) e Cinza (Dano), da mesma pessoa, no banco; Brasa entra
// no lobby do CA-02.1 e é removido, com ou sem bloqueio.
func removedFromGlast(t *testing.T, e *env, block bool) (who, string, who) {
	t.Helper()
	ana, lid := glast(t, e)
	bia := e.user(t, "bia", "Brasa:guardiao-real:172:tank", "Cinza:arquimago:180:dps")
	e.put(t, bia, "Brasa", every("18:00", "00:00"))
	e.put(t, bia, "Cinza", every("18:00", "00:00"))
	a, err := e.apps.Apply(t.Context(), bia.userID, lid, applications.ApplyInput{CharacterID: bia.chars["Brasa"]})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.apps.Accept(t.Context(), ana.userID, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := e.apps.Remove(t.Context(), ana.userID, a.ID, "mudamos o horário da run", block); err != nil {
		t.Fatal(err)
	}
	return ana, lid, bia
}

func flags(list []talents.Talent) map[string][2]bool {
	out := map[string][2]bool{}
	for _, x := range list {
		out[x.Nick] = [2]bool{x.Removed, x.Blocked}
	}
	return out
}

// CA-02.5 / RN-13: removido sem bloqueio aparece, com o selo de removido.
func TestForLobby_CA02_5_RemovedWithoutBlock(t *testing.T) {
	e := setup(t)
	ana, lid, _ := removedFromGlast(t, e, false)
	got, err := e.svc.ForLobby(t.Context(), ana.userID, lid)
	if err != nil || !slices.Equal(nicks(got), []string{"Cinza", "Brasa"}) {
		t.Fatalf("lista = %v, %v", nicks(got), err)
	}
	if f := flags(got); f["Brasa"] != [2]bool{true, false} || f["Cinza"] != [2]bool{true, false} {
		t.Errorf("selos = %v", f)
	}
}

// CA-02.6 / RN-13: com bloqueio, os dois personagens da pessoa aparecem bloqueados; depois
// do desbloqueio, só removidos; outro lobby não leva o selo.
func TestForLobby_CA02_6_RemovedBlockedAndUnblock(t *testing.T) {
	e := setup(t)
	ana, lid, bia := removedFromGlast(t, e, true)
	got, err := e.svc.ForLobby(t.Context(), ana.userID, lid)
	if err != nil {
		t.Fatal(err)
	}
	if f := flags(got); f["Brasa"] != [2]bool{true, true} || f["Cinza"] != [2]bool{true, true} {
		t.Errorf("bloqueados = %v", f)
	}
	if err := e.apps.Unblock(t.Context(), ana.userID, lid, bia.chars["Cinza"]); err != nil {
		t.Fatal(err)
	}
	got, _ = e.svc.ForLobby(t.Context(), ana.userID, lid)
	if f := flags(got); f["Brasa"] != [2]bool{true, false} || f["Cinza"] != [2]bool{true, false} {
		t.Errorf("desbloqueados = %v", f)
	}
	if _, err := e.apps.Apply(t.Context(), bia.userID, lid, applications.ApplyInput{CharacterID: bia.chars["Cinza"]}); err != nil {
		t.Errorf("candidatura depois do desbloqueio: %v", err)
	}
}
