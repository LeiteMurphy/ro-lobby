//go:build integration

package db

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

// Formação do lobby no banco (spec grupo-livre, T-01, migração 00008).

// RNF-01: lobby criado sem formação fica "por função", sem total livre.
func TestLobbies_RNF01_DefaultFormation(t *testing.T) {
	q, pool := setup(t)
	u, c := lobbyOwner(t, q)
	id, err := q.CreateLobby(t.Context(), newLobby(u.ID, c.ID))
	if err != nil {
		t.Fatal(err)
	}
	var formation string
	var free pgtype.Int2
	if err := pool.QueryRow(t.Context(), "SELECT formation, free_slots FROM lobbies WHERE id = $1", id).Scan(&formation, &free); err != nil {
		t.Fatal(err)
	}
	if formation != "roles" || free.Valid {
		t.Errorf("formação = %q, livre = %+v", formation, free)
	}
}

// CA-01.2 / RN-01 / RN-02: o banco aceita o grupo livre de 2 a 12 com as vagas por função zeradas e
// recusa as combinações erradas.
func TestLobbies_CA01_2_RN01_RN02_FormationChecks(t *testing.T) {
	q, _ := setup(t)
	u, c := lobbyOwner(t, q)
	freeLobby := func(total int16) CreateLobbyParams {
		p := newLobby(u.ID, c.ID)
		p.Formation, p.FreeSlots = "free", pgtype.Int2{Int16: total, Valid: true}
		p.SlotsTank, p.SlotsSupport, p.SlotsDps = 0, 0, 0
		return p
	}
	for _, n := range []int16{2, 12} {
		if _, err := q.CreateLobby(t.Context(), freeLobby(n)); err != nil {
			t.Errorf("livre com %d: %v", n, err)
		}
	}
	bad := map[string]CreateLobbyParams{
		"livre com 1":               freeLobby(1),
		"livre com 13":              freeLobby(13),
		"livre sem total":           func() CreateLobbyParams { p := freeLobby(6); p.FreeSlots = pgtype.Int2{}; return p }(),
		"livre com vaga por função": func() CreateLobbyParams { p := freeLobby(6); p.SlotsDps = 1; return p }(),
		"por função com total livre": func() CreateLobbyParams {
			p := newLobby(u.ID, c.ID)
			p.FreeSlots = pgtype.Int2{Int16: 6, Valid: true}
			return p
		}(),
		"formação desconhecida": func() CreateLobbyParams { p := newLobby(u.ID, c.ID); p.Formation = "misto"; return p }(),
	}
	for name, p := range bad {
		if _, err := q.CreateLobby(t.Context(), p); pgCode(err) != "23514" {
			t.Errorf("%s: err = %v, quer violação de CHECK (23514)", name, err)
		}
	}
}
