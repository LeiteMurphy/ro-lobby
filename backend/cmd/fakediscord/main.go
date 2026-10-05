// Comando fakediscord sobe o Discord falso para o teste ponta a ponta (spec
// login-discord, D-05). Não entra na imagem da API e nunca deve rodar em produção.
//
//	go run ./cmd/fakediscord -addr :8090 -client-id 123 -client-secret segredo
package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/discordfake"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8090", "endereço onde o Discord falso escuta")
	clientID := flag.String("client-id", "000000000000000000", "Client ID aceito")
	clientSecret := flag.String("client-secret", "segredo-do-discord-falso", "Client Secret aceito")
	userID := flag.String("user-id", "100000000000000001", "ID do usuário que autoriza")
	username := flag.String("username", "grimbold", "nome de usuário do Discord")
	globalName := flag.String("global-name", "Grimbold", "nome de exibição (vazio = sem nome de exibição)")
	flag.Parse()

	fake := discordfake.New(*clientID, *clientSecret, discordfake.User{
		ID: *userID, Username: *username, GlobalName: *globalName,
	})
	srv := &http.Server{Addr: *addr, Handler: fake.Handler(), ReadHeaderTimeout: 5 * time.Second}
	slog.Info("Discord falso ouvindo", "addr", *addr)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("Discord falso parou", "err", err)
		os.Exit(1)
	}
}
