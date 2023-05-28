-- +goose Up
ALTER TABLE posts ALTER COLUMN published_at SET NOT NULL;

-- +goose Down
