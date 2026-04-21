-- name: GetBoards :many
SELECT * FROM boards
ORDER BY created_at DESC;

-- name: GetBoardByID :one
SELECT * FROM boards
WHERE id = $1;

-- name: CreateBoard :one
INSERT INTO boards (title, size, author_id, lecturer_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteBoard :one
DELETE FROM boards
WHERE id = $1
RETURNING *;
