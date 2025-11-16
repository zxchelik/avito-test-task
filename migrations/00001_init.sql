-- +goose Up
-- +goose StatementBegin

CREATE TABLE teams (
                       name TEXT PRIMARY KEY
);

CREATE TABLE users (
                       id         TEXT PRIMARY KEY,
                       name       TEXT NOT NULL,
                       team_name  TEXT NOT NULL REFERENCES teams(name) ON DELETE RESTRICT,
                       is_active  BOOLEAN NOT NULL DEFAULT TRUE,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_team ON users(team_name);
CREATE INDEX idx_users_is_active ON users(is_active);

CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE pull_requests (
                               id         TEXT PRIMARY KEY,
                               title      TEXT NOT NULL,
                               author_id  TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
                               status     pr_status NOT NULL DEFAULT 'OPEN',
                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               merged_at  TIMESTAMPTZ NULL
);

CREATE INDEX idx_pr_author ON pull_requests(author_id);
CREATE INDEX idx_pr_status ON pull_requests(status);

CREATE TABLE pull_request_reviewers (
                                        pr_id       TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
                                        user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
                                        assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                        PRIMARY KEY (pr_id, user_id)
);

CREATE INDEX idx_reviews_user ON pull_request_reviewers(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS pull_request_reviewers;
DROP INDEX IF EXISTS idx_pr_status;
DROP INDEX IF EXISTS idx_pr_author;
DROP TABLE IF EXISTS pull_requests;
DROP TYPE IF EXISTS pr_status;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_team;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;

-- +goose StatementEnd
