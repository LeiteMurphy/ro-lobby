//go:build integration

package applications

import (
	"errors"
	"testing"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/lobbies"
)

// Formação do lobby com gente dentro (spec grupo-livre, T-01 e T-02).

func wantLobbyField(t *testing.T, err error, field, code string) {
	t.Helper()
	var ve *lobbies.ValidationError
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

// CA-04.3 / RN-07: com um membro aceito, ou só com uma candidatura pendente, a formação
// não muda.
func TestUpdate_CA04_3_FormationLocked(t *testing.T) {
	e, owner, player, lid := basic(t)
	toFree := lobbies.UpdateInput{StartsAt: at(1, 20, 0), MinLevel: 160, Formation: lobbies.FormationFree, FreeSlots: 12}

	pending := e.apply(t, player, "Cura", lid)
	_, err := e.lobbies.Update(t.Context(), owner.id, lid, toFree)
	wantLobbyField(t, err, lobbies.FieldFormation, lobbies.CodeLocked)

	if _, err := e.svc.Accept(t.Context(), owner.id, pending.ID); err != nil {
		t.Fatal(err)
	}
	_, err = e.lobbies.Update(t.Context(), owner.id, lid, toFree)
	wantLobbyField(t, err, lobbies.FieldFormation, lobbies.CodeLocked)
	if l, _ := e.lobbies.Get(t.Context(), lid); l.Formation != lobbies.FormationRoles {
		t.Errorf("formação mudou: %+v", l)
	}
}
