-- name: GetCellsByBoardID :many
SELECT * FROM cells
WHERE board_id = $1;

-- name: CreateCell :one
INSERT INTO cells (board_id, content)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateCell :one
UPDATE cells
SET content = $1
WHERE id = $2 AND board_id = $3
RETURNING *;

-- name: DeleteCell :one
DELETE FROM cells
WHERE id = $1 and board_id = $2
RETURNING *;
