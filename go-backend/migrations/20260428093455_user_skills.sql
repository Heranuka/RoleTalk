-- +goose Up
CREATE TABLE user_skills (
                             user_id UUID REFERENCES users(id) ON DELETE CASCADE,
                             empathy INT DEFAULT 0,
                             persuasion INT DEFAULT 0,
                             structure INT DEFAULT 0,
                             stress_resistance INT DEFAULT 0,
                             PRIMARY KEY (user_id)
);

-- +goose Down
DROP TABLE IF EXISTS user_skills;