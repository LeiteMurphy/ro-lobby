//go:build integration

package characters

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

// Servidor do personagem (spec servidor, T-01).

func (e *env) createOn(t *testing.T, userID, nick, server string) Character {
	t.Helper()
	c, err := e.svc.Create(t.Context(), userID, Input{Nick: nick, ClassID: "arcebispo", Level: 178, Role: "support", Server: server})
	if err != nil {
		t.Fatalf("criar %q em %s: %v", nick, server, err)
	}
	return c
}

func onServer(c Character, server string) Input {
	return Input{Nick: c.Nick, ClassID: c.ClassID, Level: c.Level, Role: c.Role, Server: server}
}

func wantTaken(t *testing.T, err error) {
	t.Helper()
	var ve *ValidationError
	if !errors.As(err, &ve) || !reflect.DeepEqual(ve.Fields, []FieldError{{FieldNick, CodeTaken}}) {
		t.Errorf("err = %v, quer nick taken", err)
	}
}

// CA-01.2 / RN-02: o personagem é salvo com o servidor escolhido, e nasce conferido
// (RN-15).
func TestServer_CA01_2_SavedWithServer(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1")
	c := e.createOn(t, ana, "Brasa", "nidhogg")
	if c.Server != "nidhogg" || c.ServerToConfirm {
		t.Errorf("criado = %+v", c)
	}
	if list := e.list(t, ana); len(list) != 1 || list[0].Server != "nidhogg" {
		t.Errorf("lista = %+v", list)
	}
}

// CA-01.4 / RN-03: o mesmo nick em servidores diferentes, de Usuários diferentes.
func TestServer_CA01_4_SameNickOtherServer(t *testing.T) {
	e := setup(t)
	e.createOn(t, e.user(t, "1"), "Brasa", "freya")
	if _, err := e.svc.Create(t.Context(), e.user(t, "2"), Input{Nick: "brasa", ClassID: "arcebispo", Level: 178, Role: "support", Server: "nidhogg"}); err != nil {
		t.Errorf("brasa em Nidhogg: %v", err)
	}
}

// CA-01.5 / RN-03: nick repetido no mesmo servidor, ao criar e ao editar.
func TestServer_CA01_5_SameNickSameServer(t *testing.T) {
	e := setup(t)
	e.createOn(t, e.user(t, "1"), "Brasa", "freya")
	bia := e.user(t, "2")
	_, err := e.svc.Create(t.Context(), bia, Input{Nick: "brasa", ClassID: "arcebispo", Level: 178, Role: "support", Server: "freya"})
	wantTaken(t, err)

	other := e.createOn(t, bia, "Lirien", "freya")
	in := onServer(other, "freya")
	in.Nick = "BRASA"
	_, err = e.svc.Update(t.Context(), bia, other.ID, in)
	wantTaken(t, err)
}

// CA-01.6 / RN-03, RN-04: mudar de servidor e colidir com um nick de lá.
func TestServer_CA01_6_MoveCollides(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1")
	brasa := e.createOn(t, ana, "Brasa", "freya")
	e.createOn(t, e.user(t, "2"), "Brasa", "nidhogg")

	_, err := e.svc.Update(t.Context(), ana, brasa.ID, onServer(brasa, "nidhogg"))
	wantTaken(t, err)
	if list := e.list(t, ana); list[0].Server != "freya" {
		t.Errorf("servidor mudou: %+v", list[0])
	}
}

// CA-01.7 / RN-04: sem vínculo ativo, o servidor muda.
func TestServer_CA01_7_Edit(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1")
	brasa := e.createOn(t, ana, "Brasa", "freya")
	got, err := e.svc.Update(t.Context(), ana, brasa.ID, onServer(brasa, "nidhogg"))
	if err != nil || got.Server != "nidhogg" {
		t.Errorf("editar: %+v, %v", got, err)
	}
}

// CA-01.8 / RN-04: com lobby aberto como dono, candidatura pendente ou aceita, ou pedido de
// troca pendente, o servidor não muda.
func TestServer_CA01_8_LockedByActiveLink(t *testing.T) {
	e := setup(t)
	ana := e.user(t, "1")
	bia := e.user(t, "2")
	later := e.clock.Add(48 * time.Hour)

	owner := e.createOn(t, ana, "Dono", "freya")
	e.ownLobby(t, ana, owner.ID, "support", later, false)

	host := e.createOn(t, bia, "Anfitria", "freya")
	lobby := e.ownLobby(t, bia, host.ID, "support", later, false)
	pending := e.createOn(t, ana, "Pendente", "freya")
	e.applyTo(t, lobby, ana, pending.ID, "pending")

	host2 := e.createOn(t, bia, "Anfitria2", "freya")
	lobby2 := e.ownLobby(t, bia, host2.ID, "support", later.Add(4*time.Hour), false)
	accepted := e.createOn(t, ana, "Aceito", "freya")
	app := e.applyTo(t, lobby2, ana, accepted.ID, "accepted")
	target := e.createOn(t, ana, "Troca", "freya")
	e.swapTo(t, app, target.ID, "pending")

	for _, c := range []Character{owner, pending, accepted, target} {
		if _, err := e.svc.Update(t.Context(), ana, c.ID, onServer(c, "nidhogg")); !errors.Is(err, ErrInOpenLobby) {
			t.Errorf("%s: err = %v, quer ErrInOpenLobby", c.Nick, err)
		}
	}
	for _, c := range e.list(t, ana) {
		if c.Server != "freya" {
			t.Errorf("%s mudou para %s", c.Nick, c.Server)
		}
	}
}
