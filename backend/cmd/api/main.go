// Comando api sobe o servidor HTTP do RO Lobby.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/config"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/database"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api encerrada com erro", "err", err)
		os.Exit(1)
	}
}

func run() error {
	// O .env fica na raiz do repositório; o backend roda a partir de backend/.
	for _, path := range []string{".env", "../.env"} {
		if err := config.LoadDotEnv(path); err != nil {
			return err
		}
	}

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           server.New(pool),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("api ouvindo", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
