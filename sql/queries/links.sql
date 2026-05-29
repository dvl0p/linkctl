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

-- name: UpdateLink :exec
UPDATE links
    SET updated_at = datetime('now'),
    interval_seconds = ?
WHERE id = ?;

-- name: UpdateLinkFromURL :exec
UPDATE links
    SET updated_at = datetime('now'),
    interval_seconds = ?
WHERE url = ?;

-- name: DeleteLink :exec
DELETE FROM links
WHERE id = ?;

-- name: DeleteLinkFromURL :exec
DELETE FROM links
WHERE url = ?;
