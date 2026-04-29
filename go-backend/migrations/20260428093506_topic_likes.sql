-- +goose Up
CREATE TABLE topic_likes (
                             user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                             topic_id UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
                             PRIMARY KEY (user_id, topic_id)
);

-- +goose Down
DROP TABLE IF EXISTS topic_likes;