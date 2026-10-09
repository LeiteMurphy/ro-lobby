-- Formação do lobby: por função ou grupo livre (spec grupo-livre, design D-01). Os lobbies
-- que já existem ficam "por função" (RNF-01).

-- +goose Up
ALTER TABLE lobbies
    -- RN-01: roles = vagas por função (a de antes); free = grupo livre.
    ADD COLUMN formation  text     NOT NULL DEFAULT 'roles' CHECK (formation IN ('roles', 'free')),
    -- RN-02: no grupo livre, de 2 a 12 vagas no total.
    ADD COLUMN free_slots smallint CHECK (free_slots BETWEEN 2 AND 12);

-- O CHECK da soma por função (RN-06 da lobbies) passa a valer só para "por função"; no
-- grupo livre as vagas por função ficam zeradas e o total vem de free_slots.
ALTER TABLE lobbies DROP CONSTRAINT lobbies_check;
ALTER TABLE lobbies ADD CONSTRAINT lobbies_formation_slots_check CHECK (
    (formation = 'roles' AND free_slots IS NULL
        AND slots_tank + slots_support + slots_dps BETWEEN 1 AND 12)
    OR (formation = 'free' AND free_slots IS NOT NULL
        AND slots_tank + slots_support + slots_dps = 0)
);

-- +goose Down
ALTER TABLE lobbies DROP CONSTRAINT lobbies_formation_slots_check;
DELETE FROM lobbies WHERE formation = 'free';
ALTER TABLE lobbies ADD CONSTRAINT lobbies_check
    CHECK (slots_tank + slots_support + slots_dps BETWEEN 1 AND 12);
ALTER TABLE lobbies DROP COLUMN free_slots, DROP COLUMN formation;
