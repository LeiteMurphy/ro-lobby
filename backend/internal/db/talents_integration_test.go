//go:build integration

package db

import (
	"testing"
)

// Testes da tabela de disponibilidade (spec banco-de-talentos, T-01). Só o banco: as
// regras e as mensagens ficam no serviço (internal/talents).

func availability(c Character) UpsertAvailabilityParams {
	return UpsertAvailabilityParams{
		CharacterID: c.ID, Days: 0b0111110, StartMinute: 19 * 60, EndMinute: 23 * 60,
		AnyInstance: false, InstanceIds: []string{"templo-do-demonio-rei"}, Now: t0,
	}
}

// RN-02 a RN-04: o banco recusa dias, faixa e instâncias fora das regras.
func TestAvailability_RN02_RN03_RN04_ChecksRejectInvalidValues(t *testing.T) {
	q, _ := setup(t)
	_, c := lobbyOwner(t, q)
	cases := map[string]func(p *UpsertAvailabilityParams){
		"sem dia":                      func(p *UpsertAvailabilityParams) { p.Days = 0 },
		"dia além de sábado":           func(p *UpsertAvailabilityParams) { p.Days = 128 },
		"início fora dos 30 min":       func(p *UpsertAvailabilityParams) { p.StartMinute = 19*60 + 15 },
		"fim depois de 23:30":          func(p *UpsertAvailabilityParams) { p.EndMinute = 1440 },
		"início igual ao fim":          func(p *UpsertAvailabilityParams) { p.EndMinute = p.StartMinute },
		"sem instância e sem Qualquer": func(p *UpsertAvailabilityParams) { p.InstanceIds = []string{} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := availability(c)
			mutate(&p)
			if _, err := q.UpsertAvailability(t.Context(), p); pgCode(err) != "23514" {
				t.Errorf("err = %v, quer violação de CHECK (23514)", err)
			}
		})
	}
}

// RN-01 / RN-03 / RN-04: faixa que vira a meia-noite e "Qualquer" sem lista são aceitas;
// o upsert religa e troca os dados; desligar guarda os dados.
func TestAvailability_RN01_UpsertAndDisable(t *testing.T) {
	q, _ := setup(t)
	u, c := lobbyOwner(t, q)
	p := availability(c)
	p.StartMinute, p.EndMinute = 22*60, 2*60
	p.AnyInstance, p.InstanceIds = true, []string{}
	if _, err := q.UpsertAvailability(t.Context(), p); err != nil {
		t.Fatal(err)
	}
	off, err := q.DisableAvailability(t.Context(), DisableAvailabilityParams{CharacterID: c.ID, Now: t0})
	if err != nil || off.Enabled || off.StartMinute != 22*60 || !off.AnyInstance {
		t.Fatalf("desligada = %+v, %v", off, err)
	}
	on, err := q.UpsertAvailability(t.Context(), availability(c))
	if err != nil || !on.Enabled || on.AnyInstance || on.InstanceIds[0] != "templo-do-demonio-rei" {
		t.Fatalf("religada = %+v, %v", on, err)
	}
	mine, err := q.ListAvailabilityByUser(t.Context(), u.ID)
	if err != nil || len(mine) != 1 {
		t.Fatalf("do Usuário = %+v, %v", mine, err)
	}
}

// RN-05 / D-02: excluir o personagem tira a disponibilidade junto.
func TestAvailability_RN05_CascadeOnCharacterDelete(t *testing.T) {
	q, _ := setup(t)
	u, c := lobbyOwner(t, q)
	if _, err := q.UpsertAvailability(t.Context(), availability(c)); err != nil {
		t.Fatal(err)
	}
	if _, err := q.DeleteCharacter(t.Context(), DeleteCharacterParams{ID: c.ID, UserID: u.ID}); err != nil {
		t.Fatal(err)
	}
	if mine, _ := q.ListAvailabilityByUser(t.Context(), u.ID); len(mine) != 0 {
		t.Errorf("disponibilidade sobrou: %+v", mine)
	}
}
