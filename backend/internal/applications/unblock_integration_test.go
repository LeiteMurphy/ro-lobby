//go:build integration

package applications

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

// Desbloqueio pelo dono (spec candidatura-lobby, revisão de 2026-10-09).

// CA-06.9 / RN-15: o dono desbloqueia pelo personagem; o jogador volta a se candidatar, e
// a candidatura removida fica no histórico com a justificativa.
func TestUnblock_CA06_9(t *testing.T) {
	e, owner, player, lid, a := member(t)
	if _, err := e.svc.Remove(t.Context(), owner.id, a.ID, "mudamos o horário da run", true); err != nil {
		t.Fatal(err)
	}
	if err := e.svc.Unblock(t.Context(), player.id, lid, player.chars["Fogo"]); !errors.Is(err, ErrNotFound) {
		t.Errorf("quem não é o dono: %v", err)
	}
	// Qualquer personagem da pessoa serve, porque o bloqueio é do usuário.
	if err := e.svc.Unblock(t.Context(), owner.id, lid, player.chars["Fogo"]); err != nil {
		t.Fatal(err)
	}
	var aid pgtype.UUID
	_ = aid.Scan(a.ID)
	if old, err := e.q.GetApplication(t.Context(), aid); err != nil || old.Status != StatusRemoved ||
		old.Reason.String != "mudamos o horário da run" {
		t.Errorf("histórico = %+v, %v", old, err)
	}
	again, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: player.chars["Cura"]})
	if err != nil || again.Status != StatusPending {
		t.Errorf("nova candidatura: %+v, %v", again, err)
	}
	// Sem bloqueio, desbloquear de novo não muda nada.
	if err := e.svc.Unblock(t.Context(), owner.id, lid, player.chars["Cura"]); err != nil {
		t.Errorf("sem bloqueio: %v", err)
	}
}

// CA-06.9 / RN-15: em lobby iniciado ou cancelado, desbloquear é recusado.
func TestUnblock_CA06_9_NotOpen(t *testing.T) {
	e, owner, player, lid, a := member(t)
	if _, err := e.svc.Remove(t.Context(), owner.id, a.ID, "mudamos o horário da run", true); err != nil {
		t.Fatal(err)
	}
	e.now = at(1, 21, 0) // o lobby (amanhã, 20:00) já começou
	wantRule(t, e.svc.Unblock(t.Context(), owner.id, lid, player.chars["Cura"]), CodeNotOpen)
	e.now = start
	if _, err := e.lobbies.Cancel(t.Context(), owner.id, lid, "Metade do grupo não pode"); err != nil {
		t.Fatal(err)
	}
	err := e.svc.Unblock(t.Context(), owner.id, lid, player.chars["Cura"])
	wantRule(t, err, CodeNotOpen)
	if err := e.svc.Unblock(t.Context(), owner.id, "nao-e-uuid", player.chars["Cura"]); !errors.Is(err, ErrNotFound) {
		t.Errorf("lobby inválido: %v", err)
	}
}
