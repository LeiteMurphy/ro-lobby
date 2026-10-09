//go:build integration

package talents_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/applications"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/talents"
)

// Testes da afinidade no detalhe do lobby (spec banco-de-talentos, T-03).

const wednesday = 3

// glast monta o lobby do CA-02.1: Templo na quarta (amanhã) às 20:00, nível 160, com 1
// vaga de Tank, 1 de Suporte (do dono) e 3 de Dano.
func glast(t *testing.T, e *env) (who, string) {
	t.Helper()
	ana := e.user(t, "ana", "Lirien:arcebispo:178:support", "Alt:sicario:200:dps")
	l, err := e.lobbies.Create(t.Context(), ana.userID, lobbies.Input{
		InstanceID: "templo-do-demonio-rei", StartsAt: at(1, 20, 0),
		Slots: lobbies.Slots{Tank: 1, Support: 1, Dps: 3}, MinLevel: 160, CharacterID: ana.chars["Lirien"],
	})
	if err != nil {
		t.Fatal(err)
	}
	return ana, l.ID
}

func nicks(list []talents.Talent) []string {
	out := make([]string, len(list))
	for i, x := range list {
		out[i] = x.Nick
	}
	return out
}

func temple(days []int, start, end string) talents.AvailabilityInput {
	return talents.AvailabilityInput{Days: days, Start: start, End: end, InstanceIDs: []string{"templo-do-demonio-rei"}}
}

func every(start, end string) talents.AvailabilityInput {
	return talents.AvailabilityInput{Days: []int{0, 1, 2, 3, 4, 5, 6}, Start: start, End: end, AnyInstance: true}
}

// CA-02.1 / RN-06 / RN-08 / RN-12: Fogo (200) e Brasa (172), por nível; Cura fica de
// fora porque o Suporte está cheio.
func TestForLobby_CA02_1(t *testing.T) {
	e := setup(t)
	ana, lid := glast(t, e)
	bia := e.user(t, "bia", "Brasa:guardiao-real:172:tank")
	caio := e.user(t, "caio", "Fogo:arquimago:200:dps")
	duda := e.user(t, "duda", "Cura:arcebispo:200:support")
	e.put(t, bia, "Brasa", temple([]int{wednesday}, "19:00", "23:00"))
	e.put(t, caio, "Fogo", every("18:00", "00:00"))
	e.put(t, duda, "Cura", temple([]int{wednesday}, "19:00", "23:00"))

	got, err := e.svc.ForLobby(t.Context(), ana.userID, lid)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(nicks(got), []string{"Fogo", "Brasa"}) {
		t.Fatalf("lista = %v", nicks(got))
	}
	brasa := got[1]
	if brasa.Level != 172 || brasa.Role != "tank" || brasa.DiscordUsername != "bia" || brasa.Start != "19:00" ||
		len(brasa.Instances) != 1 || brasa.Instances[0].Name != "Templo do Demônio Rei" {
		t.Errorf("Brasa = %+v", brasa)
	}
}

// CA-02.2 / RN-06 / RN-03: sem a instância, sem a quarta às 20:00 ou abaixo do nível,
// fica de fora; o fim da faixa não entra; a faixa que vira a meia-noite conta no dia
// seguinte.
func TestForLobby_CA02_2_OutOfAffinity(t *testing.T) {
	e := setup(t)
	ana, lid := glast(t, e)
	p := e.user(t, "pedro",
		"Outra:arquimago:200:dps", "Quinta:arquimago:200:dps", "Baixo:arquimago:150:dps",
		"AteVinte:arquimago:200:dps", "Virada:arquimago:200:dps")
	e.put(t, p, "Outra", talents.AvailabilityInput{Days: []int{wednesday}, Start: "19:00", End: "23:00", InstanceIDs: []string{"sonho-sombrio"}})
	e.put(t, p, "Quinta", temple([]int{4}, "19:00", "23:00"))
	e.put(t, p, "Baixo", temple([]int{wednesday}, "19:00", "23:00"))
	e.put(t, p, "AteVinte", temple([]int{wednesday}, "18:00", "20:00"))
	got, err := e.svc.ForLobby(t.Context(), ana.userID, lid)
	if err != nil || len(got) != 0 {
		t.Fatalf("lista = %v, %v", nicks(got), err)
	}

	// Terça 22:00–02:00 cobre quarta 01:00; um lobby nessa hora traz a Virada.
	e.put(t, p, "Virada", temple([]int{2}, "22:00", "02:00"))
	late, err := e.lobbies.Create(t.Context(), ana.userID, lobbies.Input{
		InstanceID: "templo-do-demonio-rei", StartsAt: at(1, 1, 0),
		Slots: lobbies.Slots{Tank: 1, Support: 1, Dps: 3}, MinLevel: 160, CharacterID: ana.chars["Lirien"],
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err = e.svc.ForLobby(t.Context(), ana.userID, late.ID)
	if err != nil || !slices.Equal(nicks(got), []string{"Virada"}) {
		t.Errorf("madrugada = %v, %v", nicks(got), err)
	}
}

// CA-02.3 / RN-06 / RN-07: candidato pendente neste lobby, aceito num lobby às 21:00 do
// mesmo dia e personagem do próprio dono ficam de fora.
func TestForLobby_CA02_3_Busy(t *testing.T) {
	e := setup(t)
	ana, lid := glast(t, e)
	e.put(t, ana, "Alt", every("18:00", "00:00"))
	bia := e.user(t, "bia", "Brasa:guardiao-real:172:tank")
	caio := e.user(t, "caio", "Fogo:arquimago:200:dps")
	duda := e.user(t, "duda", "Outro:arcebispo:200:support", "Livre:arquimago:199:dps")
	e.put(t, bia, "Brasa", every("18:00", "00:00"))
	e.put(t, caio, "Fogo", every("18:00", "00:00"))
	e.put(t, duda, "Livre", every("18:00", "00:00"))
	if got, _ := e.svc.ForLobby(t.Context(), ana.userID, lid); !slices.Equal(nicks(got), []string{"Fogo", "Livre", "Brasa"}) {
		t.Fatalf("antes = %v", nicks(got))
	}

	if _, err := e.apps.Apply(t.Context(), bia.userID, lid, applications.ApplyInput{CharacterID: bia.chars["Brasa"]}); err != nil {
		t.Fatal(err)
	}
	other, err := e.lobbies.Create(t.Context(), duda.userID, lobbies.Input{
		InstanceID: "templo-do-demonio-rei", StartsAt: at(1, 21, 0),
		Slots: lobbies.Slots{Tank: 1, Support: 1, Dps: 3}, MinLevel: 160, CharacterID: duda.chars["Outro"],
	})
	if err != nil {
		t.Fatal(err)
	}
	a, err := e.apps.Apply(t.Context(), caio.userID, other.ID, applications.ApplyInput{CharacterID: caio.chars["Fogo"]})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.apps.Accept(t.Context(), duda.userID, a.ID); err != nil {
		t.Fatal(err)
	}
	got, err := e.svc.ForLobby(t.Context(), ana.userID, lid)
	if err != nil || !slices.Equal(nicks(got), []string{"Livre"}) {
		t.Errorf("depois = %v, %v", nicks(got), err)
	}
}

// CA-02.4 (API) / RN-09: lobby de outro Usuário é ErrNotFound; cancelado, ErrNotOpen;
// sem vaga aberta, a lista vem vazia.
func TestForLobby_CA02_4_OwnerAndOpen(t *testing.T) {
	e := setup(t)
	ana, lid := glast(t, e)
	bia := e.user(t, "bia", "Brasa:guardiao-real:172:tank")
	e.put(t, bia, "Brasa", every("18:00", "00:00"))
	if _, err := e.svc.ForLobby(t.Context(), bia.userID, lid); !errors.Is(err, talents.ErrNotFound) {
		t.Errorf("de outro: %v", err)
	}
	if _, err := e.svc.ForLobby(t.Context(), ana.userID, "nao-e-uuid"); !errors.Is(err, talents.ErrNotFound) {
		t.Errorf("inexistente: %v", err)
	}
	full, err := e.lobbies.Create(t.Context(), ana.userID, lobbies.Input{
		InstanceID: "templo-do-demonio-rei", StartsAt: at(3, 20, 0),
		Slots: lobbies.Slots{Support: 1}, MinLevel: 160, CharacterID: ana.chars["Lirien"],
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := e.svc.ForLobby(t.Context(), ana.userID, full.ID); err != nil || len(got) != 0 {
		t.Errorf("lotado = %v, %v", nicks(got), err)
	}
	if _, err := e.lobbies.Cancel(t.Context(), ana.userID, lid, "Metade do grupo não pode"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.ForLobby(t.Context(), ana.userID, lid); !errors.Is(err, talents.ErrNotOpen) {
		t.Errorf("cancelado: %v", err)
	}
}
