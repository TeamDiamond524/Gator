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

-- name: GetFeed :one
SELECT
  *
FROM
  feeds
WHERE
  url = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET
  updated_at = $2,
  last_fetched_at = $3
WHERE
  id = $1;

-- name: GetNextFeedToFetch :one
SELECT
  feeds.id,
  feeds.url
FROM
  feeds
ORDER BY
  last_fetched_at ASC NULLS FIRST
LIMIT
  1;
