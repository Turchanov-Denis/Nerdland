-- +goose Up
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

-- +goose Down
DROP TABLE profiles;
DROP TABLE accounts;