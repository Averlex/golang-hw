-- +goose Up
-- Migration for test data seeding
INSERT INTO events (id, title, datetime, duration, description, user_id, remind_in) VALUES
-- On-going events
(
    '11111111-1111-1111-1111-111111111111',
    'Team Meeting',
    NOW() + INTERVAL '1 hour',
    INTERVAL '30 minutes',
    'Weekly project sync',
    'user1',
    INTERVAL '15 minutes'
),
(
    '22222222-2222-2222-2222-222222222222',
    'Lunch Break',
    NOW() - INTERVAL '15 minutes',
    INTERVAL '1 hour',
    'Team lunch at new restaurant',
    'user2',
    INTERVAL '0 minutes'
),

-- Past event, without description
(
    '33333333-3333-3333-3333-333333333333',
    'Project Deadline',
    NOW() - INTERVAL '2 days',
    INTERVAL '4 hours',
    '',
    'user1',
    INTERVAL '1 hour'
),

-- Upcoming event
(
    '44444444-4444-4444-4444-444444444444',
    'Conference Talk',
    NOW() + INTERVAL '3 days',
    INTERVAL '2 hours',
    'Annual developer conference',
    'user3',
    INTERVAL '1 day'
),

-- Long event
(
    '55555555-5555-5555-5555-555555555555',
    'Vacation',
    NOW() + INTERVAL '1 week',
    INTERVAL '5 days',
    'Summer holidays',
    'user2',
    INTERVAL '2 days'
),

-- Short-lived event
(
    '66666666-6666-6666-6666-666666666666',
    'Quick Sync',
    NOW(),
    INTERVAL '10 minutes',
    'Urgent team update',
    'user3',
    INTERVAL '5 minutes'
);

-- +goose Down
-- Cleanup test data
TRUNCATE TABLE events;