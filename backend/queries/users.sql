-- name: UpsertUserByDiscordID :one
-- RN-05: o ID do Discord identifica o Usuário; logins seguintes atualizam os nomes.
INSERT INTO users (discord_id, username, global_name, created_at, last_login_at)
VALUES (@discord_id, @username, @global_name, @now, @now)
ON CONFLICT (discord_id) DO UPDATE
SET username      = EXCLUDED.username,
    global_name   = EXCLUDED.global_name,
    last_login_at = EXCLUDED.last_login_at
RETURNING *;

-- name: CountUsersByDiscordID :one
SELECT count(*) FROM users WHERE discord_id = @discord_id;
