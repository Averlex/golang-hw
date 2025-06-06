-- +goose Up
-- This migration initializes the database schema, if it doesn't exist.
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    datetime TIMESTAMPTZ NOT NULL,
    duration INTERVAL NOT NULL,
    description TEXT NOT NULL,
    user_id TEXT NOT NULL,
    remind_in INTERVAL NOT NULL,

    CONSTRAINT duration_check CHECK (duration > INTERVAL '0 microseconds'),
    CONSTRAINT remind_in_check CHECK (remind_in >= INTERVAL '0 microseconds')
);

CREATE INDEX idx_events_user_id ON events(user_id);

-- +goose Down
-- Migration rollback.
DROP TABLE IF EXISTS events;