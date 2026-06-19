-- name: SaveSentences :copyfrom
INSERT INTO sentences (level_number, korean, english)
VALUES ($1, $2, $3);
