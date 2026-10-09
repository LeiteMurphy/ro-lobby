//go:build integration

package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/applications"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/auth"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/characters"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/migrate"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

// Fluxo das rotas de lobbies com o PostgreSQL real e duas sessões (spec lobbies, CA-01.1,
// CA-02.2, CA-03.1, CA-04.1, CA-04.4, CA-05.1, CA-05.3, CA-06.4).
func TestLobbiesFlowIntegration_CA01_1_CA04_1_CA05_1_CA06_4(t *testing.T) {
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

	q := db.New(pool)
	session := func(discordID string) string {
		u, err := q.UpsertUserByDiscordID(t.Context(), db.UpsertUserByDiscordIDParams{DiscordID: discordID, Username: "u" + discordID, Now: time.Now()})
		if err != nil {
			t.Fatal(err)
		}
		raw := make([]byte, 32)
		_, _ = rand.Read(raw)
		token := base64.RawURLEncoding.EncodeToString(raw)
		if err := q.CreateSession(t.Context(), db.CreateSessionParams{TokenHash: auth.HashToken(token), UserID: u.ID, Now: time.Now()}); err != nil {
			t.Fatal(err)
		}
		return token
	}
	ana, bia := session("1"), session("2")
	h := New(pool, auth.NewService(pool, nil), characters.NewService(pool), lobbies.NewService(pool), applications.NewService(pool), talents.NewService(pool))

	rec, lirien := call(t, h, http.MethodPost, "/characters", ana, `{"nick":"LirienL","classId":"arcebispo","level":178,"role":"support"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("personagem: %d %v", rec.Code, lirien)
	}
	tomorrow := time.Now().In(lobbies.Location).AddDate(0, 0, 1)
	day := tomorrow.Format(time.DateOnly)
	start := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 20, 0, 0, 0, lobbies.Location).UTC().Format(time.RFC3339)
	body := fmt.Sprintf(`{"instanceId":"templo-do-demonio-rei","startsAt":%q,"slots":{"tank":1,"support":2,"dps":3},"minLevel":160,"characterId":%q}`, start, lirien["id"])

	// CA-01.1: criar.
	rec, lobby := call(t, h, http.MethodPost, "/lobbies", ana, body)
	if rec.Code != http.StatusCreated || lobby["status"] != "open" {
		t.Fatalf("criar: %d %v", rec.Code, lobby)
	}
	path := "/lobbies/" + lobby["id"].(string)

	// RN-22 / CA-03.1: lista e detalhe públicos.
	rec, _ = call(t, h, http.MethodGet, "/lobbies?from="+day+"&to="+day, "", "")
	var list []map[string]any
	decode(t, rec.Body.Bytes(), &list)
	if rec.Code != http.StatusOK || len(list) != 1 || list[0]["id"] != lobby["id"] {
		t.Errorf("lista: %d %v", rec.Code, list)
	}
	if rec, got := call(t, h, http.MethodGet, path, "", ""); rec.Code != http.StatusOK || got["owner"].(map[string]any)["nick"] != "LirienL" {
		t.Errorf("detalhe: %d %v", rec.Code, got)
	}

	// CA-06.4: o personagem do dono fica travado.
	charPath := "/characters/" + lirien["id"].(string)
	if rec, got := call(t, h, http.MethodDelete, charPath, ana, ""); rec.Code != http.StatusConflict || got["error"] != "character_in_open_lobby" {
		t.Errorf("excluir personagem: %d %v", rec.Code, got)
	}

	// CA-04.4 / CA-05.3: a outra conta não edita nem cancela.
	upd := fmt.Sprintf(`{"startsAt":%q,"slots":{"tank":1,"support":2,"dps":5},"minLevel":170}`, start)
	if rec, _ := call(t, h, http.MethodPut, path, bia, upd); rec.Code != http.StatusNotFound {
		t.Errorf("editar de outra conta: %d", rec.Code)
	}
	if rec, _ := call(t, h, http.MethodPost, path+"/cancel", bia, `{"reason":"Cancelando o dos outros"}`); rec.Code != http.StatusNotFound {
		t.Errorf("cancelar de outra conta: %d", rec.Code)
	}

	// CA-04.1: o dono edita.
	if rec, got := call(t, h, http.MethodPut, path, ana, upd); rec.Code != http.StatusOK || got["minLevel"] != float64(170) {
		t.Errorf("editar: %d %v", rec.Code, got)
	}

	// CA-05.1 / CA-02.2: o dono cancela; o lobby sai da lista e o personagem fica livre.
	if rec, got := call(t, h, http.MethodPost, path+"/cancel", ana, `{"reason":"Metade do grupo não pode"}`); rec.Code != http.StatusOK || got["status"] != "cancelled" {
		t.Errorf("cancelar: %d %v", rec.Code, got)
	}
	rec, _ = call(t, h, http.MethodGet, "/lobbies?from="+day+"&to="+day, "", "")
	if rec.Body.String() != "[]\n" {
		t.Errorf("lista depois de cancelar: %s", rec.Body)
	}
	if rec, _ := call(t, h, http.MethodDelete, charPath, ana, ""); rec.Code != http.StatusNoContent {
		t.Errorf("excluir personagem depois de cancelar: %d", rec.Code)
	}
}
