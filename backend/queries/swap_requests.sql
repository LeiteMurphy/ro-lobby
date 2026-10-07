-- name: CreateSwapRequest :one
INSERT INTO swap_requests (application_id, from_character_id, to_character_id, to_role, reason, status, created_at)
VALUES (@application_id, @from_character_id, @to_character_id, @to_role, @reason, 'pending', @now)
RETURNING *;

-- name: InsertSwapRequestEvent :exec
-- RN-18: histórico de transições do pedido.
INSERT INTO swap_request_events (swap_request_id, from_status, to_status, actor_id, reason, at)
VALUES (@swap_request_id, @from_status, @to_status, @actor_id, @reason, @now);

-- name: GetSwapRequest :one
-- Leitura sem trava, para descobrir o membro antes de travar (D-10).
SELECT * FROM swap_requests WHERE id = @id;

-- name: GetSwapRequestForUpdate :one
SELECT * FROM swap_requests WHERE id = @id FOR UPDATE;

-- name: SetSwapRequestStatus :one
UPDATE swap_requests
SET status = @status, decision_reason = @decision_reason, decided_at = sqlc.arg(now)::timestamptz
WHERE id = @id
RETURNING *;

-- name: GetPendingSwapRequest :one
-- RN-20: o pedido pendente da candidatura, se houver.
SELECT * FROM swap_requests
WHERE application_id = @application_id AND status = 'pending'
FOR UPDATE;

-- name: GetLatestSwapRequest :one
-- O pedido mais recente da candidatura (o "meu pedido" do detalhe, D-12).
SELECT * FROM swap_requests
WHERE application_id = @application_id
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: ExpirePendingSwapsForLobby :many
-- RN-16, D-12: no cancelamento, os pedidos pendentes expiram.
UPDATE swap_requests SET status = 'expired', decided_at = sqlc.arg(now)::timestamptz
WHERE status = 'pending'
  AND application_id IN (SELECT id FROM applications WHERE lobby_id = @lobby_id)
RETURNING id;

-- name: ListLobbySwapRequests :many
-- D-12: pedidos pendentes do lobby, com o membro, o personagem atual e o novo (só o dono vê).
SELECT sqlc.embed(swap_requests),
       a.user_id, a.role AS from_role,
       u.username, u.global_name,
       fc.nick AS from_nick, fc.class_id AS from_class_id, fc.level AS from_level,
       fc.portrait AS from_portrait,
       tc.nick AS to_nick, tc.class_id AS to_class_id, tc.level AS to_level,
       tc.portrait AS to_portrait
FROM swap_requests
JOIN applications a ON a.id = swap_requests.application_id
JOIN users u ON u.id = a.user_id
LEFT JOIN characters fc ON fc.id = swap_requests.from_character_id
LEFT JOIN characters tc ON tc.id = swap_requests.to_character_id
WHERE a.lobby_id = @lobby_id AND swap_requests.status = 'pending'
ORDER BY swap_requests.created_at, swap_requests.id;

-- name: ListSwapRequestEvents :many
SELECT * FROM swap_request_events WHERE swap_request_id = @swap_request_id ORDER BY at, id;
