-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS oauth_connections
(
    id                  uuid PRIMARY KEY       DEFAULT gen_random_uuid(),
    user_id             uuid          NOT NULL REFERENCES users (id) ON DELETE CASCADE,

    provider            varchar(32)   NOT NULL,
    provider_user_id    varchar(255)  NOT NULL,

    created_at          timestamptz   NOT NULL DEFAULT now(),

    UNIQUE (provider, provider_user_id)
);

CREATE INDEX IF NOT EXISTS idx_oauth_connections_user_id ON oauth_connections (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS oauth_connections;
-- +goose StatementEnd