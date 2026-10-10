//go:build integration

package db

import (
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Testes das tabelas de candidatura (spec candidatura-lobby, T-01). Só o banco: as regras
// e as mensagens ficam no serviço (internal/applications).

type appFixture struct {
	q      *Queries
	lobby  pgtype.UUID
	owner  User
	player User
	char   Character
}

func appSetup(t *testing.T) appFixture {
	t.Helper()
	q, _ := setup(t)
	owner, ownerChar := lobbyOwner(t, q)
	lobby, err := q.CreateLobby(t.Context(), newLobby(owner.ID, ownerChar.ID))
	if err != nil {
		t.Fatal(err)
	}
	player := upsert(t, q, "2", "bia", "Bia", t0)
	char, err := q.CreateCharacter(t.Context(), CreateCharacterParams{
		UserID: player.ID, Nick: "Brasa", ClassID: "guardiao-real", Level: 200, Role: "tank",
		Portrait: "retrato-3", IsMain: true, Now: t0,
	})
	if err != nil {
		t.Fatal(err)
	}
	return appFixture{q: q, lobby: lobby, owner: owner, player: player, char: char}
}

func (f appFixture) apply(t *testing.T) Application {
	t.Helper()
	a, err := f.q.CreateApplication(t.Context(), CreateApplicationParams{
		LobbyID: f.lobby, UserID: f.player.ID, CharacterID: f.char.ID, Role: "tank",
		Message: pgtype.Text{String: "tenho buff", Valid: true}, Now: t0,
	})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func (f appFixture) set(t *testing.T, id pgtype.UUID, status string) {
	t.Helper()
	if _, err := f.q.SetApplicationStatus(t.Context(), SetApplicationStatusParams{ID: id, Status: status, Now: t0}); err != nil {
		t.Fatal(err)
	}
}

// RN-02: no máximo uma candidatura ativa (pendente ou aceita) por Usuário e lobby; depois
// de retirada, pode haver outra.
func TestApplications_RN02_OneActivePerUser(t *testing.T) {
	f := appSetup(t)
	first := f.apply(t)
	if _, err := f.q.CreateApplication(t.Context(), CreateApplicationParams{
		LobbyID: f.lobby, UserID: f.player.ID, CharacterID: f.char.ID, Role: "tank", Now: t0,
	}); pgCode(err) != "23505" || pgConstraint(err) != "applications_one_active_per_user" {
		t.Errorf("segunda ativa: err = %v", err)
	}
	f.set(t, first.ID, "withdrawn")
	f.apply(t)
	if ok, _ := f.q.HasActiveApplication(t.Context(), HasActiveApplicationParams{LobbyID: f.lobby, UserID: f.player.ID}); !ok {
		t.Error("a nova candidatura não ficou ativa")
	}
}

// RN-06, RN-09, RN-17: mensagem até 250, justificativa de 10 a 250 e só os estados
// conhecidos.
func TestApplications_RN06_RN09_RN17_Checks(t *testing.T) {
	f := appSetup(t)
	if _, err := f.q.CreateApplication(t.Context(), CreateApplicationParams{
		LobbyID: f.lobby, UserID: f.player.ID, CharacterID: f.char.ID, Role: "tank",
		Message: pgtype.Text{String: strings.Repeat("a", 251), Valid: true}, Now: t0,
	}); pgCode(err) != "23514" {
		t.Errorf("mensagem com 251: err = %v", err)
	}
	a := f.apply(t)
	for _, p := range []SetApplicationStatusParams{
		{ID: a.ID, Status: "rejected", Reason: pgtype.Text{String: "  curta  ", Valid: true}, Now: t0},
		{ID: a.ID, Status: "rejected", Reason: pgtype.Text{String: strings.Repeat("a", 251), Valid: true}, Now: t0},
		{ID: a.ID, Status: "talvez", Now: t0},
	} {
		if _, err := f.q.SetApplicationStatus(t.Context(), p); pgCode(err) != "23514" {
			t.Errorf("%+v: err = %v", p, err)
		}
	}
	got, err := f.q.SetApplicationStatus(t.Context(), SetApplicationStatusParams{
		ID: a.ID, Status: "rejected", Reason: pgtype.Text{String: "já temos tank", Valid: true}, Now: t0,
	})
	if err != nil || got.Status != "rejected" || !got.DecidedAt.Valid {
		t.Errorf("recusa válida: %+v, %v", got, err)
	}
	if ok, _ := f.q.WasRejected(t.Context(), WasRejectedParams{LobbyID: f.lobby, UserID: f.player.ID}); !ok {
		t.Error("WasRejected deveria ser verdadeiro")
	}
}

// RN-18: o histórico guarda cada transição em ordem, com autor e justificativa.
func TestApplications_RN18_Events(t *testing.T) {
	f := appSetup(t)
	a := f.apply(t)
	for _, e := range []InsertApplicationEventParams{
		{ApplicationID: a.ID, ToStatus: "pending", ActorID: f.player.ID, Now: t0},
		{ApplicationID: a.ID, FromStatus: pgtype.Text{String: "pending", Valid: true}, ToStatus: "rejected",
			ActorID: f.owner.ID, Reason: pgtype.Text{String: "já temos tank", Valid: true}, Now: t0.Add(time.Minute)},
	} {
		if err := f.q.InsertApplicationEvent(t.Context(), e); err != nil {
			t.Fatal(err)
		}
	}
	events, err := f.q.ListApplicationEvents(t.Context(), a.ID)
	if err != nil || len(events) != 2 || events[1].ToStatus != "rejected" || events[1].Reason.String != "já temos tank" || events[1].ActorID != f.owner.ID {
		t.Errorf("eventos = %+v, %v", events, err)
	}
}

// D-03: o lobby traz os aceitos por função e as pendentes; o cancelamento expira as
// pendentes (RN-16).
func TestApplications_D03_LobbyCountsAndExpire(t *testing.T) {
	f := appSetup(t)
	a := f.apply(t)
	row, err := f.q.GetLobby(t.Context(), f.lobby)
	if err != nil || row.PendingCount != 1 || row.AcceptedTank != 0 {
		t.Fatalf("antes do aceite: %+v, %v", row, err)
	}
	f.set(t, a.ID, "accepted")
	row, _ = f.q.GetLobby(t.Context(), f.lobby)
	if row.PendingCount != 0 || row.AcceptedTank != 1 || row.AcceptedSupport != 0 {
		t.Errorf("depois do aceite: tank %d, pendentes %d", row.AcceptedTank, row.PendingCount)
	}

	other := upsert(t, f.q, "3", "cai", "", t0)
	c, err := f.q.CreateCharacter(t.Context(), CreateCharacterParams{UserID: other.ID, Nick: "Fa", ClassID: "feiticeiro", Level: 200, Role: "dps", Portrait: "retrato-1", IsMain: true, Now: t0})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.q.CreateApplication(t.Context(), CreateApplicationParams{LobbyID: f.lobby, UserID: other.ID, CharacterID: c.ID, Role: "dps", Now: t0}); err != nil {
		t.Fatal(err)
	}
	expired, err := f.q.ExpirePendingForLobby(t.Context(), ExpirePendingForLobbyParams{LobbyID: f.lobby, Now: t0})
	if err != nil || len(expired) != 1 {
		t.Errorf("expiradas = %v, %v", expired, err)
	}
	row, _ = f.q.GetLobby(t.Context(), f.lobby)
	if row.PendingCount != 0 || row.AcceptedTank != 1 {
		t.Errorf("depois de expirar: tank %d, pendentes %d", row.AcceptedTank, row.PendingCount)
	}
}

// D-04: o personagem aceito como membro conta para o conflito de horário.
func TestApplications_D04_ConflictCountsMembers(t *testing.T) {
	f := appSetup(t)
	a := f.apply(t)
	start := t0.Add(24 * time.Hour) // início do lobby de newLobby
	conflict := func() bool {
		ok, err := f.q.HasScheduleConflict(t.Context(), HasScheduleConflictParams{
			CharacterID: f.char.ID, WindowStart: start.Add(-time.Hour), WindowEnd: start.Add(3 * time.Hour),
		})
		if err != nil {
			t.Fatal(err)
		}
		return ok
	}
	if conflict() {
		t.Error("candidatura pendente não deveria contar")
	}
	f.set(t, a.ID, "accepted")
	if !conflict() {
		t.Error("membro aceito deveria contar")
	}
}

// RN-33: "minhas candidaturas" traz o lobby e o personagem; o detalhe acha a mais recente.
func TestApplications_RN33_ListMine(t *testing.T) {
	f := appSetup(t)
	a := f.apply(t)
	mine, err := f.q.ListMyApplications(t.Context(), f.player.ID)
	if err != nil || len(mine) != 1 || mine[0].InstanceName.String != "Templo do Demônio Rei" || mine[0].Nick.String != "Brasa" {
		t.Fatalf("minhas = %+v, %v", mine, err)
	}
	last, err := f.q.GetUserApplicationInLobby(t.Context(), GetUserApplicationInLobbyParams{LobbyID: f.lobby, UserID: f.player.ID})
	if err != nil || last.ID != a.ID {
		t.Errorf("última = %+v, %v", last, err)
	}
	list, err := f.q.ListLobbyApplications(t.Context(), f.lobby)
	if err != nil || len(list) != 1 || list[0].Username != "bia" || list[0].Portrait.String != "retrato-3" {
		t.Errorf("do lobby = %+v, %v", list, err)
	}
}
