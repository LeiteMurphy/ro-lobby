package lobbies

import "testing"

// Vaga por formação (spec grupo-livre, RN-03, RN-04, design seção 3).

func TestHasRoom_CA02_1_CA02_2(t *testing.T) {
	roles := Lobby{Formation: FormationRoles, Slots: Slots{Tank: 1, Support: 2}, Occupied: Slots{Tank: 1, Support: 1}}
	free := Lobby{Formation: FormationFree, FreeSlots: 3, Occupied: Slots{Support: 2}}
	full := Lobby{Formation: FormationFree, FreeSlots: 3, Occupied: Slots{Support: 2, Dps: 1}}
	cases := []struct {
		name string
		l    Lobby
		role string
		want bool
	}{
		{"por função, Tank cheio", roles, "tank", false},
		{"por função, Suporte com vaga", roles, "support", true},
		{"por função, Dano sem vaga", roles, "dps", false},
		{"livre, qualquer função com vaga no total", free, "dps", true},
		{"livre, terceiro Suporte", free, "support", true},
		{"livre cheio", full, "tank", false},
	}
	for _, c := range cases {
		if got := HasRoom(c.l, c.role); got != c.want {
			t.Errorf("%s: %v, quer %v", c.name, got, c.want)
		}
	}
}

// CA-01.2 / RN-01 / RN-02: formação e vagas válidas.
func TestCheckFormation_CA01_2(t *testing.T) {
	if f, errs := checkFormation("", Slots{Tank: 1, Support: 2, Dps: 3}, 0); f != FormationRoles || errs != nil {
		t.Errorf("padrão = %q, %v", f, errs)
	}
	for _, n := range []int{2, 12} {
		if _, errs := checkFormation(FormationFree, Slots{}, n); errs != nil {
			t.Errorf("livre com %d: %v", n, errs)
		}
	}
	for _, c := range []struct {
		slots Slots
		free  int
	}{{Slots{}, 1}, {Slots{}, 13}, {Slots{Dps: 1}, 12}} {
		_, errs := checkFormation(FormationFree, c.slots, c.free)
		if len(errs) != 1 || errs[0] != (FieldError{FieldFreeSlots, CodeInvalid}) {
			t.Errorf("livre %+v %d: %v", c.slots, c.free, errs)
		}
	}
	if _, errs := checkFormation("misto", Slots{}, 0); len(errs) != 1 || errs[0].Field != FieldFormation {
		t.Errorf("formação desconhecida: %v", errs)
	}
}
