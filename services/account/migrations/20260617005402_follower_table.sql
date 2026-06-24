-- +goose Up
CREATE TABLE follows (
                         follower_id BIGINT NOT NULL,
                         following_id BIGINT NOT NULL,
                         created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

                         PRIMARY KEY (follower_id, following_id),

                         CHECK (follower_id <> following_id),
                         FOREIGN KEY (follower_id) REFERENCES accounts(id) ON DELETE CASCADE,
                         FOREIGN KEY (following_id) REFERENCES accounts(id) ON DELETE CASCADE
);
-- +goose Down
DROP TABLE IF EXISTS follows;
