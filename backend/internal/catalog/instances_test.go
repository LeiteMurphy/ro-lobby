package catalog

import (
	"strings"
	"testing"
)

// As instâncias "Solo" do bROWiki (consultado em 2026-10-06), que não entram no catálogo.
var soloInstances = []string{
	"Edda do Quarto Crescente", "Torneio de Magia", "Salão de Ymir", "Palácio das Mágoas",
	"Invasão ao Aeroplano",
}

// CA-06.1 / RN-01: o catálogo tem as 51 instâncias de grupo do bROWiki, sem as "Solo", e
// cada uma tem nome, nível de entrada e retorno.
func TestInstances_CA06_1_GroupInstancesFromBROWiki(t *testing.T) {
	got := Instances()
	if len(got) != 51 {
		t.Fatalf("%d instâncias, quer 51", len(got))
	}
	names := map[string]bool{}
	for _, i := range got {
		names[i.Name] = true
		if i.Name == "" || i.Level < 1 || i.Level > 275 {
			t.Errorf("instância incompleta: %+v", i)
		}
		switch i.Reset {
		case ResetDaily, ResetThreeDays, ResetHours, ResetWeekly:
		default:
			t.Errorf("%s: retorno %q", i.ID, i.Reset)
		}
	}
	for _, solo := range soloInstances {
		if names[solo] {
			t.Errorf("%q é Solo e não deveria estar no catálogo", solo)
		}
	}
	for _, want := range []struct {
		id    string
		level int
		reset Reset
	}{
		{"torre-da-constelacao", 240, ResetThreeDays},
		{"templo-do-demonio-rei", 160, ResetDaily},
		{"sarah-vs-fenrir", 145, ResetWeekly},
		{"altar-do-selo", 75, ResetHours},
		{"vila-dos-porings", 30, ResetDaily},
	} {
		i, ok := InstanceByID(want.id)
		if !ok || i.Level != want.level || i.Reset != want.reset {
			t.Errorf("%s = %+v, %v", want.id, i, ok)
		}
	}
}

// RN-01 / D-02: IDs únicos em kebab-case, tirados do nome sem acento.
func TestInstances_RN01_IDsAreUniqueKebabCase(t *testing.T) {
	seen := map[string]bool{}
	for _, i := range Instances() {
		if !kebab.MatchString(i.ID) {
			t.Errorf("ID %q não está em kebab-case", i.ID)
		}
		want := strings.Join(strings.FieldsFunc(unaccent.Replace(strings.ToLower(i.Name)), func(r rune) bool {
			return r == ' ' || r == '-'
		}), "-")
		if i.ID != want {
			t.Errorf("ID de %q = %q, quer %q", i.Name, i.ID, want)
		}
		if seen[i.ID] {
			t.Errorf("ID repetido: %q", i.ID)
		}
		seen[i.ID] = true
	}
	if _, ok := InstanceByID("caverna-de-gelo"); ok {
		t.Error("um nome inventado foi aceito")
	}
	if _, ok := InstanceByID(""); ok {
		t.Error("ID vazio foi aceito")
	}
}

// CA-06.2 / RN-02: nível 130+ primeiro, depois as de nível menor; em cada parte, nível
// decrescente e nome no empate.
func TestInstances_CA06_2_ChoiceOrder(t *testing.T) {
	got := Instances()
	if got[0].ID != "torre-da-constelacao" {
		t.Errorf("primeira = %s", got[0].ID)
	}
	crossed := false
	for k := 1; k < len(got); k++ {
		prev, cur := got[k-1], got[k]
		prevHigh, curHigh := prev.Level >= HighLevel, cur.Level >= HighLevel
		switch {
		case prevHigh && !curHigh:
			if crossed {
				t.Fatalf("o separador aparece duas vezes, em %s", cur.ID)
			}
			crossed = true
		case !prevHigh && curHigh:
			t.Fatalf("%s (nível %d) veio depois das de nível menor", cur.ID, cur.Level)
		case prev.Level < cur.Level:
			t.Errorf("%s (%d) antes de %s (%d)", prev.ID, prev.Level, cur.ID, cur.Level)
		case prev.Level == cur.Level && prev.Name > cur.Name:
			t.Errorf("empate fora da ordem de nome: %s antes de %s", prev.Name, cur.Name)
		}
	}
	if !crossed {
		t.Error("faltou a parte de nível menor")
	}
	// A parte de nível menor começa pela mais alta abaixo de 130.
	for _, i := range got {
		if i.Level < HighLevel {
			if i.ID != "sonho-sombrio" {
				t.Errorf("primeira de nível menor = %s", i.ID)
			}
			break
		}
	}
}

// Instances devolve uma cópia; mexer nela não muda o catálogo (RN-01).
func TestInstances_RN01_ReturnsACopy(t *testing.T) {
	Instances()[0].Name = "mudou"
	if Instances()[0].Name != "Torre da Constelação" {
		t.Error("o catálogo mudou por fora")
	}
}
