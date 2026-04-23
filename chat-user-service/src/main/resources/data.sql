CREATE EXTENSION IF NOT EXISTS "pgcrypto";

INSERT INTO users (
	id,
    user_id,
    username,
    email,
    created_at,
    updated_at,
    is_active
)
VALUES (
	nextval('users_seq'),
    '089989b5-1ef7-47c2-be09-184fff3d5ddd',
    'dude',
    'dude@gmail.com',
    NOW(),
    null,
    TRUE
) ON CONFLICT DO NOTHING;

INSERT INTO users (
	id,
    user_id,
    username,
    email,
    created_at,
    updated_at,
    is_active
)
VALUES (
	nextval('users_seq'),
    '68afaa0a-618d-4f5a-8d1f-5a35eebed7fa',
    'mary',
    'mary@gmail.com',
    NOW(),
    null,
    TRUE
) ON CONFLICT DO NOTHING;

INSERT INTO users (
	id,
    user_id,
    username,
    email,
    created_at,
    updated_at,
    is_active
)
VALUES (
	nextval('users_seq'),
    '68afaa0a-618d-4f5a-8d1f-5a35eebedbbb',
    'bob',
    'bob@gmail.com',
    NOW(),
    null,
    TRUE
) ON CONFLICT DO NOTHING;