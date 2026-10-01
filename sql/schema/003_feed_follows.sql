-- +goose Up
CREATE TABLE feed_follows (
  id INTEGER PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  user_id INTEGER NOT NULL,
  feed_id INTEGER NOT NULL,
  UNIQUE (user_id, feed_id),
  CONSTRAINT fk_users FOREIGN key (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT fk_feeds FOREIGN key (feed_id) REFERENCES feeds (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feed_follows;
