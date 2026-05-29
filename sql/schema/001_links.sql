-- +goose Up
CREATE TABLE links(
    id INTEGER PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    url TEXT UNIQUE NOT NULL,
    interval_seconds INTEGER NOT NULL
);

-- +goose Down
DROP TABLE links;
