-- +goose Up
ALTER TABLE visits ADD COLUMN referrer TEXT;

-- +goose Down
ALTER TABLE visits DROP COLUMN referrer;
