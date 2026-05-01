-- +goose Up
-- +goose StatementBegin
CREATE TYPE token_purpose AS ENUM ('email_verification', 'password_reset');

CREATE TABLE IF NOT EXISTS verification_tokens
(
    token_hash          text PRIMARY KEY,
    user_id             uuid          NOT NULL REFERENCES users (id) ON DELETE CASCADE,

    purpose             token_purpose NOT NULL,
    expires_at          timestamptz   NOT NULL,

    created_at          timestamptz   NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_verification_tokens_expires_at ON verification_tokens (expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS verification_tokens;
DROP TYPE IF EXISTS token_purpose;
-- +goose StatementEnd