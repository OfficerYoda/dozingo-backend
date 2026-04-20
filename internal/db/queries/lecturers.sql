-- name: GetLecturers :many
SELECT * FROM lecturers
ORDER BY name ASC;

-- name: GetLecturersById :one
SELECT * FROM lecturers
WHERE id = $1;

-- name: GetLecturersBySlug :one
SELECT * FROM lecturers
WHERE slug = $1;

-- name: CreateLecturer :one
INSERT INTO lecturers (name, slug)
VALUES ($1, $2)
RETURNING *;

-- name: DeleteLecturer :one
DELETE FROM lecturers
WHERE id = $1
RETURNING *;
