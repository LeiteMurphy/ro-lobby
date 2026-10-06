-- name: CreateLobby :one
INSERT INTO lobbies (owner_id, instance_id, instance_name, instance_level, starts_at,
    slots_tank, slots_support, slots_dps, min_level, owner_character_id, owner_role, note, created_at)
VALUES (@owner_id, @instance_id, @instance_name, @instance_level, @starts_at,
    @slots_tank, @slots_support, @slots_dps, @min_level, @owner_character_id, @owner_role, @note, @now)
RETURNING id;

-- name: GetLobby :one
-- RN-15: o lobby em qualquer estado, com o personagem e o Discord do dono.
SELECT sqlc.embed(lobbies),
       c.nick AS owner_nick, c.class_id AS owner_class_id, c.level AS owner_level,
       u.username AS owner_username, u.global_name AS owner_global_name
FROM lobbies
JOIN users u ON u.id = lobbies.owner_id
LEFT JOIN characters c ON c.id = lobbies.owner_character_id
WHERE lobbies.id = @id;

-- name: ListOpenLobbies :many
-- RN-13 / RN-14: só abertos (não cancelados e ainda não iniciados), por início.
SELECT sqlc.embed(lobbies),
       c.nick AS owner_nick, c.class_id AS owner_class_id, c.level AS owner_level,
       u.username AS owner_username, u.global_name AS owner_global_name
FROM lobbies
JOIN users u ON u.id = lobbies.owner_id
LEFT JOIN characters c ON c.id = lobbies.owner_character_id
WHERE lobbies.cancelled_at IS NULL
  AND lobbies.starts_at > @now
  AND lobbies.starts_at >= @from_at
  AND lobbies.starts_at < @to_at
ORDER BY lobbies.starts_at, lobbies.created_at, lobbies.id;

-- name: CountOpenLobbiesByOwner :one
-- RN-11: lobbies abertos do Usuário como dono.
SELECT count(*) FROM lobbies
WHERE owner_id = @owner_id AND cancelled_at IS NULL AND starts_at > @now;

-- name: HasScheduleConflict :one
-- RN-10 / D-03: outro lobby não cancelado do mesmo personagem com início dentro da
-- janela (início - 2 h, início + 2 h), calculada pelo serviço.
SELECT EXISTS (
    SELECT 1 FROM lobbies
    WHERE owner_character_id = @character_id
      AND cancelled_at IS NULL
      AND id IS DISTINCT FROM @exclude_id
      AND starts_at > @window_start
      AND starts_at < @window_end
);

-- name: GetOwnLobbyForUpdate :one
-- RN-20 / D-08: só o lobby do próprio Usuário, travado para editar ou cancelar.
SELECT * FROM lobbies WHERE id = @id AND owner_id = @owner_id FOR UPDATE;

-- name: UpdateLobby :exec
UPDATE lobbies
SET starts_at = @starts_at, slots_tank = @slots_tank, slots_support = @slots_support,
    slots_dps = @slots_dps, min_level = @min_level, note = @note
WHERE id = @id;

-- name: CancelLobby :exec
UPDATE lobbies SET cancelled_at = @now::timestamptz, cancel_reason = @reason::text WHERE id = @id;

-- name: CharacterOwnsOpenLobby :one
-- RN-21: o personagem é dono de um lobby aberto.
SELECT EXISTS (
    SELECT 1 FROM lobbies
    WHERE owner_character_id = @character_id AND cancelled_at IS NULL AND starts_at > @now
);
