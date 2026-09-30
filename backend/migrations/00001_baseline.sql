-- Linha de base (D-04): fixa o fuso do banco em UTC, sem tabela de domínio.
-- As tabelas entram nas specs de cada feature.

-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    EXECUTE format('ALTER DATABASE %I SET timezone TO %L', current_database(), 'UTC');
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    EXECUTE format('ALTER DATABASE %I RESET timezone', current_database());
END
$$;
-- +goose StatementEnd
