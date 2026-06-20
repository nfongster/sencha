-- name: LevelsInPhase :many
SELECT number, phase_number, grammar_md
FROM levels
WHERE phase_number = $1
ORDER BY number;

-- name: GetLevel :one
SELECT number, phase_number, grammar_md
FROM levels
WHERE number = $1;

-- name: CreateLevel :exec
INSERT INTO levels (number, phase_number, grammar_md)
VALUES ($1, $2, $3);

-- name: UpdateLevel :exec
UPDATE levels SET grammar_md = $2 WHERE number = $1;

-- name: MaxLevelNumber :one
SELECT COALESCE(MAX(number), 0) FROM levels;

-- name: LevelsUpTo :many
SELECT number, phase_number, grammar_md
FROM levels
WHERE number <= $1
ORDER BY number;
