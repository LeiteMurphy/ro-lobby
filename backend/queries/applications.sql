-- name: CreateApplication :one
INSERT INTO applications (lobby_id, user_id, character_id, role, message, status, created_at)
VALUES (@lobby_id, @user_id, @character_id, @role, @message, 'pending', @now)
RETURNING *;

-- name: InsertApplicationEvent :exec
-- RN-18: histórico de transições.
INSERT INTO application_events (application_id, from_status, to_status, actor_id, reason, at)
VALUES (@application_id, @from_status, @to_status, @actor_id, @reason, @now);

-- name: GetApplicationForUpdate :one
SELECT * FROM applications WHERE id = @id FOR UPDATE;

-- name: SetApplicationStatus :one
UPDATE applications
SET status = @status, reason = @reason, decided_at = sqlc.arg(now)::timestamptz
WHERE id = @id
RETURNING *;

-- name: HasActiveApplication :one
-- RN-02: candidatura pendente ou aceita do Usuário no lobby.
SELECT EXISTS (
    SELECT 1 FROM applications
    WHERE lobby_id = @lobby_id AND user_id = @user_id AND status IN ('pending', 'accepted')
);

-- name: WasRejected :one
-- RN-07: recusado não se candidata de novo ao mesmo lobby.
SELECT EXISTS (
    SELECT 1 FROM applications
    WHERE lobby_id = @lobby_id AND user_id = @user_id AND status = 'rejected'
);

-- name: ExpirePendingForLobby :many
-- RN-16: no cancelamento, as pendentes expiram.
UPDATE applications SET status = 'expired', decided_at = sqlc.arg(now)::timestamptz
WHERE lobby_id = @lobby_id AND status = 'pending'
RETURNING id;

-- name: ListLobbyApplications :many
-- Candidaturas ativas do lobby, com o personagem e o Discord (a visibilidade é do serviço).
SELECT sqlc.embed(applications),
       c.nick, c.class_id, c.level, c.portrait, c.link,
       u.username, u.global_name
FROM applications
JOIN users u ON u.id = applications.user_id
LEFT JOIN characters c ON c.id = applications.character_id
WHERE applications.lobby_id = @lobby_id AND applications.status IN ('pending', 'accepted')
ORDER BY applications.created_at, applications.id;

-- name: GetUserApplicationInLobby :one
-- A candidatura mais recente do Usuário no lobby (a "minha candidatura" do detalhe).
SELECT * FROM applications
WHERE lobby_id = @lobby_id AND user_id = @user_id
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: ListMyApplications :many
-- RN-33: candidaturas do Usuário, com o lobby e o personagem, as mais recentes primeiro.
SELECT sqlc.embed(applications),
       l.instance_name, l.starts_at, l.cancelled_at AS lobby_cancelled_at,
       c.nick, c.class_id, c.level, c.portrait
FROM applications
JOIN lobbies l ON l.id = applications.lobby_id
LEFT JOIN characters c ON c.id = applications.character_id
WHERE applications.user_id = @user_id
ORDER BY applications.created_at DESC, applications.id DESC;

-- name: ListApplicationEvents :many
SELECT * FROM application_events WHERE application_id = @application_id ORDER BY at, id;

-- name: LockLobby :one
-- D-05: trava o lobby no aceite, depois do Usuário do candidato.
SELECT id FROM lobbies WHERE id = @id FOR UPDATE;

-- name: GetApplication :one
-- Leitura sem trava, para descobrir o candidato antes de travar (D-05).
SELECT * FROM applications WHERE id = @id;

-- name: SetApplicationBlocked :exec
-- RN-15: remoção com bloqueio.
UPDATE applications SET blocked = true WHERE id = @id;

-- name: UnblockUserInLobby :exec
-- RN-15 / CA-06.9: o dono desbloqueia o usuário do personagem neste lobby; a candidatura
-- removida continua no histórico.
UPDATE applications SET blocked = false
WHERE lobby_id = @lobby_id AND blocked
  AND user_id = (SELECT c.user_id FROM characters c WHERE c.id = @character_id);

-- name: IsBlocked :one
-- RN-07, RN-15: removido com bloqueio não se candidata de novo ao mesmo lobby.
SELECT EXISTS (
    SELECT 1 FROM applications
    WHERE lobby_id = @lobby_id AND user_id = @user_id AND status = 'removed' AND blocked
);

-- name: SetApplicationCharacter :one
-- RN-23: o aceite da troca muda o personagem e a função de uma vez.
UPDATE applications SET character_id = @character_id, role = @role
WHERE id = @id
RETURNING *;
