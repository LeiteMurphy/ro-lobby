//go:build integration

package applications

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// Formação do lobby com gente dentro (spec grupo-livre, T-01 e T-02).

func wantLobbyField(t *testing.T, err error, field, code string) {
	t.Helper()
	var ve *lobbies.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v, quer erro em %s", err, field)
	}
	for _, f := range ve.Fields {
		if f.Field == field && f.Code == code {
			return
		}
	}
	t.Errorf("erros = %+v, quer %s/%s", ve.Fields, field, code)
}

// CA-04.3 / RN-07: com um membro aceito, ou só com uma candidatura pendente, a formação
// não muda.
func TestUpdate_CA04_3_FormationLocked(t *testing.T) {
	e, owner, player, lid := basic(t)
	toFree := lobbies.UpdateInput{StartsAt: at(1, 20, 0), MinLevel: 160, Formation: lobbies.FormationFree, FreeSlots: 12}

	pending := e.apply(t, player, "Cura", lid)
	_, err := e.lobbies.Update(t.Context(), owner.id, lid, toFree)
	wantLobbyField(t, err, lobbies.FieldFormation, lobbies.CodeLocked)

	if _, err := e.svc.Accept(t.Context(), owner.id, pending.ID); err != nil {
		t.Fatal(err)
	}
	_, err = e.lobbies.Update(t.Context(), owner.id, lid, toFree)
	wantLobbyField(t, err, lobbies.FieldFormation, lobbies.CodeLocked)
	if l, _ := e.lobbies.Get(t.Context(), lid); l.Formation != lobbies.FormationRoles {
		t.Errorf("formação mudou: %+v", l)
	}
}

// freeLobby cria um grupo livre do Templo (nível 160) com o personagem do dono.
func (e *env) freeLobby(t *testing.T, owner who, nick string, total int) string {
	t.Helper()
	l, err := e.lobbies.Create(t.Context(), owner.id, lobbies.Input{
		InstanceID: "templo-do-demonio-rei", StartsAt: at(1, 20, 0), MinLevel: 160,
		CharacterID: owner.chars[nick], Formation: lobbies.FormationFree, FreeSlots: total,
	})
	if err != nil {
		t.Fatal(err)
	}
	return l.ID
}

func (e *env) join(t *testing.T, owner, w who, nick, lid string) Application {
	t.Helper()
	a, err := e.svc.Accept(t.Context(), owner.id, e.apply(t, w, nick, lid).ID)
	if err != nil {
		t.Fatalf("%s: %v", nick, err)
	}
	return a
}

// CA-02.1 / RN-03 / RN-04: três Suportes no mesmo grupo livre de 3 vagas.
func TestFree_CA02_1_AnyRole(t *testing.T) {
	e := setup(t)
	owner := e.user(t, "Dono:200:support")
	a := e.user(t, "CuraA:200:support")
	b := e.user(t, "CuraB:200:support")
	lid := e.freeLobby(t, owner, "Dono", 3)
	e.join(t, owner, a, "CuraA", lid)
	e.join(t, owner, b, "CuraB", lid)
	l, err := e.lobbies.Get(t.Context(), lid)
	if err != nil || l.Occupied.Support != 3 || lobbies.HasRoom(l, "tank") {
		t.Errorf("grupo = %+v, %v", l, err)
	}
}

// CA-02.2 / RN-04: com o grupo cheio, candidatar e aceitar dão group_full.
func TestFree_CA02_2_Full(t *testing.T) {
	e := setup(t)
	owner := e.user(t, "Dono:200:support")
	a := e.user(t, "Escudo:200:tank")
	b := e.user(t, "Fogo:200:dps")
	c := e.user(t, "Cura:200:support")
	lid := e.freeLobby(t, owner, "Dono", 2)
	pending := e.apply(t, b, "Fogo", lid)
	e.join(t, owner, a, "Escudo", lid)
	_, err := e.svc.Apply(t.Context(), c.id, lid, ApplyInput{CharacterID: c.chars["Cura"]})
	wantRule(t, err, CodeGroupFull)
	_, err = e.svc.Accept(t.Context(), owner.id, pending.ID)
	wantRule(t, err, CodeGroupFull)
}

// CA-02.3 / RN-04: aceites ao mesmo tempo na última vaga; só um passa.
func TestFree_CA02_3_Concurrent(t *testing.T) {
	e := setup(t)
	owner := e.user(t, "Dono:200:support")
	lid := e.freeLobby(t, owner, "Dono", 2)
	var apps []Application
	for i := range 6 {
		w := e.user(t, fmt.Sprintf("P%d:200:%s", i, []string{"tank", "support", "dps"}[i%3]))
		apps = append(apps, e.apply(t, w, fmt.Sprintf("P%d", i), lid))
	}
	var wg sync.WaitGroup
	errs := make([]error, len(apps))
	for i, a := range apps {
		wg.Go(func() { _, errs[i] = e.svc.Accept(t.Context(), owner.id, a.ID) })
	}
	wg.Wait()
	ok := 0
	for _, err := range errs {
		if err == nil {
			ok++
			continue
		}
		wantRule(t, err, CodeGroupFull)
	}
	l, _ := e.lobbies.Get(t.Context(), lid)
	if ok != 1 || l.Occupied.Tank+l.Occupied.Support+l.Occupied.Dps != 2 {
		t.Errorf("aceites %d, ocupantes %+v", ok, l.Occupied)
	}
}

// CA-02.4 / RN-05: com o grupo cheio, o membro troca de Dano para Tank.
func TestFree_CA02_4_SwapWithoutRoleSlot(t *testing.T) {
	e := setup(t)
	owner := e.user(t, "Dono:200:support")
	p := e.user(t, "Fogo:200:dps", "Escudo:200:tank")
	lid := e.freeLobby(t, owner, "Dono", 2)
	member := e.join(t, owner, p, "Fogo", lid)
	r, err := e.svc.RequestSwap(t.Context(), p.id, member.ID, SwapInput{CharacterID: p.chars["Escudo"], Reason: "o grupo precisa de tank"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.AcceptSwap(t.Context(), owner.id, r.ID); err != nil {
		t.Fatalf("troca: %v", err)
	}
	l, _ := e.lobbies.Get(t.Context(), lid)
	if l.Occupied.Tank != 1 || l.Occupied.Dps != 0 || l.Occupied.Support != 1 {
		t.Errorf("ocupantes = %+v", l.Occupied)
	}
}

// CA-04.1 / RN-06: com 5 ocupantes, o grupo livre não fica com 4 vagas.
func TestFree_CA04_1_BelowOccupied(t *testing.T) {
	e := setup(t)
	owner := e.user(t, "Dono:200:support")
	lid := e.freeLobby(t, owner, "Dono", 12)
	for i := range 4 {
		w := e.user(t, fmt.Sprintf("M%d:200:dps", i))
		e.join(t, owner, w, fmt.Sprintf("M%d", i), lid)
	}
	upd := lobbies.UpdateInput{StartsAt: at(1, 20, 0), MinLevel: 160, FreeSlots: 4}
	_, err := e.lobbies.Update(t.Context(), owner.id, lid, upd)
	wantLobbyField(t, err, lobbies.FieldFreeSlots, lobbies.CodeBelowOccupied)
	upd.FreeSlots = 5
	if _, err := e.lobbies.Update(t.Context(), owner.id, lid, upd); err != nil {
		t.Errorf("5 vagas: %v", err)
	}
}
