-- +goose Up
-- +goose StatementBegin
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    account_number VARCHAR(60) NOT NULL UNIQUE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    balance NUMERIC(19, 4) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    account_status VARCHAR(10) NOT NULL CHECK (account_status IN ('ACTIVE', 'FROZEN', 'CLOSED')),
    account_type VARCHAR(10) NOT NULL CHECK (account_type IN ('CHECKING', 'SAVINGS'))
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS accounts;
-- +goose StatementEnd
