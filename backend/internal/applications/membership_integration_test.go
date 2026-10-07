//go:build integration

package applications

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// Testes da saída do grupo, da remoção e do bloqueio (spec candidatura-lobby, Parte 2,
// T-10).

// member: o basic com o jogador aceito como Suporte (Cura).
func member(t *testing.T) (*env, who, who, string, Application) {
	t.Helper()
	e, owner, player, lid := basic(t)
	a := e.apply(t, player, "Cura", lid)
	if _, err := e.svc.Accept(t.Context(), owner.id, a.ID); err != nil {
		t.Fatal(err)
	}
	return e, owner, player, lid, a
}

func uuid(s string) pgtype.UUID {
	var id pgtype.UUID
	_ = id.Scan(s)
	return id
}

// pendingSwap abre um pedido de troca pendente direto no banco, para os testes que só
// precisam do pedido existir.
func (e *env) pendingSwap(t *testing.T, a Application, toCharacterID, toRole string) db.SwapRequest {
	t.Helper()
	r, err := e.q.CreateSwapRequest(t.Context(), db.CreateSwapRequestParams{
		ApplicationID: uuid(a.ID), FromCharacterID: uuid(a.CharacterID), ToCharacterID: uuid(toCharacterID),
		ToRole: toRole, Reason: "ninguém apareceu de tank", Now: e.now,
	})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func (e *env) swapStatus(t *testing.T, id pgtype.UUID) string {
	t.Helper()
	r, err := e.q.GetSwapRequest(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	return r.Status
}

// CA-05.1: o membro sai antes do início; a candidatura vira "saiu" e a vaga volta a ficar
// livre, com o evento no histórico.
func TestLeave_CA05_1_BeforeStart(t *testing.T) {
	e, _, player, lid, a := member(t)
	left, err := e.svc.Leave(t.Context(), player.id, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if left.Status != StatusLeft || left.DecidedAt.IsZero() {
		t.Errorf("saída = %+v", left)
	}
	if occ, _ := e.occupied(t, lid); occ != (lobbies.Slots{Dps: 1}) {
		t.Errorf("ocupantes = %+v", occ)
	}
	events, _ := e.q.ListApplicationEvents(t.Context(), uuid(a.ID))
	last := events[len(events)-1]
	if last.FromStatus.String != StatusAccepted || last.ToStatus != StatusLeft || last.ActorID.String() != player.id {
		t.Errorf("evento = %+v", last)
	}
}

// CA-05.2: depois do início (ou do cancelamento), não sai mais.
func TestLeave_CA05_2_NotOpen(t *testing.T) {
	e, owner, player, lid, a := member(t)
	e.now = at(1, 20, 0)
	_, err := e.svc.Leave(t.Context(), player.id, a.ID)
	wantRule(t, err, CodeNotOpen)

	e.now = start
	if _, err := e.lobbies.Cancel(t.Context(), owner.id, lid, "imprevisto no trabalho"); err != nil {
		t.Fatal(err)
	}
	_, err = e.svc.Leave(t.Context(), player.id, a.ID)
	wantRule(t, err, CodeNotOpen)
	if got := e.status(t, a.ID); got != StatusAccepted {
		t.Errorf("estado = %s", got)
	}
}

// RN-14: só o próprio membro sai, e só de candidatura aceita.
func TestLeave_RN14_OnlyOwnAcceptedApplication(t *testing.T) {
	e, owner, player, lid, a := member(t)
	_, err := e.svc.Leave(t.Context(), owner.id, a.ID)
	wantRule(t, err, CodeNotYours)

	other := e.user(t, "Outro:200:dps")
	pending := e.apply(t, other, "Outro", lid)
	_, err = e.svc.Leave(t.Context(), other.id, pending.ID)
	wantRule(t, err, CodeNotMember)

	if _, err := e.svc.Leave(t.Context(), player.id, a.ID); err != nil {
		t.Fatal(err)
	}
	_, err = e.svc.Leave(t.Context(), player.id, a.ID)
	wantRule(t, err, CodeNotMember)

	_, err = e.svc.Leave(t.Context(), player.id, "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("inexistente: %v", err)
	}
}

// CA-05.3: a saída cancela o pedido de troca pendente, com evento.
func TestLeave_CA05_3_CancelsSwap(t *testing.T) {
	e, _, player, _, a := member(t)
	r := e.pendingSwap(t, a, player.chars["Escudo"], "tank")
	if _, err := e.svc.Leave(t.Context(), player.id, a.ID); err != nil {
		t.Fatal(err)
	}
	if got := e.swapStatus(t, r.ID); got != StatusCancelled {
		t.Errorf("pedido = %s", got)
	}
	events, _ := e.q.ListSwapRequestEvents(t.Context(), r.ID)
	if len(events) != 1 || events[0].ToStatus != StatusCancelled || events[0].ActorID.String() != player.id {
		t.Errorf("eventos = %+v", events)
	}
}

// CA-05.4: quem saiu se candidata de novo ao mesmo lobby (RN-37).
func TestLeave_CA05_4_ApplyAgain(t *testing.T) {
	e, _, player, lid, a := member(t)
	if _, err := e.svc.Leave(t.Context(), player.id, a.ID); err != nil {
		t.Fatal(err)
	}
	again, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: player.chars["Escudo"]})
	if err != nil || again.Status != StatusPending || again.ID == a.ID {
		t.Errorf("nova candidatura: %+v, %v", again, err)
	}
}

// CA-06.1: o dono remove com justificativa; a vaga volta a ficar livre e a justificativa
// fica registrada.
func TestRemove_CA06_1_WithReason(t *testing.T) {
	e, owner, _, lid, a := member(t)
	removed, err := e.svc.Remove(t.Context(), owner.id, a.ID, "  mudamos o horário da run  ", false)
	if err != nil {
		t.Fatal(err)
	}
	if removed.Status != StatusRemoved || removed.Reason != "mudamos o horário da run" || removed.Blocked {
		t.Errorf("remoção = %+v", removed)
	}
	if occ, _ := e.occupied(t, lid); occ != (lobbies.Slots{Dps: 1}) {
		t.Errorf("ocupantes = %+v", occ)
	}
	events, _ := e.q.ListApplicationEvents(t.Context(), uuid(a.ID))
	last := events[len(events)-1]
	if last.ToStatus != StatusRemoved || last.ActorID.String() != owner.id || last.Reason.String != "mudamos o horário da run" {
		t.Errorf("evento = %+v", last)
	}
}

// CA-06.2: sem justificativa (ou fora de 10 a 250), a remoção é rejeitada e o membro
// continua aceito.
func TestRemove_CA06_2_ReasonRequired(t *testing.T) {
	e, owner, _, _, a := member(t)
	for reason, code := range map[string]string{
		"   ":                    lobbies.CodeRequired,
		"curta":                  lobbies.CodeTooShort,
		strings.Repeat("é", 251): lobbies.CodeTooLong,
		"\t\n  \t ":              lobbies.CodeRequired,
	} {
		_, err := e.svc.Remove(t.Context(), owner.id, a.ID, reason, false)
		wantField(t, err, lobbies.FieldReason, code)
	}
	if got := e.status(t, a.ID); got != StatusAccepted {
		t.Errorf("estado = %s", got)
	}
}

// CA-06.3: quem não é dono não remove.
func TestRemove_CA06_3_NotOwner(t *testing.T) {
	e, _, player, lid, a := member(t)
	other := e.user(t, "Outro:200:dps")
	e.apply(t, other, "Outro", lid)
	for _, w := range []who{player, other} {
		_, err := e.svc.Remove(t.Context(), w.id, a.ID, "mudamos o horário da run", false)
		wantRule(t, err, CodeNotOwner)
	}
	if got := e.status(t, a.ID); got != StatusAccepted {
		t.Errorf("estado = %s", got)
	}
}

// CA-06.4: depois do início, o dono não remove; candidatura pendente não é "membro".
func TestRemove_CA06_4_NotOpen(t *testing.T) {
	e, owner, _, lid, a := member(t)
	other := e.user(t, "Outro:200:dps")
	pending := e.apply(t, other, "Outro", lid)
	_, err := e.svc.Remove(t.Context(), owner.id, pending.ID, "mudamos o horário da run", false)
	wantRule(t, err, CodeNotMember)

	e.now = at(1, 20, 0)
	_, err = e.svc.Remove(t.Context(), owner.id, a.ID, "mudamos o horário da run", false)
	wantRule(t, err, CodeNotOpen)
}

// CA-06.5: removido sem bloqueio se candidata de novo ao mesmo lobby.
func TestRemove_CA06_5_ApplyAgainWithoutBlock(t *testing.T) {
	e, owner, player, lid, a := member(t)
	if _, err := e.svc.Remove(t.Context(), owner.id, a.ID, "mudamos o horário da run", false); err != nil {
		t.Fatal(err)
	}
	again, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: player.chars["Cura"]})
	if err != nil || again.Status != StatusPending {
		t.Errorf("nova candidatura: %+v, %v", again, err)
	}
}

// CA-06.6 e CA-06.7: com bloqueio, não se candidata de novo ao mesmo lobby com nenhum
// personagem, mas se candidata a outro lobby do mesmo dono.
func TestRemove_CA06_6_CA06_7_Blocked(t *testing.T) {
	e, owner, player, lid, a := member(t)
	removed, err := e.svc.Remove(t.Context(), owner.id, a.ID, "mudamos o horário da run", true)
	if err != nil || !removed.Blocked {
		t.Fatalf("remoção com bloqueio: %+v, %v", removed, err)
	}
	for _, nick := range []string{"Cura", "Escudo", "Fogo"} {
		_, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: player.chars[nick]})
		wantRule(t, err, CodeBlocked)
	}
	other := e.lobby(t, owner, "Dono", at(3, 20, 0), std, 160)
	b, err := e.svc.Apply(t.Context(), player.id, other, ApplyInput{CharacterID: player.chars["Cura"]})
	if err != nil || b.Status != StatusPending {
		t.Errorf("outro lobby: %+v, %v", b, err)
	}
}

// CA-08.10: a remoção cancela o pedido de troca pendente.
func TestRemove_CA08_10_CancelsSwap(t *testing.T) {
	e, owner, player, _, a := member(t)
	r := e.pendingSwap(t, a, player.chars["Escudo"], "tank")
	if _, err := e.svc.Remove(t.Context(), owner.id, a.ID, "mudamos o horário da run", false); err != nil {
		t.Fatal(err)
	}
	if got := e.swapStatus(t, r.ID); got != StatusCancelled {
		t.Errorf("pedido = %s", got)
	}
}

// D-10: a saída e a remoção do mesmo membro ao mesmo tempo; só uma passa e a outra vê que
// ele não é mais membro.
func TestLeaveRemove_D10_Concurrent(t *testing.T) {
	for range 5 {
		e, owner, player, _, a := member(t)
		var wg sync.WaitGroup
		var leaveErr, removeErr error
		wg.Go(func() { _, leaveErr = e.svc.Leave(t.Context(), player.id, a.ID) })
		wg.Go(func() {
			_, removeErr = e.svc.Remove(t.Context(), owner.id, a.ID, "mudamos o horário da run", false)
		})
		wg.Wait()
		switch {
		case leaveErr == nil && removeErr != nil:
			wantRule(t, removeErr, CodeNotMember)
		case removeErr == nil && leaveErr != nil:
			wantRule(t, leaveErr, CodeNotMember)
		default:
			t.Errorf("sair: %v, remover: %v", leaveErr, removeErr)
		}
	}
}
