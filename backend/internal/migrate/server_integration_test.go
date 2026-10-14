//go:build integration

package migrate

import (
	"testing"
)

// Migração 00010 sobre um banco com personagem e lobby (spec servidor, T-01).

// CA-06.7 / RN-15: o que já existia fica em Freya e o personagem fica a conferir; o criado
// depois nasce conferido; o nick passa a ser único por servidor; a volta mantém os dados.
func TestMigration00010_CA06_7_ExistingDataGoesToFreya(t *testing.T) {
	db, p := setup(t)
	ctx := t.Context()
	if _, err := p.UpTo(ctx, 9); err != nil {
		t.Fatal(err)
	}
	mustExec(t, db, `INSERT INTO users (id, discord_id, username, created_at, last_login_at)
		VALUES ('00000000-0000-0000-0000-000000000001', '1', 'ana', now(), now())`)
	mustExec(t, db, `INSERT INTO characters (id, user_id, nick, class_id, level, role, portrait, created_at)
		VALUES ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001',
		        'Lirien', 'arcebispo', 178, 'support', 'retrato-1', now())`)
	mustExec(t, db, `INSERT INTO lobbies (id, owner_id, instance_id, instance_name, instance_level,
		    starts_at, slots_tank, slots_support, slots_dps, min_level, owner_character_id, owner_role, created_at)
		VALUES ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001',
		        'templo-do-demonio-rei', 'Templo do Demônio Rei', 160, now() + interval '1 day', 1, 2, 3, 160,
		        '00000000-0000-0000-0000-000000000002', 'support', now())`)

	if _, err := p.UpTo(ctx, 10); err != nil {
		t.Fatal(err)
	}
	var charServer, lobbyServer string
	var toConfirm bool
	if err := db.QueryRowContext(ctx, `SELECT server, server_to_confirm FROM characters`).Scan(&charServer, &toConfirm); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT server FROM lobbies`).Scan(&lobbyServer); err != nil {
		t.Fatal(err)
	}
	if charServer != "freya" || !toConfirm || lobbyServer != "freya" {
		t.Errorf("depois da 00010: personagem %s (a conferir %v), lobby %s", charServer, toConfirm, lobbyServer)
	}

	mustExec(t, db, `INSERT INTO characters (id, user_id, nick, class_id, level, role, portrait, created_at, server)
		VALUES ('00000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001',
		        'lirien', 'arcebispo', 178, 'support', 'retrato-1', now(), 'nidhogg')`)
	if err := db.QueryRowContext(ctx, `SELECT server_to_confirm FROM characters WHERE server = 'nidhogg'`).Scan(&toConfirm); err != nil || toConfirm {
		t.Errorf("personagem novo: a conferir %v, %v", toConfirm, err)
	}
	for name, query := range map[string]string{
		"sem servidor": `INSERT INTO characters (user_id, nick, class_id, level, role, portrait, created_at)
			VALUES ('00000000-0000-0000-0000-000000000001', 'Novo', 'arcebispo', 1, 'dps', 'retrato-1', now())`,
		"servidor fora do catálogo": `INSERT INTO characters (user_id, nick, class_id, level, role, portrait, created_at, server)
			VALUES ('00000000-0000-0000-0000-000000000001', 'Novo', 'arcebispo', 1, 'dps', 'retrato-1', now(), 'thor')`,
		"nick repetido no servidor": `INSERT INTO characters (user_id, nick, class_id, level, role, portrait, created_at, server)
			VALUES ('00000000-0000-0000-0000-000000000001', 'LIRIEN', 'arcebispo', 1, 'dps', 'retrato-1', now(), 'freya')`,
	} {
		if _, err := db.ExecContext(ctx, query); err == nil {
			t.Errorf("%s: aceito", name)
		}
	}

	// A volta só passa sem o mesmo nick em servidores diferentes.
	mustExec(t, db, `DELETE FROM characters WHERE server = 'nidhogg'`)
	if _, err := p.DownTo(ctx, 9); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM characters) + (SELECT count(*) FROM lobbies)`).Scan(&n); err != nil || n != 2 {
		t.Errorf("depois da volta: %d linhas, %v", n, err)
	}
}
