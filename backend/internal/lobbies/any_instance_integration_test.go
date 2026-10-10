//go:build integration

package lobbies

import (
	"strings"
	"testing"
)

// Lobby sem instância no serviço (spec lobby-sem-instancia, T-01).

func anyLobby(characterID, title string, minLevel int) Input {
	in := temple(characterID, at(1, 20, 0))
	in.InstanceID, in.AnyInstance, in.Title, in.MinLevel = "", true, title, minLevel
	return in
}

// CA-01.1 / CA-01.2 / RN-02 / RN-04: com título, o nome é o título; sem título (ou só
// espaços), "Qualquer instância"; nível de instância 1 e sem retorno.
func TestCreate_CA01_1_CA01_2_AnyInstance(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	l, err := e.svc.Create(t.Context(), ana.userID, anyLobby(ana.chars["Lirien"], "  Caça ao MVP  ", 1))
	if err != nil {
		t.Fatal(err)
	}
	if !l.AnyInstance || l.InstanceID != "" || l.InstanceName != "Caça ao MVP" || l.Title != "Caça ao MVP" ||
		l.InstanceLevel != 1 || l.InstanceReset != "" || l.MinLevel != 1 {
		t.Errorf("com título = %+v", l)
	}
	for _, title := range []string{"", "   "} {
		l, err := e.svc.Create(t.Context(), ana.userID, func() Input {
			in := anyLobby(ana.chars["Lirien"], title, 1)
			in.StartsAt = at(2+len(title), 20, 0)
			return in
		}())
		if err != nil || l.InstanceName != AnyInstanceName || l.Title != "" {
			t.Errorf("título %q = %+v, %v", title, l, err)
		}
	}
}

// CA-01.3 / RN-02: título de 41 caracteres é recusado; 40 passa.
func TestCreate_CA01_3_TitleTooLong(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	_, err := e.svc.Create(t.Context(), ana.userID, anyLobby(ana.chars["Lirien"], strings.Repeat("é", 41), 1))
	wantField(t, err, FieldTitle, CodeTooLong)
	if _, err := e.svc.Create(t.Context(), ana.userID, anyLobby(ana.chars["Lirien"], strings.Repeat("é", 40), 1)); err != nil {
		t.Errorf("40 caracteres: %v", err)
	}
}

// CA-01.4 / RN-03: sem instância, o nível vai de 1 até o do personagem do dono.
func TestCreate_CA01_4_MinLevel(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Novato:arcebispo:120:support")
	if _, err := e.svc.Create(t.Context(), ana.userID, anyLobby(ana.chars["Novato"], "", 120)); err != nil {
		t.Errorf("nível 120: %v", err)
	}
	in := anyLobby(ana.chars["Novato"], "", 121)
	in.StartsAt = at(3, 20, 0)
	_, err := e.svc.Create(t.Context(), ana.userID, in)
	wantField(t, err, FieldCharacterID, CodeLevelTooLow)
	in.MinLevel = 0
	_, err = e.svc.Create(t.Context(), ana.userID, in)
	wantField(t, err, FieldMinLevel, CodeInvalid)
}

// CA-03.1 / RN-06: de instância para "sem instância" com título, e de volta para outra
// instância com o nível de entrada dela.
func TestUpdate_CA03_1_SwitchAnyInstance(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1", "Lirien:arcebispo:178:support")
	l, err := e.svc.Create(t.Context(), ana.userID, temple(ana.chars["Lirien"], at(1, 20, 0)))
	if err != nil {
		t.Fatal(err)
	}
	got, err := e.svc.Update(t.Context(), ana.userID, l.ID, UpdateInput{
		AnyInstance: true, Title: "Farm", StartsAt: at(1, 20, 0), Slots: Slots{1, 2, 3}, MinLevel: 10,
	})
	if err != nil || !got.AnyInstance || got.InstanceName != "Farm" || got.MinLevel != 10 {
		t.Fatalf("sem instância = %+v, %v", got, err)
	}
	// Sem AnyInstance e sem instância, continua sem instância, com o título.
	got, err = e.svc.Update(t.Context(), ana.userID, l.ID, UpdateInput{StartsAt: at(1, 21, 0), Slots: Slots{1, 2, 3}, MinLevel: 10})
	if err != nil || !got.AnyInstance || got.Title != "Farm" {
		t.Errorf("mantém = %+v, %v", got, err)
	}
	got, err = e.svc.Update(t.Context(), ana.userID, l.ID, UpdateInput{
		InstanceID: "sonho-sombrio", StartsAt: at(1, 20, 0), Slots: Slots{1, 2, 3}, MinLevel: 120,
	})
	if err != nil || got.AnyInstance || got.InstanceID != "sonho-sombrio" || got.InstanceLevel != 120 {
		t.Errorf("de volta = %+v, %v", got, err)
	}
	_, err = e.svc.Update(t.Context(), ana.userID, l.ID, UpdateInput{
		InstanceID: "sonho-sombrio", StartsAt: at(1, 20, 0), Slots: Slots{1, 2, 3}, MinLevel: 10,
	})
	wantField(t, err, FieldMinLevel, CodeInvalid)
}
