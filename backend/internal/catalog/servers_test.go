package catalog

import "testing"

// CA-01.1 / RN-01: Freya e Nidhogg, nesta ordem, com os códigos freya e nidhogg.
func TestServers_CA01_1_FreyaAndNidhoggInOrder(t *testing.T) {
	got := Servers()
	want := []Server{{ID: "freya", Name: "Freya"}, {ID: "nidhogg", Name: "Nidhogg"}}
	if len(got) != len(want) {
		t.Fatalf("Servers() = %+v, quer %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Servers()[%d] = %+v, quer %+v", i, got[i], want[i])
		}
	}
	got[0].Name = "mexido"
	if Servers()[0].Name != "Freya" {
		t.Error("Servers() devolve o catálogo em vez de uma cópia")
	}
}

// RN-02: só os códigos do catálogo valem.
func TestServerByID_RN02(t *testing.T) {
	if s, ok := ServerByID("nidhogg"); !ok || s.Name != "Nidhogg" {
		t.Errorf("ServerByID(nidhogg) = %+v, %v", s, ok)
	}
	for _, id := range []string{"", "Freya", "thor"} {
		if _, ok := ServerByID(id); ok {
			t.Errorf("ServerByID(%q) aceito", id)
		}
	}
	if _, ok := ServerByID(DefaultServer); !ok {
		t.Error("DefaultServer fora do catálogo")
	}
}
