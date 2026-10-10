-- Lobby sem instância (spec lobby-sem-instancia, D-01). Sem instância, instance_id fica
-- nulo, instance_name guarda o título (ou nulo) e instance_level fica 1, para o CHECK
-- min_level >= instance_level continuar valendo com o nível de 1 a 275 (RN-03). Os
-- lobbies existentes não mudam (RNF-01).

-- +goose Up
ALTER TABLE lobbies
    ALTER COLUMN instance_id DROP NOT NULL,
    ALTER COLUMN instance_name DROP NOT NULL;

-- RN-01 / RN-02: com instância, id e nome; sem instância, nível 1 e título opcional de
-- 1 a 40 caracteres, sem espaço nas pontas.
ALTER TABLE lobbies ADD CONSTRAINT lobbies_any_instance_check CHECK (
    (instance_id IS NOT NULL AND instance_name IS NOT NULL)
    OR (instance_id IS NULL AND instance_level = 1
        AND (instance_name IS NULL
             OR (char_length(instance_name) BETWEEN 1 AND 40 AND instance_name = btrim(instance_name))))
);

-- +goose Down
ALTER TABLE lobbies DROP CONSTRAINT lobbies_any_instance_check;
DELETE FROM lobbies WHERE instance_id IS NULL;
ALTER TABLE lobbies
    ALTER COLUMN instance_id SET NOT NULL,
    ALTER COLUMN instance_name SET NOT NULL;
