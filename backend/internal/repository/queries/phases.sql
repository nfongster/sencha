-- name: ListPhases :many
SELECT number, name FROM phases ORDER BY number;

-- name: GetPhase :one
SELECT number, name FROM phases WHERE number = $1;

-- name: CreatePhase :exec
INSERT INTO phases (number, name) VALUES ($1, $2);

-- name: MaxPhaseNumber :one
SELECT COALESCE(MAX(number), 0) FROM phases;
