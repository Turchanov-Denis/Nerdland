-- +goose Up
SELECT 'up SQL query';

CREATE TABLE accounts (
                          id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                          email TEXT NOT NULL,
                          password_hash TEXT NOT NULL,
                          created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                          CONSTRAINT accounts_email_unique UNIQUE (email)
);

CREATE TABLE profiles (
                          account_id BIGINT PRIMARY KEY,
                          username TEXT NOT NULL,
                          display_name TEXT NOT NULL,
                          bio TEXT NOT NULL DEFAULT '',
                          avatar_url TEXT NOT NULL DEFAULT '',
                          FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
                          CONSTRAINT profiles_username_unique UNIQUE (username)
);

CREATE TABLE sessions (
                          id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                          account_id BIGINT NOT NULL,
                          refresh_token_hash TEXT NOT NULL,
                          user_agent TEXT DEFAULT NULL,
                          created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                          expires_at TIMESTAMPTZ NOT NULL,
                          revoked_at TIMESTAMPTZ DEFAULT NULL,
                          FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
                          CONSTRAINT session_refresh_token_hash_unique UNIQUE (refresh_token_hash)
);
-- +goose Down
SELECT 'down SQL query';

DROP TABLE sessions;
DROP TABLE profiles;
DROP TABLE accounts;