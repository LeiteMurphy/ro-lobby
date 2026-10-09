//go:build integration

package lobbies

import "testing"

// Grupo livre no serviço de lobbies (spec grupo-livre, T-01).

func free(characterID string, slots int) Input {
	in := temple(characterID, at(1, 20, 0))
	in.Formation, in.Slots, in.FreeSlots = FormationFree, Slots{}, slots
	return in
}

// CA-01.1 / RN-01 / RN-02: grupo livre de 12, com o dono numa das vagas.
func TestCreate_CA01_1_Free(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	l, err := e.svc.Create(t.Context(), ana.userID, free(ana.chars["Lirien"], 12))
	if err != nil {
		t.Fatal(err)
	}
	if l.Formation != FormationFree || l.FreeSlots != 12 || l.Slots != (Slots{}) || l.Occupied.total() != 1 {
		t.Errorf("criado = %+v", l)
	}
	if !HasRoom(l, "tank") {
		t.Error("sem vaga logo após criar")
	}
	// RNF-01: sem formação, continua por função.
	r, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(2, 20, 0)))
	if err != nil || r.Formation != FormationRoles || r.FreeSlots != 0 {
		t.Errorf("por função = %+v, %v", r, err)
	}
}

// CA-01.2 / RN-02: 1 ou 13 vagas, ou vagas por função no grupo livre, são recusados.
func TestCreate_CA01_2_FreeLimits(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	for _, n := range []int{1, 13} {
		_, err := e.svc.Create(t.Context(), ana.userID, free(ana.chars["Lirien"], n))
		wantField(t, err, FieldFreeSlots, CodeInvalid)
	}
	mixed := free(ana.chars["Lirien"], 6)
	mixed.Slots = Slots{Tank: 1}
	_, err := e.svc.Create(t.Context(), ana.userID, mixed)
	wantField(t, err, FieldFreeSlots, CodeInvalid)
}

// CA-04.2 / RN-07: sozinho, o dono troca de livre para por função (com vaga na função
// dele) e de volta; por função sem vaga na função do dono é recusado.
func TestUpdate_CA04_2_SwitchFormationAlone(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	l, err := e.svc.Create(t.Context(), ana.userID, free(ana.chars["Lirien"], 12))
	if err != nil {
		t.Fatal(err)
	}
	upd := UpdateInput{StartsAt: at(1, 20, 0), MinLevel: 160, Formation: FormationRoles, Slots: Slots{Tank: 1, Dps: 3}}
	got, err := e.svc.Update(t.Context(), ana.userID, l.ID, upd)
	if err == nil {
		t.Fatalf("sem vaga de Suporte para o dono: %+v", got)
	}
	upd.Slots = Slots{Tank: 1, Support: 2, Dps: 3}
	got, err = e.svc.Update(t.Context(), ana.userID, l.ID, upd)
	if err != nil || got.Formation != FormationRoles || got.FreeSlots != 0 || got.Occupied.Support != 1 {
		t.Fatalf("por função = %+v, %v", got, err)
	}
	got, err = e.svc.Update(t.Context(), ana.userID, l.ID, UpdateInput{
		StartsAt: at(1, 20, 0), MinLevel: 160, Formation: FormationFree, FreeSlots: 5,
	})
	if err != nil || got.Formation != FormationFree || got.FreeSlots != 5 || got.Slots != (Slots{}) {
		t.Errorf("de volta ao livre = %+v, %v", got, err)
	}
	// Sem formação, mantém a atual e valida as vagas dela.
	_, err = e.svc.Update(t.Context(), ana.userID, l.ID, UpdateInput{StartsAt: at(1, 20, 0), MinLevel: 160, FreeSlots: 1})
	wantField(t, err, FieldFreeSlots, CodeInvalid)
}
