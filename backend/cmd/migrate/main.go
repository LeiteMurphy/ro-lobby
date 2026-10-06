// Comando migrate aplica ou reverte as migrações do banco (RN-09).
//
//	go run ./cmd/migrate up      aplica todas as pendentes
//	go run ./cmd/migrate down    reverte a última
//	go run ./cmd/migrate reset   reverte todas
//	go run ./cmd/migrate status  mostra o estado de cada migração
//	go run ./cmd/migrate fresh   apaga, recria e migra o banco (só bancos *_e2e)
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"

	"github.com/LeiteMurphy/ro-lobby/backend/internal/config"
	"github.com/LeiteMurphy/ro-lobby/backend/internal/migrate"
)

const usage = "uso: go run ./cmd/migrate up|down|reset|status|fresh"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 1 {
		return errors.New(usage)
	}

	for _, path := range []string{".env", "../.env"} {
		if err := config.LoadDotEnv(path); err != nil {
			return err
		}
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL não definida")
	}

	ctx := context.Background()
	if args[0] == "fresh" {
		if err := migrate.Fresh(ctx, databaseURL); err != nil {
			return err
		}
		fmt.Println("banco recriado e migrado")
		return nil
	}

	db, err := migrate.OpenDB(databaseURL)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	provider, err := migrate.NewProvider(db)
	if err != nil {
		return err
	}

	switch args[0] {
	case "up":
		return report(provider.Up(ctx))
	case "down":
		result, err := provider.Down(ctx)
		if result != nil {
			fmt.Println(result)
		}
		return err
	case "reset":
		return report(provider.DownTo(ctx, 0))
	case "status":
		statuses, err := provider.Status(ctx)
		if err != nil {
			return err
		}
		for _, s := range statuses {
			fmt.Printf("%-8s %s\n", s.State, s.Source.Path)
		}
		return nil
	default:
		return errors.New(usage)
	}
}

func report(results []*goose.MigrationResult, err error) error {
	for _, r := range results {
		fmt.Println(r)
	}
	if err == nil && len(results) == 0 {
		fmt.Println("nada a fazer")
	}
	return err
}
