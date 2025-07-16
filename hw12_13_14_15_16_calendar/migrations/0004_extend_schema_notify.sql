-- +goose Up
-- Extend event schema with is_notified field
ALTER TABLE events
ADD is_notified bool NOT NULL DEFAULT false;


-- +goose Down
-- Remove is_notified field
ALTER TABLE events
DROP COLUMN IF EXISTS is_notified;
