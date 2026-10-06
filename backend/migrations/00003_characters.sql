-- Personagens do Usuário (spec personagens, design D-02, D-03, D-06).
-- Os CHECK repetem as regras do serviço como última defesa; quem dá a mensagem ao
-- usuário é o serviço. A classe e o retrato são validados no catálogo em Go (D-03, D-04),
-- por isso não têm FK: um personagem salvo sobrevive a uma classe removida.

-- +goose Up
CREATE TABLE characters (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Desempate da ordem de cadastro quando dois personagens têm o mesmo created_at.
    seq        bigint      NOT NULL GENERATED ALWAYS AS IDENTITY UNIQUE,
    user_id    uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- RN-04: 1 a 24 caracteres, sem espaço nas pontas e sem caracteres de controle.
    nick       text        NOT NULL CHECK (
                               char_length(nick) BETWEEN 1 AND 24
                               AND nick = btrim(nick)
                               AND nick !~ '[[:cntrl:]]'
                           ),
    class_id   text        NOT NULL CHECK (class_id <> ''),
    level      smallint    NOT NULL CHECK (level BETWEEN 1 AND 275),        -- RN-07
    role       text        NOT NULL CHECK (role IN ('tank', 'support', 'dps')), -- RN-08
    portrait   text        NOT NULL CHECK (portrait <> ''),
    -- RN-09: opcional; quando existe, https:// e até 300 caracteres.
    link       text        CHECK (char_length(link) <= 300 AND link LIKE 'https://%'),
    is_main    boolean     NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL
);

-- RN-05 / D-02: nick único em todo o RO Lobby, sem diferenciar maiúsculas. A collation
-- pg_c_utf8 (provedor builtin) faz o lower() de letras acentuadas igual em qualquer
-- sistema operacional, sem depender do locale do banco.
CREATE UNIQUE INDEX characters_nick_key ON characters (lower(nick COLLATE pg_c_utf8));

-- RN-12: no máximo um principal por Usuário.
CREATE UNIQUE INDEX characters_one_main_per_user ON characters (user_id) WHERE is_main;

-- RN-17: listagem do Usuário por ordem de cadastro.
CREATE INDEX characters_user_id_created_idx ON characters (user_id, created_at, seq);

-- +goose Down
DROP TABLE characters;
