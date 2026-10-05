// Package config lê a configuração do backend a partir de variáveis de ambiente (RN-01).
package config

import (
	"errors"
	"fmt"
	"strconv"
)

// Config é a configuração do processo da API.
type Config struct {
	// Port é a porta HTTP da API (API_PORT, padrão 8080).
	Port int
	// DatabaseURL é a URL de conexão do PostgreSQL (DATABASE_URL, obrigatória).
	DatabaseURL string
	// Discord é o aplicativo do Discord usado no login (spec login-discord).
	Discord Discord
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
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL não definida")
	}
	if cfg.Discord.ClientID == "" || cfg.Discord.ClientSecret == "" {
		return Config{}, errors.New("DISCORD_CLIENT_ID e DISCORD_CLIENT_SECRET precisam estar definidas")
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
