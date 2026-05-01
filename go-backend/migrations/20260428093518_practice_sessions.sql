-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS practice_sessions (
                                                 id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
                                                 user_id    UUID NOT NULL        REFERENCES users(id) ON DELETE CASCADE,
                                                 topic_id   UUID NOT NULL        REFERENCES topics(id) ON DELETE CASCADE,

    -- Статусы: 'active', 'completed', 'abandoned'
                                                 status     VARCHAR(20) NOT NULL DEFAULT 'active',

                                                 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                                 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индекс для получения истории практик конкретного юзера
CREATE INDEX idx_practice_sessions_user_id ON practice_sessions(user_id);
-- Индекс для аналитики по темам (какие темы чаще всего выбирают)
CREATE INDEX idx_practice_sessions_topic_id ON practice_sessions(topic_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS practice_sessions;
-- +goose StatementEnd