-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE IF NOT EXISTS users
(
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    email             VARCHAR(255) NOT NULL,
    password_hash     TEXT NOT NULL,

    username          VARCHAR(32),
    display_name      VARCHAR(100),
    photo_url         TEXT,
    interface_lang    VARCHAR(10) NOT NULL DEFAULT 'ru',
    practice_lang     VARCHAR(10) NOT NULL DEFAULT 'en',

    role              user_role NOT NULL DEFAULT 'user',
    is_email_verified BOOLEAN NOT NULL DEFAULT false,

    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT users_email_key UNIQUE (email),
    CONSTRAINT users_username_key UNIQUE (username)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd