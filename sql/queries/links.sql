-- name: CreateLink :one
INSERT INTO links (
    created_at, updated_at, url, interval_seconds
) VALUES (
    datetime('now'), datetime('now'), ?, ?
)
RETURNING *;

-- name: ListLinks :many
SELECT * FROM links
ORDER BY updated_at DESC;

-- name: GetLink :one
SELECT * FROM links
WHERE id = ?;

-- name: GetLinkFromURL :one
SELECT * FROM links
WHERE url = ?;

-- name: UpdateLink :one
UPDATE links
    SET updated_at = datetime('now'),
    interval_seconds = ?
WHERE id = ?
RETURNING *;

-- name: UpdateLinkFromURL :one
UPDATE links
    SET updated_at = datetime('now'),
    interval_seconds = ?
WHERE url = ?
RETURNING *;

-- name: DeleteLink :one
DELETE FROM links
WHERE id = ?
RETURNING *;

-- name: DeleteLinkFromURL :one
DELETE FROM links
WHERE url = ?
RETURNING *;
