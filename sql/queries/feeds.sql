-- name: CreateFeed :one
INSERT INTO
  feeds (id, created_at, updated_at, name, url, user_id)
VALUES
  ($1, $2, $3, $4, $5, $6)
RETURNING
  *;

-- name: GetFeeds :many
SELECT
  feeds.name AS Feed_name,
  feeds.url AS URL,
  users.name AS User
FROM
  feeds
  JOIN users ON users.id = feeds.user_id;
