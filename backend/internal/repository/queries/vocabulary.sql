-- name: VocabularyUpTo :many
SELECT korean, english
FROM vocabulary
WHERE level_number <= $1
ORDER BY level_number, id;

-- name: AddVocabulary :copyfrom
INSERT INTO vocabulary (level_number, korean, english)
VALUES ($1, $2, $3);
