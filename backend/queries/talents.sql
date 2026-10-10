-- Banco de talentos (spec banco-de-talentos, design D-01 a D-05).

-- name: UpsertAvailability :one
-- RN-01 a RN-04: grava a disponibilidade ligada do personagem.
INSERT INTO character_availability
    (character_id, enabled, days, start_minute, end_minute, any_instance, instance_ids, updated_at)
VALUES (@character_id, true, @days, @start_minute, @end_minute, @any_instance, @instance_ids::text[], @now)
ON CONFLICT (character_id) DO UPDATE
SET enabled = true, days = EXCLUDED.days, start_minute = EXCLUDED.start_minute,
    end_minute = EXCLUDED.end_minute, any_instance = EXCLUDED.any_instance,
    instance_ids = EXCLUDED.instance_ids, updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: DisableAvailability :one
-- RN-01: desligar tira do banco e guarda os dados.
UPDATE character_availability SET enabled = false, updated_at = @now
WHERE character_id = @character_id
RETURNING *;

-- name: ListAvailabilityByUser :many
-- A disponibilidade de cada personagem do Usuário, para o perfil (CA-01.4).
SELECT a.* FROM character_availability a
JOIN characters c ON c.id = a.character_id
WHERE c.user_id = @user_id;

-- name: Affinity :many
-- RN-06 a RN-08 / D-04: personagens do banco com afinidade com um lobby (gravado ou na
-- criação). dow e minute são o início do lobby em Brasília (D-01); a janela é o início
-- ± 2 h (D-03 da lobbies).
SELECT c.id, c.nick, c.class_id, c.level, c.role, c.portrait, c.link,
       a.days, a.start_minute, a.end_minute, a.any_instance, a.instance_ids, u.username,
       -- RN-13: a pessoa já foi removida deste lobby, e se com bloqueio.
       EXISTS (SELECT 1 FROM applications ap
               WHERE ap.lobby_id = sqlc.narg(lobby_id)::uuid AND ap.user_id = c.user_id
                 AND ap.status = 'removed')::boolean AS removed,
       EXISTS (SELECT 1 FROM applications ap
               WHERE ap.lobby_id = sqlc.narg(lobby_id)::uuid AND ap.user_id = c.user_id
                 AND ap.status = 'removed' AND ap.blocked)::boolean AS blocked
FROM character_availability a
JOIN characters c ON c.id = a.character_id
JOIN users u ON u.id = c.user_id
WHERE a.enabled
  AND c.user_id <> sqlc.arg(exclude_user_id)
  -- RN-07 da lobby-sem-instancia: lobby sem instância (instance_id vazio) aceita todos.
  AND (sqlc.arg(instance_id)::text = '' OR a.any_instance OR sqlc.arg(instance_id)::text = ANY (a.instance_ids))
  AND ((a.start_minute < a.end_minute
        AND a.days & (1 << sqlc.arg(dow)::int) <> 0
        AND sqlc.arg(minute)::int >= a.start_minute AND sqlc.arg(minute)::int < a.end_minute)
    OR (a.start_minute > a.end_minute
        AND ((a.days & (1 << sqlc.arg(dow)::int) <> 0 AND sqlc.arg(minute)::int >= a.start_minute)
          OR (a.days & (1 << ((sqlc.arg(dow)::int + 6) % 7)) <> 0 AND sqlc.arg(minute)::int < a.end_minute))))
  AND c.level >= sqlc.arg(min_level)::int
  AND c.role = ANY (sqlc.arg(roles)::text[])
  AND NOT EXISTS (
      SELECT 1 FROM lobbies l
      WHERE l.cancelled_at IS NULL
        AND l.starts_at > sqlc.arg(window_start) AND l.starts_at < sqlc.arg(window_end)
        AND (l.owner_character_id = c.id
             OR EXISTS (SELECT 1 FROM applications ap
                        WHERE ap.lobby_id = l.id AND ap.character_id = c.id
                          AND ap.status IN ('pending', 'accepted'))))
ORDER BY c.level DESC, lower(c.nick), c.seq;

-- name: Catalog :many
-- RN-11 / RN-08: o banco com filtros opcionais, combinados com "E". Com dia e hora, a
-- hora cai na faixa daquele dia (RN-03); só com o dia, a faixa toca o dia; só com a hora,
-- a hora cai na faixa em qualquer dia.
SELECT c.id, c.nick, c.class_id, c.level, c.role, c.portrait, c.link,
       a.days, a.start_minute, a.end_minute, a.any_instance, a.instance_ids, u.username
FROM character_availability a
JOIN characters c ON c.id = a.character_id
JOIN users u ON u.id = c.user_id
WHERE a.enabled
  AND (sqlc.narg(instance_id)::text IS NULL OR a.any_instance
       OR sqlc.narg(instance_id)::text = ANY (a.instance_ids))
  AND (sqlc.narg(role)::text IS NULL OR c.role = sqlc.narg(role)::text)
  AND (sqlc.narg(dow)::int IS NULL
       OR a.days & (1 << sqlc.narg(dow)::int) <> 0
       OR (a.start_minute > a.end_minute AND sqlc.narg(minute)::int IS NULL
           AND a.days & (1 << ((sqlc.narg(dow)::int + 6) % 7)) <> 0)
       OR (a.start_minute > a.end_minute AND sqlc.narg(minute)::int < a.end_minute
           AND a.days & (1 << ((sqlc.narg(dow)::int + 6) % 7)) <> 0))
  AND (sqlc.narg(minute)::int IS NULL
       OR (a.start_minute < a.end_minute
           AND sqlc.narg(minute)::int >= a.start_minute AND sqlc.narg(minute)::int < a.end_minute
           AND (sqlc.narg(dow)::int IS NULL OR a.days & (1 << sqlc.narg(dow)::int) <> 0))
       OR (a.start_minute > a.end_minute
           AND ((sqlc.narg(minute)::int >= a.start_minute
                 AND (sqlc.narg(dow)::int IS NULL OR a.days & (1 << sqlc.narg(dow)::int) <> 0))
             OR (sqlc.narg(minute)::int < a.end_minute
                 AND (sqlc.narg(dow)::int IS NULL
                      OR a.days & (1 << ((sqlc.narg(dow)::int + 6) % 7)) <> 0)))))
ORDER BY c.level DESC, lower(c.nick), c.seq
LIMIT @max_rows;
