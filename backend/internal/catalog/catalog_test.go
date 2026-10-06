package catalog

import (
	"regexp"
	"strings"
	"testing"
)

// Títulos das páginas de classe em https://browiki.org/wiki/Classes, como estavam em
// 2026-10-01, na ordem da página. É a mesma lista do teste do catálogo do web.
var browikiPages = []string{
	"Aprendizes",
	"Espadachins",
	"Cavaleiros",
	"Lordes",
	"Cavaleiros Rúnicos",
	"Cavaleiros Draconianos",
	"Templários",
	"Paladinos",
	"Guardiões Reais",
	"Guardiões Imperiais",
	"Magos",
	"Bruxos",
	"Arquimagos",
	"Arcanos",
	"Magus",
	"Sábios",
	"Professores",
	"Feiticeiros",
	"Elementalistas",
	"Gatunos",
	"Mercenários",
	"Algozes",
	"Sicários",
	"Executores",
	"Arruaceiros",
	"Desordeiros",
	"Renegados",
	"Mandraques",
	"Mercadores",
	"Ferreiros",
	"Mestres-Ferreiros",
	"Mecânicos",
	"Engenheiros",
	"Alquimistas",
	"Criadores",
	"Bioquímicos",
	"Cientistas",
	"Noviços",
	"Sacerdotes",
	"Sumo Sacerdotes",
	"Arcebispos",
	"Cardeais",
	"Monges",
	"Mestres",
	"Shuras",
	"Inquisidores",
	"Arqueiros",
	"Caçadores",
	"Atiradores de Elite",
	"Sentinelas",
	"Falcões do Vento",
	"Bardos",
	"Odaliscas",
	"Menestréis",
	"Ciganas",
	"Trovadores",
	"Musas",
	"Maestros",
	"Divas",
	"Taekwons",
	"Mestres Taekwons",
	"Mestres Estelares",
	"Mestres Celestiais",
	"Espiritualistas",
	"Ceifadores de Almas",
	"Ascetas das Almas",
	"Superaprendizes",
	"Superaprendizes EX",
	"Hiperaprendizes",
	"Justiceiros",
	"Insurgentes",
	"Guerrilheiros",
	"Ninjas",
	"Kagerou",
	"Oboro",
	"Shinkiro",
	"Shiranui",
	"Druidas",
	"Karnos",
	"Alitea",
	"Invocadores",
	"Animistas",
}

// CA-07.1 / RN-20: o catálogo tem as 82 classes do bROWiki, na ordem da página.
func TestClasses_CA07_1_MatchBROWiki(t *testing.T) {
	got := Classes()
	if len(got) != len(browikiPages) {
		t.Fatalf("%d classes, quer %d", len(got), len(browikiPages))
	}
	for i, c := range got {
		if c.Plural != browikiPages[i] {
			t.Errorf("classe %d: plural %q, quer %q", i, c.Plural, browikiPages[i])
		}
		if c.Name == "" || c.Family == "" || c.Tier == "" {
			t.Errorf("classe %q incompleta: %+v", c.ID, c)
		}
	}
}

var kebab = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

var unaccent = strings.NewReplacer(
	"á", "a", "â", "a", "ã", "a", "à", "a", "é", "e", "ê", "e", "í", "i",
	"ó", "o", "ô", "o", "õ", "o", "ú", "u", "ç", "c",
)

// D-03: o ID é o nome sem acento em kebab-case, e não se repete.
func TestClasses_D03_IDsAreUniqueKebabCase(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range Classes() {
		if !kebab.MatchString(c.ID) {
			t.Errorf("ID %q não está em kebab-case", c.ID)
		}
		want := strings.Join(strings.FieldsFunc(unaccent.Replace(strings.ToLower(c.Name)), func(r rune) bool {
			return r == ' ' || r == '-'
		}), "-")
		if c.ID != want {
			t.Errorf("ID de %q = %q, quer %q", c.Name, c.ID, want)
		}
		if seen[c.ID] {
			t.Errorf("ID repetido: %q", c.ID)
		}
		seen[c.ID] = true
	}
}

// RN-06: só classes do catálogo são encontradas.
func TestClassByID_RN06(t *testing.T) {
	if c, ok := ClassByID("guardiao-real"); !ok || c.Name != "Guardião Real" {
		t.Errorf("guardiao-real = %+v, %v", c, ok)
	}
	for _, id := range []string{"", "paladino-supremo", "Guardião Real", "GUARDIAO-REAL"} {
		if _, ok := ClassByID(id); ok {
			t.Errorf("%q foi aceito", id)
		}
	}
}

// Classes devolve uma cópia: mexer nela não muda o catálogo.
func TestClasses_ReturnsACopy(t *testing.T) {
	Classes()[0].Name = "mudou"
	if Classes()[0].Name != "Aprendiz" {
		t.Error("o catálogo mudou por fora")
	}
}

// RN-10: 4 retratos de exemplo, com o primeiro como padrão.
func TestPortraits_RN10(t *testing.T) {
	got := Portraits()
	if strings.Join(got, ",") != "retrato-1,retrato-2,retrato-3,retrato-4" {
		t.Errorf("retratos = %v", got)
	}
	if got[0] != DefaultPortrait {
		t.Errorf("o padrão %q não é o primeiro da lista", DefaultPortrait)
	}
	for _, p := range got {
		if !HasPortrait(p) {
			t.Errorf("HasPortrait(%q) = false", p)
		}
	}
	for _, p := range []string{"", "retrato-5", "Retrato-1", "../retrato-1"} {
		if HasPortrait(p) {
			t.Errorf("HasPortrait(%q) = true", p)
		}
	}
}
