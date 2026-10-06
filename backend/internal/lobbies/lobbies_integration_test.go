//go:build integration

package lobbies

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/migrate"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

// "Agora" dos testes: terça, 6 out 2026, 16:40 em São Paulo (19:40 UTC).
var now = time.Date(2026, 10, 6, 19, 40, 0, 0, time.UTC)

// at monta um horário de São Paulo, daqui a `days` dias.
func at(days, hour, minute int) time.Time {
	y, m, d := now.In(Location).Date()
	return time.Date(y, m, d+days, hour, minute, 0, 0, Location)
}

type env struct {
	svc  *Service
	pool *pgxpool.Pool
	q    *db.Queries
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
	svc := NewService(pool)
	svc.Now = func() time.Time { return now }
	return &env{svc: svc, pool: pool, q: db.New(pool)}
}

type who struct {
	userID string
	chars  map[string]string // nick → id
}

// user cria um Usuário com personagens "Nick:classe:nível:função".
func (e *env) user(t *testing.T, discordID string, chars ...string) who {
	t.Helper()
	u, err := e.q.UpsertUserByDiscordID(t.Context(), db.UpsertUserByDiscordIDParams{
		DiscordID: discordID, Username: "u" + discordID, GlobalName: pgtype.Text{String: "Jogador " + discordID, Valid: true}, Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	w := who{userID: u.ID.String(), chars: map[string]string{}}
	for i, spec := range chars {
		var nick, class, role string
		var level int16
		parts := strings.Split(spec, ":")
		nick, class, role = parts[0], parts[1], parts[3]
		_, _ = fmt.Sscan(parts[2], &level)
		c, err := e.q.CreateCharacter(t.Context(), db.CreateCharacterParams{
			UserID: u.ID, Nick: nick, ClassID: class, Level: level, Role: role,
			Portrait: "retrato-1", IsMain: i == 0, Now: now,
		})
		if err != nil {
			t.Fatal(err)
		}
		w.chars[nick] = c.ID.String()
	}
	return w
}

func temple(characterID string, startsAt time.Time) Input {
	return Input{
		InstanceID:  "templo-do-demonio-rei",
		StartsAt:    startsAt,
		Slots:       Slots{Tank: 1, Support: 2, Dps: 3},
		MinLevel:    160,
		CharacterID: characterID,
	}
}

func wantField(t *testing.T, err error, field, code string) {
	t.Helper()
	var ve *ValidationError
	if !errors.As(err, &ve) || !reflect.DeepEqual(ve.Fields, []FieldError{{field, code}}) {
		t.Errorf("err = %v, quer %s/%s", err, field, code)
	}
}

// CA-01.1 / RN-04 / RN-06 / RN-07 / RN-08: criação com os padrões; o dono ocupa a vaga
// da função dele.
func TestCreate_CA01_1_Defaults(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")

	l, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(1, 20, 0)))
	if err != nil {
		t.Fatal(err)
	}
	if l.Status != StatusOpen || l.Slots != (Slots{1, 2, 3}) || l.MinLevel != 160 {
		t.Errorf("lobby = %+v", l)
	}
	if l.Occupied != (Slots{Support: 1}) {
		t.Errorf("ocupados = %+v", l.Occupied)
	}
	if l.InstanceName != "Templo do Demônio Rei" || l.InstanceLevel != 160 || l.InstanceReset != "daily" {
		t.Errorf("instância = %s/%d/%s", l.InstanceName, l.InstanceLevel, l.InstanceReset)
	}
	wantOwner := Owner{UserID: ana.userID, DiscordName: "Jogador 1", CharacterID: ana.chars["Lirien"],
		Nick: "Lirien", ClassID: "arcebispo", Level: 178, Role: "support"}
	if l.Owner != wantOwner {
		t.Errorf("dono = %+v", l.Owner)
	}
	if !l.StartsAt.Equal(time.Date(2026, 10, 7, 23, 0, 0, 0, time.UTC)) {
		t.Errorf("início em UTC = %v", l.StartsAt)
	}
}

// CA-01.2 / RN-05: horário no passado ou fora dos 14 dias é recusado; o último dia vale.
func TestCreate_CA01_2_StartWindow(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	for _, start := range []time.Time{at(0, 16, 0), now, at(14, 0, 0), {}} {
		_, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], start))
		code := CodeInvalid
		if start.IsZero() {
			code = CodeRequired
		}
		wantField(t, err, FieldStartsAt, code)
	}
	if _, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(13, 23, 30))); err != nil {
		t.Errorf("último dia: %v", err)
	}
}

// Borda de RN-05: 23:30 de hoje em São Paulo é hoje, mesmo sendo amanhã em UTC.
func TestCreate_RN05_DayTurnInSaoPaulo(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	l, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(0, 23, 30)))
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.svc.List(t.Context(), "2026-10-06", "2026-10-06")
	if err != nil || len(got) != 1 || got[0].ID != l.ID {
		t.Errorf("lista de hoje = %+v, %v", got, err)
	}
	if got, _ := e.svc.List(t.Context(), "2026-10-07", "2026-10-07"); len(got) != 0 {
		t.Errorf("apareceu amanhã: %+v", got)
	}
}

// CA-01.3 a CA-01.7, CA-01.11, CA-01.12: erros de campo, sem gravar nada.
func TestCreate_CA01_3_to_CA01_12_FieldErrors(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support", "Brasa:guardiao-real:172:tank")
	bia := e.user(t, "2", "Outro:paladino:200:tank")
	cases := []struct {
		name        string
		mutate      func(*Input)
		field, code string
	}{
		{"CA-01.3 total 0", func(in *Input) { in.Slots = Slots{} }, FieldSlots, CodeInvalid},
		{"CA-01.3 total 13", func(in *Input) { in.Slots = Slots{1, 4, 8} }, FieldSlots, CodeInvalid},
		{"CA-01.4 nível abaixo da instância", func(in *Input) { in.InstanceID = "torre-da-constelacao"; in.MinLevel = 200 }, FieldMinLevel, CodeInvalid},
		{"CA-01.5 personagem abaixo do nível", func(in *Input) { in.CharacterID = ana.chars["Brasa"]; in.MinLevel = 180 }, FieldCharacterID, CodeLevelTooLow},
		{"CA-01.6 função do dono sem vaga", func(in *Input) { in.CharacterID = ana.chars["Brasa"]; in.Slots = Slots{0, 2, 3} }, FieldSlots, CodeInvalid},
		{"CA-01.7 personagem de outro", func(in *Input) { in.CharacterID = bia.chars["Outro"] }, FieldCharacterID, CodeInvalid},
		{"CA-01.7 personagem malformado", func(in *Input) { in.CharacterID = "nao-e-uuid" }, FieldCharacterID, CodeInvalid},
		{"CA-01.11 observação com 251", func(in *Input) { in.Note = strings.Repeat("a", 251) }, FieldNote, CodeTooLong},
		{"CA-01.12 instância fora do catálogo", func(in *Input) { in.InstanceID = "caverna-de-gelo" }, FieldInstanceID, CodeInvalid},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := temple(ana.chars["Lirien"], at(1, 20, 0))
			c.mutate(&in)
			_, err := e.svc.Create(t.Context(), ana.userID, in)
			wantField(t, err, c.field, c.code)
		})
	}
	if got, _ := e.svc.List(t.Context(), "2026-10-06", "2026-10-19"); len(got) != 0 {
		t.Errorf("gravou lobbies inválidos: %+v", got)
	}
}

// CA-01.8 / RN-10: 21:30 conflita com o lobby das 20:00; 22:00 não.
func TestCreate_CA01_8_ScheduleConflict(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support", "Brasa:guardiao-real:172:tank")
	if _, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(0, 20, 0))); err != nil {
		t.Fatal(err)
	}
	_, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(0, 21, 30)))
	wantField(t, err, FieldStartsAt, CodeConflict)
	if _, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(0, 22, 0))); err != nil {
		t.Errorf("22:00: %v", err)
	}
	// Outro personagem do mesmo Usuário no mesmo horário pode.
	brasa := temple(ana.chars["Brasa"], at(0, 20, 0))
	if _, err := e.svc.Create(t.Context(), ana.userID, brasa); err != nil {
		t.Errorf("outro personagem: %v", err)
	}
}

// CA-01.9 / RN-11: com 5 lobbies abertos, o sexto é recusado; um cancelado libera.
func TestCreate_CA01_9_LimitOfFive(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	var first Lobby
	for d := range MaxOpenPerOwner {
		l, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(d+1, 20, 0)))
		if err != nil {
			t.Fatal(err)
		}
		if d == 0 {
			first = l
		}
	}
	if _, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(7, 20, 0))); !errors.Is(err, ErrLimitReached) {
		t.Errorf("err = %v, quer ErrLimitReached", err)
	}
	if _, err := e.svc.Cancel(t.Context(), ana.userID, first.ID, "Liberando espaço para outro"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(7, 20, 0))); err != nil {
		t.Errorf("depois de cancelar: %v", err)
	}
}

// CA-06.5 / RN-10 / RN-11 / D-05: criações simultâneas não passam do limite nem criam
// conflito.
func TestCreate_CA06_5_Concurrent(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	var wg sync.WaitGroup
	errs := make([]error, 8)
	for i := range errs {
		wg.Go(func() {
			// Dias diferentes (sem conflito): só o limite segura.
			_, errs[i] = e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(i+1, 20, 0)))
		})
	}
	wg.Wait()
	ok, limited := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, ErrLimitReached):
			limited++
		default:
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if ok != MaxOpenPerOwner || limited != 3 {
		t.Errorf("%d criados e %d recusados, quer 5 e 3", ok, limited)
	}

	bia := e.user(t, "2", "Brasa:guardiao-real:172:tank")
	results := make([]error, 4)
	for i := range results {
		wg.Go(func() {
			// Mesmo personagem, horários a 30 min um do outro: só um passa.
			_, results[i] = e.svc.Create(t.Context(), bia.userID, temple(bia.chars["Brasa"], at(1, 20, 30*i)))
		})
	}
	wg.Wait()
	created := 0
	for _, err := range results {
		var ve *ValidationError
		switch {
		case err == nil:
			created++
		case errors.As(err, &ve) && ve.Fields[0].Code == CodeConflict:
		default:
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if created != 1 {
		t.Errorf("%d lobbies criados para o mesmo personagem em conflito", created)
	}
}

// CA-02.2 / RN-13 / RN-14: iniciado e cancelado saem da listagem; o detalhe mostra o
// estado.
func TestList_CA02_2_OnlyOpen(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support", "Brasa:guardiao-real:172:tank")
	soon, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], now.Add(time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	later, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Brasa"], at(1, 20, 0)))
	if err != nil {
		t.Fatal(err)
	}
	gone, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(2, 20, 0)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.Cancel(t.Context(), ana.userID, gone.ID, "Metade do grupo não pode"); err != nil {
		t.Fatal(err)
	}
	e.svc.Now = func() time.Time { return now.Add(2 * time.Minute) } // o primeiro já começou

	got, err := e.svc.List(t.Context(), "2026-10-06", "2026-10-19")
	if err != nil || len(got) != 1 || got[0].ID != later.ID {
		t.Errorf("lista = %+v, %v", got, err)
	}
	for id, want := range map[string]string{soon.ID: StatusStarted, gone.ID: StatusCancelled, later.ID: StatusOpen} {
		l, err := e.svc.Get(t.Context(), id)
		if err != nil || l.Status != want {
			t.Errorf("%s: estado %q, quer %q (%v)", id, l.Status, want, err)
		}
	}
	if l, _ := e.svc.Get(t.Context(), gone.ID); l.CancelReason != "Metade do grupo não pode" {
		t.Errorf("motivo = %q", l.CancelReason)
	}
	if _, err := e.svc.List(t.Context(), "2026-10-19", "2026-10-06"); !errors.Is(err, ErrInvalidRange) {
		t.Errorf("intervalo invertido: %v", err)
	}
}

// CA-03.2 / RN-15: id inexistente ou malformado é ErrNotFound.
func TestGet_CA03_2_NotFound(t *testing.T) {
	e := setup(t)
	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "nao-e-uuid"} {
		if _, err := e.svc.Get(t.Context(), id); !errors.Is(err, ErrNotFound) {
			t.Errorf("%s: %v", id, err)
		}
	}
}

// CA-04.1 / RN-17: editar horário e vagas.
func TestUpdate_CA04_1(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	l, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(1, 20, 0)))
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.svc.Update(t.Context(), ana.userID, l.ID, UpdateInput{
		StartsAt: at(1, 21, 0), Slots: Slots{1, 2, 5}, MinLevel: 170, Note: "  Chamar no Discord  ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !got.StartsAt.Equal(at(1, 21, 0)) || got.Slots.Dps != 5 || got.MinLevel != 170 || got.Note != "Chamar no Discord" {
		t.Errorf("editado = %+v", got)
	}
	if got.InstanceID != l.InstanceID || got.Owner != l.Owner {
		t.Errorf("instância ou dono mudaram: %+v", got)
	}
	// Mover para perto do próprio horário não conflita consigo mesmo.
	if _, err := e.svc.Update(t.Context(), ana.userID, l.ID, UpdateInput{StartsAt: at(1, 21, 30), Slots: Slots{1, 2, 3}, MinLevel: 160}); err != nil {
		t.Errorf("mover 30 min: %v", err)
	}
}

// CA-04.2 / CA-04.3 / RN-18: vagas abaixo dos ocupantes e nível acima do dono.
func TestUpdate_CA04_2_CA04_3_Limits(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	l, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(1, 20, 0)))
	if err != nil {
		t.Fatal(err)
	}
	_, err = e.svc.Update(t.Context(), ana.userID, l.ID, UpdateInput{StartsAt: at(1, 20, 0), Slots: Slots{1, 0, 3}, MinLevel: 160})
	wantField(t, err, FieldSlots, CodeBelowOccupied)
	_, err = e.svc.Update(t.Context(), ana.userID, l.ID, UpdateInput{StartsAt: at(1, 20, 0), Slots: Slots{1, 2, 3}, MinLevel: 180})
	wantField(t, err, FieldMinLevel, CodeAboveOwner)
	_, err = e.svc.Update(t.Context(), ana.userID, l.ID, UpdateInput{StartsAt: at(1, 20, 0), Slots: Slots{1, 2, 3}, MinLevel: 150})
	wantField(t, err, FieldMinLevel, CodeInvalid)
}

// CA-04.4 / RN-17 / RN-20 / D-08: lobby de outro é ErrNotFound; iniciado ou cancelado é
// ErrNotOpen; nada muda.
func TestUpdate_CA04_4_OtherOrNotOpen(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support", "Brasa:guardiao-real:172:tank")
	bia := e.user(t, "2", "Outro:paladino:200:tank")
	l, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], now.Add(30*time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	upd := UpdateInput{StartsAt: at(1, 21, 0), Slots: Slots{1, 2, 3}, MinLevel: 160}
	if _, err := e.svc.Update(t.Context(), bia.userID, l.ID, upd); !errors.Is(err, ErrNotFound) {
		t.Errorf("de outro: %v", err)
	}
	if _, err := e.svc.Cancel(t.Context(), bia.userID, l.ID, "Cancelando o dos outros"); !errors.Is(err, ErrNotFound) {
		t.Errorf("cancelar de outro (CA-05.3): %v", err)
	}
	e.svc.Now = func() time.Time { return now.Add(31 * time.Minute) } // começou
	if _, err := e.svc.Update(t.Context(), ana.userID, l.ID, UpdateInput{StartsAt: at(1, 21, 0).Add(time.Hour), Slots: Slots{1, 2, 3}, MinLevel: 160}); !errors.Is(err, ErrNotOpen) {
		t.Errorf("iniciado: %v", err)
	}
	if _, err := e.svc.Cancel(t.Context(), ana.userID, l.ID, "Já começou, mas quero cancelar"); !errors.Is(err, ErrNotOpen) {
		t.Errorf("cancelar iniciado: %v", err)
	}
	got, _ := e.svc.Get(t.Context(), l.ID)
	if got.Status != StatusStarted || !got.StartsAt.Equal(l.StartsAt) {
		t.Errorf("lobby mudou: %+v", got)
	}
}

// CA-05.1 / CA-05.2 / RN-19: cancelar com motivo; motivo curto ou longo é recusado;
// cancelado não se cancela de novo.
func TestCancel_CA05_1_CA05_2(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	l, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(1, 20, 0)))
	if err != nil {
		t.Fatal(err)
	}
	_, err = e.svc.Cancel(t.Context(), ana.userID, l.ID, "  não dá  ")
	wantField(t, err, FieldReason, CodeTooShort)
	_, err = e.svc.Cancel(t.Context(), ana.userID, l.ID, strings.Repeat("a", 251))
	wantField(t, err, FieldReason, CodeTooLong)

	got, err := e.svc.Cancel(t.Context(), ana.userID, l.ID, "  Metade do grupo não pode hoje  ")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusCancelled || got.CancelReason != "Metade do grupo não pode hoje" {
		t.Errorf("cancelado = %+v", got)
	}
	if _, err := e.svc.Cancel(t.Context(), ana.userID, l.ID, "Cancelando de novo, por engano"); !errors.Is(err, ErrNotOpen) {
		t.Errorf("cancelar de novo: %v", err)
	}
	// RN-10: o lobby cancelado não conta para conflito.
	if _, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(1, 20, 0))); err != nil {
		t.Errorf("mesmo horário depois de cancelar: %v", err)
	}
}

// Borda de RN-01 / D-02: o lobby guarda o nome da instância; se ela sair do catálogo, o
// nome continua e o retorno fica vazio.
func TestGet_RN01_InstanceLeftCatalog(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	l, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(1, 20, 0)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.pool.Exec(t.Context(), "UPDATE lobbies SET instance_id = 'instancia-removida'"); err != nil {
		t.Fatal(err)
	}
	got, err := e.svc.Get(t.Context(), l.ID)
	if err != nil || got.InstanceName != "Templo do Demônio Rei" || got.InstanceReset != "" {
		t.Errorf("lobby = %+v, %v", got, err)
	}
}
