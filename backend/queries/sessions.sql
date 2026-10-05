-- name: CreateSession :exec
INSERT INTO sessions (token_hash, user_id, created_at, last_used_at)
VALUES (@token_hash, @user_id, @now, @now);

-- name: GetActiveSession :one
-- RN-09 e RN-10: a Sessão vale enquanto o último uso for posterior a @valid_after
-- (agora menos 30 dias). Vencida ou desconhecida, não volta nada.
SELECT sqlc.embed(users), sessions.last_used_at
FROM sessions
JOIN users ON users.id = sessions.user_id
WHERE sessions.token_hash = @token_hash
  AND sessions.last_used_at > @valid_after;

-- name: TouchSession :execrows
-- RN-09: renova o último uso, mas só grava se o registrado for anterior a @stale_before
-- (agora menos 1 hora).
UPDATE sessions
SET last_used_at = @now
WHERE token_hash = @token_hash
  AND last_used_at < @stale_before;

-- name: DeleteSession :exec
-- RN-11: sair apaga só esta Sessão.
DELETE FROM sessions WHERE token_hash = @token_hash;
