package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func envFrom(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

// RN-01: toda configuração vem de variáveis de ambiente.
func TestLoad_RN01_ReadsEnvironment(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{
		"DATABASE_URL": "postgres://u:p@localhost:5432/db",
		"API_PORT":     "9090",
	}))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cfg.DatabaseURL != "postgres://u:p@localhost:5432/db" || cfg.Port != 9090 {
		t.Fatalf("config inesperada: %+v", cfg)
	}
}

func TestLoad_RN01_DefaultPort8080(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{"DATABASE_URL": "postgres://x"}))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("porta = %d, esperado 8080", cfg.Port)
	}
}

func TestLoad_RN01_RequiresDatabaseURL(t *testing.T) {
	if _, err := Load(envFrom(nil)); err == nil {
		t.Fatal("esperado erro sem DATABASE_URL")
	}
}

func TestLoad_RN01_RejectsInvalidPort(t *testing.T) {
	for _, port := range []string{"abc", "0", "70000"} {
		_, err := Load(envFrom(map[string]string{"DATABASE_URL": "postgres://x", "API_PORT": port}))
		if err == nil {
			t.Errorf("API_PORT=%q: esperado erro", port)
		}
	}
}

func TestParseDotEnv_RN02_EnvExampleFormat(t *testing.T) {
	input := "# comentário\n\nPOSTGRES_USER=ro_lobby\nQUOTED=\"com espaço\"\nSINGLE='x'\nEMPTY=\n"
	vars, err := parseDotEnv(bufio.NewScanner(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	want := map[string]string{"POSTGRES_USER": "ro_lobby", "QUOTED": "com espaço", "SINGLE": "x", "EMPTY": ""}
	if len(vars) != len(want) {
		t.Fatalf("vars = %v, esperado %v", vars, want)
	}
	for k, v := range want {
		if vars[k] != v {
			t.Errorf("%s = %q, esperado %q", k, vars[k], v)
		}
	}
}

// RN-02: o .env segue o formato KEY=VALUE do .env.example.
func TestParseDotEnv_RN02_RejectsLineWithoutEquals(t *testing.T) {
	if _, err := parseDotEnv(bufio.NewScanner(strings.NewReader("SEM_IGUAL\n"))); err == nil {
		t.Fatal("esperado erro")
	}
}

// RN-01: variáveis já definidas no ambiente têm precedência sobre o .env.
func TestLoadDotEnv_RN01_DoesNotOverrideEnvironment(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("RO_LOBBY_TEST_A=do_arquivo\nRO_LOBBY_TEST_B=do_arquivo\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RO_LOBBY_TEST_A", "do_ambiente")
	// t.Setenv registra a restauração; o Unsetenv deixa B de fato indefinida.
	t.Setenv("RO_LOBBY_TEST_B", "")
	if err := os.Unsetenv("RO_LOBBY_TEST_B"); err != nil {
		t.Fatal(err)
	}

	if err := LoadDotEnv(path); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got := os.Getenv("RO_LOBBY_TEST_A"); got != "do_ambiente" {
		t.Errorf("A = %q, esperado do_ambiente", got)
	}
	if got := os.Getenv("RO_LOBBY_TEST_B"); got != "do_arquivo" {
		t.Errorf("B = %q, esperado do_arquivo", got)
	}
}

// RN-01: sem .env, a configuração vem só do ambiente (caso da CI).
func TestLoadDotEnv_RN01_MissingFileIsNotAnError(t *testing.T) {
	if err := LoadDotEnv(filepath.Join(t.TempDir(), "nao-existe")); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
}
