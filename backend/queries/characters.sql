-- name: LockUser :one
-- D-06: trava o Usuário enquanto cria, exclui ou troca o principal, para o limite de 10
-- e o principal único valerem com pedidos simultâneos.
SELECT id FROM users WHERE id = @user_id FOR UPDATE;

-- name: CountCharacters :one
SELECT count(*) FROM characters WHERE user_id = @user_id;

-- name: CreateCharacter :one
INSERT INTO characters (user_id, nick, class_id, level, role, portrait, link, is_main, created_at)
VALUES (@user_id, @nick, @class_id, @level, @role, @portrait, @link, @is_main, @now)
RETURNING *;

-- name: ListCharacters :many
-- RN-17: o principal primeiro, depois por ordem de cadastro.
SELECT * FROM characters
WHERE user_id = @user_id
ORDER BY is_main DESC, created_at, seq;

-- name: UpdateCharacter :one
-- D-08: filtra pelo dono; personagem de outro Usuário não volta nada (RN-02).
UPDATE characters
SET nick = @nick, class_id = @class_id, level = @level, role = @role,
    portrait = @portrait, link = @link
WHERE id = @id AND user_id = @user_id
RETURNING *;

-- name: DeleteCharacter :one
-- Devolve se o excluído era o principal, para passar a vez ao mais antigo (RN-14).
DELETE FROM characters
WHERE id = @id AND user_id = @user_id
RETURNING is_main;

-- name: ClearMain :exec
UPDATE characters SET is_main = false WHERE user_id = @user_id AND is_main;

-- name: SetMain :execrows
UPDATE characters SET is_main = true WHERE id = @id AND user_id = @user_id;

-- name: PromoteOldest :exec
-- RN-14: o personagem mais antigo que sobrou vira o principal.
UPDATE characters SET is_main = true
WHERE id = (
    SELECT c.id FROM characters c
    WHERE c.user_id = @user_id
    ORDER BY c.created_at, c.seq
    LIMIT 1
);

-- name: GetOwnCharacter :one
-- Personagem do próprio Usuário (RN-08 da spec lobbies); de outro Usuário não volta nada.
SELECT * FROM characters WHERE id = @id AND user_id = @user_id;
