//go:build integration

package applications

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/migrate"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

// "Agora" dos testes: terça, 6 out 2026, 16:40 em São Paulo (19:40 UTC).
var start = time.Date(2026, 10, 6, 19, 40, 0, 0, time.UTC)

// at monta um horário de São Paulo, daqui a `days` dias.
func at(days, hour, minute int) time.Time {
	y, m, d := start.In(lobbies.Location).Date()
	return time.Date(y, m, d+days, hour, minute, 0, 0, lobbies.Location)
}

type env struct {
	svc     *Service
	lobbies *lobbies.Service
	q       *db.Queries
	now     time.Time
	next    int
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
	e := &env{svc: NewService(pool), lobbies: lobbies.NewService(pool), q: db.New(pool), now: start}
	e.svc.Now = func() time.Time { return e.now }
	e.lobbies.Now = func() time.Time { return e.now }
	return e
}

type who struct {
	id    string
	chars map[string]string // nick → id
}

// user cria um Usuário com personagens "Nick:nível:função".
func (e *env) user(t *testing.T, chars ...string) who {
	t.Helper()
	e.next++
	u, err := e.q.UpsertUserByDiscordID(t.Context(), db.UpsertUserByDiscordIDParams{
		DiscordID: fmt.Sprint(e.next), Username: fmt.Sprint("u", e.next), Now: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	w := who{id: u.ID.String(), chars: map[string]string{}}
	for i, spec := range chars {
		parts := strings.Split(spec, ":")
		var level int16
		_, _ = fmt.Sscan(parts[1], &level)
		c, err := e.q.CreateCharacter(t.Context(), db.CreateCharacterParams{
			UserID: u.ID, Nick: parts[0], ClassID: "cavaleiro-runico", Level: level, Role: parts[2],
			Portrait: "retrato-1", IsMain: i == 0, Now: start,
		})
		if err != nil {
			t.Fatal(err)
		}
		w.chars[parts[0]] = c.ID.String()
	}
	return w
}

// lobby cria um lobby do Templo (nível 160) com o personagem do dono.
func (e *env) lobby(t *testing.T, owner who, nick string, startsAt time.Time, slots lobbies.Slots, minLevel int) string {
	t.Helper()
	l, err := e.lobbies.Create(t.Context(), owner.id, lobbies.Input{
		InstanceID: "templo-do-demonio-rei", StartsAt: startsAt, Slots: slots, MinLevel: minLevel,
		CharacterID: owner.chars[nick],
	})
	if err != nil {
		t.Fatal(err)
	}
	return l.ID
}

func (e *env) apply(t *testing.T, w who, nick, lobbyID string) Application {
	t.Helper()
	a, err := e.svc.Apply(t.Context(), w.id, lobbyID, ApplyInput{CharacterID: w.chars[nick]})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func (e *env) occupied(t *testing.T, lobbyID string) (lobbies.Slots, int) {
	t.Helper()
	l, err := e.lobbies.Get(t.Context(), lobbyID)
	if err != nil {
		t.Fatal(err)
	}
	return l.Occupied, l.PendingCount
}

func (e *env) status(t *testing.T, id string) string {
	t.Helper()
	var aid pgtype.UUID
	_ = aid.Scan(id)
	a, err := e.q.GetApplication(t.Context(), aid)
	if err != nil {
		t.Fatal(err)
	}
	return a.Status
}

func wantRule(t *testing.T, err error, code string) {
	t.Helper()
	var re *RuleError
	if !errors.As(err, &re) || re.Code != code {
		t.Errorf("err = %v, quer a regra %s", err, code)
	}
}

func wantField(t *testing.T, err error, field, code string) {
	t.Helper()
	var ve *lobbies.ValidationError
	if !errors.As(err, &ve) || !reflect.DeepEqual(ve.Fields, []lobbies.FieldError{{Field: field, Code: code}}) {
		t.Errorf("err = %v, quer %s/%s", err, field, code)
	}
}

var std = lobbies.Slots{Tank: 1, Support: 2, Dps: 3}

// basic: dono (Dps), um lobby amanhã às 20:00 e um candidato com Tank, Suporte e Dano.
func basic(t *testing.T) (*env, who, who, string) {
	e := setup(t)
	owner := e.user(t, "Dono:200:dps")
	player := e.user(t, "Escudo:200:tank", "Cura:200:support", "Fogo:200:dps", "Novato:170:dps")
	return e, owner, player, e.lobby(t, owner, "Dono", at(1, 20, 0), std, 160)
}

// CA-01.1: candidatura pendente com a função do personagem; a vaga continua livre.
func TestApply_CA01_1_Pending(t *testing.T) {
	e, _, player, lid := basic(t)
	a, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: player.chars["Cura"]})
	if err != nil {
		t.Fatal(err)
	}
	if a.Status != StatusPending || a.Role != "support" || a.CharacterID != player.chars["Cura"] || a.UserID != player.id {
		t.Errorf("candidatura = %+v", a)
	}
	occ, pending := e.occupied(t, lid)
	if occ != (lobbies.Slots{Dps: 1}) || pending != 1 {
		t.Errorf("ocupantes %+v, pendentes %d", occ, pending)
	}
}

// CA-01.2 e CA-01.3: mensagem de até 250 caracteres.
func TestApply_CA01_2_CA01_3_Message(t *testing.T) {
	e, _, player, lid := basic(t)
	_, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: player.chars["Cura"], Message: strings.Repeat("é", 251)})
	wantField(t, err, FieldMessage, lobbies.CodeTooLong)
	if _, pending := e.occupied(t, lid); pending != 0 {
		t.Error("não devia criar candidatura")
	}
	a, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: player.chars["Cura"], Message: strings.Repeat("é", 250)})
	if err != nil || a.Message != strings.Repeat("é", 250) {
		t.Errorf("250: %v", err)
	}
}

// CA-01.4: personagem de outro Usuário.
func TestApply_CA01_4_OtherUsersCharacter(t *testing.T) {
	e, owner, player, lid := basic(t)
	other := e.user(t, "Outro:200:support")
	_, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: other.chars["Outro"]})
	wantField(t, err, lobbies.FieldCharacterID, lobbies.CodeInvalid)
	_, err = e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: owner.chars["Dono"]})
	wantField(t, err, lobbies.FieldCharacterID, lobbies.CodeInvalid)
}

// CA-01.5: uma candidatura ativa por Usuário e lobby, com qualquer personagem.
func TestApply_CA01_5_AlreadyActive(t *testing.T) {
	e, _, player, lid := basic(t)
	e.apply(t, player, "Cura", lid)
	_, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: player.chars["Fogo"]})
	wantRule(t, err, CodeAlreadyActive)
}

// CA-01.6: o dono não se candidata ao próprio lobby.
func TestApply_CA01_6_OwnLobby(t *testing.T) {
	e := setup(t)
	owner := e.user(t, "Dono:200:dps", "Alt:200:tank")
	lid := e.lobby(t, owner, "Dono", at(1, 20, 0), std, 160)
	_, err := e.svc.Apply(t.Context(), owner.id, lid, ApplyInput{CharacterID: owner.chars["Alt"]})
	wantRule(t, err, CodeOwnLobby)
}

// CA-01.7: função sem vaga, contando dono e aceitos.
func TestApply_CA01_7_RoleFull(t *testing.T) {
	e := setup(t)
	owner := e.user(t, "Dono:200:tank")
	player := e.user(t, "Escudo:200:tank")
	lid := e.lobby(t, owner, "Dono", at(1, 20, 0), std, 160) // 1 vaga de Tank, a do dono
	_, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: player.chars["Escudo"]})
	wantRule(t, err, CodeRoleFull)
}

// CA-01.8: lobby iniciado, cancelado ou inexistente.
func TestApply_CA01_8_NotOpen(t *testing.T) {
	e, owner, player, lid := basic(t)
	if _, err := e.lobbies.Cancel(t.Context(), owner.id, lid, "imprevisto no trabalho"); err != nil {
		t.Fatal(err)
	}
	_, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: player.chars["Cura"]})
	wantRule(t, err, CodeNotOpen)

	lid2 := e.lobby(t, owner, "Dono", at(2, 20, 0), std, 160)
	e.now = at(2, 20, 0)
	_, err = e.svc.Apply(t.Context(), player.id, lid2, ApplyInput{CharacterID: player.chars["Cura"]})
	wantRule(t, err, CodeNotOpen)

	_, err = e.svc.Apply(t.Context(), player.id, "00000000-0000-0000-0000-000000000000", ApplyInput{CharacterID: player.chars["Cura"]})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("inexistente: %v", err)
	}
}

// CA-01.9: recusado não se candidata de novo ao mesmo lobby, com nenhum personagem.
func TestApply_CA01_9_RejectedBefore(t *testing.T) {
	e, owner, player, lid := basic(t)
	a := e.apply(t, player, "Cura", lid)
	if _, err := e.svc.Reject(t.Context(), owner.id, a.ID, "precisamos de mais dano"); err != nil {
		t.Fatal(err)
	}
	_, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: player.chars["Fogo"]})
	wantRule(t, err, CodeRejectedBefore)
}

// CA-01.10: pendência em lobby de horário próximo é permitida.
func TestApply_CA01_10_PendingNearbyAllowed(t *testing.T) {
	e, owner, player, lid := basic(t) // 20:00
	other := e.user(t, "Outro:200:dps")
	e.apply(t, player, "Cura", lid)
	if _, err := e.svc.Accept(t.Context(), owner.id, e.apply(t, player, "Cura", e.lobby(t, owner, "Dono", at(3, 20, 0), std, 160)).ID); err != nil {
		t.Fatal(err)
	}
	near := e.lobby(t, other, "Outro", at(1, 21, 0), std, 160)
	if a, err := e.svc.Apply(t.Context(), player.id, near, ApplyInput{CharacterID: player.chars["Cura"]}); err != nil || a.Status != StatusPending {
		t.Errorf("perto: %+v, %v", a, err)
	}
}

// CA-01.11: personagem abaixo do nível mínimo do lobby (RN-30).
func TestApply_CA01_11_BelowMinLevel(t *testing.T) {
	e := setup(t)
	owner := e.user(t, "Dono:250:dps")
	player := e.user(t, "Baixo:199:support", "Certo:200:support")
	lid := e.lobby(t, owner, "Dono", at(1, 20, 0), std, 200)
	_, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: player.chars["Baixo"]})
	wantRule(t, err, CodeBelowMinLevel)
	e.apply(t, player, "Certo", lid)
}

// CA-02.1: aceite ocupa a vaga.
func TestAccept_CA02_1(t *testing.T) {
	e := setup(t)
	owner := e.user(t, "Dono:200:tank")
	player := e.user(t, "Fogo:200:dps")
	lid := e.lobby(t, owner, "Dono", at(1, 20, 0), lobbies.Slots{Tank: 1, Dps: 1}, 160)
	a := e.apply(t, player, "Fogo", lid)
	got, err := e.svc.Accept(t.Context(), owner.id, a.ID)
	if err != nil || got.Status != StatusAccepted || got.DecidedAt.IsZero() {
		t.Fatalf("aceite: %+v, %v", got, err)
	}
	occ, pending := e.occupied(t, lid)
	if occ != (lobbies.Slots{Tank: 1, Dps: 1}) || pending != 0 {
		t.Errorf("ocupantes %+v, pendentes %d", occ, pending)
	}
}

// CA-02.2 a CA-02.5: recusa com justificativa de 10 a 250 caracteres, sem contar os
// espaços nas pontas.
func TestReject_CA02_2_to_CA02_5(t *testing.T) {
	e, owner, player, lid := basic(t)
	a := e.apply(t, player, "Cura", lid)
	for _, c := range []struct{ reason, code string }{
		{"", lobbies.CodeRequired},
		{"    ", lobbies.CodeRequired},
		{"  123456789  ", lobbies.CodeTooShort},
		{strings.Repeat("a", 251), lobbies.CodeTooLong},
	} {
		_, err := e.svc.Reject(t.Context(), owner.id, a.ID, c.reason)
		wantField(t, err, lobbies.FieldReason, c.code)
	}
	if got := e.status(t, a.ID); got != StatusPending {
		t.Errorf("continua pendente: %s", got)
	}
	got, err := e.svc.Reject(t.Context(), owner.id, a.ID, "  precisamos de mais dano mágico ")
	if err != nil || got.Status != StatusRejected || got.Reason != "precisamos de mais dano mágico" {
		t.Errorf("recusa: %+v, %v", got, err)
	}

	// Limites exatos: 10 e 250.
	for i, reason := range []string{strings.Repeat("b", 10), strings.Repeat("c", 250)} {
		p := e.user(t, fmt.Sprintf("Lim%d:200:dps", i))
		app := e.apply(t, p, fmt.Sprintf("Lim%d", i), lid)
		if got, err := e.svc.Reject(t.Context(), owner.id, app.ID, reason); err != nil || got.Status != StatusRejected {
			t.Errorf("%d caracteres: %v", len(reason), err)
		}
	}
}

// CA-02.6: só o dono decide.
func TestDecide_CA02_6_NotOwner(t *testing.T) {
	e, _, player, lid := basic(t)
	a := e.apply(t, player, "Cura", lid)
	stranger := e.user(t)
	_, err := e.svc.Accept(t.Context(), stranger.id, a.ID)
	wantRule(t, err, CodeNotOwner)
	_, err = e.svc.Reject(t.Context(), player.id, a.ID, "eu mesmo me recuso")
	wantRule(t, err, CodeNotOwner)
	if got := e.status(t, a.ID); got != StatusPending {
		t.Errorf("continua pendente: %s", got)
	}
	_, err = e.svc.Accept(t.Context(), stranger.id, "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("inexistente: %v", err)
	}
}

// CA-02.7 e CA-04.3: recusada, retirada ou expirada não é decidida.
func TestDecide_CA02_7_CA04_3_NotPending(t *testing.T) {
	e, owner, player, lid := basic(t)
	rejected := e.apply(t, player, "Cura", lid)
	if _, err := e.svc.Reject(t.Context(), owner.id, rejected.ID, "já temos suporte"); err != nil {
		t.Fatal(err)
	}
	_, err := e.svc.Accept(t.Context(), owner.id, rejected.ID)
	wantRule(t, err, CodeNotPending)

	p2 := e.user(t, "Dois:200:dps")
	withdrawn := e.apply(t, p2, "Dois", lid)
	if _, err := e.svc.Withdraw(t.Context(), p2.id, withdrawn.ID); err != nil {
		t.Fatal(err)
	}
	_, err = e.svc.Accept(t.Context(), owner.id, withdrawn.ID)
	wantRule(t, err, CodeNotPending)

	// Expirada pelo início (D-02) e pelo cancelamento (RN-16).
	p3 := e.user(t, "Tres:200:dps")
	byStart := e.apply(t, p3, "Tres", lid)
	cancelledLobby := e.lobby(t, owner, "Dono", at(5, 20, 0), std, 160)
	byCancel := e.apply(t, p3, "Tres", cancelledLobby)
	if _, err := e.lobbies.Cancel(t.Context(), owner.id, cancelledLobby, "imprevisto no trabalho"); err != nil {
		t.Fatal(err)
	}
	_, err = e.svc.Reject(t.Context(), owner.id, byCancel.ID, "já não dá mais")
	wantRule(t, err, CodeNotPending)
	e.now = at(1, 20, 0)
	_, err = e.svc.Accept(t.Context(), owner.id, byStart.ID)
	wantRule(t, err, CodeNotPending)
	_, err = e.svc.Reject(t.Context(), owner.id, byStart.ID, "já começou o lobby")
	wantRule(t, err, CodeNotPending)
}

// CA-02.8: segundo aceite com a função lotada falha e a candidatura continua pendente.
func TestAccept_CA02_8_RoleFull(t *testing.T) {
	e, owner, _, lid := basic(t) // 1 vaga de Tank
	a := e.user(t, "A:200:tank")
	b := e.user(t, "B:200:tank")
	first := e.apply(t, a, "A", lid)
	second := e.apply(t, b, "B", lid)
	if _, err := e.svc.Accept(t.Context(), owner.id, first.ID); err != nil {
		t.Fatal(err)
	}
	_, err := e.svc.Accept(t.Context(), owner.id, second.ID)
	wantRule(t, err, CodeRoleFull)
	if got := e.status(t, second.ID); got != StatusPending {
		t.Errorf("segunda: %s", got)
	}
}

// CA-02.9: aceites simultâneos para a última vaga; só um conclui (RN-12, D-05).
func TestAccept_CA02_9_Concurrent(t *testing.T) {
	e, owner, _, lid := basic(t) // 1 vaga de Tank
	var apps []Application
	for i := range 6 {
		w := e.user(t, fmt.Sprintf("T%d:200:tank", i))
		apps = append(apps, e.apply(t, w, fmt.Sprintf("T%d", i), lid))
	}
	var wg sync.WaitGroup
	errs := make([]error, len(apps))
	for i, a := range apps {
		wg.Go(func() { _, errs[i] = e.svc.Accept(t.Context(), owner.id, a.ID) })
	}
	wg.Wait()
	ok := 0
	for _, err := range errs {
		if err == nil {
			ok++
			continue
		}
		wantRule(t, err, CodeRoleFull)
	}
	occ, pending := e.occupied(t, lid)
	if ok != 1 || occ.Tank != 1 || pending != 5 {
		t.Errorf("aceites %d, tanks %d, pendentes %d", ok, occ.Tank, pending)
	}
}

// CA-02.10 a CA-02.12: conflito de horário no aceite, com a janela de 2 h (RN-11, D-04).
func TestAccept_CA02_10_to_CA02_12_Schedule(t *testing.T) {
	e := setup(t)
	o1 := e.user(t, "Um:200:dps")
	o2 := e.user(t, "Dois:200:dps")
	player := e.user(t, "Cura:200:support")
	first := e.lobby(t, o1, "Um", at(1, 20, 0), std, 160)
	if _, err := e.svc.Accept(t.Context(), o1.id, e.apply(t, player, "Cura", first).ID); err != nil {
		t.Fatal(err)
	}

	near := e.lobby(t, o2, "Dois", at(1, 21, 30), std, 160)
	a := e.apply(t, player, "Cura", near)
	_, err := e.svc.Accept(t.Context(), o2.id, a.ID)
	wantRule(t, err, CodeScheduleConflict) // CA-02.10
	if got := e.status(t, a.ID); got != StatusPending {
		t.Errorf("continua pendente: %s", got)
	}

	o4 := e.user(t, "Quatro:200:dps")
	twoHours := e.lobby(t, o4, "Quatro", at(1, 22, 0), std, 160)
	if _, err := e.svc.Accept(t.Context(), o4.id, e.apply(t, player, "Cura", twoHours).ID); err != nil {
		t.Errorf("CA-02.11: %v", err)
	}

	// CA-02.12: o primeiro é cancelado; o das 21:00 deixa de conflitar.
	o3 := e.user(t, "Tres:200:dps")
	at21 := e.lobby(t, o3, "Tres", at(1, 21, 0), std, 160)
	b := e.apply(t, player, "Cura", at21)
	_, err = e.svc.Accept(t.Context(), o3.id, b.ID)
	wantRule(t, err, CodeScheduleConflict)
	if _, err := e.lobbies.Cancel(t.Context(), o1.id, first, "imprevisto no trabalho"); err != nil {
		t.Fatal(err)
	}
	// O das 22:00 continua a menos de 2 h do das 21:00; sai da conta para isolar o caso.
	if _, err := e.lobbies.Cancel(t.Context(), o4.id, twoHours, "imprevisto no trabalho"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.Accept(t.Context(), o3.id, b.ID); err != nil {
		t.Errorf("CA-02.12: %v", err)
	}
}

// RN-30: o nível mínimo vale também no aceite (o dono pode ter subido o mínimo).
func TestAccept_RN30_MinLevelAtAccept(t *testing.T) {
	e := setup(t)
	owner := e.user(t, "Dono:250:dps")
	player := e.user(t, "Cura:200:support")
	lid := e.lobby(t, owner, "Dono", at(1, 20, 0), std, 160)
	a := e.apply(t, player, "Cura", lid)
	if _, err := e.lobbies.Update(t.Context(), owner.id, lid, lobbies.UpdateInput{StartsAt: at(1, 20, 0), Slots: std, MinLevel: 210}); err != nil {
		t.Fatal(err)
	}
	_, err := e.svc.Accept(t.Context(), owner.id, a.ID)
	wantRule(t, err, CodeBelowMinLevel)
}

// CA-03.1: as próprias candidaturas com lobby, personagem, estado e justificativa.
func TestListMine_CA03_1(t *testing.T) {
	e, owner, player, lid := basic(t)
	rejected := e.apply(t, player, "Cura", lid)
	if _, err := e.svc.Reject(t.Context(), owner.id, rejected.ID, "já temos suporte"); err != nil {
		t.Fatal(err)
	}
	accepted := e.apply(t, player, "Escudo", e.lobby(t, owner, "Dono", at(2, 20, 0), std, 160))
	if _, err := e.svc.Accept(t.Context(), owner.id, accepted.ID); err != nil {
		t.Fatal(err)
	}
	e.now = e.now.Add(time.Minute)
	pending := e.apply(t, player, "Fogo", e.lobby(t, owner, "Dono", at(3, 20, 0), std, 160))

	mine, err := e.svc.ListMine(t.Context(), player.id)
	if err != nil || len(mine) != 3 {
		t.Fatalf("minhas = %d, %v", len(mine), err)
	}
	byID := map[string]Mine{}
	for _, m := range mine {
		byID[m.ID] = m
	}
	if m := byID[rejected.ID]; m.Status != StatusRejected || m.Reason != "já temos suporte" || m.Nick != "Cura" || m.InstanceName != "Templo do Demônio Rei" || m.LobbyStatus != lobbies.StatusOpen {
		t.Errorf("recusada = %+v", m)
	}
	if m := byID[accepted.ID]; m.Status != StatusAccepted || m.Level != 200 {
		t.Errorf("aceita = %+v", m)
	}
	if mine[0].ID != pending.ID {
		t.Errorf("a mais recente primeiro: %s", mine[0].ID)
	}
}

// CA-03.2 a CA-03.5: retirar a própria pendente e candidatar de novo.
func TestWithdraw_CA03_2_to_CA03_5(t *testing.T) {
	e, owner, player, lid := basic(t)
	a := e.apply(t, player, "Cura", lid)
	stranger := e.user(t)
	_, err := e.svc.Withdraw(t.Context(), stranger.id, a.ID)
	wantRule(t, err, CodeNotYours) // CA-03.4
	got, err := e.svc.Withdraw(t.Context(), player.id, a.ID)
	if err != nil || got.Status != StatusWithdrawn {
		t.Fatalf("CA-03.2: %+v, %v", got, err)
	}
	again := e.apply(t, player, "Cura", lid) // CA-03.5
	if _, err := e.svc.Accept(t.Context(), owner.id, again.ID); err != nil {
		t.Fatal(err)
	}
	_, err = e.svc.Withdraw(t.Context(), player.id, again.ID)
	wantRule(t, err, CodeNotPending) // CA-03.3
	if got := e.status(t, again.ID); got != StatusAccepted {
		t.Errorf("continua aceita: %s", got)
	}
}

// CA-03.6: um evento por transição, com autor, data em UTC e justificativa.
func TestHistory_CA03_6(t *testing.T) {
	e, owner, player, lid := basic(t)
	a, err := e.svc.Apply(t.Context(), player.id, lid, ApplyInput{CharacterID: player.chars["Cura"], Message: "tenho buff de ASPD"})
	if err != nil {
		t.Fatal(err)
	}
	e.now = e.now.Add(time.Hour)
	if _, err := e.svc.Reject(t.Context(), owner.id, a.ID, "já temos suporte"); err != nil {
		t.Fatal(err)
	}
	var aid pgtype.UUID
	_ = aid.Scan(a.ID)
	events, err := e.q.ListApplicationEvents(t.Context(), aid)
	if err != nil || len(events) != 2 {
		t.Fatalf("eventos = %+v, %v", events, err)
	}
	created, rejected := events[0], events[1]
	if created.FromStatus.Valid || created.ToStatus != StatusPending || created.ActorID.String() != player.id ||
		created.Reason.String != "tenho buff de ASPD" || !created.At.Equal(start) {
		t.Errorf("criação = %+v", created)
	}
	if rejected.FromStatus.String != StatusPending || rejected.ToStatus != StatusRejected || rejected.ActorID.String() != owner.id ||
		rejected.Reason.String != "já temos suporte" || !rejected.At.Equal(start.Add(time.Hour)) {
		t.Errorf("recusa = %+v", rejected)
	}
}

// CA-04.1: no início, as pendentes contam como expiradas e as aceitas continuam.
func TestExpire_CA04_1_ByStart(t *testing.T) {
	e, owner, player, lid := basic(t)
	other := e.user(t, "Outro:200:dps")
	pending := e.apply(t, other, "Outro", lid)
	accepted := e.apply(t, player, "Cura", lid)
	if _, err := e.svc.Accept(t.Context(), owner.id, accepted.ID); err != nil {
		t.Fatal(err)
	}
	e.now = at(1, 20, 0)
	for _, c := range []struct {
		w    who
		id   string
		want string
	}{{other, pending.ID, StatusExpired}, {player, accepted.ID, StatusAccepted}} {
		mine, err := e.svc.ListMine(t.Context(), c.w.id)
		if err != nil || len(mine) != 1 || mine[0].ID != c.id || mine[0].Status != c.want || mine[0].LobbyStatus != lobbies.StatusStarted {
			t.Errorf("%s: %+v, %v", c.want, mine, err)
		}
	}
	_, err := e.svc.Withdraw(t.Context(), other.id, pending.ID)
	wantRule(t, err, CodeNotPending)
}

// CA-04.2: o cancelamento grava as pendentes como expiradas, com evento.
func TestExpire_CA04_2_ByCancel(t *testing.T) {
	e, owner, player, lid := basic(t)
	other := e.user(t, "Outro:200:dps")
	pending := e.apply(t, other, "Outro", lid)
	accepted := e.apply(t, player, "Cura", lid)
	if _, err := e.svc.Accept(t.Context(), owner.id, accepted.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := e.lobbies.Cancel(t.Context(), owner.id, lid, "imprevisto no trabalho"); err != nil {
		t.Fatal(err)
	}
	if got := e.status(t, pending.ID); got != StatusExpired {
		t.Errorf("pendente: %s", got)
	}
	if got := e.status(t, accepted.ID); got != StatusAccepted {
		t.Errorf("aceita: %s", got)
	}
	var aid pgtype.UUID
	_ = aid.Scan(pending.ID)
	events, _ := e.q.ListApplicationEvents(t.Context(), aid)
	if len(events) != 2 || events[1].ToStatus != StatusExpired || events[1].ActorID.String() != owner.id {
		t.Errorf("eventos = %+v", events)
	}
}

// RN-18 da lobbies com os membros: as vagas não descem abaixo do dono e dos aceitos.
func TestUpdate_RN18_SlotsBelowMembers(t *testing.T) {
	e, owner, player, lid := basic(t)
	if _, err := e.svc.Accept(t.Context(), owner.id, e.apply(t, player, "Cura", lid).ID); err != nil {
		t.Fatal(err)
	}
	_, err := e.lobbies.Update(t.Context(), owner.id, lid, lobbies.UpdateInput{
		StartsAt: at(1, 20, 0), Slots: lobbies.Slots{Tank: 1, Support: 0, Dps: 3}, MinLevel: 160,
	})
	var ve *lobbies.ValidationError
	if !errors.As(err, &ve) || ve.Fields[0].Code != lobbies.CodeBelowOccupied {
		t.Errorf("err = %v", err)
	}
	if _, err := e.lobbies.Update(t.Context(), owner.id, lid, lobbies.UpdateInput{
		StartsAt: at(1, 20, 0), Slots: lobbies.Slots{Tank: 0, Support: 1, Dps: 1}, MinLevel: 160,
	}); err != nil {
		t.Errorf("no limite: %v", err)
	}
}
