-- Sala do grupo no Discord (spec discord-bot, design seção 4, D-02).

-- name: EnqueueDiscordJob :exec
-- D-02: o efeito nasce na mesma transação da ação do domínio.
INSERT INTO discord_jobs (lobby_id, kind, user_id, payload, run_at, created_at)
VALUES (@lobby_id, @kind, sqlc.narg(user_id), @payload, @now, @now);

-- name: NextDiscordJob :one
-- RN-21 / RNF-04: o próximo job vencido, travado para um trabalhador só.
SELECT j.id, j.lobby_id, j.kind, j.user_id, j.payload, j.attempts,
       r.status AS room_status, r.category_id, r.text_channel_id, r.voice_channel_id
FROM discord_jobs j
JOIN discord_rooms r ON r.lobby_id = j.lobby_id
WHERE j.done_at IS NULL AND j.run_at <= @now
ORDER BY j.run_at, j.id
LIMIT 1
FOR UPDATE OF j SKIP LOCKED;

-- name: FinishDiscordJob :exec
UPDATE discord_jobs SET done_at = sqlc.arg(now)::timestamptz, last_error = sqlc.narg(last_error) WHERE id = @id;

-- name: RetryDiscordJob :exec
-- RN-21: falhou; tenta de novo em run_at.
UPDATE discord_jobs SET attempts = attempts + 1, run_at = @run_at, last_error = @last_error WHERE id = @id;

-- name: RoomMember :one
-- RN-07 / D-02: na hora de rodar o sync_access, o Usuário é membro (dono ou aceito) e está
-- conectado?
SELECT u.discord_id,
       (u.discord_connected_at IS NOT NULL)::boolean AS connected,
       (EXISTS (SELECT 1 FROM lobbies l WHERE l.id = @lobby_id AND l.owner_id = u.id)
        OR EXISTS (SELECT 1 FROM applications a
                   WHERE a.lobby_id = @lobby_id AND a.user_id = u.id AND a.status = 'accepted'))::boolean AS member
FROM users u
WHERE u.id = @user_id;

-- name: RoomMembers :many
-- RN-08: o dono e os membros aceitos, com o personagem e se estão conectados.
SELECT x.user_id, u.discord_id, (u.discord_connected_at IS NOT NULL)::boolean AS connected,
       c.nick, c.class_id, x.role, x.is_owner
FROM (
    SELECT l.owner_id AS user_id, l.owner_character_id AS character_id, l.owner_role AS role,
           true AS is_owner, 0 AS ord, l.created_at AS at
    FROM lobbies l WHERE l.id = @lobby_id
    UNION ALL
    SELECT a.user_id, a.character_id, a.role, false, 1, a.created_at
    FROM applications a WHERE a.lobby_id = @lobby_id AND a.status = 'accepted'
) x
JOIN users u ON u.id = x.user_id
LEFT JOIN characters c ON c.id = x.character_id
ORDER BY x.ord, x.at;

-- name: DueReminders :many
-- RN-14: salas abertas na janela do lembrete (window_end = agora + 15 min), de lobby não
-- cancelado, abertas antes da janela e que ainda não lembraram. Depois do início, o
-- lembrete não serve mais.
SELECT r.lobby_id, r.text_channel_id, l.starts_at
FROM discord_rooms r
JOIN lobbies l ON l.id = r.lobby_id
WHERE r.status = 'open' AND r.reminder_sent_at IS NULL AND l.cancelled_at IS NULL
  AND l.starts_at > @now
  AND l.starts_at <= @window_end
  AND r.opened_at <= l.starts_at - interval '15 minutes'
FOR UPDATE OF r SKIP LOCKED;

-- name: MarkReminderSent :exec
UPDATE discord_rooms SET reminder_sent_at = sqlc.arg(now)::timestamptz WHERE lobby_id = @lobby_id;

-- name: DueRoomDeletions :many
-- RN-16 / RN-17: aberta 6 h depois do início (started_before = agora - 6 h); encerrando
-- 1 h depois do cancelamento (closed_before = agora - 1 h).
SELECT r.lobby_id, r.category_id, r.text_channel_id, r.voice_channel_id
FROM discord_rooms r
JOIN lobbies l ON l.id = r.lobby_id
WHERE (r.status = 'open' AND l.starts_at <= @started_before)
   OR (r.status = 'closing' AND r.closing_at <= sqlc.arg(closed_before)::timestamptz);

-- name: MarkRoomDeleted :exec
-- RN-09: apagada é final.
UPDATE discord_rooms SET status = 'deleted', deleted_at = sqlc.arg(now)::timestamptz
WHERE lobby_id = @lobby_id AND status <> 'deleted';
