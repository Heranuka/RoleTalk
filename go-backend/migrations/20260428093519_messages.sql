-- +goose Up
CREATE TABLE messages (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          session_id UUID REFERENCES practice_sessions(id) ON DELETE CASCADE,
                          sender_role VARCHAR(20),
                          text_content TEXT,
                          audio_url TEXT,
                          created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS messages;