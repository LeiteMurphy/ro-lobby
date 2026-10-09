//go:build integration

package talents_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/db"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
)

// Testes do catálogo e da contagem da criação (spec banco-de-talentos, T-04).

// trio põe no banco os personagens do CA-02.1: Brasa (Tank, 172, quarta 19–23, Templo),
// Fogo (Dano, 200, todo dia 18–00, qualquer instância) e Cura (Suporte, 200, quarta
// 19–23, Templo).
func trio(t *testing.T, e *env) {
	t.Helper()
	bia := e.user(t, "bia", "Brasa:guardiao-real:172:tank")
	caio := e.user(t, "caio", "Fogo:arquimago:200:dps")
	duda := e.user(t, "duda", "Cura:arcebispo:200:support")
	e.put(t, bia, "Brasa", temple([]int{wednesday}, "19:00", "23:00"))
	e.put(t, caio, "Fogo", every("18:00", "00:00"))
	e.put(t, duda, "Cura", temple([]int{wednesday}, "19:00", "23:00"))
}

func day(d int) *int { return &d }

// CA-01.1 / CA-04.1 / RN-11: sem filtro, todos por nível; por função e por instância.
func TestCatalog_CA01_1_CA04_1_Filters(t *testing.T) {
	e := setup(t)
	trio(t, e)
	all, err := e.svc.Catalog(t.Context(), talents.CatalogFilter{})
	if err != nil || !slices.Equal(nicks(all), []string{"Cura", "Fogo", "Brasa"}) {
		t.Fatalf("todos = %v, %v", nicks(all), err)
	}
	brasa := all[2]
	if !slices.Equal(brasa.Days, []int{wednesday}) || brasa.Start != "19:00" || brasa.End != "23:00" ||
		brasa.Instances[0].ID != "templo-do-demonio-rei" || brasa.DiscordUsername != "bia" {
		t.Errorf("Brasa = %+v", brasa)
	}
	cases := map[string]struct {
		f    talents.CatalogFilter
		want []string
	}{
		"Tank":          {talents.CatalogFilter{Role: "tank"}, []string{"Brasa"}},
		"Sonho Sombrio": {talents.CatalogFilter{InstanceID: "sonho-sombrio"}, []string{"Fogo"}},
		"Templo e Dano": {talents.CatalogFilter{InstanceID: "templo-do-demonio-rei", Role: "dps"}, []string{"Fogo"}},
		"quinta":        {talents.CatalogFilter{Day: day(4)}, []string{"Fogo"}},
		"18:30":         {talents.CatalogFilter{Time: "18:30"}, []string{"Fogo"}},
		"quarta 22:00":  {talents.CatalogFilter{Day: day(wednesday), Time: "22:00"}, []string{"Cura", "Fogo", "Brasa"}},
		"quarta 23:00":  {talents.CatalogFilter{Day: day(wednesday), Time: "23:00"}, []string{"Fogo"}},
	}
	for name, c := range cases {
		got, err := e.svc.Catalog(t.Context(), c.f)
		if err != nil || !slices.Equal(nicks(got), c.want) {
			t.Errorf("%s = %v, %v; quer %v", name, nicks(got), err, c.want)
		}
	}
}

// CA-01.2 / RN-03: sexta 22:00–02:00 aparece no sábado à 01:00, não no sábado às 22:00;
// só com o dia, aparece na sexta e no sábado.
func TestCatalog_CA01_2_Midnight(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "ana", "Lirien:arcebispo:178:support")
	e.put(t, ana, "Lirien", talents.AvailabilityInput{Days: []int{5}, Start: "22:00", End: "02:00", AnyInstance: true})
	cases := map[string]struct {
		f    talents.CatalogFilter
		want int
	}{
		"sábado 01:00": {talents.CatalogFilter{Day: day(6), Time: "01:00"}, 1},
		"sábado 02:00": {talents.CatalogFilter{Day: day(6), Time: "02:00"}, 0},
		"sábado 22:00": {talents.CatalogFilter{Day: day(6), Time: "22:00"}, 0},
		"sexta 23:30":  {talents.CatalogFilter{Day: day(5), Time: "23:30"}, 1},
		"sexta 01:00":  {talents.CatalogFilter{Day: day(5), Time: "01:00"}, 0},
		"sábado":       {talents.CatalogFilter{Day: day(6)}, 1},
		"sexta":        {talents.CatalogFilter{Day: day(5)}, 1},
		"domingo":      {talents.CatalogFilter{Day: day(0)}, 0},
		"01:00":        {talents.CatalogFilter{Time: "01:00"}, 1},
		"12:00":        {talents.CatalogFilter{Time: "12:00"}, 0},
	}
	for name, c := range cases {
		got, err := e.svc.Catalog(t.Context(), c.f)
		if err != nil || len(got) != c.want {
			t.Errorf("%s = %v, %v; quer %d", name, nicks(got), err, c.want)
		}
	}
}

// CA-04.3 / RN-01: desligado não aparece; filtro inválido é recusado.
func TestCatalog_CA04_3_DisabledAndInvalid(t *testing.T) {
	e := setup(t)
	trio(t, e)
	bia := e.user(t, "bia2", "Sombra:sicario:180:dps")
	e.put(t, bia, "Sombra", every("18:00", "00:00"))
	if _, err := e.svc.SetAvailability(t.Context(), bia.userID, bia.chars["Sombra"], talents.AvailabilityInput{}); err != nil {
		t.Fatal(err)
	}
	if got, _ := e.svc.Catalog(t.Context(), talents.CatalogFilter{}); slices.Contains(nicks(got), "Sombra") {
		t.Errorf("desligado apareceu: %v", nicks(got))
	}
	_, err := e.svc.Catalog(t.Context(), talents.CatalogFilter{Day: day(7), Time: "01:15", Role: "healer", InstanceID: "nao-existe"})
	for _, field := range []string{talents.FieldDay, talents.FieldTime, talents.FieldRole, talents.FieldInstanceID} {
		wantField(t, err, field, talents.CodeInvalid)
	}
}

// Risco do design: o catálogo devolve no máximo 100 personagens.
func TestCatalog_MaxRows(t *testing.T) {
	e := setup(t)
	w := e.user(t, "muitos")
	var uid pgtype.UUID
	if err := uid.Scan(w.userID); err != nil {
		t.Fatal(err)
	}
	for i := range talents.MaxCatalog + 1 {
		c, err := e.q.CreateCharacter(t.Context(), db.CreateCharacterParams{
			UserID: uid, Nick: fmt.Sprintf("Char%03d", i), ClassID: "arquimago", Level: 200, Role: "dps",
			Portrait: "retrato-1", Now: now,
		})
		if err != nil {
			t.Fatal(err)
		}
		w.chars[c.Nick] = c.ID.String()
		e.put(t, w, c.Nick, every("18:00", "00:00"))
	}
	if got, _ := e.svc.Catalog(t.Context(), talents.CatalogFilter{}); len(got) != talents.MaxCatalog {
		t.Errorf("%d personagens, quer %d", len(got), talents.MaxCatalog)
	}
}

// CA-03.1 (API) / RN-10: a contagem da criação usa as vagas do formulário menos a do
// dono; às 17:00 ninguém tem afinidade.
func TestCount_CA03_1(t *testing.T) {
	e := setup(t)
	trio(t, e)
	ana := e.user(t, "ana", "Lirien:arcebispo:178:support")
	in := talents.CountInput{
		InstanceID: "templo-do-demonio-rei", StartsAt: at(1, 20, 0), MinLevel: 160,
		Tank: 1, Support: 2, Dps: 3, CharacterID: ana.chars["Lirien"],
	}
	if n, err := e.svc.Count(t.Context(), ana.userID, in); err != nil || n != 3 {
		t.Errorf("20:00 = %d, %v", n, err)
	}
	in.Support = 1 // a vaga de Suporte é do dono
	if n, _ := e.svc.Count(t.Context(), ana.userID, in); n != 2 {
		t.Errorf("Suporte cheio = %d", n)
	}
	in.Support, in.StartsAt = 2, at(1, 17, 0)
	if n, _ := e.svc.Count(t.Context(), ana.userID, in); n != 0 {
		t.Errorf("17:00 = %d", n)
	}
	bia := e.user(t, "bia3", "Outro:arcebispo:200:support")
	in.CharacterID = bia.chars["Outro"]
	_, err := e.svc.Count(t.Context(), ana.userID, in)
	wantField(t, err, talents.FieldCharacterID, talents.CodeInvalid)
}
