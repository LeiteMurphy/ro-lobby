//go:build integration

package server

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/auth"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/characters"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/migrate"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

// Fluxo das rotas de personagem com o PostgreSQL real e duas sessões (spec personagens,
// CA-01.3, CA-02.1, CA-02.3, CA-03.3, CA-04.3, CA-07.2).
func TestCharactersFlowIntegration(t *testing.T) {
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
		now := time.Now()
		u, err := q.UpsertUserByDiscordID(t.Context(), db.UpsertUserByDiscordIDParams{DiscordID: discordID, Username: "u" + discordID, Now: now})
		if err != nil {
			t.Fatal(err)
		}
		raw := make([]byte, 32)
		_, _ = rand.Read(raw)
		token := base64.RawURLEncoding.EncodeToString(raw)
		if err := q.CreateSession(t.Context(), db.CreateSessionParams{TokenHash: auth.HashToken(token), UserID: u.ID, Now: now}); err != nil {
			t.Fatal(err)
		}
		return token
	}
	ana, bia := session("1"), session("2")
	h := New(pool, auth.NewService(pool, nil), characters.NewService(pool))

	// CA-02.1: o primeiro personagem nasce principal.
	rec, brasa := call(t, h, http.MethodPost, "/characters", ana, `{"nick":"Brasa","classId":"guardiao-real","level":172,"role":"tank"}`)
	if rec.Code != http.StatusCreated || brasa["isMain"] != true || brasa["portrait"] != "retrato-1" {
		t.Fatalf("criar Brasa: status %d, corpo %v", rec.Code, brasa)
	}
	rec, lirien := call(t, h, http.MethodPost, "/characters", ana, `{"nick":"Lirien","classId":"arcebispo","level":178,"role":"support","portrait":"retrato-2"}`)
	if rec.Code != http.StatusCreated || lirien["isMain"] != false {
		t.Fatalf("criar Lirien: status %d, corpo %v", rec.Code, lirien)
	}

	// CA-02.3: o nick é recusado para outro Usuário.
	rec, body := call(t, h, http.MethodPost, "/characters", bia, `{"nick":"BRASA","classId":"paladino","level":99,"role":"tank"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("nick repetido: status %d, corpo %v", rec.Code, body)
	}

	// CA-01.3: cada um vê só os seus.
	rec, _ = call(t, h, http.MethodGet, "/characters", bia, "")
	if rec.Code != http.StatusOK || rec.Body.String() != "[]\n" {
		t.Errorf("lista da bia: status %d, corpo %s", rec.Code, rec.Body)
	}

	// CA-03.3: a bia não edita, não exclui e não marca o personagem da ana.
	lirienPath := "/characters/" + lirien["id"].(string)
	for _, r := range []struct{ method, path, body string }{
		{http.MethodPut, lirienPath, `{"nick":"Roubado","classId":"aprendiz","level":1,"role":"dps"}`},
		{http.MethodDelete, lirienPath, ""},
		{http.MethodPut, lirienPath + "/main", ""},
	} {
		if rec, _ := call(t, h, r.method, r.path, bia, r.body); rec.Code != http.StatusNotFound {
			t.Errorf("%s %s pela bia: status %d", r.method, r.path, rec.Code)
		}
	}

	// CA-05.1 e CA-04.3: Lirien vira principal; excluída, o principal volta para Brasa.
	if rec, _ := call(t, h, http.MethodPut, lirienPath+"/main", ana, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("marcar principal: status %d", rec.Code)
	}
	if rec, _ := call(t, h, http.MethodDelete, lirienPath, ana, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("excluir: status %d", rec.Code)
	}
	rec, _ = call(t, h, http.MethodGet, "/characters", ana, "")
	var list []map[string]any
	decode(t, rec.Body.Bytes(), &list)
	if len(list) != 1 || list[0]["nick"] != "Brasa" || list[0]["isMain"] != true {
		t.Errorf("lista da ana = %v", list)
	}

	// CA-07.2: sem sessão, 401.
	if rec, _ := call(t, h, http.MethodGet, "/characters", "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("sem sessão: status %d", rec.Code)
	}
}
