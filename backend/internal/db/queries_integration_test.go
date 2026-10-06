//go:build integration

package db

import (
	"context"
	"crypto/sha256"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/migrate"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

var t0 = time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)

func setup(t *testing.T) (*Queries, *pgxpool.Pool) {
	t.Helper()
	url := testdb.New(t)
	sqlDB, err := migrate.OpenDB(url)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := migrate.NewProvider(sqlDB)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(t.Context()); err != nil {
		t.Fatal(err)
	}
	_ = sqlDB.Close()

	pool, err := database.Open(t.Context(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return New(pool), pool
}

func hash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func upsert(t *testing.T, q *Queries, discordID, username, globalName string, now time.Time) User {
	t.Helper()
	u, err := q.UpsertUserByDiscordID(t.Context(), UpsertUserByDiscordIDParams{
		DiscordID:  discordID,
		Username:   username,
		GlobalName: pgtype.Text{String: globalName, Valid: globalName != ""},
		Now:        now,
	})
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// RN-05 / CA-01.2: o ID do Discord identifica o Usuário; o segundo login atualiza os nomes.
func TestUsers_RN05_UpsertByDiscordID(t *testing.T) {
	q, _ := setup(t)

	first := upsert(t, q, "111", "grimbold", "Grimbold", t0)
	second := upsert(t, q, "111", "grimbold", "Grimbold, o Sábio", t0.Add(48*time.Hour))

	if first.ID != second.ID {
		t.Fatalf("segundo login criou outro Usuário: %v != %v", first.ID, second.ID)
	}
	if second.GlobalName.String != "Grimbold, o Sábio" {
		t.Errorf("nome de exibição = %q", second.GlobalName.String)
	}
	if !second.CreatedAt.Equal(t0) || !second.LastLoginAt.Equal(t0.Add(48*time.Hour)) {
		t.Errorf("datas = criado %v, último login %v", second.CreatedAt, second.LastLoginAt)
	}
	n, err := q.CountUsersByDiscordID(t.Context(), "111")
	if err != nil || n != 1 {
		t.Fatalf("usuários com o ID 111 = %d (%v), esperado 1", n, err)
	}
}

// RN-06 / RN-07 / CA-01.4 / CA-02.4: as tabelas não têm coluna para token em texto,
// tokens do Discord nem avatar.
func TestSchema_CA01_4_NoTokensNorAvatar(t *testing.T) {
	_, pool := setup(t)
	columns := func(table string) []string {
		rows, err := pool.Query(t.Context(),
			`SELECT column_name FROM information_schema.columns WHERE table_name = $1 ORDER BY column_name`, table)
		if err != nil {
			t.Fatal(err)
		}
		names, err := pgx.CollectRows(rows, pgx.RowTo[string])
		if err != nil {
			t.Fatal(err)
		}
		return names
	}
	if got, want := columns("users"), []string{"created_at", "discord_id", "global_name", "id", "last_login_at", "username"}; !slices.Equal(got, want) {
		t.Errorf("colunas de users = %v, esperado %v", got, want)
	}
	if got, want := columns("sessions"), []string{"created_at", "last_used_at", "token_hash", "user_id"}; !slices.Equal(got, want) {
		t.Errorf("colunas de sessions = %v, esperado %v", got, want)
	}
}

// RN-07: o banco só aceita o hash SHA-256 (32 bytes), não o token em texto.
func TestSessions_RN07_OnlyHashIsAccepted(t *testing.T) {
	q, _ := setup(t)
	u := upsert(t, q, "222", "mirai.exe", "", t0)
	err := q.CreateSession(t.Context(), CreateSessionParams{TokenHash: []byte("token-em-texto"), UserID: u.ID, Now: t0})
	if err == nil {
		t.Fatal("o banco aceitou um token em texto no lugar do hash")
	}
	if err := q.CreateSession(t.Context(), CreateSessionParams{TokenHash: hash("t"), UserID: u.ID, Now: t0}); err != nil {
		t.Fatalf("hash de 32 bytes recusado: %v", err)
	}
}

// RN-09 / RN-10: a Sessão vale enquanto o último uso for posterior a agora menos 30 dias.
func TestSessions_RN09_ActiveUntil30DaysAfterLastUse(t *testing.T) {
	q, _ := setup(t)
	u := upsert(t, q, "333", "kaizen", "Kaizen", t0)
	if err := q.CreateSession(t.Context(), CreateSessionParams{TokenHash: hash("a"), UserID: u.ID, Now: t0}); err != nil {
		t.Fatal(err)
	}

	row, err := q.GetActiveSession(t.Context(), GetActiveSessionParams{TokenHash: hash("a"), ValidAfter: t0.Add(-time.Second)})
	if err != nil {
		t.Fatalf("sessão válida não encontrada: %v", err)
	}
	if row.User.ID != u.ID || row.User.Username != "kaizen" {
		t.Errorf("usuário da sessão = %+v", row.User)
	}

	_, err = q.GetActiveSession(t.Context(), GetActiveSessionParams{TokenHash: hash("a"), ValidAfter: t0})
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("sessão vencida deveria sumir, err = %v", err)
	}
	_, err = q.GetActiveSession(t.Context(), GetActiveSessionParams{TokenHash: hash("desconhecido"), ValidAfter: t0.Add(-time.Hour)})
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("sessão desconhecida deveria sumir, err = %v", err)
	}
}

// RN-09: a renovação só grava se o último uso registrado for mais antigo que o corte.
func TestSessions_RN09_TouchOnlyWhenStale(t *testing.T) {
	q, _ := setup(t)
	u := upsert(t, q, "444", "nhoque", "", t0)
	if err := q.CreateSession(t.Context(), CreateSessionParams{TokenHash: hash("b"), UserID: u.ID, Now: t0}); err != nil {
		t.Fatal(err)
	}

	n, err := q.TouchSession(t.Context(), TouchSessionParams{TokenHash: hash("b"), Now: t0.Add(30 * time.Minute), StaleBefore: t0.Add(-30 * time.Minute)})
	if err != nil || n != 0 {
		t.Fatalf("renovou sessão usada há 30 min: linhas = %d (%v)", n, err)
	}
	n, err = q.TouchSession(t.Context(), TouchSessionParams{TokenHash: hash("b"), Now: t0.Add(2 * time.Hour), StaleBefore: t0.Add(time.Hour)})
	if err != nil || n != 1 {
		t.Fatalf("não renovou sessão usada há 2 h: linhas = %d (%v)", n, err)
	}
	row, err := q.GetActiveSession(t.Context(), GetActiveSessionParams{TokenHash: hash("b"), ValidAfter: t0})
	if err != nil || !row.LastUsedAt.Equal(t0.Add(2*time.Hour)) {
		t.Fatalf("último uso = %v (%v), esperado %v", row.LastUsedAt, err, t0.Add(2*time.Hour))
	}
}

// RN-11 / CA-03.2: sair apaga só a Sessão deste navegador.
func TestSessions_RN11_DeleteOnlyThisSession(t *testing.T) {
	q, _ := setup(t)
	u := upsert(t, q, "555", "valkyrja", "", t0)
	for _, tok := range []string{"navegador-1", "navegador-2"} {
		if err := q.CreateSession(t.Context(), CreateSessionParams{TokenHash: hash(tok), UserID: u.ID, Now: t0}); err != nil {
			t.Fatal(err)
		}
	}
	if err := q.DeleteSession(t.Context(), hash("navegador-1")); err != nil {
		t.Fatal(err)
	}
	valid := GetActiveSessionParams{ValidAfter: t0.Add(-time.Hour)}
	valid.TokenHash = hash("navegador-1")
	if _, err := q.GetActiveSession(t.Context(), valid); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("sessão apagada continua valendo: %v", err)
	}
	valid.TokenHash = hash("navegador-2")
	if _, err := q.GetActiveSession(t.Context(), valid); err != nil {
		t.Errorf("a outra sessão deveria continuar: %v", err)
	}
}

// RN-10: apagar o Usuário apaga as sessões dele.
func TestSessions_RN10_DeletedUserEndsSessions(t *testing.T) {
	q, pool := setup(t)
	u := upsert(t, q, "666", "lyrae", "", t0)
	if err := q.CreateSession(t.Context(), CreateSessionParams{TokenHash: hash("c"), UserID: u.ID, Now: t0}); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, u.ID); err != nil {
		t.Fatal(err)
	}
	_, err := q.GetActiveSession(t.Context(), GetActiveSessionParams{TokenHash: hash("c"), ValidAfter: t0.Add(-time.Hour)})
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("sessão de usuário apagado continua valendo: %v", err)
	}
}
