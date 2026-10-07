//go:build integration

package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/applications"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/auth"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/characters"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/migrate"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

// Fluxo da candidatura pelas rotas, com o PostgreSQL real e quatro sessões: dono,
// candidato, outro candidato e visitante (spec candidatura-lobby, T-04 e T-08: CA-03.1,
// CA-03.7, CA-03.8, CA-10.2, CA-10.3, RN-28, RN-29, RN-32, D-06, D-07).
// flow é a API inteira sobre um banco descartável, com sessões criadas direto no banco.
type flow struct {
	h http.Handler
	q *db.Queries
}

func newFlow(t *testing.T) flow {
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
	h := New(pool, auth.NewService(pool, nil), characters.NewService(pool), lobbies.NewService(pool), applications.NewService(pool))
	return flow{h: h, q: db.New(pool)}
}

// session cria o Usuário com o nome do Discord e devolve o token de uma sessão dele.
func (f flow) session(t *testing.T, discordID, name string) string {
	t.Helper()
	u, err := f.q.UpsertUserByDiscordID(t.Context(), db.UpsertUserByDiscordIDParams{
		DiscordID: discordID, Username: "u" + discordID, GlobalName: pgtype.Text{String: name, Valid: true}, Now: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	raw := make([]byte, 32)
	_, _ = rand.Read(raw)
	token := base64.RawURLEncoding.EncodeToString(raw)
	if err := f.q.CreateSession(t.Context(), db.CreateSessionParams{TokenHash: auth.HashToken(token), UserID: u.ID, Now: time.Now()}); err != nil {
		t.Fatal(err)
	}
	return token
}

func (f flow) character(t *testing.T, token, body string) string {
	t.Helper()
	rec, got := call(t, f.h, http.MethodPost, "/characters", token, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("personagem: %d %v", rec.Code, got)
	}
	return got["id"].(string)
}

// lobby cria o lobby do Templo amanhã às 20:00 (1 Tank, 2 Suportes, 3 Danos) e devolve o
// caminho dele e o dia em São Paulo.
func (f flow) lobby(t *testing.T, token, characterID string) (string, string) {
	t.Helper()
	tomorrow := time.Now().In(lobbies.Location).AddDate(0, 0, 1)
	start := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 20, 0, 0, 0, lobbies.Location).UTC().Format(time.RFC3339)
	rec, lobby := call(t, f.h, http.MethodPost, "/lobbies", token,
		fmt.Sprintf(`{"instanceId":"templo-do-demonio-rei","startsAt":%q,"slots":{"tank":1,"support":2,"dps":3},"minLevel":160,"characterId":%q}`, start, characterID))
	if rec.Code != http.StatusCreated {
		t.Fatalf("lobby: %d %v", rec.Code, lobby)
	}
	return "/lobbies/" + lobby["id"].(string), tomorrow.Format(time.DateOnly)
}

func TestApplicationsFlowIntegration_CA03_7_CA03_8_CA10_2_CA10_3(t *testing.T) {
	f := newFlow(t)
	h := f.h
	ana, bia, caio, duda := f.session(t, "1", "Ana"), f.session(t, "2", "Bia"), f.session(t, "3", "Caio"), f.session(t, "4", "Duda")
	host := f.character(t, ana, `{"nick":"Anfitria","classId":"arcebispo","level":200,"role":"support"}`)
	brasa := f.character(t, bia, `{"nick":"Brasa","classId":"guardiao-real","level":200,"role":"tank","link":"https://ragnaplace.com/brasa"}`)
	fogo := f.character(t, caio, `{"nick":"Fogo","classId":"arquimago","level":200,"role":"dps"}`)
	path, day := f.lobby(t, ana, host)

	// Candidaturas de Bia (com mensagem) e Caio.
	rec, app := call(t, h, http.MethodPost, path+"/applications", bia, fmt.Sprintf(`{"characterId":%q,"message":"tenho buff de ASPD"}`, brasa))
	if rec.Code != http.StatusCreated || app["status"] != "pending" || app["role"] != "tank" {
		t.Fatalf("candidatar: %d %v", rec.Code, app)
	}
	rec, other := call(t, h, http.MethodPost, path+"/applications", caio, fmt.Sprintf(`{"characterId":%q}`, fogo))
	if rec.Code != http.StatusCreated {
		t.Fatalf("candidatar Caio: %d %v", rec.Code, other)
	}
	// D-07: segunda candidatura ativa é 409 com código; personagem de outro, 422.
	if rec, got := call(t, h, http.MethodPost, path+"/applications", bia, fmt.Sprintf(`{"characterId":%q}`, brasa)); rec.Code != http.StatusConflict || got["code"] != "already_active" {
		t.Errorf("segunda ativa: %d %v", rec.Code, got)
	}
	if rec, got := call(t, h, http.MethodPost, path+"/applications", duda, fmt.Sprintf(`{"characterId":%q}`, brasa)); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("personagem de outro: %d %v", rec.Code, got)
	}

	// CA-03.7 / CA-10.3: o visitante e o terceiro veem só a quantidade; o dono vê os
	// pendentes com mensagem e Discord; quem se candidatou vê a própria.
	for name, token := range map[string]string{"sem sessão": "", "Duda": duda} {
		_, got := call(t, h, http.MethodGet, path, token, "")
		if got["pendingCount"] != float64(2) || got["pending"] != nil || got["myApplication"] != nil || len(got["members"].([]any)) != 0 {
			t.Errorf("%s: %v", name, got)
		}
	}
	_, got := call(t, h, http.MethodGet, path, ana, "")
	pending, _ := got["pending"].([]any)
	if len(pending) != 2 {
		t.Fatalf("pendentes do dono: %v", got["pending"])
	}
	first := pending[0].(map[string]any)
	if first["nick"] != "Brasa" || first["message"] != "tenho buff de ASPD" || first["discordName"] != "Bia" || first["link"] != "https://ragnaplace.com/brasa" {
		t.Errorf("pendente = %v", first)
	}
	_, got = call(t, h, http.MethodGet, path, bia, "")
	if mine, _ := got["myApplication"].(map[string]any); mine["id"] != app["id"] || mine["status"] != "pending" || got["pending"] != nil {
		t.Errorf("candidato: %v", got)
	}

	// CA-02.6: quem não é dono não decide.
	appPath := "/applications/" + app["id"].(string)
	if rec, got := call(t, h, http.MethodPost, appPath+"/accept", duda, ""); rec.Code != http.StatusConflict || got["code"] != "not_owner" {
		t.Errorf("aceite de terceiro: %d %v", rec.Code, got)
	}
	// CA-02.1: o dono aceita Bia; CA-02.2: recusa Caio com justificativa.
	if rec, got := call(t, h, http.MethodPost, appPath+"/accept", ana, ""); rec.Code != http.StatusOK || got["status"] != "accepted" {
		t.Fatalf("aceitar: %d %v", rec.Code, got)
	}
	otherPath := "/applications/" + other["id"].(string)
	if rec, got := call(t, h, http.MethodPost, otherPath+"/reject", ana, `{"reason":"   curta  "}`); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("justificativa curta: %d %v", rec.Code, got)
	}
	if rec, got := call(t, h, http.MethodPost, otherPath+"/reject", ana, `{"reason":"precisamos de mais suporte"}`); rec.Code != http.StatusOK || got["reason"] != "precisamos de mais suporte" {
		t.Errorf("recusar: %d %v", rec.Code, got)
	}

	// CA-10.2: o Discord do membro só para o dono e os membros.
	discordOf := func(token string) any {
		_, got := call(t, h, http.MethodGet, path, token, "")
		members := got["members"].([]any)
		if len(members) != 1 || got["occupied"].(map[string]any)["tank"] != float64(1) || got["pendingCount"] != float64(0) {
			t.Fatalf("membros: %v", got)
		}
		return members[0].(map[string]any)["discordName"]
	}
	for name, c := range map[string]struct {
		token string
		want  any
	}{"sem sessão": {"", nil}, "Duda": {duda, nil}, "Caio (recusado)": {caio, nil}, "Bia (membro)": {bia, "Bia"}, "Ana (dono)": {ana, "Bia"}} {
		if got := discordOf(c.token); got != c.want {
			t.Errorf("%s vê Discord %v, quer %v", name, got, c.want)
		}
	}

	// RN-32: o Discord do anfitrião segue a mesma regra, na lista e no detalhe.
	for name, c := range map[string]struct {
		token string
		want  any
	}{"sem sessão": {"", nil}, "Duda": {duda, nil}, "Caio (recusado)": {caio, nil}, "Bia (membro)": {bia, "Ana"}, "Ana (dono)": {ana, "Ana"}} {
		_, got := call(t, h, http.MethodGet, path, c.token, "")
		if d := got["owner"].(map[string]any)["discordName"]; d != c.want {
			t.Errorf("%s vê o Discord do anfitrião %v, quer %v", name, d, c.want)
		}
	}
	rec, _ = call(t, h, http.MethodGet, "/lobbies?from="+day+"&to="+day, ana, "")
	var public []map[string]any
	decode(t, rec.Body.Bytes(), &public)
	if len(public) != 1 || public[0]["owner"].(map[string]any)["discordName"] != nil {
		t.Errorf("lista pública com Discord: %v", public)
	}

	// CA-03.8 / RN-29: a justificativa só para o candidato afetado.
	_, got = call(t, h, http.MethodGet, path, caio, "")
	if mine, _ := got["myApplication"].(map[string]any); mine["status"] != "rejected" || mine["reason"] != "precisamos de mais suporte" {
		t.Errorf("recusado: %v", got["myApplication"])
	}
	_, got = call(t, h, http.MethodGet, path, duda, "")
	if got["myApplication"] != nil {
		t.Errorf("terceiro: %v", got)
	}

	// CA-03.1: minhas candidaturas; CA-03.3: retirar aceita é 409; candidatura
	// inexistente é 404.
	rec, _ = call(t, h, http.MethodGet, "/me/applications", caio, "")
	var list []map[string]any
	decode(t, rec.Body.Bytes(), &list)
	if rec.Code != http.StatusOK || len(list) != 1 || list[0]["application"].(map[string]any)["status"] != "rejected" ||
		list[0]["character"].(map[string]any)["nick"] != "Fogo" || list[0]["lobby"].(map[string]any)["instanceName"] != "Templo do Demônio Rei" {
		t.Errorf("minhas: %d %v", rec.Code, list)
	}
	if rec, got := call(t, h, http.MethodPost, appPath+"/withdraw", bia, ""); rec.Code != http.StatusConflict || got["code"] != "not_pending" {
		t.Errorf("retirar aceita: %d %v", rec.Code, got)
	}
	if rec, _ := call(t, h, http.MethodPost, "/applications/00000000-0000-0000-0000-000000000000/accept", ana, ""); rec.Code != http.StatusNotFound {
		t.Errorf("inexistente: %d", rec.Code)
	}

	// CA-09.1: o personagem aceito fica com nível travado.
	if rec, got := call(t, h, http.MethodPut, "/characters/"+brasa, bia, `{"nick":"Brasa","classId":"guardiao-real","level":201,"role":"tank"}`); rec.Code != http.StatusConflict || got["error"] != "character_in_open_lobby" {
		t.Errorf("mudar nível do membro: %d %v", rec.Code, got)
	}
}
