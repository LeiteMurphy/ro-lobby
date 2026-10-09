//go:build integration

package talents_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/applications"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/characters"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/migrate"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/testdb"
)

// "Agora" dos testes: terça, 6 out 2026, 16:40 em São Paulo (19:40 UTC).
var now = time.Date(2026, 10, 6, 19, 40, 0, 0, time.UTC)

var nextDiscordID atomic.Int64

// at monta um horário de São Paulo, daqui a `days` dias.
func at(days, hour, minute int) time.Time {
	y, m, d := now.In(lobbies.Location).Date()
	return time.Date(y, m, d+days, hour, minute, 0, 0, lobbies.Location)
}

type env struct {
	svc     *talents.Service
	lobbies *lobbies.Service
	apps    *applications.Service
	chars   *characters.Service
	pool    *pgxpool.Pool
	q       *db.Queries
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
	clock := func() time.Time { return now }
	e := &env{
		svc: talents.NewService(pool), lobbies: lobbies.NewService(pool), apps: applications.NewService(pool),
		chars: characters.NewService(pool), pool: pool, q: db.New(pool),
	}
	e.svc.Now, e.lobbies.Now, e.apps.Now, e.chars.Now = clock, clock, clock, clock
	return e
}

type who struct {
	userID string
	chars  map[string]string // nick → id
}

// user cria um Usuário (username = name) com personagens "Nick:classe:nível:função".
func (e *env) user(t *testing.T, name string, chars ...string) who {
	t.Helper()
	u, err := e.q.UpsertUserByDiscordID(t.Context(), db.UpsertUserByDiscordIDParams{
		DiscordID: fmt.Sprint(nextDiscordID.Add(1)), Username: name,
		GlobalName: pgtype.Text{String: name, Valid: true}, Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	w := who{userID: u.ID.String(), chars: map[string]string{}}
	for i, spec := range chars {
		parts := strings.Split(spec, ":")
		var level int16
		_, _ = fmt.Sscan(parts[2], &level)
		c, err := e.q.CreateCharacter(t.Context(), db.CreateCharacterParams{
			UserID: u.ID, Nick: parts[0], ClassID: parts[1], Level: level, Role: parts[3],
			Portrait: "retrato-1", IsMain: i == 0, Now: now,
		})
		if err != nil {
			t.Fatal(err)
		}
		w.chars[parts[0]] = c.ID.String()
	}
	return w
}

// put liga a disponibilidade e falha o teste se não der.
func (e *env) put(t *testing.T, w who, nick string, in talents.AvailabilityInput) {
	t.Helper()
	in.Enabled = true
	if _, err := e.svc.SetAvailability(t.Context(), w.userID, w.chars[nick], in); err != nil {
		t.Fatalf("%s: %v", nick, err)
	}
}

func wantField(t *testing.T, err error, field, code string) {
	t.Helper()
	var ve *talents.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v, quer erro em %s", err, field)
	}
	for _, f := range ve.Fields {
		if f.Field == field && f.Code == code {
			return
		}
	}
	t.Errorf("erros = %+v, quer %s/%s", ve.Fields, field, code)
}

var weekdays = []int{1, 2, 3, 4, 5}

// CA-01.1 / CA-01.4 / RN-01: ligar grava; desligar guarda os dados; a lista do perfil
// traz a disponibilidade.
func TestSetAvailability_CA01_1_CA01_4(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "ana", "Lirien:arcebispo:178:support")
	got, err := e.svc.SetAvailability(t.Context(), ana.userID, ana.chars["Lirien"], talents.AvailabilityInput{
		Enabled: true, Days: weekdays, Start: "19:00", End: "23:00",
		InstanceIDs: []string{"templo-do-demonio-rei", "sonho-sombrio", "templo-do-demonio-rei"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := talents.Availability{Enabled: true, Days: weekdays, Start: "19:00", End: "23:00", InstanceIDs: []string{"templo-do-demonio-rei", "sonho-sombrio"}}
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("gravada = %+v", *got)
	}
	off, err := e.svc.SetAvailability(t.Context(), ana.userID, ana.chars["Lirien"], talents.AvailabilityInput{Enabled: false})
	if err != nil || off.Enabled || off.Start != "19:00" || len(off.InstanceIDs) != 2 {
		t.Fatalf("desligada = %+v, %v", off, err)
	}
	list, err := e.chars.List(t.Context(), ana.userID)
	if err != nil || len(list) != 1 || list[0].Availability == nil || list[0].Availability.Enabled || list[0].Availability.End != "23:00" {
		t.Errorf("perfil = %+v, %v", list, err)
	}
}

// RN-01: desligar quem nunca entrou não grava nada e devolve nil.
func TestSetAvailability_RN01_DisableNeverEnabled(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "ana", "Lirien:arcebispo:178:support")
	got, err := e.svc.SetAvailability(t.Context(), ana.userID, ana.chars["Lirien"], talents.AvailabilityInput{Enabled: false})
	if err != nil || got != nil {
		t.Errorf("= %+v, %v", got, err)
	}
}

// CA-01.3 / RN-02 a RN-04: campos obrigatórios e inválidos; nada é gravado.
func TestSetAvailability_CA01_3_FieldErrors(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "ana", "Lirien:arcebispo:178:support")
	ok := talents.AvailabilityInput{Enabled: true, Days: weekdays, Start: "19:00", End: "23:00", AnyInstance: true}
	cases := []struct {
		name        string
		mutate      func(in *talents.AvailabilityInput)
		field, code string
	}{
		{"sem dia", func(in *talents.AvailabilityInput) { in.Days = nil }, talents.FieldDays, talents.CodeRequired},
		{"dia 7", func(in *talents.AvailabilityInput) { in.Days = []int{7} }, talents.FieldDays, talents.CodeInvalid},
		{"início fora dos 30 min", func(in *talents.AvailabilityInput) { in.Start = "19:15" }, talents.FieldStart, talents.CodeInvalid},
		{"fim 24:00", func(in *talents.AvailabilityInput) { in.End = "24:00" }, talents.FieldEnd, talents.CodeInvalid},
		{"início igual ao fim", func(in *talents.AvailabilityInput) { in.End = "19:00" }, talents.FieldEnd, talents.CodeSameAsStart},
		{"sem instância e sem Qualquer", func(in *talents.AvailabilityInput) { in.AnyInstance = false }, talents.FieldInstanceIDs, talents.CodeRequired},
		{"instância fora do catálogo", func(in *talents.AvailabilityInput) {
			in.AnyInstance, in.InstanceIDs = false, []string{"nao-existe"}
		}, talents.FieldInstanceIDs, talents.CodeInvalid},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := ok
			c.mutate(&in)
			_, err := e.svc.SetAvailability(t.Context(), ana.userID, ana.chars["Lirien"], in)
			wantField(t, err, c.field, c.code)
		})
	}
	if list, _ := e.chars.List(t.Context(), ana.userID); list[0].Availability != nil {
		t.Errorf("gravou com erro: %+v", list[0].Availability)
	}
}

// CA-01.5 / RN-05: personagem de outro Usuário, ou inexistente, é talents.ErrNotFound.
func TestSetAvailability_CA01_5_OtherUser(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "ana", "Lirien:arcebispo:178:support")
	bia := e.user(t, "bia", "Brasa:guardiao-real:172:tank")
	in := talents.AvailabilityInput{Enabled: true, Days: weekdays, Start: "19:00", End: "23:00", AnyInstance: true}
	for _, id := range []string{bia.chars["Brasa"], "00000000-0000-0000-0000-000000000000", "nao-e-uuid"} {
		if _, err := e.svc.SetAvailability(t.Context(), ana.userID, id, in); !errors.Is(err, talents.ErrNotFound) {
			t.Errorf("%s: %v", id, err)
		}
	}
	if list, _ := e.chars.List(t.Context(), bia.userID); list[0].Availability != nil {
		t.Errorf("mudou o de outro: %+v", list[0].Availability)
	}
}
