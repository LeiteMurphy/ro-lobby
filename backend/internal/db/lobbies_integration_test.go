//go:build integration

package db

import (
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Testes da tabela de lobbies (spec lobbies, T-02). Só o banco: as mensagens e a ordem das
// regras ficam no serviço (internal/lobbies).

func newLobby(ownerID, characterID pgtype.UUID) CreateLobbyParams {
	return CreateLobbyParams{
		OwnerID:          ownerID,
		InstanceID:       pgtype.Text{String: "templo-do-demonio-rei", Valid: true},
		InstanceName:     pgtype.Text{String: "Templo do Demônio Rei", Valid: true},
		InstanceLevel:    160,
		StartsAt:         t0.Add(24 * time.Hour),
		SlotsTank:        1,
		SlotsSupport:     2,
		SlotsDps:         3,
		MinLevel:         160,
		OwnerCharacterID: characterID,
		OwnerRole:        "support",
		Now:              t0,
	}
}

func lobbyOwner(t *testing.T, q *Queries) (User, Character) {
	t.Helper()
	u := upsert(t, q, "1", "ana", "Ana", t0)
	c, err := q.CreateCharacter(t.Context(), CreateCharacterParams{
		UserID: u.ID, Nick: "Lirien", ClassID: "arcebispo", Level: 178, Role: "support",
		Portrait: "retrato-1", IsMain: true, Now: t0,
	})
	if err != nil {
		t.Fatal(err)
	}
	return u, c
}

// RN-06, RN-07, RN-09, RN-19: o banco recusa valores fora das regras.
func TestLobbies_RN06_RN07_RN09_RN19_ChecksRejectInvalidValues(t *testing.T) {
	q, _ := setup(t)
	u, c := lobbyOwner(t, q)

	cases := map[string]func(p *CreateLobbyParams){
		"vagas de Tank 13":            func(p *CreateLobbyParams) { p.SlotsTank = 13 },
		"vagas negativas":             func(p *CreateLobbyParams) { p.SlotsDps = -1 },
		"total 0":                     func(p *CreateLobbyParams) { p.SlotsTank, p.SlotsSupport, p.SlotsDps = 0, 0, 0 },
		"total 13":                    func(p *CreateLobbyParams) { p.SlotsTank, p.SlotsSupport, p.SlotsDps = 1, 4, 8 },
		"nível mínimo 276":            func(p *CreateLobbyParams) { p.MinLevel = 276 },
		"nível mínimo abaixo da inst": func(p *CreateLobbyParams) { p.MinLevel = 150 },
		"função desconhecida":         func(p *CreateLobbyParams) { p.OwnerRole = "healer" },
		"instância vazia":             func(p *CreateLobbyParams) { p.InstanceID = pgtype.Text{String: "", Valid: true} },
		"observação com 251":          func(p *CreateLobbyParams) { p.Note = pgtype.Text{String: strings.Repeat("a", 251), Valid: true} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := newLobby(u.ID, c.ID)
			mutate(&p)
			if _, err := q.CreateLobby(t.Context(), p); pgCode(err) != "23514" {
				t.Errorf("err = %v, quer violação de CHECK (23514)", err)
			}
		})
	}

	t.Run("limites aceitos", func(t *testing.T) {
		p := newLobby(u.ID, c.ID)
		p.SlotsTank, p.SlotsSupport, p.SlotsDps = 0, 12, 0
		p.MinLevel = 275
		p.Note = pgtype.Text{String: strings.Repeat("a", 250), Valid: true}
		if _, err := q.CreateLobby(t.Context(), p); err != nil {
			t.Fatal(err)
		}
	})
}

// RN-19: cancelar exige motivo de 10 a 250 caracteres, e o motivo só existe com o
// cancelamento.
func TestLobbies_RN19_CancelNeedsReason(t *testing.T) {
	q, pool := setup(t)
	u, c := lobbyOwner(t, q)
	id, err := q.CreateLobby(t.Context(), newLobby(u.ID, c.ID))
	if err != nil {
		t.Fatal(err)
	}
	for _, reason := range []string{"curto", "   curto   ", strings.Repeat("a", 251)} {
		if err := q.CancelLobby(t.Context(), CancelLobbyParams{ID: id, Now: t0, Reason: reason}); pgCode(err) != "23514" {
			t.Errorf("motivo %q: err = %v", reason, err)
		}
	}
	if _, err := pool.Exec(t.Context(), "UPDATE lobbies SET cancel_reason = 'motivo sem data' WHERE id = $1", id); pgCode(err) != "23514" {
		t.Errorf("motivo sem cancelamento: err = %v", err)
	}
	if err := q.CancelLobby(t.Context(), CancelLobbyParams{ID: id, Now: t0, Reason: "Metade do grupo não pode"}); err != nil {
		t.Fatal(err)
	}
}

// RN-13 / RN-14: a listagem traz só abertos, por início, dentro do intervalo; cancelado e
// iniciado ficam de fora.
func TestLobbies_RN13_RN14_ListOnlyOpen(t *testing.T) {
	q, _ := setup(t)
	u, c := lobbyOwner(t, q)
	now := t0.Add(time.Hour)
	create := func(startsAt time.Time) pgtype.UUID {
		p := newLobby(u.ID, c.ID)
		p.StartsAt = startsAt
		id, err := q.CreateLobby(t.Context(), p)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	started := create(now.Add(-time.Minute))
	later := create(now.Add(5 * time.Hour))
	sooner := create(now.Add(2 * time.Hour))
	cancelled := create(now.Add(3 * time.Hour))
	outside := create(now.Add(72 * time.Hour))
	if err := q.CancelLobby(t.Context(), CancelLobbyParams{ID: cancelled, Now: now, Reason: "Não vai rolar hoje"}); err != nil {
		t.Fatal(err)
	}

	rows, err := q.ListOpenLobbies(t.Context(), ListOpenLobbiesParams{Now: now, FromAt: t0, ToAt: t0.Add(48 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	var got []pgtype.UUID
	for _, r := range rows {
		got = append(got, r.Lobby.ID)
		if r.OwnerNick.String != "Lirien" || r.OwnerUsername != "ana" {
			t.Errorf("dono = %+v", r)
		}
	}
	if len(got) != 2 || got[0] != sooner || got[1] != later {
		t.Errorf("lista = %v, quer [%v %v] (sem %v, %v, %v)", got, sooner, later, started, cancelled, outside)
	}
}

// RN-10 / D-03: a janela de conflito é aberta nas pontas e ignora cancelados e o próprio
// lobby.
func TestLobbies_RN10_ScheduleConflictWindow(t *testing.T) {
	q, _ := setup(t)
	u, c := lobbyOwner(t, q)
	start := t0.Add(24 * time.Hour)
	p := newLobby(u.ID, c.ID)
	p.StartsAt = start
	id, err := q.CreateLobby(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}
	conflict := func(at time.Time, exclude pgtype.UUID) bool {
		ok, err := q.HasScheduleConflict(t.Context(), HasScheduleConflictParams{
			CharacterID: c.ID, ExcludeID: exclude, WindowStart: at.Add(-2 * time.Hour), WindowEnd: at.Add(2 * time.Hour),
		})
		if err != nil {
			t.Fatal(err)
		}
		return ok
	}
	none := pgtype.UUID{}
	if !conflict(start.Add(90*time.Minute), none) {
		t.Error("21:30 deveria conflitar com 20:00")
	}
	if conflict(start.Add(2*time.Hour), none) || conflict(start.Add(-2*time.Hour), none) {
		t.Error("2 h de distância não conflita")
	}
	if conflict(start, id) {
		t.Error("o próprio lobby não conta")
	}
	if err := q.CancelLobby(t.Context(), CancelLobbyParams{ID: id, Now: t0, Reason: "Cancelado no teste"}); err != nil {
		t.Fatal(err)
	}
	if conflict(start, none) {
		t.Error("lobby cancelado não conta")
	}
}

// RN-21 / D-01: o personagem excluído vira nulo no lobby, e o lobby sai junto com o
// Usuário.
func TestLobbies_RN21_CharacterSetNullAndUserCascade(t *testing.T) {
	q, pool := setup(t)
	u, c := lobbyOwner(t, q)
	id, err := q.CreateLobby(t.Context(), newLobby(u.ID, c.ID))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(t.Context(), "DELETE FROM characters WHERE id = $1", c.ID); err != nil {
		t.Fatal(err)
	}
	row, err := q.GetLobby(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	if row.Lobby.OwnerCharacterID.Valid || row.OwnerNick.Valid {
		t.Errorf("personagem continua no lobby: %+v", row)
	}
	if _, err := pool.Exec(t.Context(), "DELETE FROM users WHERE id = $1", u.ID); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM lobbies").Scan(&n); err != nil || n != 0 {
		t.Errorf("sobraram %d lobbies (err %v)", n, err)
	}
}
