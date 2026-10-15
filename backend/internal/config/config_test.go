package config

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func envFrom(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

func withDiscord(vars map[string]string) map[string]string {
	vars["DISCORD_CLIENT_ID"] = "123"
	vars["DISCORD_CLIENT_SECRET"] = "segredo"
	return vars
}

// RN-03 / RN-17 (login-discord): o aplicativo do Discord vem do ambiente, com a API
// oficial como padrão.
func TestLoad_RN17_DiscordDefaults(t *testing.T) {
	cfg, err := Load(envFrom(withDiscord(map[string]string{"DATABASE_URL": "postgres://x"})))
	if err != nil {
		t.Fatal(err)
	}
	want := Discord{ClientID: "123", ClientSecret: "segredo", APIBaseURL: "https://discord.com/api"}
	if cfg.Discord != want {
		t.Fatalf("Discord = %+v, esperado %+v", cfg.Discord, want)
	}
	cfg, err = Load(envFrom(withDiscord(map[string]string{"DATABASE_URL": "postgres://x", "DISCORD_API_BASE_URL": "http://localhost:8090/api"})))
	if err != nil || cfg.Discord.APIBaseURL != "http://localhost:8090/api" {
		t.Fatalf("APIBaseURL = %q (%v)", cfg.Discord.APIBaseURL, err)
	}
}

// RN-03 (login-discord): sem Client ID ou Client Secret, a API não sobe.
func TestLoad_RN03_RequiresDiscordCredentials(t *testing.T) {
	for _, missing := range []string{"DISCORD_CLIENT_ID", "DISCORD_CLIENT_SECRET"} {
		vars := withDiscord(map[string]string{"DATABASE_URL": "postgres://x"})
		delete(vars, missing)
		if _, err := Load(envFrom(vars)); err == nil {
			t.Errorf("sem %s: esperado erro", missing)
		}
	}
}

// RN-01: toda configuração vem de variáveis de ambiente.
func TestLoad_RN01_ReadsEnvironment(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{
		"DATABASE_URL":          "postgres://u:p@localhost:5432/db",
		"API_PORT":              "9090",
		"DISCORD_CLIENT_ID":     "123",
		"DISCORD_CLIENT_SECRET": "segredo",
	}))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cfg.DatabaseURL != "postgres://u:p@localhost:5432/db" || cfg.Port != 9090 {
		t.Fatalf("config inesperada: %+v", cfg)
	}
}

func TestLoad_RN01_DefaultPort8080(t *testing.T) {
	cfg, err := Load(envFrom(withDiscord(map[string]string{"DATABASE_URL": "postgres://x"})))
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
		_, err := Load(envFrom(withDiscord(map[string]string{"DATABASE_URL": "postgres://x", "API_PORT": port})))
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

// RN-22 (discord-bot): sem token e sem servidor, a sala no Discord fica desligada; com os
// dois, ligada; só um dos dois é erro de configuração.
func TestLoad_RN22_BotOnlyWithTokenAndGuild(t *testing.T) {
	base := func() map[string]string { return withDiscord(map[string]string{"DATABASE_URL": "postgres://x"}) }
	cfg, err := Load(envFrom(base()))
	if err != nil || cfg.Bot.Enabled() {
		t.Fatalf("sem as duas: %+v, %v", cfg.Bot, err)
	}
	both := base()
	both["DISCORD_BOT_TOKEN"], both["DISCORD_GUILD_ID"] = "token-do-bot", "1560117933705797643"
	cfg, err = Load(envFrom(both))
	if err != nil || !cfg.Bot.Enabled() || cfg.Bot.GuildID != "1560117933705797643" {
		t.Fatalf("com as duas: %+v, %v", cfg.Bot, err)
	}
	for _, only := range []string{"DISCORD_BOT_TOKEN", "DISCORD_GUILD_ID"} {
		vars := base()
		vars[only] = "123"
		if _, err := Load(envFrom(vars)); err == nil {
			t.Errorf("só %s: esperado erro", only)
		}
	}
	bad := base()
	bad["DISCORD_BOT_TOKEN"], bad["DISCORD_GUILD_ID"] = "token-do-bot", "servidor"
	if _, err := Load(envFrom(bad)); err == nil {
		t.Error("DISCORD_GUILD_ID fora do formato: esperado erro")
	}
}

// RNF-01 (discord-bot): o token do bot não aparece ao imprimir nem ao logar a config.
func TestBot_RNF01_TokenHidden(t *testing.T) {
	b := Bot{Token: "token-super-secreto", GuildID: "1560117933705797643"}
	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("config", "bot", b)
	for name, out := range map[string]string{
		"String": b.String(),
		"%v":     fmt.Sprintf("%v", b),
		"%+v":    fmt.Sprintf("%+v", b),
		"%#v":    fmt.Sprintf("%#v", b),
		"Config": fmt.Sprintf("%+v", Config{Bot: b}),
		"slog":   buf.String(),
	} {
		if bytes.Contains([]byte(out), []byte("token-super-secreto")) {
			t.Errorf("%s mostra o token: %s", name, out)
		}
	}
}

// RN-15 (endurecimento): o limite de requisições fica ligado por padrão; só
// API_RATE_LIMIT=off desliga, e nunca com origem pública.
func TestLoad_RN15_RateLimitSwitch(t *testing.T) {
	cfg, err := Load(envFrom(withDiscord(map[string]string{"DATABASE_URL": "postgres://x"})))
	if err != nil || !cfg.RateLimit {
		t.Fatalf("RateLimit = %v (%v), esperado ligado por padrão", cfg.RateLimit, err)
	}
	cfg, err = Load(envFrom(withDiscord(map[string]string{"DATABASE_URL": "postgres://x", "API_RATE_LIMIT": "off"})))
	if err != nil || cfg.RateLimit {
		t.Fatalf("RateLimit = %v (%v), esperado desligado com off", cfg.RateLimit, err)
	}
	_, err = Load(envFrom(withDiscord(map[string]string{
		"DATABASE_URL": "postgres://x", "API_RATE_LIMIT": "off", "APP_ORIGIN": "https://rolobby.com.br",
	})))
	if err == nil {
		t.Fatal("aceitou API_RATE_LIMIT=off com origem pública")
	}
}
