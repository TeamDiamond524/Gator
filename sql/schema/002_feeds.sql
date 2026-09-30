-- +goose Up
CREATE TABLE feeds (
  id INTEGER PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  name TEXT NOT NULL,
  url TEXT NOT NULL,
  user_id INTEGER NOT NULL,
  CONSTRAINT fk_feeds_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;
