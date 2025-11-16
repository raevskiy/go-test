-- +goose Up
-- +goose StatementBegin
-- required for gen_random_uuid
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

ALTER TABLE users
    ADD COLUMN uuid UUID NOT NULL DEFAULT gen_random_uuid();

CREATE UNIQUE INDEX users_uuid_idx ON users(uuid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users_uuid_idx;

ALTER TABLE users
DROP COLUMN uuid;
-- +goose StatementEnd
