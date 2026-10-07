-- name: CreateLobby :one
INSERT INTO lobbies (owner_id, instance_id, instance_name, instance_level, starts_at,
    slots_tank, slots_support, slots_dps, min_level, owner_character_id, owner_role, note, created_at)
VALUES (@owner_id, @instance_id, @instance_name, @instance_level, @starts_at,
    @slots_tank, @slots_support, @slots_dps, @min_level, @owner_character_id, @owner_role, @note, @now)
RETURNING id;

-- name: GetLobby :one
-- RN-15: o lobby em qualquer estado, com o personagem e o Discord do dono.
SELECT sqlc.embed(lobbies),
       c.nick AS owner_nick, c.class_id AS owner_class_id, c.level AS owner_level, c.portrait AS owner_portrait, c.link AS owner_link,
       u.username AS owner_username, u.global_name AS owner_global_name,
       -- D-03 (candidatura): membros aceitos por função e pendentes.
       (SELECT count(*) FROM applications a WHERE a.lobby_id = lobbies.id AND a.status = 'accepted' AND a.role = 'tank') AS accepted_tank,
       (SELECT count(*) FROM applications a WHERE a.lobby_id = lobbies.id AND a.status = 'accepted' AND a.role = 'support') AS accepted_support,
       (SELECT count(*) FROM applications a WHERE a.lobby_id = lobbies.id AND a.status = 'accepted' AND a.role = 'dps') AS accepted_dps,
       (SELECT count(*) FROM applications a WHERE a.lobby_id = lobbies.id AND a.status = 'pending') AS pending_count
FROM lobbies
JOIN users u ON u.id = lobbies.owner_id
LEFT JOIN characters c ON c.id = lobbies.owner_character_id
WHERE lobbies.id = @id;

-- name: ListOpenLobbies :many
-- RN-13 / RN-14: só abertos (não cancelados e ainda não iniciados), por início.
SELECT sqlc.embed(lobbies),
       c.nick AS owner_nick, c.class_id AS owner_class_id, c.level AS owner_level, c.portrait AS owner_portrait, c.link AS owner_link,
       u.username AS owner_username, u.global_name AS owner_global_name,
       -- D-03 (candidatura): membros aceitos por função e pendentes.
       (SELECT count(*) FROM applications a WHERE a.lobby_id = lobbies.id AND a.status = 'accepted' AND a.role = 'tank') AS accepted_tank,
       (SELECT count(*) FROM applications a WHERE a.lobby_id = lobbies.id AND a.status = 'accepted' AND a.role = 'support') AS accepted_support,
       (SELECT count(*) FROM applications a WHERE a.lobby_id = lobbies.id AND a.status = 'accepted' AND a.role = 'dps') AS accepted_dps,
       (SELECT count(*) FROM applications a WHERE a.lobby_id = lobbies.id AND a.status = 'pending') AS pending_count
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
-- RN-10 da lobbies e RN-11 da candidatura (D-03 e D-04): outro lobby não cancelado, com
-- início dentro da janela (início - 2 h, início + 2 h) calculada pelo serviço, em que o
-- personagem é dono ou membro aceito.
SELECT EXISTS (
    SELECT 1 FROM lobbies l
    WHERE l.cancelled_at IS NULL
      AND l.id IS DISTINCT FROM @exclude_id
      AND l.starts_at > @window_start
      AND l.starts_at < @window_end
      AND (l.owner_character_id = @character_id
           OR EXISTS (SELECT 1 FROM applications a
                      WHERE a.lobby_id = l.id AND a.character_id = @character_id
                        AND a.status = 'accepted'))
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

-- name: CharacterInOpenLobby :one
-- RN-21 da lobbies e RN-25/RN-26 da candidatura: o personagem é dono de um lobby aberto
-- ou tem candidatura pendente ou aceita num lobby aberto.
SELECT EXISTS (
    SELECT 1 FROM lobbies l
    WHERE l.cancelled_at IS NULL AND l.starts_at > @now
      AND (l.owner_character_id = @character_id
           OR EXISTS (SELECT 1 FROM applications a
                      WHERE a.lobby_id = l.id AND a.character_id = @character_id
                        AND a.status IN ('pending', 'accepted')))
);

-- name: SetLobbyOwnerCharacter :exec
-- RN-19 da candidatura: o dono troca o próprio personagem, e a vaga passa a ser da função
-- do personagem novo.
UPDATE lobbies SET owner_character_id = @character_id, owner_role = @role WHERE id = @id;
