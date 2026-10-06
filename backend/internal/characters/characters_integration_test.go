//go:build integration

package characters

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

type env struct {
	svc   *Service
	pool  *pgxpool.Pool
	clock time.Time
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

	e := &env{pool: pool, clock: time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)}
	e.svc = NewService(pool)
	// Cada leitura do relógio avança um minuto, para a ordem de cadastro ficar clara.
	e.svc.Now = func() time.Time {
		e.clock = e.clock.Add(time.Minute)
		return e.clock
	}
	return e
}

func (e *env) user(t *testing.T, discordID string) string {
	t.Helper()
	u, err := db.New(e.pool).UpsertUserByDiscordID(t.Context(), db.UpsertUserByDiscordIDParams{
		DiscordID: discordID, Username: "u" + discordID, Now: e.clock,
	})
	if err != nil {
		t.Fatal(err)
	}
	return u.ID.String()
}

func (e *env) create(t *testing.T, userID, nick string) Character {
	t.Helper()
	c, err := e.svc.Create(t.Context(), userID, Input{Nick: nick, ClassID: "arcebispo", Level: 178, Role: "support"})
	if err != nil {
		t.Fatalf("criar %q: %v", nick, err)
	}
	return c
}

func (e *env) list(t *testing.T, userID string) []Character {
	t.Helper()
	list, err := e.svc.List(t.Context(), userID)
	if err != nil {
		t.Fatal(err)
	}
	return list
}

// order resume a lista como "Nick*" para o principal e "Nick" para os outros.
func order(list []Character) string {
	names := make([]string, len(list))
	for i, c := range list {
		names[i] = c.Nick
		if c.IsMain {
			names[i] += "*"
		}
	}
	return strings.Join(names, ",")
}

func wantField(t *testing.T, err error, field, code string) {
	t.Helper()
	var ve *ValidationError
	if !errors.As(err, &ve) || !reflect.DeepEqual(ve.Fields, []FieldError{{field, code}}) {
		t.Errorf("err = %v, quer %s/%s", err, field, code)
	}
}

// CA-02.1 / CA-02.2 / RN-12: o primeiro personagem vira principal e fica com o primeiro
// retrato quando o Usuário não escolhe; o segundo respeita o retrato escolhido.
func TestCreate_CA02_1_CA02_2_FirstIsMainWithDefaultPortrait(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1")

	first, err := e.svc.Create(t.Context(), ana, Input{Nick: " Brasa ", ClassID: "guardiao-real", Level: 172, Role: "tank"})
	if err != nil {
		t.Fatal(err)
	}
	want := Character{ID: first.ID, Nick: "Brasa", ClassID: "guardiao-real", Level: 172, Role: "tank",
		Portrait: "retrato-1", IsMain: true, CreatedAt: first.CreatedAt}
	if first != want {
		t.Errorf("primeiro = %+v, quer %+v", first, want)
	}

	second, err := e.svc.Create(t.Context(), ana, Input{Nick: "Lirien", ClassID: "arcebispo", Level: 178,
		Role: "support", Portrait: "retrato-3", Link: "https://exemplo.com/char/123"})
	if err != nil {
		t.Fatal(err)
	}
	if second.IsMain || second.Portrait != "retrato-3" || second.Link != "https://exemplo.com/char/123" {
		t.Errorf("segundo = %+v", second)
	}
}

// CA-02.3 / RN-05: o nick de outro Usuário é recusado, sem diferenciar maiúsculas.
func TestCreate_CA02_3_NickTakenByAnotherUser(t *testing.T) {
	e := setup(t)
	ana, bia := e.user(t, "1"), e.user(t, "2")
	e.create(t, ana, "Brasa")

	_, err := e.svc.Create(t.Context(), bia, Input{Nick: "brasa", ClassID: "guardiao-real", Level: 1, Role: "tank"})
	wantField(t, err, FieldNick, CodeTaken)
	if got := e.list(t, bia); len(got) != 0 {
		t.Errorf("criou mesmo assim: %+v", got)
	}
}

// CA-02.4 a CA-02.7: os erros de campo chegam do serviço sem gravar nada.
func TestCreate_CA02_4_to_CA02_7_InvalidFieldsSaveNothing(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1")
	cases := []struct {
		in          Input
		field, code string
	}{
		{Input{Nick: "Brasa", ClassID: "paladino-supremo", Level: 1, Role: "tank"}, FieldClassID, CodeInvalid},
		{Input{Nick: "Brasa", ClassID: "paladino", Level: 0, Role: "tank"}, FieldLevel, CodeInvalid},
		{Input{Nick: "Brasa", ClassID: "paladino", Level: 276, Role: "tank"}, FieldLevel, CodeInvalid},
		{Input{Nick: "   ", ClassID: "paladino", Level: 1, Role: "tank"}, FieldNick, CodeRequired},
		{Input{Nick: strings.Repeat("a", 25), ClassID: "paladino", Level: 1, Role: "tank"}, FieldNick, CodeTooLong},
		{Input{Nick: "Brasa", ClassID: "paladino", Level: 1, Role: "tank", Link: "http://exemplo.com"}, FieldLink, CodeInvalid},
		{Input{Nick: "Brasa", ClassID: "paladino", Level: 1, Role: "tank", Link: "javascript:alert(1)"}, FieldLink, CodeInvalid},
	}
	for _, c := range cases {
		_, err := e.svc.Create(t.Context(), ana, c.in)
		wantField(t, err, c.field, c.code)
	}
	if got := e.list(t, ana); len(got) != 0 {
		t.Errorf("gravou personagens inválidos: %+v", got)
	}
}

// CA-02.8 / RN-11: com 10 personagens, o décimo primeiro é recusado.
func TestCreate_CA02_8_LimitOfTen(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1")
	for i := range MaxPerUser {
		e.create(t, ana, fmt.Sprintf("Char%d", i))
	}
	_, err := e.svc.Create(t.Context(), ana, Input{Nick: "Sobra", ClassID: "aprendiz", Level: 1, Role: "dps"})
	if !errors.Is(err, ErrLimitReached) {
		t.Errorf("err = %v, quer ErrLimitReached", err)
	}
	if n := len(e.list(t, ana)); n != MaxPerUser {
		t.Errorf("%d personagens", n)
	}
}

// CA-07.3 / RN-05 / RN-11 / RN-12 / D-06: cadastros simultâneos não passam de 10, não
// repetem o nick e não criam dois principais.
func TestCreate_CA07_3_Concurrent(t *testing.T) {
	e := setup(t)
	ana, bia := e.user(t, "1"), e.user(t, "2")
	e.svc.Now = time.Now // o relógio de teste não é seguro entre goroutines

	var wg sync.WaitGroup
	errs := make([]error, 15)
	for i := range errs {
		wg.Go(func() {
			_, errs[i] = e.svc.Create(t.Context(), ana, Input{Nick: fmt.Sprintf("Ana%d", i), ClassID: "aprendiz", Level: 1, Role: "dps"})
		})
	}
	wg.Wait()
	limited := 0
	for _, err := range errs {
		switch {
		case errors.Is(err, ErrLimitReached):
			limited++
		case err != nil:
			t.Errorf("erro inesperado: %v", err)
		}
	}
	list := e.list(t, ana)
	if len(list) != MaxPerUser || limited != 5 {
		t.Errorf("%d criados e %d recusados, quer 10 e 5", len(list), limited)
	}
	mains := 0
	for _, c := range list {
		if c.IsMain {
			mains++
		}
	}
	if mains != 1 {
		t.Errorf("%d principais", mains)
	}

	// Dois Usuários pedindo o mesmo nick ao mesmo tempo: só um fica com ele. A ana está no
	// limite, então abre espaço antes.
	if err := e.svc.Delete(t.Context(), ana, list[MaxPerUser-1].ID); err != nil {
		t.Fatal(err)
	}
	results := make([]error, 2)
	for i, uid := range []string{ana, bia} {
		wg.Go(func() {
			_, results[i] = e.svc.Create(t.Context(), uid, Input{Nick: "Disputado", ClassID: "aprendiz", Level: 1, Role: "dps"})
		})
	}
	wg.Wait()
	ok, taken := 0, 0
	for _, err := range results {
		var ve *ValidationError
		switch {
		case err == nil:
			ok++
		case errors.As(err, &ve) && ve.Fields[0].Code == CodeTaken:
			taken++
		default:
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if ok != 1 || taken != 1 {
		t.Errorf("%d aceitos e %d recusados, quer 1 e 1", ok, taken)
	}
}

// CA-01.3 / RN-01 / RN-17: cada Usuário vê só os seus, o principal primeiro.
func TestList_CA01_3_RN17_OnlyOwnMainFirst(t *testing.T) {
	e := setup(t)
	ana, bia := e.user(t, "1"), e.user(t, "2")
	e.create(t, ana, "Brasa")
	lirien := e.create(t, ana, "Lirien")
	e.create(t, ana, "Faísca")
	e.create(t, bia, "Outro")

	if err := e.svc.SetMain(t.Context(), ana, lirien.ID); err != nil {
		t.Fatal(err)
	}
	if got := order(e.list(t, ana)); got != "Lirien*,Brasa,Faísca" {
		t.Errorf("ana vê %s", got)
	}
	if got := order(e.list(t, bia)); got != "Outro*" {
		t.Errorf("bia vê %s", got)
	}
}

// CA-03.1 / CA-03.2 / RN-15: editar muda os campos e aceita o próprio nick (até com
// outra caixa).
func TestUpdate_CA03_1_CA03_2(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1")
	faisca, err := e.svc.Create(t.Context(), ana, Input{Nick: "Faísca", ClassID: "feiticeiro", Level: 165, Role: "dps"})
	if err != nil {
		t.Fatal(err)
	}

	got, err := e.svc.Update(t.Context(), ana, faisca.ID, Input{Nick: "Faísca", ClassID: "elementalista", Level: 170, Role: "dps", Portrait: "retrato-2"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ClassID != "elementalista" || got.Level != 170 || got.Portrait != "retrato-2" || !got.IsMain || got.CreatedAt != faisca.CreatedAt {
		t.Errorf("editado = %+v", got)
	}

	got, err = e.svc.Update(t.Context(), ana, faisca.ID, Input{Nick: "FAÍSCA", ClassID: "elementalista", Level: 170, Role: "dps"})
	if err != nil || got.Nick != "FAÍSCA" {
		t.Errorf("mudar só a caixa do próprio nick: %+v, %v", got, err)
	}
}

// RN-05 / RN-15: editar para o nick de outro personagem é recusado.
func TestUpdate_RN05_NickOfAnotherCharacter(t *testing.T) {
	e := setup(t)
	ana, bia := e.user(t, "1"), e.user(t, "2")
	e.create(t, bia, "Brasa")
	mine := e.create(t, ana, "Lirien")

	_, err := e.svc.Update(t.Context(), ana, mine.ID, Input{Nick: "BRASA", ClassID: "arcebispo", Level: 178, Role: "support"})
	wantField(t, err, FieldNick, CodeTaken)
}

// Borda de RN-06: um personagem com classe fora do catálogo só é editado com uma classe
// válida.
func TestUpdate_RN06_RemovedClassAsksForValidOne(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1")
	c := e.create(t, ana, "Antigo")
	if _, err := e.pool.Exec(t.Context(), "UPDATE characters SET class_id = 'classe-removida'"); err != nil {
		t.Fatal(err)
	}
	if got := e.list(t, ana); got[0].ClassID != "classe-removida" {
		t.Fatalf("lista = %+v", got)
	}
	_, err := e.svc.Update(t.Context(), ana, c.ID, Input{Nick: "Antigo", ClassID: "classe-removida", Level: 2, Role: "support"})
	wantField(t, err, FieldClassID, CodeInvalid)
}

// CA-03.3 / CA-04.4 / RN-02: personagem de outro Usuário (ou id que não existe) é 404,
// e nada muda.
func TestOtherUsersCharacter_CA03_3_CA04_4_NotFound(t *testing.T) {
	e := setup(t)
	ana, bia := e.user(t, "1"), e.user(t, "2")
	brasa := e.create(t, ana, "Brasa")
	e.create(t, ana, "Lirien")
	e.create(t, bia, "Outro")

	in := Input{Nick: "Roubado", ClassID: "aprendiz", Level: 1, Role: "dps"}
	for _, id := range []string{brasa.ID, "00000000-0000-0000-0000-000000000000", "nao-e-uuid"} {
		if _, err := e.svc.Update(t.Context(), bia, id, in); !errors.Is(err, ErrNotFound) {
			t.Errorf("editar %s: err = %v", id, err)
		}
		if err := e.svc.Delete(t.Context(), bia, id); !errors.Is(err, ErrNotFound) {
			t.Errorf("excluir %s: err = %v", id, err)
		}
		if err := e.svc.SetMain(t.Context(), bia, id); !errors.Is(err, ErrNotFound) {
			t.Errorf("principal %s: err = %v", id, err)
		}
	}
	if got := order(e.list(t, ana)); got != "Brasa*,Lirien" {
		t.Errorf("ana ficou com %s", got)
	}
	if got := order(e.list(t, bia)); got != "Outro*" {
		t.Errorf("bia ficou com %s (perdeu o principal?)", got)
	}
}

// CA-04.1: excluir tira o personagem e libera o nick para outro Usuário.
func TestDelete_CA04_1_FreesNick(t *testing.T) {
	e := setup(t)
	ana, bia := e.user(t, "1"), e.user(t, "2")
	faisca := e.create(t, ana, "Faísca")

	if err := e.svc.Delete(t.Context(), ana, faisca.ID); err != nil {
		t.Fatal(err)
	}
	if got := e.list(t, ana); len(got) != 0 {
		t.Errorf("sobrou %+v", got)
	}
	e.create(t, bia, "faísca")
}

// CA-04.3 / RN-14: excluir o principal passa a vez ao mais antigo que sobrou.
func TestDelete_CA04_3_OldestBecomesMain(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1")
	e.create(t, ana, "Brasa")
	lirien := e.create(t, ana, "Lirien")
	e.create(t, ana, "Faísca")
	if err := e.svc.SetMain(t.Context(), ana, lirien.ID); err != nil {
		t.Fatal(err)
	}

	if err := e.svc.Delete(t.Context(), ana, lirien.ID); err != nil {
		t.Fatal(err)
	}
	if got := order(e.list(t, ana)); got != "Brasa*,Faísca" {
		t.Errorf("depois de excluir o principal: %s", got)
	}
}

// Borda de RN-12: excluir o único personagem deixa o Usuário sem principal, e o próximo
// cadastrado vira o principal.
func TestDelete_RN12_OnlyCharacter(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1")
	c := e.create(t, ana, "Brasa")
	if err := e.svc.Delete(t.Context(), ana, c.ID); err != nil {
		t.Fatal(err)
	}
	if got := e.create(t, ana, "Novo"); !got.IsMain {
		t.Error("o próximo personagem não virou principal")
	}
}

// RN-14: excluir um personagem que não é o principal não mexe no principal.
func TestDelete_RN14_NonMainKeepsMain(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1")
	e.create(t, ana, "Brasa")
	lirien := e.create(t, ana, "Lirien")
	if err := e.svc.Delete(t.Context(), ana, lirien.ID); err != nil {
		t.Fatal(err)
	}
	if got := order(e.list(t, ana)); got != "Brasa*" {
		t.Errorf("lista = %s", got)
	}
}

// CA-05.1 / RN-13: marcar outro como principal desmarca o anterior.
func TestSetMain_CA05_1(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1")
	e.create(t, ana, "Lirien")
	faisca := e.create(t, ana, "Faísca")

	if err := e.svc.SetMain(t.Context(), ana, faisca.ID); err != nil {
		t.Fatal(err)
	}
	if got := order(e.list(t, ana)); got != "Faísca*,Lirien" {
		t.Errorf("lista = %s", got)
	}
	// Marcar de novo o que já é principal não muda nada.
	if err := e.svc.SetMain(t.Context(), ana, faisca.ID); err != nil {
		t.Fatal(err)
	}
	if got := order(e.list(t, ana)); got != "Faísca*,Lirien" {
		t.Errorf("lista = %s", got)
	}
}

// RN-01: um Usuário apagado não cadastra nada.
func TestCreate_RN01_UnknownUser(t *testing.T) {
	e := setup(t)
	var id pgtype.UUID
	_ = id.Scan("00000000-0000-0000-0000-000000000001")
	_, err := e.svc.Create(t.Context(), id.String(), Input{Nick: "Fantasma", ClassID: "aprendiz", Level: 1, Role: "dps"})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v", err)
	}
}
