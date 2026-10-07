//go:build integration

package applications

import (
	"errors"
	"sync"
	"testing"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// Testes da troca do dono e dos pedidos de troca do membro (spec candidatura-lobby,
// Parte 2, T-11).

// swapBasic: dono com quatro personagens (Dano, Tank, Suporte e um Dano de nível 170), um
// lobby amanhã às 20:00 (1 Tank, 2 Suportes, 3 Danos, nível mínimo 160) e o jogador aceito
// como Dano (Fogo).
func swapBasic(t *testing.T) (*env, who, who, string, Application) {
	t.Helper()
	e := setup(t)
	owner := e.user(t, "Dono:200:dps", "DonoTank:200:tank", "DonoCura:200:support", "DonoBaixo:199:dps")
	player := e.user(t, "Escudo:200:tank", "Cura:200:support", "Fogo:200:dps", "Novato:199:tank")
	lid := e.lobby(t, owner, "Dono", at(1, 20, 0), std, 160)
	a := e.apply(t, player, "Fogo", lid)
	if _, err := e.svc.Accept(t.Context(), owner.id, a.ID); err != nil {
		t.Fatal(err)
	}
	return e, owner, player, lid, a
}

// fillTank ocupa a vaga de Tank do lobby com outro jogador aceito.
func (e *env) fillTank(t *testing.T, owner who, lid string) {
	t.Helper()
	w := e.user(t, "Muralha:200:tank")
	a := e.apply(t, w, "Muralha", lid)
	if _, err := e.svc.Accept(t.Context(), owner.id, a.ID); err != nil {
		t.Fatal(err)
	}
}

// busyElsewhere põe o personagem como membro aceito de outro lobby às 21:00 do mesmo dia,
// dentro da janela de 2 h (RN-11).
func (e *env) busyElsewhere(t *testing.T, w who, nick string) {
	t.Helper()
	host := e.user(t, "Anfitriao:200:dps")
	other := e.lobby(t, host, "Anfitriao", at(1, 21, 0), std, 160)
	a := e.apply(t, w, nick, other)
	if _, err := e.svc.Accept(t.Context(), host.id, a.ID); err != nil {
		t.Fatal(err)
	}
}

func (e *env) ownerOf(t *testing.T, lid string) (string, lobbies.Slots) {
	t.Helper()
	l, err := e.lobbies.Get(t.Context(), lid)
	if err != nil {
		t.Fatal(err)
	}
	return l.Owner.CharacterID, l.Occupied
}

func (e *env) request(t *testing.T, w who, a Application, nick string) SwapRequest {
	t.Helper()
	r, err := e.svc.RequestSwap(t.Context(), w.id, a.ID, SwapInput{CharacterID: w.chars[nick], Reason: "ninguém apareceu de tank"})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func (e *env) application(t *testing.T, id string) Application {
	t.Helper()
	a, err := e.q.GetApplication(t.Context(), uuid(id))
	if err != nil {
		t.Fatal(err)
	}
	return toApplication(a)
}

// CA-07.1: o dono troca de Dano para Tank com vaga; a vaga de Dano fica livre.
func TestOwnerSwap_CA07_1_RoleWithSlot(t *testing.T) {
	e, owner, _, lid, _ := swapBasic(t)
	if err := e.svc.SwapOwnerCharacter(t.Context(), owner.id, lid, owner.chars["DonoTank"]); err != nil {
		t.Fatal(err)
	}
	char, occ := e.ownerOf(t, lid)
	if char != owner.chars["DonoTank"] || occ != (lobbies.Slots{Tank: 1, Dps: 1}) {
		t.Errorf("personagem %s, ocupantes %+v", char, occ)
	}
}

// CA-07.2: dentro da mesma função lotada, a troca passa (a vaga que sai conta como livre).
func TestOwnerSwap_CA07_2_SameRoleFull(t *testing.T) {
	e := setup(t)
	owner := e.user(t, "DonoCura:200:support", "DonoCura2:200:support")
	lid := e.lobby(t, owner, "DonoCura", at(1, 20, 0), lobbies.Slots{Support: 1, Dps: 1}, 160)
	if err := e.svc.SwapOwnerCharacter(t.Context(), owner.id, lid, owner.chars["DonoCura2"]); err != nil {
		t.Fatal(err)
	}
	if char, occ := e.ownerOf(t, lid); char != owner.chars["DonoCura2"] || occ != (lobbies.Slots{Support: 1}) {
		t.Errorf("personagem %s, ocupantes %+v", char, occ)
	}
}

// CA-07.3: para função sem vaga, a troca é rejeitada e o dono continua com o atual.
func TestOwnerSwap_CA07_3_RoleFull(t *testing.T) {
	e, owner, _, lid, _ := swapBasic(t)
	e.fillTank(t, owner, lid)
	err := e.svc.SwapOwnerCharacter(t.Context(), owner.id, lid, owner.chars["DonoTank"])
	wantRule(t, err, CodeRoleFull)
	if char, _ := e.ownerOf(t, lid); char != owner.chars["Dono"] {
		t.Errorf("personagem = %s", char)
	}
}

// CA-07.4: personagem com vaga noutro lobby no mesmo horário não entra na troca.
func TestOwnerSwap_CA07_4_ScheduleConflict(t *testing.T) {
	e, owner, _, lid, _ := swapBasic(t)
	e.busyElsewhere(t, owner, "DonoTank")
	err := e.svc.SwapOwnerCharacter(t.Context(), owner.id, lid, owner.chars["DonoTank"])
	wantRule(t, err, CodeScheduleConflict)
}

// CA-07.5: o membro não troca direto, sem pedido.
func TestOwnerSwap_CA07_5_NotOwner(t *testing.T) {
	e, _, player, lid, _ := swapBasic(t)
	err := e.svc.SwapOwnerCharacter(t.Context(), player.id, lid, player.chars["Escudo"])
	wantRule(t, err, CodeNotOwner)
}

// CA-07.6: personagem abaixo do nível mínimo não entra na troca (RN-36).
func TestOwnerSwap_CA07_6_BelowMinLevel(t *testing.T) {
	e := setup(t)
	owner := e.user(t, "Dono:200:dps", "DonoBaixo:199:dps")
	lid := e.lobby(t, owner, "Dono", at(1, 20, 0), std, 200)
	err := e.svc.SwapOwnerCharacter(t.Context(), owner.id, lid, owner.chars["DonoBaixo"])
	wantRule(t, err, CodeBelowMinLevel)
	if char, _ := e.ownerOf(t, lid); char != owner.chars["Dono"] {
		t.Errorf("personagem = %s", char)
	}
}

// RN-19: só com lobby aberto, com personagem do próprio dono; trocar pelo mesmo não muda
// nada.
func TestOwnerSwap_RN19_Guards(t *testing.T) {
	e, owner, player, lid, _ := swapBasic(t)
	err := e.svc.SwapOwnerCharacter(t.Context(), owner.id, lid, player.chars["Escudo"])
	wantField(t, err, lobbies.FieldCharacterID, lobbies.CodeInvalid)
	err = e.svc.SwapOwnerCharacter(t.Context(), owner.id, lid, "")
	wantField(t, err, lobbies.FieldCharacterID, lobbies.CodeRequired)
	if err := e.svc.SwapOwnerCharacter(t.Context(), owner.id, lid, owner.chars["Dono"]); err != nil {
		t.Errorf("mesmo personagem: %v", err)
	}
	err = e.svc.SwapOwnerCharacter(t.Context(), owner.id, "00000000-0000-0000-0000-000000000000", owner.chars["DonoTank"])
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("lobby inexistente: %v", err)
	}
	e.now = at(1, 20, 0)
	err = e.svc.SwapOwnerCharacter(t.Context(), owner.id, lid, owner.chars["DonoTank"])
	wantRule(t, err, CodeNotOpen)
}

// CA-08.1: o pedido fica pendente e o membro continua na vaga de Dano com o atual.
func TestRequestSwap_CA08_1_Pending(t *testing.T) {
	e, _, player, lid, a := swapBasic(t)
	r, err := e.svc.RequestSwap(t.Context(), player.id, a.ID, SwapInput{
		CharacterID: player.chars["Escudo"], Reason: "  ninguém apareceu de tank  ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != StatusPending || r.FromCharacterID != player.chars["Fogo"] || r.ToCharacterID != player.chars["Escudo"] ||
		r.ToRole != "tank" || r.Reason != "ninguém apareceu de tank" {
		t.Errorf("pedido = %+v", r)
	}
	if got := e.application(t, a.ID); got.CharacterID != player.chars["Fogo"] || got.Role != "dps" {
		t.Errorf("candidatura = %+v", got)
	}
	if _, occ := e.ownerOf(t, lid); occ != (lobbies.Slots{Dps: 2}) {
		t.Errorf("ocupantes = %+v", occ)
	}
}

// CA-08.2: sem motivo, o pedido é rejeitado.
func TestRequestSwap_CA08_2_ReasonRequired(t *testing.T) {
	e, _, player, _, a := swapBasic(t)
	_, err := e.svc.RequestSwap(t.Context(), player.id, a.ID, SwapInput{CharacterID: player.chars["Escudo"], Reason: "  "})
	wantField(t, err, lobbies.FieldReason, lobbies.CodeRequired)
}

// CA-08.3: um segundo pedido pendente no mesmo lobby é rejeitado.
func TestRequestSwap_CA08_3_SecondPending(t *testing.T) {
	e, _, player, _, a := swapBasic(t)
	e.request(t, player, a, "Escudo")
	_, err := e.svc.RequestSwap(t.Context(), player.id, a.ID, SwapInput{CharacterID: player.chars["Cura"], Reason: "melhor de suporte"})
	wantRule(t, err, CodeSwapPending)
}

// CA-08.4: personagem de outro Usuário no pedido é rejeitado.
func TestRequestSwap_CA08_4_OtherUsersCharacter(t *testing.T) {
	e, owner, player, _, a := swapBasic(t)
	_, err := e.svc.RequestSwap(t.Context(), player.id, a.ID, SwapInput{CharacterID: owner.chars["DonoTank"], Reason: "ninguém apareceu de tank"})
	wantField(t, err, lobbies.FieldCharacterID, lobbies.CodeInvalid)
}

// CA-08.13: personagem abaixo do nível mínimo é rejeitado já no pedido (RN-36).
func TestRequestSwap_CA08_13_BelowMinLevel(t *testing.T) {
	e := setup(t)
	owner := e.user(t, "Dono:200:dps")
	player := e.user(t, "Fogo:200:dps", "Novato:199:tank")
	lid := e.lobby(t, owner, "Dono", at(1, 20, 0), std, 200)
	a := e.apply(t, player, "Fogo", lid)
	if _, err := e.svc.Accept(t.Context(), owner.id, a.ID); err != nil {
		t.Fatal(err)
	}
	_, err := e.svc.RequestSwap(t.Context(), player.id, a.ID, SwapInput{CharacterID: player.chars["Novato"], Reason: "ninguém apareceu de tank"})
	wantRule(t, err, CodeBelowMinLevel)
}

// RN-20: só o próprio membro aceito pede, com lobby aberto e personagem diferente do atual.
func TestRequestSwap_RN20_Guards(t *testing.T) {
	e, owner, player, lid, a := swapBasic(t)
	_, err := e.svc.RequestSwap(t.Context(), owner.id, a.ID, SwapInput{CharacterID: owner.chars["DonoTank"], Reason: "ninguém apareceu de tank"})
	wantRule(t, err, CodeNotYours)
	_, err = e.svc.RequestSwap(t.Context(), player.id, a.ID, SwapInput{CharacterID: player.chars["Fogo"], Reason: "ninguém apareceu de tank"})
	wantField(t, err, lobbies.FieldCharacterID, lobbies.CodeInvalid)

	other := e.user(t, "Outro:200:dps", "OutroTank:200:tank")
	pending := e.apply(t, other, "Outro", lid)
	_, err = e.svc.RequestSwap(t.Context(), other.id, pending.ID, SwapInput{CharacterID: other.chars["OutroTank"], Reason: "ninguém apareceu de tank"})
	wantRule(t, err, CodeNotMember)

	e.now = at(1, 20, 0)
	_, err = e.svc.RequestSwap(t.Context(), player.id, a.ID, SwapInput{CharacterID: player.chars["Escudo"], Reason: "ninguém apareceu de tank"})
	wantRule(t, err, CodeNotOpen)
}

// CA-08.5: o dono aceita; o membro passa para Tank com o personagem novo, a vaga de Dano
// fica livre e o pedido vira aceito, com evento.
func TestAcceptSwap_CA08_5(t *testing.T) {
	e, owner, player, lid, a := swapBasic(t)
	r := e.request(t, player, a, "Escudo")
	got, err := e.svc.AcceptSwap(t.Context(), owner.id, r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusAccepted || got.DecidedAt.IsZero() {
		t.Errorf("pedido = %+v", got)
	}
	if app := e.application(t, a.ID); app.CharacterID != player.chars["Escudo"] || app.Role != "tank" || app.Status != StatusAccepted {
		t.Errorf("candidatura = %+v", app)
	}
	if _, occ := e.ownerOf(t, lid); occ != (lobbies.Slots{Tank: 1, Dps: 1}) {
		t.Errorf("ocupantes = %+v", occ)
	}
	events, _ := e.q.ListSwapRequestEvents(t.Context(), uuid(r.ID))
	if len(events) != 2 || events[0].ToStatus != StatusPending || events[0].ActorID.String() != player.id ||
		events[1].ToStatus != StatusAccepted || events[1].ActorID.String() != owner.id {
		t.Errorf("eventos = %+v", events)
	}
}

// CA-08.6: com a função lotada, o aceite é rejeitado; o pedido continua pendente e o
// membro, com o atual.
func TestAcceptSwap_CA08_6_RoleFull(t *testing.T) {
	e, owner, player, _, a := swapBasic(t)
	r := e.request(t, player, a, "Escudo") // o pedido pode esperar a vaga (D-11)
	e.fillTank(t, owner, a.LobbyID)
	_, err := e.svc.AcceptSwap(t.Context(), owner.id, r.ID)
	wantRule(t, err, CodeRoleFull)
	if got := e.swapStatus(t, uuid(r.ID)); got != StatusPending {
		t.Errorf("pedido = %s", got)
	}
	if app := e.application(t, a.ID); app.CharacterID != player.chars["Fogo"] {
		t.Errorf("candidatura = %+v", app)
	}
}

// CA-08.7: com conflito de horário do personagem novo, o aceite é rejeitado.
func TestAcceptSwap_CA08_7_ScheduleConflict(t *testing.T) {
	e, owner, player, _, a := swapBasic(t)
	r := e.request(t, player, a, "Escudo")
	e.busyElsewhere(t, player, "Escudo")
	_, err := e.svc.AcceptSwap(t.Context(), owner.id, r.ID)
	wantRule(t, err, CodeScheduleConflict)
	if got := e.swapStatus(t, uuid(r.ID)); got != StatusPending {
		t.Errorf("pedido = %s", got)
	}
}

// RN-36 no aceite: o nível do personagem novo é conferido de novo (o dono pode ter subido
// o nível mínimo depois do pedido).
func TestAcceptSwap_RN36_MinLevelAtAccept(t *testing.T) {
	e, owner, player, lid, a := swapBasic(t)
	r := e.request(t, player, a, "Novato") // 199, acima do mínimo de 160 no pedido
	if _, err := e.lobbies.Update(t.Context(), owner.id, lid, lobbies.UpdateInput{StartsAt: at(1, 20, 0), Slots: std, MinLevel: 200}); err != nil {
		t.Fatal(err)
	}
	_, err := e.svc.AcceptSwap(t.Context(), owner.id, r.ID)
	wantRule(t, err, CodeBelowMinLevel)
}

// CA-08.8: o dono recusa com justificativa; o membro continua com o atual.
func TestRejectSwap_CA08_8(t *testing.T) {
	e, owner, player, _, a := swapBasic(t)
	r := e.request(t, player, a, "Escudo")
	got, err := e.svc.RejectSwap(t.Context(), owner.id, r.ID, "já achamos um tank")
	if err != nil || got.Status != StatusRejected || got.DecisionReason != "já achamos um tank" {
		t.Fatalf("recusa: %+v, %v", got, err)
	}
	if app := e.application(t, a.ID); app.CharacterID != player.chars["Fogo"] || app.Role != "dps" {
		t.Errorf("candidatura = %+v", app)
	}
}

// CA-08.9: recusa sem justificativa é rejeitada e o pedido continua pendente.
func TestRejectSwap_CA08_9_ReasonRequired(t *testing.T) {
	e, owner, player, _, a := swapBasic(t)
	r := e.request(t, player, a, "Escudo")
	_, err := e.svc.RejectSwap(t.Context(), owner.id, r.ID, "")
	wantField(t, err, lobbies.FieldReason, lobbies.CodeRequired)
	if got := e.swapStatus(t, uuid(r.ID)); got != StatusPending {
		t.Errorf("pedido = %s", got)
	}
}

// RN-22: só o dono decide o pedido, e só pedido pendente.
func TestDecideSwap_RN22_Guards(t *testing.T) {
	e, owner, player, _, a := swapBasic(t)
	r := e.request(t, player, a, "Escudo")
	_, err := e.svc.AcceptSwap(t.Context(), player.id, r.ID)
	wantRule(t, err, CodeNotOwner)
	_, err = e.svc.RejectSwap(t.Context(), player.id, r.ID, "já achamos um tank")
	wantRule(t, err, CodeNotOwner)
	if _, err := e.svc.RejectSwap(t.Context(), owner.id, r.ID, "já achamos um tank"); err != nil {
		t.Fatal(err)
	}
	_, err = e.svc.AcceptSwap(t.Context(), owner.id, r.ID)
	wantRule(t, err, CodeNotPending)
	_, err = e.svc.AcceptSwap(t.Context(), owner.id, "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("inexistente: %v", err)
	}
}

// CA-08.11: o membro retira o pedido, continua com o atual e pode abrir outro.
func TestWithdrawSwap_CA08_11(t *testing.T) {
	e, _, player, _, a := swapBasic(t)
	r := e.request(t, player, a, "Escudo")
	got, err := e.svc.WithdrawSwap(t.Context(), player.id, r.ID)
	if err != nil || got.Status != StatusWithdrawn {
		t.Fatalf("retirada: %+v, %v", got, err)
	}
	if app := e.application(t, a.ID); app.CharacterID != player.chars["Fogo"] {
		t.Errorf("candidatura = %+v", app)
	}
	e.now = e.now.Add(1)
	e.request(t, player, a, "Cura")
}

// CA-08.12: ninguém retira o pedido de outra pessoa, nem o dono.
func TestWithdrawSwap_CA08_12_NotYours(t *testing.T) {
	e, owner, player, lid, a := swapBasic(t)
	r := e.request(t, player, a, "Escudo")
	other := e.user(t, "Outro:200:dps")
	b := e.apply(t, other, "Outro", lid)
	if _, err := e.svc.Accept(t.Context(), owner.id, b.ID); err != nil {
		t.Fatal(err)
	}
	for _, w := range []who{other, owner} {
		_, err := e.svc.WithdrawSwap(t.Context(), w.id, r.ID)
		wantRule(t, err, CodeNotYours)
	}
	if got := e.swapStatus(t, uuid(r.ID)); got != StatusPending {
		t.Errorf("pedido = %s", got)
	}
}

// RN-16 e D-12: o pedido pendente expira com o início (lido) e com o cancelamento
// (gravado, com evento); expirado não é decidido nem retirado.
func TestSwap_D12_Expire(t *testing.T) {
	e, owner, player, lid, a := swapBasic(t)
	r := e.request(t, player, a, "Escudo")
	e.now = at(1, 20, 0)
	_, err := e.svc.AcceptSwap(t.Context(), owner.id, r.ID)
	wantRule(t, err, CodeNotPending)
	_, err = e.svc.WithdrawSwap(t.Context(), player.id, r.ID)
	wantRule(t, err, CodeNotPending)

	e.now = start
	if _, err := e.lobbies.Cancel(t.Context(), owner.id, lid, "imprevisto no trabalho"); err != nil {
		t.Fatal(err)
	}
	if got := e.swapStatus(t, uuid(r.ID)); got != StatusExpired {
		t.Errorf("pedido = %s", got)
	}
	events, _ := e.q.ListSwapRequestEvents(t.Context(), uuid(r.ID))
	if last := events[len(events)-1]; last.ToStatus != StatusExpired || last.ActorID.String() != owner.id {
		t.Errorf("evento = %+v", last)
	}
}

// D-10: o aceite da troca para Tank e o aceite de um candidato Tank, ao mesmo tempo, na
// última vaga: só um passa.
func TestAcceptSwap_D10_ConcurrentLastSlot(t *testing.T) {
	for range 5 {
		e, owner, player, lid, a := swapBasic(t)
		r := e.request(t, player, a, "Escudo")
		w := e.user(t, "Muralha:200:tank")
		b := e.apply(t, w, "Muralha", lid)
		var wg sync.WaitGroup
		var swapErr, acceptErr error
		wg.Go(func() { _, swapErr = e.svc.AcceptSwap(t.Context(), owner.id, r.ID) })
		wg.Go(func() { _, acceptErr = e.svc.Accept(t.Context(), owner.id, b.ID) })
		wg.Wait()
		switch {
		case swapErr == nil && acceptErr != nil:
			wantRule(t, acceptErr, CodeRoleFull)
		case acceptErr == nil && swapErr != nil:
			wantRule(t, swapErr, CodeRoleFull)
		default:
			t.Errorf("troca: %v, aceite: %v", swapErr, acceptErr)
		}
		if _, occ := e.ownerOf(t, lid); occ.Tank != 1 {
			t.Errorf("tanks = %d", occ.Tank)
		}
	}
}

// RN-16 e D-10 (validação, ciclo 1): pedir troca e cancelar o lobby ao mesmo tempo não deixa
// pedido pendente num lobby cancelado.
func TestRequestSwap_RN16_ConcurrentWithCancel(t *testing.T) {
	for range 5 {
		e, owner, player, lid, a := swapBasic(t)
		var wg sync.WaitGroup
		var reqErr, cancelErr error
		wg.Go(func() {
			_, reqErr = e.svc.RequestSwap(t.Context(), player.id, a.ID, SwapInput{
				CharacterID: player.chars["Escudo"], Reason: "ninguém apareceu de tank",
			})
		})
		wg.Go(func() { _, cancelErr = e.lobbies.Cancel(t.Context(), owner.id, lid, "imprevisto no trabalho") })
		wg.Wait()
		if cancelErr != nil {
			t.Fatalf("cancelar: %v", cancelErr)
		}
		if reqErr != nil {
			wantRule(t, reqErr, CodeNotOpen)
			continue
		}
		r, err := e.q.GetLatestSwapRequest(t.Context(), uuid(a.ID))
		if err != nil || r.Status != StatusExpired {
			t.Errorf("pedido depois do cancelamento: %+v, %v", r, err)
		}
	}
}

// RN-16: pedido que ficou pendente num lobby cancelado (dado antigo) não é aceito nem
// retirado.
func TestSwap_RN16_CancelledLobbyGuards(t *testing.T) {
	e, owner, player, lid, a := swapBasic(t)
	if _, err := e.lobbies.Cancel(t.Context(), owner.id, lid, "imprevisto no trabalho"); err != nil {
		t.Fatal(err)
	}
	r := e.pendingSwap(t, a, player.chars["Escudo"], "tank")
	_, err := e.svc.AcceptSwap(t.Context(), owner.id, r.ID.String())
	wantRule(t, err, CodeNotPending)
	_, err = e.svc.WithdrawSwap(t.Context(), player.id, r.ID.String())
	wantRule(t, err, CodeNotPending)
	if app := e.application(t, a.ID); app.CharacterID != player.chars["Fogo"] {
		t.Errorf("candidatura = %+v", app)
	}
}
