//go:build integration

package discordroom

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordbot"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordfake"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/migrate"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

// Trabalhador da sala contra o PostgreSQL real e o Discord falso (spec discord-bot, T-03).

const guildID = "1560117933705797643"

// O lobby dos testes começa às 20:00 de Brasília de 20/10/2026 (23:00 UTC).
var startsAt = time.Date(2026, 10, 20, 23, 0, 0, 0, time.UTC)

type env struct {
	svc   *Service
	fake  *discordfake.Fake
	bot   *discordbot.Client
	pool  *pgxpool.Pool
	q     *db.Queries
	clock time.Time
	ids   atomic.Int64
}

func setup(t *testing.T) *env {
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

	fake := discordfake.New("id", "segredo", discordfake.User{ID: "1", Username: "u"})
	fake.BotToken = "token-do-bot"
	srv := httptest.NewServer(fake.Handler())
	t.Cleanup(srv.Close)
	bot := discordbot.New(srv.URL+"/api", "token-do-bot", guildID, "111", discordbot.WithTimeout(2*time.Second))

	e := &env{fake: fake, bot: bot, pool: pool, q: db.New(pool), clock: startsAt.Add(-3 * time.Hour)}
	e.ids.Store(500)
	e.svc = NewService(pool, bot)
	e.svc.Now = func() time.Time { return e.clock }
	return e
}

func (e *env) exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := e.pool.Exec(t.Context(), sql, args...); err != nil {
		t.Fatal(err)
	}
}

type person struct {
	id        pgtype.UUID
	discordID string
	character pgtype.UUID
}

// user cria um Usuário com um personagem; connected marca a conexão ao servidor.
func (e *env) user(t *testing.T, nick string, connected bool) person {
	t.Helper()
	n := e.ids.Add(1)
	p := person{discordID: fmt.Sprint(n)}
	var conn any
	if connected {
		conn = e.clock
	}
	if err := e.pool.QueryRow(t.Context(), `INSERT INTO users (discord_id, username, created_at, last_login_at, discord_connected_at)
		VALUES ($1, $2, now(), now(), $3) RETURNING id`, p.discordID, strings.ToLower(nick), conn).Scan(&p.id); err != nil {
		t.Fatal(err)
	}
	if err := e.pool.QueryRow(t.Context(), `INSERT INTO characters (user_id, nick, class_id, level, role, portrait, created_at, server)
		VALUES ($1, $2, 'arcebispo', 200, 'support', 'retrato-1', now(), 'freya') RETURNING id`, p.id, nick).Scan(&p.character); err != nil {
		t.Fatal(err)
	}
	return p
}

// lobby cria o lobby do dono às startsAt.
func (e *env) lobby(t *testing.T, owner person) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := e.pool.QueryRow(t.Context(), `INSERT INTO lobbies (owner_id, instance_id, instance_name, instance_level, starts_at,
		    slots_tank, slots_support, slots_dps, min_level, owner_character_id, owner_role, created_at, server)
		VALUES ($1, 'templo-do-demonio-rei', 'Templo do Demônio Rei', 160, $2, 1, 2, 3, 160, $3, 'support', now(), 'freya')
		RETURNING id`, owner.id, startsAt, owner.character).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func (e *env) member(t *testing.T, lobby pgtype.UUID, p person) {
	t.Helper()
	e.exec(t, `INSERT INTO applications (lobby_id, user_id, character_id, role, status, created_at, decided_at)
		VALUES ($1, $2, $3, 'support', 'accepted', now(), now())`, lobby, p.id, p.character)
}

// room cria a sala no Discord falso e grava aberta em openedAt.
func (e *env) room(t *testing.T, lobby pgtype.UUID, openedAt time.Time, allow ...string) discordbot.Room {
	t.Helper()
	r, err := e.bot.CreateRoom(t.Context(), "Templo", allow)
	if err != nil {
		t.Fatal(err)
	}
	e.exec(t, `INSERT INTO discord_rooms (lobby_id, category_id, text_channel_id, voice_channel_id, status, opened_at)
		VALUES ($1, $2, $3, $4, 'open', $5)`, lobby, r.CategoryID, r.TextID, r.VoiceID, openedAt)
	return r
}

func (e *env) tick(t *testing.T) {
	t.Helper()
	if err := e.svc.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func (e *env) status(t *testing.T, lobby pgtype.UUID) string {
	t.Helper()
	var s string
	if err := e.pool.QueryRow(t.Context(), `SELECT status FROM discord_rooms WHERE lobby_id = $1`, lobby).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

func brt(h, m int) time.Time {
	return startsAt.Add(time.Duration(h-20)*time.Hour + time.Duration(m)*time.Minute)
}

// CA-06.1 / RN-14: às 19:45, o lembrete menciona só os conectados, uma vez só.
func TestReminder_CA06_1_OnceAt1945(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "Ana", true)
	bia := e.user(t, "Bia", true)
	caio := e.user(t, "Caio", false)
	lobby := e.lobby(t, ana)
	e.member(t, lobby, bia)
	e.member(t, lobby, caio)
	r := e.room(t, lobby, brt(18, 0))

	e.clock = brt(19, 44)
	e.tick(t)
	if n := len(e.fake.Messages(r.TextID)); n != 0 {
		t.Fatalf("às 19:44: %d mensagens", n)
	}
	e.clock = brt(19, 45)
	e.tick(t)
	e.clock = brt(19, 50)
	e.tick(t)
	msgs := e.fake.Messages(r.TextID)
	if len(msgs) != 1 {
		t.Fatalf("lembretes = %d, quer 1", len(msgs))
	}
	m := msgs[0]
	if !strings.Contains(m.Content, "15 minutos") || !strings.Contains(m.Content, "20:00") {
		t.Errorf("texto = %q", m.Content)
	}
	if got := strings.Join(m.AllowedMentions.Users, ","); got != ana.discordID+","+bia.discordID {
		t.Errorf("menções = %s", got)
	}
	if discordfake.HasMention(m.Content, caio.discordID) {
		t.Error("menciona quem não está conectado")
	}
}

// CA-06.2 / RN-14: a sala aberta às 19:50 não recebe lembrete.
func TestReminder_CA06_2_OpenedLate(t *testing.T) {
	e := setup(t)
	lobby := e.lobby(t, e.user(t, "Ana", true))
	r := e.room(t, lobby, brt(19, 50))
	e.clock = brt(19, 55)
	e.tick(t)
	if n := len(e.fake.Messages(r.TextID)); n != 0 {
		t.Errorf("%d mensagens", n)
	}
}

// CA-06.3 / RN-14: lobby cancelado às 19:30 não lembra; depois do início, também não.
func TestReminder_CA06_3_CancelledOrStarted(t *testing.T) {
	e := setup(t)
	cancelled := e.lobby(t, e.user(t, "Ana", true))
	rc := e.room(t, cancelled, brt(18, 0))
	e.exec(t, `UPDATE lobbies SET cancelled_at = $1, cancel_reason = 'Faltou o tank, fica para amanhã' WHERE id = $2`, brt(19, 30), cancelled)
	late := e.lobby(t, e.user(t, "Bia", true))
	rl := e.room(t, late, brt(18, 0))

	e.clock = brt(19, 45)
	e.exec(t, `UPDATE discord_rooms SET reminder_sent_at = NULL`)
	e.tick(t)
	if n := len(e.fake.Messages(rc.TextID)); n != 0 {
		t.Errorf("cancelado: %d mensagens", n)
	}
	// O trabalhador ficou parado e só volta às 20:05: o lembrete já não serve.
	e.exec(t, `UPDATE discord_rooms SET reminder_sent_at = NULL WHERE lobby_id = $1`, late)
	e.fake.SetBotFailures(discordfake.BotFailures{})
	before := len(e.fake.Messages(rl.TextID))
	e.clock = brt(20, 5)
	e.tick(t)
	if n := len(e.fake.Messages(rl.TextID)); n != before {
		t.Errorf("depois do início: %d mensagens novas", n-before)
	}
}

// CA-05.1 / CA-05.2 / RN-16: apagada às 02:00, não às 01:59.
func TestDeletion_CA05_1_CA05_2_SixHoursAfterStart(t *testing.T) {
	e := setup(t)
	lobby := e.lobby(t, e.user(t, "Ana", true))
	r := e.room(t, lobby, brt(18, 0))
	e.clock = startsAt.Add(6*time.Hour - time.Minute)
	e.tick(t)
	if e.status(t, lobby) != StatusOpen || e.fake.ChannelCount() != 3 {
		t.Fatalf("às 01:59: %s, %d canais", e.status(t, lobby), e.fake.ChannelCount())
	}
	e.clock = startsAt.Add(6 * time.Hour)
	e.tick(t)
	if e.status(t, lobby) != StatusDeleted {
		t.Errorf("às 02:00: %s", e.status(t, lobby))
	}
	for _, ch := range r.Channels() {
		if _, ok := e.fake.Channel(ch); ok {
			t.Errorf("canal %s ainda existe", ch)
		}
	}
}

// RN-17: a sala encerrando é apagada 1 h depois do cancelamento.
func TestDeletion_RN17_ClosingAfterOneHour(t *testing.T) {
	e := setup(t)
	lobby := e.lobby(t, e.user(t, "Ana", true))
	e.room(t, lobby, brt(18, 0))
	e.exec(t, `UPDATE discord_rooms SET status = 'closing', closing_at = $1`, brt(17, 0))
	e.clock = brt(17, 59)
	e.tick(t)
	if e.status(t, lobby) != StatusClosing {
		t.Fatalf("antes de 1 h: %s", e.status(t, lobby))
	}
	e.clock = brt(18, 0)
	e.tick(t)
	if e.status(t, lobby) != StatusDeleted || e.fake.ChannelCount() != 0 {
		t.Errorf("depois de 1 h: %s, %d canais", e.status(t, lobby), e.fake.ChannelCount())
	}
}

// CA-05.4 / RN-21: com o Discord fora, a remoção fica pendente e é feita quando ele volta.
func TestDeletion_CA05_4_RetriedWhenDiscordIsBack(t *testing.T) {
	e := setup(t)
	lobby := e.lobby(t, e.user(t, "Ana", true))
	e.room(t, lobby, brt(18, 0))
	e.clock = startsAt.Add(7 * time.Hour)
	e.fake.SetBotFailures(discordfake.BotFailures{Down: true})
	e.tick(t)
	if e.status(t, lobby) != StatusOpen {
		t.Fatalf("com o Discord fora: %s", e.status(t, lobby))
	}
	e.fake.SetBotFailures(discordfake.BotFailures{})
	e.tick(t)
	if e.status(t, lobby) != StatusDeleted || e.fake.ChannelCount() != 0 {
		t.Errorf("com o Discord de volta: %s, %d canais", e.status(t, lobby), e.fake.ChannelCount())
	}
}

// CA-05.5 / RN-09: canais apagados à mão não travam a remoção.
func TestDeletion_CA05_5_ChannelsDeletedByHand(t *testing.T) {
	e := setup(t)
	lobby := e.lobby(t, e.user(t, "Ana", true))
	r := e.room(t, lobby, brt(18, 0))
	for _, ch := range r.Channels() {
		e.fake.DeleteChannelDirect(ch)
	}
	e.clock = startsAt.Add(6 * time.Hour)
	e.tick(t)
	if e.status(t, lobby) != StatusDeleted {
		t.Errorf("status = %s", e.status(t, lobby))
	}
}

// RN-21 / RNF-04: o job que falha é refeito depois, com espera crescente, e postado uma vez só.
func TestJobs_RN21_RetryWithBackoffAndRunOnce(t *testing.T) {
	e := setup(t)
	lobby := e.lobby(t, e.user(t, "Ana", true))
	r := e.room(t, lobby, brt(17, 0))
	if err := enqueuePost(t.Context(), e.q, lobby, postPayload{Content: "oi"}, e.clock); err != nil {
		t.Fatal(err)
	}
	e.fake.SetBotFailures(discordfake.BotFailures{Down: true})
	e.tick(t)
	var attempts int
	var runAt time.Time
	if err := e.pool.QueryRow(t.Context(), `SELECT attempts, run_at FROM discord_jobs`).Scan(&attempts, &runAt); err != nil {
		t.Fatal(err)
	}
	if attempts != 1 || !runAt.Equal(e.clock.Add(2*time.Second)) {
		t.Fatalf("depois da falha: tentativa %d, próxima em %v", attempts, runAt.Sub(e.clock))
	}
	e.fake.SetBotFailures(discordfake.BotFailures{})
	e.tick(t) // ainda não venceu
	if n := len(e.fake.Messages(r.TextID)); n != 0 {
		t.Fatalf("antes da espera: %d mensagens", n)
	}
	e.clock = e.clock.Add(2 * time.Second)
	e.tick(t)
	e.tick(t)
	if n := len(e.fake.Messages(r.TextID)); n != 1 {
		t.Errorf("mensagens = %d, quer 1", n)
	}
	if backoff(1) != 4*time.Second || backoff(20) != maxBackoff {
		t.Errorf("espera: %v, %v", backoff(1), backoff(20))
	}
}

// RN-21: o job de uma sala já apagada é descartado sem chamar o Discord.
func TestJobs_RN21_DiscardedWhenRoomDeleted(t *testing.T) {
	e := setup(t)
	lobby := e.lobby(t, e.user(t, "Ana", true))
	e.room(t, lobby, brt(17, 0))
	if err := enqueuePost(t.Context(), e.q, lobby, postPayload{Content: "oi"}, e.clock); err != nil {
		t.Fatal(err)
	}
	e.exec(t, `UPDATE discord_rooms SET status = 'deleted', deleted_at = now()`)
	calls := e.fake.BotCalls()
	e.tick(t)
	var done bool
	if err := e.pool.QueryRow(t.Context(), `SELECT done_at IS NOT NULL FROM discord_jobs`).Scan(&done); err != nil || !done {
		t.Errorf("job não descartado: %v", err)
	}
	if e.fake.BotCalls() != calls {
		t.Error("chamou o Discord para uma sala apagada")
	}
}

// RN-07 / D-02: o sync_access decide na hora: membro conectado ganha acesso; quem não é
// membro perde.
func TestJobs_D02_SyncAccessDecidesAtRunTime(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "Ana", true)
	bia := e.user(t, "Bia", true)
	caio := e.user(t, "Caio", true)
	lobby := e.lobby(t, ana)
	e.member(t, lobby, bia)
	r := e.room(t, lobby, brt(17, 0), caio.discordID) // Caio tinha acesso, mas não é membro
	for _, p := range []person{bia, caio} {
		if err := enqueueSync(t.Context(), e.q, lobby, p.id, e.clock); err != nil {
			t.Fatal(err)
		}
	}
	e.tick(t)
	for _, ch := range r.Channels() {
		c, _ := e.fake.Channel(ch)
		if _, ok := c.Overwrites[bia.discordID]; !ok {
			t.Errorf("%s: Bia sem acesso", ch)
		}
		if _, ok := c.Overwrites[caio.discordID]; ok {
			t.Errorf("%s: Caio ainda com acesso", ch)
		}
	}
}
