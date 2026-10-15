-- Sala do grupo no Discord (spec discord-bot, design seção 4, D-02 e D-04). O usuário guarda
-- só a data da conexão ao servidor oficial, nunca o token (RN-02). A sala tem os IDs dos
-- canais no Discord e o estado (RN-09). Os efeitos assíncronos ficam na fila
-- discord_jobs (D-02); lembrete e remoção são deduzidos do horário e não viram job.

-- +goose Up
ALTER TABLE users ADD COLUMN discord_connected_at timestamptz; -- RN-02, RN-04

-- RN-23: link opcional do Discord do dono; o formato fica no serviço, aqui só a defesa.
ALTER TABLE lobbies ADD COLUMN discord_link text
    CHECK (char_length(discord_link) <= 200 AND discord_link LIKE 'https://%');

-- RN-05: no máximo uma sala por lobby. D-04: uma categoria por sala, com texto e voz.
CREATE TABLE discord_rooms (
    lobby_id         uuid        PRIMARY KEY REFERENCES lobbies (id) ON DELETE CASCADE,
    category_id      text        NOT NULL CHECK (category_id ~ '^[0-9]+$'),
    text_channel_id  text        NOT NULL CHECK (text_channel_id ~ '^[0-9]+$'),
    voice_channel_id text        NOT NULL CHECK (voice_channel_id ~ '^[0-9]+$'),
    status           text        NOT NULL CHECK (status IN ('open', 'closing', 'deleted')), -- RN-09
    opened_at        timestamptz NOT NULL,
    closing_at       timestamptz,          -- RN-17: quando o lobby foi cancelado
    deleted_at       timestamptz,          -- RN-16, RN-17
    reminder_sent_at timestamptz,          -- RN-14; volta a nulo quando o início muda (RN-15)
    last_alert_at    timestamptz,          -- RN-27
    CHECK (status <> 'open' OR (closing_at IS NULL AND deleted_at IS NULL)),
    CHECK (status <> 'closing' OR (closing_at IS NOT NULL AND deleted_at IS NULL)),
    CHECK (status <> 'deleted' OR deleted_at IS NOT NULL)
);

-- D-02: a fila de efeitos. sync_access confere na hora de rodar se o Usuário é membro e
-- está conectado, por isso só guarda quem; post leva o texto pronto e os IDs a mencionar.
CREATE TABLE discord_jobs (
    id         bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    lobby_id   uuid        NOT NULL REFERENCES discord_rooms (lobby_id) ON DELETE CASCADE,
    kind       text        NOT NULL CHECK (kind IN ('sync_access', 'post')),
    user_id    uuid        REFERENCES users (id) ON DELETE CASCADE,
    payload    jsonb       NOT NULL DEFAULT '{}',
    run_at     timestamptz NOT NULL,
    attempts   integer     NOT NULL DEFAULT 0,
    last_error text,
    done_at    timestamptz,
    created_at timestamptz NOT NULL,
    CHECK ((kind = 'sync_access') = (user_id IS NOT NULL))
);

-- RN-21 / RNF-04: o trabalhador pega os jobs vencidos e ainda não feitos.
CREATE INDEX discord_jobs_pending_idx ON discord_jobs (run_at) WHERE done_at IS NULL;

-- +goose Down
DROP TABLE discord_jobs;
DROP TABLE discord_rooms;
ALTER TABLE lobbies DROP COLUMN discord_link;
ALTER TABLE users DROP COLUMN discord_connected_at;
