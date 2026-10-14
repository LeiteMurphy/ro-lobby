-- Servidor do jogo (spec servidor, RN-01 a RN-07, RN-15, RNF-01, RNF-02). O personagem e o
-- lobby guardam o servidor; o lobby guarda o próprio, porque o personagem do dono pode
-- sumir (RN-07). Os dados que já existem ficam em Freya, e os personagens ficam a
-- conferir (RN-15); os criados depois nascem conferidos.

-- +goose Up
ALTER TABLE characters
    ADD COLUMN server text NOT NULL DEFAULT 'freya' CHECK (server IN ('freya', 'nidhogg')),
    ADD COLUMN server_to_confirm boolean NOT NULL DEFAULT true;
ALTER TABLE characters
    ALTER COLUMN server DROP DEFAULT,
    ALTER COLUMN server_to_confirm SET DEFAULT false;

ALTER TABLE lobbies
    ADD COLUMN server text NOT NULL DEFAULT 'freya' CHECK (server IN ('freya', 'nidhogg'));
ALTER TABLE lobbies ALTER COLUMN server DROP DEFAULT;

-- RN-03 / RNF-01: o nick passa a ser único por servidor, sem diferenciar maiúsculas. Se o
-- índice antigo deixou passar nicks iguais, a criação falha e a migração inteira volta.
DROP INDEX characters_nick_key;
CREATE UNIQUE INDEX characters_server_nick_key ON characters (server, lower(nick COLLATE pg_c_utf8));

-- +goose Down
-- A volta só passa se não houver o mesmo nick em servidores diferentes.
DROP INDEX characters_server_nick_key;
CREATE UNIQUE INDEX characters_nick_key ON characters (lower(nick COLLATE pg_c_utf8));
ALTER TABLE lobbies DROP COLUMN server;
ALTER TABLE characters DROP COLUMN server_to_confirm, DROP COLUMN server;
