-- Usuário e Sessão do login com Discord (spec login-discord, D-01).
-- RN-06: o Usuário guarda só o ID do Discord, os nomes e as datas; nada de avatar nem
-- tokens do Discord. RN-07: a Sessão guarda só o hash SHA-256 do token.

-- +goose Up
CREATE TABLE users (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    discord_id    text        NOT NULL UNIQUE CHECK (discord_id ~ '^[0-9]+$'),
    username      text        NOT NULL CHECK (username <> ''),
    global_name   text,
    created_at    timestamptz NOT NULL,
    last_login_at timestamptz NOT NULL
);

CREATE TABLE sessions (
    token_hash   bytea       PRIMARY KEY CHECK (octet_length(token_hash) = 32),
    user_id      uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at   timestamptz NOT NULL,
    last_used_at timestamptz NOT NULL
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);

-- +goose Down
DROP TABLE sessions;
DROP TABLE users;
