-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS auth_sessions (
                                             id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
                                             user_id    UUID NOT NULL        REFERENCES users(id) ON DELETE CASCADE,

                                             token_hash TEXT NOT NULL UNIQUE,
                                             expires_at TIMESTAMPTZ NOT NULL,
                                             revoked    BOOLEAN NOT NULL     DEFAULT FALSE,

                                             created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                             updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы для быстрого поиска токена и сессий юзера
CREATE INDEX idx_auth_sessions_token_hash ON auth_sessions(token_hash);
CREATE INDEX idx_auth_sessions_user_id ON auth_sessions(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auth_sessions;
-- +goose StatementEnd