// Package config lê a configuração do backend a partir de variáveis de ambiente (RN-01).
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
)

// snowflake é o formato de um ID do Discord.
var snowflake = regexp.MustCompile(`^[0-9]+$`)

// Config é a configuração do processo da API.
type Config struct {
	// Port é a porta HTTP da API (API_PORT, padrão 8080).
	Port int
	// DatabaseURL é a URL de conexão do PostgreSQL (DATABASE_URL, obrigatória).
	DatabaseURL string
	// Discord é o aplicativo do Discord usado no login (spec login-discord).
	Discord Discord
	// Bot é o bot da sala do grupo no Discord (spec discord-bot, RN-22).
	Bot Bot
}

// Bot: DISCORD_BOT_TOKEN e DISCORD_GUILD_ID juntas ligam a sala do grupo; nenhuma das duas a
// desliga (RN-22 da discord-bot). O token só existe no ambiente da API e nunca vai para o
// log (RNF-01): String, GoString e LogValue escondem o valor.
type Bot struct {
	Token   string
	GuildID string
}

// Enabled diz se a sala do grupo está ligada (RN-22 da discord-bot).
func (b Bot) Enabled() bool { return b.Token != "" && b.GuildID != "" }

func (b Bot) String() string {
	if b.Token == "" {
		return fmt.Sprintf("Bot{Token: \"\", GuildID: %q}", b.GuildID)
	}
	return fmt.Sprintf("Bot{Token: [oculto], GuildID: %q}", b.GuildID)
}

func (b Bot) GoString() string { return b.String() }

// LogValue esconde o token quando a config vai para o slog.
func (b Bot) LogValue() slog.Value {
	return slog.GroupValue(slog.Bool("enabled", b.Enabled()), slog.String("guild_id", b.GuildID))
}

// Discord: DISCORD_CLIENT_ID e DISCORD_CLIENT_SECRET são obrigatórias; o secret só existe
// no ambiente da API (RN-03). DISCORD_API_BASE_URL aponta para o Discord falso nos testes
// (RN-17).
type Discord struct {
	ClientID     string
	ClientSecret string
	APIBaseURL   string
}

const (
	defaultPort              = 8080
	defaultDiscordAPIBaseURL = "https://discord.com/api"
)

// Load monta a Config usando getenv, normalmente os.Getenv.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Port:        defaultPort,
		DatabaseURL: getenv("DATABASE_URL"),
		Discord: Discord{
			ClientID:     getenv("DISCORD_CLIENT_ID"),
			ClientSecret: getenv("DISCORD_CLIENT_SECRET"),
			APIBaseURL:   getenv("DISCORD_API_BASE_URL"),
		},
		Bot: Bot{Token: getenv("DISCORD_BOT_TOKEN"), GuildID: getenv("DISCORD_GUILD_ID")},
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL não definida")
	}
	if cfg.Discord.ClientID == "" || cfg.Discord.ClientSecret == "" {
		return Config{}, errors.New("DISCORD_CLIENT_ID e DISCORD_CLIENT_SECRET precisam estar definidas")
	}
	if (cfg.Bot.Token == "") != (cfg.Bot.GuildID == "") {
		return Config{}, errors.New("DISCORD_BOT_TOKEN e DISCORD_GUILD_ID precisam estar definidas juntas (ou nenhuma, para desligar a sala no Discord)")
	}
	if cfg.Bot.GuildID != "" && !snowflake.MatchString(cfg.Bot.GuildID) {
		return Config{}, fmt.Errorf("DISCORD_GUILD_ID inválida: %q", cfg.Bot.GuildID)
	}
	if cfg.Discord.APIBaseURL == "" {
		cfg.Discord.APIBaseURL = defaultDiscordAPIBaseURL
	}

	if raw := getenv("API_PORT"); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil || port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("API_PORT inválida: %q", raw)
		}
		cfg.Port = port
	}

	return cfg, nil
}
