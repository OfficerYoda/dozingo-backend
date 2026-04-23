-- name: GetVotesByBoardID :one
SELECT
    COALESCE(SUM(vote_value), 0)::int AS score,
    COUNT(*)::int                     AS vote_count,
    COALESCE(MAX(CASE WHEN user_id = $2 THEN vote_value END), 0)::int AS user_vote
FROM votes
WHERE board_id = $1;

-- name: UpsertVote :one
INSERT INTO votes (user_id, board_id, vote_value)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, board_id)
DO UPDATE SET vote_value = EXCLUDED.vote_value
RETURNING *;

-- name: DeleteVote :one
DELETE FROM votes
WHERE user_id = $1 and board_id = $2
RETURNING *;
