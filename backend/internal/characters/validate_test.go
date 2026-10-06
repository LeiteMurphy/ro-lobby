package characters

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func valid() Input {
	return Input{Nick: "Brasa", ClassID: "guardiao-real", Level: 172, Role: "tank"}
}

func fieldErrors(t *testing.T, err error) []FieldError {
	t.Helper()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v, quer *ValidationError", err)
	}
	return ve.Fields
}

// CA-02.2 / RN-10: entrada válida é aceita, e sem retrato fica o primeiro da lista.
func TestValidate_CA02_2_AcceptsValidInputWithDefaultPortrait(t *testing.T) {
	got, err := Validate(valid())
	if err != nil {
		t.Fatal(err)
	}
	if got.Portrait != "retrato-1" {
		t.Errorf("retrato padrão = %q", got.Portrait)
	}
}

// Borda de RN-04: " Brasa " é salvo como "Brasa".
func TestValidate_RN04_TrimsNick(t *testing.T) {
	in := valid()
	in.Nick = "  Brasa  "
	got, err := Validate(in)
	if err != nil || got.Nick != "Brasa" {
		t.Errorf("nick = %q, err = %v", got.Nick, err)
	}
}

// CA-02.4, CA-02.5, CA-02.6, CA-02.7: cada campo errado volta com o código certo.
func TestValidate_CA02_4_to_CA02_7_FieldErrors(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Input)
		want   FieldError
	}{
		{"CA-02.6 nick só com espaços", func(in *Input) { in.Nick = "   " }, FieldError{FieldNick, CodeRequired}},
		{"CA-02.6 nick com 25 caracteres", func(in *Input) { in.Nick = strings.Repeat("a", 25) }, FieldError{FieldNick, CodeTooLong}},
		{"RN-04 nick com caractere de controle", func(in *Input) { in.Nick = "Bra\u0000sa" }, FieldError{FieldNick, CodeInvalid}},
		{"CA-02.4 classe fora do catálogo", func(in *Input) { in.ClassID = "paladino-supremo" }, FieldError{FieldClassID, CodeInvalid}},
		{"CA-02.4 classe pelo nome", func(in *Input) { in.ClassID = "Paladino Supremo" }, FieldError{FieldClassID, CodeInvalid}},
		{"RN-06 sem classe", func(in *Input) { in.ClassID = "" }, FieldError{FieldClassID, CodeRequired}},
		{"CA-02.5 nível 0", func(in *Input) { in.Level = 0 }, FieldError{FieldLevel, CodeInvalid}},
		{"CA-02.5 nível 276", func(in *Input) { in.Level = 276 }, FieldError{FieldLevel, CodeInvalid}},
		{"RN-08 sem função", func(in *Input) { in.Role = "" }, FieldError{FieldRole, CodeRequired}},
		{"RN-08 função desconhecida", func(in *Input) { in.Role = "healer" }, FieldError{FieldRole, CodeInvalid}},
		{"RN-10 retrato fora da lista", func(in *Input) { in.Portrait = "retrato-9" }, FieldError{FieldPortrait, CodeInvalid}},
		{"CA-02.7 link http", func(in *Input) { in.Link = "http://exemplo.com" }, FieldError{FieldLink, CodeInvalid}},
		{"CA-02.7 link javascript", func(in *Input) { in.Link = "javascript:alert(1)" }, FieldError{FieldLink, CodeInvalid}},
		{"D-09 link sem host", func(in *Input) { in.Link = "https://" }, FieldError{FieldLink, CodeInvalid}},
		{"D-09 link com usuário", func(in *Input) { in.Link = "https://eu:senha@exemplo.com" }, FieldError{FieldLink, CodeInvalid}},
		{"D-09 link com espaço", func(in *Input) { in.Link = "https://exemplo.com/a b" }, FieldError{FieldLink, CodeInvalid}},
		{"RN-09 link com 301 caracteres", func(in *Input) { in.Link = "https://exemplo.com/" + strings.Repeat("a", 281) }, FieldError{FieldLink, CodeTooLong}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := valid()
			c.mutate(&in)
			_, err := Validate(in)
			if got := fieldErrors(t, err); !reflect.DeepEqual(got, []FieldError{c.want}) {
				t.Errorf("erros = %v, quer %v", got, c.want)
			}
		})
	}
}

// RN-19: todos os campos errados voltam juntos.
func TestValidate_RN19_ReportsAllFields(t *testing.T) {
	_, err := Validate(Input{Level: 999, Link: "ftp://x"})
	got := fieldErrors(t, err)
	want := []FieldError{
		{FieldNick, CodeRequired}, {FieldClassID, CodeRequired}, {FieldLevel, CodeInvalid},
		{FieldRole, CodeRequired}, {FieldLink, CodeInvalid},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("erros = %v, quer %v", got, want)
	}
}

// RN-04, RN-07, RN-09: os limites são aceitos.
func TestValidate_RN04_RN07_RN09_AcceptsLimits(t *testing.T) {
	for _, in := range []Input{
		{Nick: strings.Repeat("ã", 24), ClassID: "aprendiz", Level: 1, Role: "dps", Portrait: "retrato-4"},
		{Nick: "B", ClassID: "animista", Level: 275, Role: "support", Link: "https://exemplo.com/" + strings.Repeat("a", 280)},
		{Nick: "Faísca", ClassID: "feiticeiro", Level: 165, Role: "dps", Link: "  https://exemplo.com/char/123  "},
	} {
		if _, err := Validate(in); err != nil {
			t.Errorf("%+v: %v", in, err)
		}
	}
}
