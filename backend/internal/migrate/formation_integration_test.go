//go:build integration

package migrate

import (
	"database/sql"
	"testing"
)

// Migração 00008 sobre um banco com lobby (spec grupo-livre, T-01, RNF-01).

// RNF-01: o lobby que já existia fica "por função", sem total livre; a volta da migração
// devolve o CHECK da soma por função.
func TestMigration00008_RNF01_ExistingLobby(t *testing.T) {
	db, p := setup(t)
	ctx := t.Context()
	if _, err := p.UpTo(ctx, 7); err != nil {
		t.Fatal(err)
	}
	mustExec(t, db, `INSERT INTO users (id, discord_id, username, created_at, last_login_at)
		VALUES ('00000000-0000-0000-0000-000000000001', '1', 'ana', now(), now())`)
	mustExec(t, db, `INSERT INTO characters (id, user_id, nick, class_id, level, role, portrait, created_at)
		VALUES ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001',
		        'Lirien', 'arcebispo', 178, 'support', 'retrato-1', now())`)
	mustExec(t, db, `INSERT INTO lobbies (id, owner_id, instance_id, instance_name, instance_level, starts_at,
		    slots_tank, slots_support, slots_dps, min_level, owner_character_id, owner_role, created_at)
		VALUES ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001',
		        'templo-do-demonio-rei', 'Templo do Demônio Rei', 160, now() + interval '1 day',
		        1, 2, 3, 160, '00000000-0000-0000-0000-000000000002', 'support', now())`)

	if _, err := p.UpTo(ctx, 8); err != nil {
		t.Fatal(err)
	}
	var formation string
	var free sql.NullInt16
	if err := db.QueryRowContext(ctx, `SELECT formation, free_slots FROM lobbies`).Scan(&formation, &free); err != nil {
		t.Fatal(err)
	}
	if formation != "roles" || free.Valid {
		t.Errorf("depois da 00008: formação %q, livre %+v", formation, free)
	}

	if _, err := p.DownTo(ctx, 7); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM lobbies`).Scan(&n); err != nil || n != 1 {
		t.Errorf("lobby depois da volta: %d, %v", n, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE lobbies SET slots_tank = 0, slots_support = 0, slots_dps = 0`); err == nil {
		t.Error("a volta não devolveu o CHECK da soma por função")
	}
}

func mustExec(t *testing.T, db *sql.DB, query string) {
	t.Helper()
	if _, err := db.ExecContext(t.Context(), query); err != nil {
		t.Fatal(err)
	}
}
