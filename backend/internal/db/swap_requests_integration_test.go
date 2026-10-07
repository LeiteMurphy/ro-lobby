//go:build integration

package db

import (
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Testes das tabelas de pedido de troca e do bloqueio (spec candidatura-lobby, Parte 2,
// T-09). Só o banco: as regras e as mensagens ficam no serviço (internal/applications).

type swapFixture struct {
	appFixture
	app    Application
	toChar Character
}

func swapSetup(t *testing.T) swapFixture {
	t.Helper()
	f := appSetup(t)
	app := f.apply(t)
	f.set(t, app.ID, "accepted")
	toChar, err := f.q.CreateCharacter(t.Context(), CreateCharacterParams{
		UserID: f.player.ID, Nick: "Faisca", ClassID: "arcebispo", Level: 200, Role: "support",
		Portrait: "retrato-1", Now: t0,
	})
	if err != nil {
		t.Fatal(err)
	}
	return swapFixture{appFixture: f, app: app, toChar: toChar}
}

func (f swapFixture) request(t *testing.T) SwapRequest {
	t.Helper()
	r, err := f.q.CreateSwapRequest(t.Context(), CreateSwapRequestParams{
		ApplicationID: f.app.ID, FromCharacterID: f.char.ID, ToCharacterID: f.toChar.ID,
		ToRole: "support", Reason: "ninguém apareceu de suporte", Now: t0,
	})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// RN-15, D-09: a candidatura nasce sem bloqueio; o bloqueio vale só para o lobby da
// remoção e só para a candidatura removida.
func TestApplications_RN15_Blocked(t *testing.T) {
	f := appSetup(t)
	a := f.apply(t)
	if a.Blocked {
		t.Fatal("candidatura nova já bloqueada")
	}
	params := IsBlockedParams{LobbyID: f.lobby, UserID: f.player.ID}
	f.set(t, a.ID, "removed")
	if ok, _ := f.q.IsBlocked(t.Context(), params); ok {
		t.Error("removida sem bloqueio não deveria bloquear")
	}
	if err := f.q.SetApplicationBlocked(t.Context(), a.ID); err != nil {
		t.Fatal(err)
	}
	if ok, err := f.q.IsBlocked(t.Context(), params); err != nil || !ok {
		t.Errorf("removida com bloqueio: %v, %v", ok, err)
	}
	other, err := f.q.CreateLobby(t.Context(), newLobby(f.owner.ID, pgtype.UUID{}))
	if err != nil {
		t.Fatal(err)
	}
	if ok, _ := f.q.IsBlocked(t.Context(), IsBlockedParams{LobbyID: other, UserID: f.player.ID}); ok {
		t.Error("o bloqueio vazou para outro lobby (CA-06.7)")
	}
}

// RN-20: no máximo um pedido pendente por candidatura; depois de retirado, pode haver
// outro.
func TestSwapRequests_RN20_OnePending(t *testing.T) {
	f := swapSetup(t)
	first := f.request(t)
	if _, err := f.q.CreateSwapRequest(t.Context(), CreateSwapRequestParams{
		ApplicationID: f.app.ID, ToCharacterID: f.toChar.ID, ToRole: "support",
		Reason: "outro motivo qualquer", Now: t0,
	}); pgCode(err) != "23505" || pgConstraint(err) != "swap_requests_one_pending" {
		t.Errorf("segundo pendente: err = %v", err)
	}
	got, err := f.q.GetPendingSwapRequest(t.Context(), f.app.ID)
	if err != nil || got.ID != first.ID {
		t.Fatalf("pendente: %+v, %v", got, err)
	}
	if _, err := f.q.SetSwapRequestStatus(t.Context(), SetSwapRequestStatusParams{ID: first.ID, Status: "withdrawn", Now: t0}); err != nil {
		t.Fatal(err)
	}
	second, err := f.q.CreateSwapRequest(t.Context(), CreateSwapRequestParams{
		ApplicationID: f.app.ID, ToCharacterID: f.toChar.ID, ToRole: "support",
		Reason: "agora sim, de suporte", Now: t0.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	latest, err := f.q.GetLatestSwapRequest(t.Context(), f.app.ID)
	if err != nil || latest.ID != second.ID {
		t.Errorf("mais recente: %+v, %v", latest, err)
	}
}

// RN-20, RN-22, RN-24: motivo e justificativa de 10 a 250 caracteres e só os estados
// conhecidos.
func TestSwapRequests_RN20_RN22_RN24_Checks(t *testing.T) {
	f := swapSetup(t)
	for _, reason := range []string{"  curto   ", strings.Repeat("a", 251)} {
		if _, err := f.q.CreateSwapRequest(t.Context(), CreateSwapRequestParams{
			ApplicationID: f.app.ID, ToCharacterID: f.toChar.ID, ToRole: "support", Reason: reason, Now: t0,
		}); pgCode(err) != "23514" {
			t.Errorf("motivo %q: err = %v", reason, err)
		}
	}
	r := f.request(t)
	for _, p := range []SetSwapRequestStatusParams{
		{ID: r.ID, Status: "rejected", DecisionReason: pgtype.Text{String: "  curta  ", Valid: true}, Now: t0},
		{ID: r.ID, Status: "rejected", DecisionReason: pgtype.Text{String: strings.Repeat("a", 251), Valid: true}, Now: t0},
		{ID: r.ID, Status: "talvez", Now: t0},
		{ID: r.ID, Status: "left", Now: t0},
	} {
		if _, err := f.q.SetSwapRequestStatus(t.Context(), p); pgCode(err) != "23514" {
			t.Errorf("%+v: err = %v", p, err)
		}
	}
	for _, status := range []string{"accepted", "rejected", "withdrawn", "expired", "cancelled"} {
		p := SetSwapRequestStatusParams{ID: r.ID, Status: status, Now: t0}
		if status == "rejected" {
			p.DecisionReason = pgtype.Text{String: "já achamos um tank", Valid: true}
		}
		got, err := f.q.SetSwapRequestStatus(t.Context(), p)
		if err != nil || got.Status != status || !got.DecidedAt.Valid {
			t.Errorf("%s: %+v, %v", status, got, err)
		}
	}
}

// RN-18: o histórico do pedido guarda cada transição em ordem, com autor e justificativa.
func TestSwapRequests_RN18_Events(t *testing.T) {
	f := swapSetup(t)
	r := f.request(t)
	for _, e := range []InsertSwapRequestEventParams{
		{SwapRequestID: r.ID, ToStatus: "pending", ActorID: f.player.ID, Reason: pgtype.Text{String: r.Reason, Valid: true}, Now: t0},
		{SwapRequestID: r.ID, FromStatus: pgtype.Text{String: "pending", Valid: true}, ToStatus: "rejected",
			ActorID: f.owner.ID, Reason: pgtype.Text{String: "já achamos um tank", Valid: true}, Now: t0.Add(time.Minute)},
	} {
		if err := f.q.InsertSwapRequestEvent(t.Context(), e); err != nil {
			t.Fatal(err)
		}
	}
	events, err := f.q.ListSwapRequestEvents(t.Context(), r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].ToStatus != "pending" || events[1].ToStatus != "rejected" ||
		events[1].ActorID != f.owner.ID || events[1].Reason.String != "já achamos um tank" {
		t.Errorf("histórico: %+v", events)
	}
}

// D-08, D-12, RN-16, RN-23: a lista do dono traz o personagem atual e o novo; o
// cancelamento expira os pendentes; o aceite muda personagem e função da candidatura.
func TestSwapRequests_D12_ListExpireAndApply(t *testing.T) {
	f := swapSetup(t)
	r := f.request(t)
	rows, err := f.q.ListLobbySwapRequests(t.Context(), f.lobby)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].SwapRequest.ID != r.ID || rows[0].FromNick.String != "Brasa" ||
		rows[0].ToNick.String != "Faisca" || rows[0].FromRole != "tank" || rows[0].UserID != f.player.ID {
		t.Errorf("lista: %+v", rows)
	}

	updated, err := f.q.SetApplicationCharacter(t.Context(), SetApplicationCharacterParams{
		ID: f.app.ID, CharacterID: f.toChar.ID, Role: "support",
	})
	if err != nil || updated.CharacterID != f.toChar.ID || updated.Role != "support" || updated.Status != "accepted" {
		t.Errorf("troca aplicada: %+v, %v", updated, err)
	}

	expired, err := f.q.ExpirePendingSwapsForLobby(t.Context(), ExpirePendingSwapsForLobbyParams{LobbyID: f.lobby, Now: t0})
	if err != nil || len(expired) != 1 || expired[0] != r.ID {
		t.Errorf("expirados: %v, %v", expired, err)
	}
	if rows, _ := f.q.ListLobbySwapRequests(t.Context(), f.lobby); len(rows) != 0 {
		t.Errorf("pendentes depois de expirar: %d", len(rows))
	}
}
