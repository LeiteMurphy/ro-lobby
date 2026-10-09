-- Disponibilidade do personagem no banco de talentos (spec banco-de-talentos, design D-01
-- a D-03). Os CHECK repetem as regras do serviço como última defesa.

-- +goose Up
CREATE TABLE character_availability (
    -- D-02: 1:1 com o personagem; excluído o personagem, ele sai do banco (RN-05).
    character_id uuid        PRIMARY KEY REFERENCES characters (id) ON DELETE CASCADE,
    -- RN-01: desligado, sai do banco e guarda os dados.
    enabled      boolean     NOT NULL,
    -- RN-02 / D-03: bit d ligado = dia d da semana (domingo = 0); pelo menos um dia.
    days         smallint    NOT NULL CHECK (days BETWEEN 1 AND 127),
    -- RN-03 / D-01: minutos desde a meia-noite em Brasília, de 30 em 30. Fim menor que o
    -- início vira a meia-noite.
    start_minute smallint    NOT NULL CHECK (start_minute BETWEEN 0 AND 1410 AND start_minute % 30 = 0),
    end_minute   smallint    NOT NULL CHECK (end_minute BETWEEN 0 AND 1410 AND end_minute % 30 = 0),
    -- RN-04: "Qualquer instância" ou uma ou mais do catálogo (validadas em Go).
    any_instance boolean     NOT NULL,
    instance_ids text[]      NOT NULL,
    updated_at   timestamptz NOT NULL,
    CHECK (end_minute <> start_minute),
    CHECK (any_instance OR cardinality(instance_ids) > 0)
);

CREATE INDEX character_availability_enabled_idx
    ON character_availability (character_id) WHERE enabled;

-- +goose Down
DROP TABLE character_availability;
