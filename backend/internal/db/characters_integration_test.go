//go:build integration

package db

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// Testes da tabela de personagens (spec personagens, T-01). Aqui só o banco: as mensagens
// e a ordem das regras ficam no serviço (internal/characters).

func newCharacter(userID pgtype.UUID, nick string) CreateCharacterParams {
	return CreateCharacterParams{
		UserID:   userID,
		Nick:     nick,
		ClassID:  "guardiao-real",
		Level:    172,
		Role:     "tank",
		Portrait: "retrato-1",
		Now:      t0,
	}
}

func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

func pgConstraint(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName
	}
	return ""
}

// CA-07.3 / RN-05: o nick é único em todo o RO Lobby, sem diferenciar maiúsculas, mesmo
// entre Usuários diferentes.
func TestCharacters_CA07_3_NickUniqueIgnoringCase(t *testing.T) {
	q, _ := setup(t)
	ana := upsert(t, q, "1", "ana", "", t0)
	bia := upsert(t, q, "2", "bia", "", t0)

	cases := []struct{ first, second string }{
		{"Brasa", "brasa"},
		{"FAÍSCA", "faísca"},
		{"Ômega", "ômega"},
	}
	for _, c := range cases {
		if _, err := q.CreateCharacter(t.Context(), newCharacter(ana.ID, c.first)); err != nil {
			t.Fatalf("criar %q: %v", c.first, err)
		}
		_, err := q.CreateCharacter(t.Context(), newCharacter(bia.ID, c.second))
		if pgCode(err) != "23505" || pgConstraint(err) != "characters_nick_key" {
			t.Errorf("%q depois de %q: err = %v, quer violação de characters_nick_key", c.second, c.first, err)
		}
	}
}

// RN-05 (borda): "Faísca" e "Faisca" são nicks diferentes.
func TestCharacters_RN05_AccentMakesADifferentNick(t *testing.T) {
	q, _ := setup(t)
	ana := upsert(t, q, "1", "ana", "", t0)

	for _, nick := range []string{"Faísca", "Faisca"} {
		if _, err := q.CreateCharacter(t.Context(), newCharacter(ana.ID, nick)); err != nil {
			t.Errorf("criar %q: %v", nick, err)
		}
	}
}

// RN-12: no máximo um principal por Usuário; Usuários diferentes têm cada um o seu.
func TestCharacters_RN12_OneMainPerUser(t *testing.T) {
	q, _ := setup(t)
	ana := upsert(t, q, "1", "ana", "", t0)
	bia := upsert(t, q, "2", "bia", "", t0)

	first := newCharacter(ana.ID, "Lirien")
	first.IsMain = true
	if _, err := q.CreateCharacter(t.Context(), first); err != nil {
		t.Fatal(err)
	}
	other := newCharacter(bia.ID, "Brasa")
	other.IsMain = true
	if _, err := q.CreateCharacter(t.Context(), other); err != nil {
		t.Fatalf("principal de outro Usuário: %v", err)
	}

	second := newCharacter(ana.ID, "Faísca")
	second.IsMain = true
	_, err := q.CreateCharacter(t.Context(), second)
	if pgCode(err) != "23505" || pgConstraint(err) != "characters_one_main_per_user" {
		t.Errorf("segundo principal: err = %v, quer violação de characters_one_main_per_user", err)
	}
}

// RN-04, RN-07, RN-08, RN-09: o banco recusa valores fora das regras dos campos.
func TestCharacters_RN04_RN07_RN08_RN09_ChecksRejectInvalidValues(t *testing.T) {
	q, _ := setup(t)
	ana := upsert(t, q, "1", "ana", "", t0)

	cases := map[string]func(p *CreateCharacterParams){
		"nick vazio":                     func(p *CreateCharacterParams) { p.Nick = "" },
		"nick com espaço na ponta":       func(p *CreateCharacterParams) { p.Nick = " Brasa" },
		"nick com 25 caracteres":         func(p *CreateCharacterParams) { p.Nick = strings.Repeat("a", 25) },
		"nick com caractere de controle": func(p *CreateCharacterParams) { p.Nick = "Bra\tsa" },
		"classe vazia":                   func(p *CreateCharacterParams) { p.ClassID = "" },
		"nível 0":                        func(p *CreateCharacterParams) { p.Level = 0 },
		"nível 276":                      func(p *CreateCharacterParams) { p.Level = 276 },
		"função desconhecida":            func(p *CreateCharacterParams) { p.Role = "healer" },
		"retrato vazio":                  func(p *CreateCharacterParams) { p.Portrait = "" },
		"link http":                      func(p *CreateCharacterParams) { p.Link = pgtype.Text{String: "http://exemplo.com", Valid: true} },
		"link javascript":                func(p *CreateCharacterParams) { p.Link = pgtype.Text{String: "javascript:alert(1)", Valid: true} },
		"link com 301 caracteres": func(p *CreateCharacterParams) {
			p.Link = pgtype.Text{String: "https://exemplo.com/" + strings.Repeat("a", 281), Valid: true}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			p := newCharacter(ana.ID, "Brasa")
			mutate(&p)
			_, err := q.CreateCharacter(t.Context(), p)
			if pgCode(err) != "23514" {
				t.Errorf("err = %v, quer violação de CHECK (23514)", err)
			}
		})
	}

	t.Run("valores nos limites são aceitos", func(t *testing.T) {
		p := newCharacter(ana.ID, strings.Repeat("a", 24))
		p.Level = 275
		p.Link = pgtype.Text{String: "https://exemplo.com/" + strings.Repeat("a", 280), Valid: true}
		if _, err := q.CreateCharacter(t.Context(), p); err != nil {
			t.Fatal(err)
		}
		p = newCharacter(ana.ID, "B")
		p.Level = 1
		if _, err := q.CreateCharacter(t.Context(), p); err != nil {
			t.Fatal(err)
		}
	})
}

// RN-01: os personagens saem junto com o Usuário.
func TestCharacters_RN01_DeletedWithUser(t *testing.T) {
	q, pool := setup(t)
	ana := upsert(t, q, "1", "ana", "", t0)
	if _, err := q.CreateCharacter(t.Context(), newCharacter(ana.ID, "Brasa")); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(t.Context(), "DELETE FROM users WHERE id = $1", ana.ID); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM characters").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("sobraram %d personagens", n)
	}
}

// RN-02: atualizar, excluir e marcar principal filtram pelo dono.
func TestCharacters_RN02_OtherUsersCharacterIsUntouched(t *testing.T) {
	q, _ := setup(t)
	ana := upsert(t, q, "1", "ana", "", t0)
	bia := upsert(t, q, "2", "bia", "", t0)
	c, err := q.CreateCharacter(t.Context(), newCharacter(ana.ID, "Brasa"))
	if err != nil {
		t.Fatal(err)
	}

	upd := UpdateCharacterParams{ID: c.ID, UserID: bia.ID, Nick: "Roubado", ClassID: c.ClassID, Level: 1, Role: "dps", Portrait: c.Portrait}
	if _, err := q.UpdateCharacter(t.Context(), upd); err == nil {
		t.Error("atualizou o personagem de outro Usuário")
	}
	if _, err := q.DeleteCharacter(t.Context(), DeleteCharacterParams{ID: c.ID, UserID: bia.ID}); err == nil {
		t.Error("excluiu o personagem de outro Usuário")
	}
	if n, err := q.SetMain(t.Context(), SetMainParams{ID: c.ID, UserID: bia.ID}); err != nil || n != 0 {
		t.Errorf("SetMain de outro Usuário: n = %d, err = %v", n, err)
	}

	list, err := q.ListCharacters(t.Context(), ana.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Nick != "Brasa" || list[0].IsMain {
		t.Errorf("personagem mudou: %+v", list)
	}
}

// RN-17 e RN-14: a lista traz o principal primeiro e depois a ordem de cadastro; o mais
// antigo é o primeiro cadastrado, mesmo com o mesmo horário.
func TestCharacters_RN17_RN14_OrderAndOldest(t *testing.T) {
	q, _ := setup(t)
	ana := upsert(t, q, "1", "ana", "", t0)

	for i, nick := range []string{"Brasa", "Lirien", "Faísca"} {
		p := newCharacter(ana.ID, nick)
		p.Now = t0.Add(time.Duration(i/2) * time.Minute) // Brasa e Lirien no mesmo instante
		p.IsMain = nick == "Lirien"
		if _, err := q.CreateCharacter(t.Context(), p); err != nil {
			t.Fatal(err)
		}
	}
	if got := nicks(t, q, ana.ID); got != "Lirien,Brasa,Faísca" {
		t.Errorf("ordem = %s", got)
	}

	if err := q.ClearMain(t.Context(), ana.ID); err != nil {
		t.Fatal(err)
	}
	if err := q.PromoteOldest(t.Context(), ana.ID); err != nil {
		t.Fatal(err)
	}
	if got := nicks(t, q, ana.ID); got != "Brasa,Lirien,Faísca" {
		t.Errorf("depois de promover o mais antigo, ordem = %s", got)
	}
}

func nicks(t *testing.T, q *Queries, userID pgtype.UUID) string {
	t.Helper()
	list, err := q.ListCharacters(t.Context(), userID)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(list))
	for i, c := range list {
		names[i] = c.Nick
	}
	return strings.Join(names, ",")
}
