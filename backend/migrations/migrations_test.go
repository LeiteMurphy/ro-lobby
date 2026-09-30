package migrations

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

var (
	upAnnotation   = regexp.MustCompile(`(?m)^--\s*\+goose\s+Up\b`)
	downAnnotation = regexp.MustCompile(`(?m)^--\s*\+goose\s+Down\b`)
)

// CA-03.3: todo arquivo de migração tem uma seção up e uma seção down.
func TestMigracoes_CA03_3_TodasTemUpEDown(t *testing.T) {
	files, err := fs.Glob(FS, "*.sql")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("nenhuma migração encontrada")
	}
	for _, name := range files {
		content, err := fs.ReadFile(FS, name)
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		up := upAnnotation.FindStringIndex(text)
		down := downAnnotation.FindStringIndex(text)
		switch {
		case up == nil:
			t.Errorf("%s: sem -- +goose Up", name)
		case down == nil:
			t.Errorf("%s: sem -- +goose Down", name)
		case down[0] < up[0]:
			t.Errorf("%s: a seção Down vem antes da Up", name)
		case strings.TrimSpace(text[down[1]:]) == "":
			t.Errorf("%s: a seção Down está vazia", name)
		}
	}
}
