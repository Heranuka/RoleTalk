-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE topics (
                        id                    UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
                        author_id             UUID                 REFERENCES users(id) ON DELETE SET NULL,
                        title                 VARCHAR(255)         NOT NULL,
                        description           TEXT,
                        emoji                 VARCHAR(10),
                        difficulty_level      VARCHAR(10),
                        is_official           BOOLEAN              NOT NULL DEFAULT FALSE,
                        likes_count           INT                  NOT NULL DEFAULT 0,

    -- Roleplay specific fields
                        my_role               VARCHAR(255)         NOT NULL DEFAULT 'Student',
                        partner_role          VARCHAR(255)         NOT NULL DEFAULT 'AI Tutor',
                        partner_emoji         VARCHAR(10)          NOT NULL DEFAULT '🤖',
                        goal                  TEXT                 NOT NULL DEFAULT 'Practice conversation',
                        partner_secret_motive TEXT,

                        created_at            TIMESTAMP            NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_topics_is_official_created_at ON topics (is_official, created_at DESC);
CREATE INDEX idx_topics_likes_count_created_at ON topics (likes_count DESC, created_at DESC);
CREATE INDEX idx_topics_author_id ON topics (author_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS topics;
-- +goose StatementEnd