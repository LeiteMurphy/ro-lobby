//go:build integration

package migrate

import (
	"database/sql"
	"testing"
)

// Migração 00009 sobre um banco com lobby (spec lobby-sem-instancia, T-01, RNF-01).

// RNF-01 / RN-02: o lobby que já existia continua com a instância; o lobby sem instância
// é aceito com nível 1 e título de até 40; a volta apaga os sem instância.
func TestMigration00009_RNF01_ExistingLobby(t *testing.T) {
	db, p := setup(t)
	ctx := t.Context()
	if _, err := p.UpTo(ctx, 8); err != nil {
		t.Fatal(err)
	}
	mustExec(t, db, `INSERT INTO users (id, discord_id, username, created_at, last_login_at)
		VALUES ('00000000-0000-0000-0000-000000000001', '1', 'ana', now(), now())`)
	mustExec(t, db, `INSERT INTO characters (id, user_id, nick, class_id, level, role, portrait, created_at)
		VALUES ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001',
		        'Lirien', 'arcebispo', 178, 'support', 'retrato-1', now())`)
	insert := func(id, instanceID, name string, level int) error {
		_, err := db.ExecContext(ctx, `INSERT INTO lobbies (id, owner_id, instance_id, instance_name, instance_level,
		    starts_at, slots_tank, slots_support, slots_dps, min_level, owner_character_id, owner_role, created_at)
		VALUES ($1, '00000000-0000-0000-0000-000000000001', NULLIF($2, ''), NULLIF($3, ''), $4,
		        now() + interval '1 day', 1, 2, 3, 160, '00000000-0000-0000-0000-000000000002', 'support', now())`,
			id, instanceID, name, level)
		return err
	}
	if err := insert("00000000-0000-0000-0000-000000000003", "templo-do-demonio-rei", "Templo do Demônio Rei", 160); err != nil {
		t.Fatal(err)
	}
	if _, err := p.UpTo(ctx, 9); err != nil {
		t.Fatal(err)
	}
	var id sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT instance_id FROM lobbies`).Scan(&id); err != nil || id.String != "templo-do-demonio-rei" {
		t.Errorf("lobby antigo: %+v, %v", id, err)
	}
	if err := insert("00000000-0000-0000-0000-000000000004", "", "Caça ao MVP", 1); err != nil {
		t.Errorf("sem instância com título: %v", err)
	}
	if err := insert("00000000-0000-0000-0000-000000000005", "", "", 1); err != nil {
		t.Errorf("sem instância sem título: %v", err)
	}
	for name, err := range map[string]error{
		"nível 160 sem instância": insert("00000000-0000-0000-0000-000000000006", "", "", 160),
		"título de 41":            insert("00000000-0000-0000-0000-000000000007", "", "abcdefghijabcdefghijabcdefghijabcdefghijk", 1),
		"título com espaço":       insert("00000000-0000-0000-0000-000000000008", "", " Farm", 1),
		"instância sem nome":      insert("00000000-0000-0000-0000-000000000009", "templo-do-demonio-rei", "", 160),
	} {
		if err == nil {
			t.Errorf("%s: aceito", name)
		}
	}

	if _, err := p.DownTo(ctx, 8); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM lobbies`).Scan(&n); err != nil || n != 1 {
		t.Errorf("depois da volta: %d lobbies, %v", n, err)
	}
}
