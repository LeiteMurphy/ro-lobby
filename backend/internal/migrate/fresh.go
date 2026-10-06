package migrate

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"
)

// FreshSuffix é o fim obrigatório do nome do banco que Fresh aceita apagar. É a trava que
// impede apagar o banco de desenvolvimento ou o da pilha app por engano.
const FreshSuffix = "_e2e"

// ErrNotFreshable: o banco não termina em FreshSuffix.
var ErrNotFreshable = errors.New("migrate: fresh só apaga bancos terminados em " + FreshSuffix)

// Fresh apaga e recria o banco da URL e aplica todas as migrações. Serve ao ponta a ponta,
// que começa sempre de um banco vazio (spec lobbies, T-10).
func Fresh(ctx context.Context, databaseURL string) error {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("migrate: DATABASE_URL inválida: %w", err)
	}
	name := strings.TrimPrefix(u.Path, "/")
	if !strings.HasSuffix(name, FreshSuffix) || name == FreshSuffix {
		return ErrNotFreshable
	}

	admin := *u
	admin.Path = "/postgres"
	conn, err := pgx.Connect(ctx, admin.String())
	if err != nil {
		return fmt.Errorf("migrate: conectar no servidor: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()
	ident := pgx.Identifier{name}.Sanitize()
	if _, err := conn.Exec(ctx, "DROP DATABASE IF EXISTS "+ident+" WITH (FORCE)"); err != nil {
		return fmt.Errorf("migrate: apagar %s: %w", name, err)
	}
	if _, err := conn.Exec(ctx, "CREATE DATABASE "+ident); err != nil {
		return fmt.Errorf("migrate: criar %s: %w", name, err)
	}

	db, err := OpenDB(databaseURL)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	provider, err := NewProvider(db)
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	return err
}
