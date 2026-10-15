//go:build integration

package migrate

import (
	"strings"
	"testing"
)

// Migração 00011 sobre um banco com usuário e lobby (spec discord-bot, T-01).

// RN-05 / RN-09 / RN-22 / RN-23 / D-02: a sala é uma por lobby, com estados coerentes; o link
// do dono tem limite; a fila só aceita os dois tipos; a volta mantém os dados antigos.
func TestMigration00011_RN05_RN09_RoomsAndJobs(t *testing.T) {
	db, p := setup(t)
	ctx := t.Context()
	if _, err := p.UpTo(ctx, 10); err != nil {
		t.Fatal(err)
	}
	mustExec(t, db, `INSERT INTO users (id, discord_id, username, created_at, last_login_at)
		VALUES ('00000000-0000-0000-0000-000000000001', '1', 'ana', now(), now())`)
	mustExec(t, db, `INSERT INTO characters (id, user_id, nick, class_id, level, role, portrait, created_at, server)
		VALUES ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001',
		        'Lirien', 'arcebispo', 178, 'support', 'retrato-1', now(), 'freya')`)
	mustExec(t, db, `INSERT INTO lobbies (id, owner_id, instance_id, instance_name, instance_level,
		    starts_at, slots_tank, slots_support, slots_dps, min_level, owner_character_id, owner_role, created_at, server)
		VALUES ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001',
		        'templo-do-demonio-rei', 'Templo do Demônio Rei', 160, now() + interval '1 day', 1, 2, 3, 160,
		        '00000000-0000-0000-0000-000000000002', 'support', now(), 'freya')`)

	if _, err := p.UpTo(ctx, 11); err != nil {
		t.Fatal(err)
	}
	var connected, link any
	if err := db.QueryRowContext(ctx, `SELECT u.discord_connected_at, l.discord_link FROM users u, lobbies l`).Scan(&connected, &link); err != nil || connected != nil || link != nil {
		t.Errorf("dados antigos: conectado %v, link %v, %v", connected, link, err)
	}

	room := func(status, closing, deleted string) error {
		_, err := db.ExecContext(ctx, `INSERT INTO discord_rooms (lobby_id, category_id, text_channel_id, voice_channel_id, status, opened_at, closing_at, deleted_at)
			VALUES ('00000000-0000-0000-0000-000000000003', '10', '11', '12', $1, now(), `+closing+`, `+deleted+`)`, status)
		return err
	}
	for name, err := range map[string]error{
		"estado inválido":     room("archived", "NULL", "NULL"),
		"aberta com fim":      room("open", "now()", "NULL"),
		"encerrando sem data": room("closing", "NULL", "NULL"),
		"apagada sem data":    room("deleted", "NULL", "NULL"),
	} {
		if err == nil {
			t.Errorf("%s: aceito", name)
		}
	}
	if err := room("open", "NULL", "NULL"); err != nil {
		t.Fatalf("sala aberta: %v", err)
	}
	if err := room("open", "NULL", "NULL"); err == nil {
		t.Error("RN-05: segunda sala no mesmo lobby aceita")
	}

	job := func(kind, user string) error {
		_, err := db.ExecContext(ctx, `INSERT INTO discord_jobs (lobby_id, kind, user_id, run_at, created_at)
			VALUES ('00000000-0000-0000-0000-000000000003', $1, `+user+`, now(), now())`, kind)
		return err
	}
	if err := job("sync_access", "'00000000-0000-0000-0000-000000000001'"); err != nil {
		t.Errorf("sync_access: %v", err)
	}
	if err := job("post", "NULL"); err != nil {
		t.Errorf("post: %v", err)
	}
	for name, err := range map[string]error{
		"tipo inválido":        job("dm", "NULL"),
		"sync_access sem quem": job("sync_access", "NULL"),
		"post com usuário":     job("post", "'00000000-0000-0000-0000-000000000001'"),
	} {
		if err == nil {
			t.Errorf("%s: aceito", name)
		}
	}

	if _, err := db.ExecContext(ctx, `UPDATE lobbies SET discord_link = $1`, "https://discord.gg/"+strings.Repeat("a", 190)); err == nil {
		t.Error("RN-23: link com mais de 200 caracteres aceito")
	}
	if _, err := db.ExecContext(ctx, `UPDATE lobbies SET discord_link = 'http://discord.gg/abc'`); err == nil {
		t.Error("RN-23: link sem https aceito")
	}
	mustExec(t, db, `UPDATE lobbies SET discord_link = 'https://discord.gg/abc123'`)

	if _, err := p.DownTo(ctx, 10); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM users) + (SELECT count(*) FROM lobbies)`).Scan(&n); err != nil || n != 2 {
		t.Errorf("depois da volta: %d linhas, %v", n, err)
	}
}
