-- +goose Up
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
DROP TABLE sessions;
