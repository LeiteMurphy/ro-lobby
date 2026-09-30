// Package migrations guarda os arquivos SQL do goose, embutidos no binário (RN-09).
// Cada arquivo tem uma seção `-- +goose Up` e uma `-- +goose Down`.
package migrations

import "embed"

// FS contém todos os arquivos .sql desta pasta.
//
//go:embed *.sql
var FS embed.FS
