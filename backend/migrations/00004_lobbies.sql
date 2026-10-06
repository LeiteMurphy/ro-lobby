-- Lobbies (spec lobbies, design D-01 a D-04). O personagem do dono fica no próprio lobby;
-- a tabela de membros entra com a candidatura (D-01). Os CHECK repetem as regras do
-- serviço como última defesa. O estado não é coluna: cancelado quando cancelled_at existe,
-- iniciado quando starts_at já passou, aberto nos outros casos (RN-13).

-- +goose Up
CREATE TABLE lobbies (
    id                 uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id           uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- D-02: cópia do catálogo, para o lobby não perder o nome se a instância sair dele.
    instance_id        text        NOT NULL CHECK (instance_id <> ''),
    instance_name      text        NOT NULL CHECK (instance_name <> ''),
    instance_level     smallint    NOT NULL CHECK (instance_level BETWEEN 1 AND 275),
    starts_at          timestamptz NOT NULL,
    -- RN-06: de 0 a 12 por função, de 1 a 12 no total.
    slots_tank         smallint    NOT NULL CHECK (slots_tank BETWEEN 0 AND 12),
    slots_support      smallint    NOT NULL CHECK (slots_support BETWEEN 0 AND 12),
    slots_dps          smallint    NOT NULL CHECK (slots_dps BETWEEN 0 AND 12),
    min_level          smallint    NOT NULL CHECK (min_level BETWEEN 1 AND 275),     -- RN-07
    -- RN-08 / P-02: o personagem do dono ocupa uma vaga da função dele. Fica nulo se o
    -- personagem for excluído depois do início (RN-21 impede antes).
    owner_character_id uuid        REFERENCES characters (id) ON DELETE SET NULL,
    owner_role         text        NOT NULL CHECK (owner_role IN ('tank', 'support', 'dps')),
    note               text        CHECK (char_length(note) <= 250),                -- RN-09
    -- RN-19: cancelado é final e sempre tem motivo de 10 a 250 caracteres.
    cancelled_at       timestamptz,
    cancel_reason      text        CHECK (char_length(btrim(cancel_reason)) BETWEEN 10 AND 250),
    created_at         timestamptz NOT NULL,
    CHECK (slots_tank + slots_support + slots_dps BETWEEN 1 AND 12),
    CHECK ((cancelled_at IS NULL) = (cancel_reason IS NULL)),
    CHECK (min_level >= instance_level)
);

CREATE INDEX lobbies_starts_at_idx ON lobbies (starts_at) WHERE cancelled_at IS NULL;
CREATE INDEX lobbies_owner_id_idx ON lobbies (owner_id);
CREATE INDEX lobbies_owner_character_id_idx ON lobbies (owner_character_id);

-- +goose Down
DROP TABLE lobbies;
