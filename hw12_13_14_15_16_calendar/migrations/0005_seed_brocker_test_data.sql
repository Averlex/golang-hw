-- +goose Up
-- Migration for test data seeding to verify notification sending
INSERT INTO events (id, title, datetime, duration, description, user_id, remind_in, is_notified) VALUES
-- Should be notified
(
    '11111111-1111-1111-1111-111111111111',
    'Urgent Sync',
    '2025-07-08 15:25:00',
    INTERVAL '15 minutes',
    'Quick team sync for urgent issue',
    'user1',
    INTERVAL '10 days',
    FALSE
),
-- Should not be notified
(
    '22222222-2222-2222-2222-222222222222',
    'Standup Meeting',
    '2025-07-18 15:35:00',
    INTERVAL '30 minutes',
    'Daily standup meeting',
    'user2',
    INTERVAL '15 minutes',
    FALSE
),
-- Should be notified, past event
(
    '33333333-3333-3333-3333-333333333333',
    'Client Call',
    '2025-06-08 18:00:00',
    INTERVAL '1 hour',
    'Discuss project updates with client',
    'user3',
    INTERVAL '30 minutes',
    FALSE
),
-- Should not be notified, past event
(
    '44444444-4444-4444-4444-444444444444',
    'Team Workshop',
    '2025-06-12 10:00:00',
    INTERVAL '2 hours',
    'Technical workshop for team',
    'user1',
    INTERVAL '1 day',
    TRUE
),
-- Should be deleted, old event. May send notification
(
    '55555555-5555-5555-5555-555555555555',
    'Training Session',
    '2023-07-15 09:00:00',
    INTERVAL '3 days',
    'Annual training for new tools',
    'user2',
    INTERVAL '2 days',
    FALSE
),
-- Should not be notified
(
    '66666666-6666-6666-6666-666666666666',
    'Company Event',
    '2025-07-22 18:00:00',
    INTERVAL '3 hours',
    'Annual company gathering',
    'user3',
    INTERVAL '30 days',
    TRUE
),
-- Should not be notified - no remind_in duration set
(
    '77777777-7777-7777-7777-777777777777',
    'Some Event',
    '2025-07-27 14:00:00',
    INTERVAL '2 hours',
    'Something important',
    'user42',
    INTERVAL '0 days',
    FALSE
);

-- +goose Down
-- Cleanup test data
TRUNCATE TABLE events;