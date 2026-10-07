-- Candidaturas a lobby e o histórico delas (spec candidatura-lobby, design D-01 e D-03).
-- Os CHECK repetem as regras do serviço como última defesa. Os estados da Parte 2
-- (left, removed, cancelled) já cabem aqui, mas só a Parte 1 os usa: pending, accepted,
-- rejected, withdrawn e expired.

-- +goose Up
CREATE TABLE applications (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    lobby_id     uuid        NOT NULL REFERENCES lobbies (id) ON DELETE CASCADE,
    user_id      uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- RN-26: o personagem não é excluído com candidatura ativa em lobby aberto; depois,
    -- a candidatura fica sem ele.
    character_id uuid        REFERENCES characters (id) ON DELETE SET NULL,
    role         text        NOT NULL CHECK (role IN ('tank', 'support', 'dps')),     -- RN-04
    message      text        CHECK (char_length(message) <= 250),                    -- RN-06
    status       text        NOT NULL CHECK (status IN (
                                 'pending', 'accepted', 'rejected', 'withdrawn', 'expired',
                                 'left', 'removed', 'cancelled')),                     -- RN-17
    -- RN-09: justificativa de 10 a 250 caracteres, quando existe.
    reason       text        CHECK (char_length(btrim(reason)) BETWEEN 10 AND 250),
    created_at   timestamptz NOT NULL,
    decided_at   timestamptz
);

-- RN-02: no máximo uma candidatura ativa (pendente ou aceita) por Usuário e lobby.
CREATE UNIQUE INDEX applications_one_active_per_user
    ON applications (lobby_id, user_id) WHERE status IN ('pending', 'accepted');
CREATE INDEX applications_lobby_id_idx ON applications (lobby_id);
CREATE INDEX applications_user_id_idx ON applications (user_id);
CREATE INDEX applications_character_active_idx
    ON applications (character_id) WHERE status IN ('pending', 'accepted');

-- RN-18: cada transição com autor (nulo quando é o sistema), data e justificativa.
CREATE TABLE application_events (
    id             bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    application_id uuid        NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    from_status    text,
    to_status      text        NOT NULL,
    actor_id       uuid        REFERENCES users (id) ON DELETE SET NULL,
    reason         text,
    at             timestamptz NOT NULL
);
CREATE INDEX application_events_application_id_idx ON application_events (application_id);

-- +goose Down
DROP TABLE application_events;
DROP TABLE applications;
