-- name: SaveSentences :copyfrom
INSERT INTO sentences (level_number, korean, english)
VALUES ($1, $2, $3);

-- name: ListSentencesByLevel :many
SELECT level_number, korean, english FROM sentences WHERE level_number = $1 ORDER BY id;

-- name: CountSentencesByLevel :one
SELECT COUNT(*) FROM sentences WHERE level_number = $1;

-- name: DeleteSentencesByLevel :exec
DELETE FROM sentences WHERE level_number = $1;

-- name: GetRandomSentencesByLevel :many
SELECT level_number, korean, english FROM sentences WHERE level_number = $1 ORDER BY RANDOM() LIMIT $2;
