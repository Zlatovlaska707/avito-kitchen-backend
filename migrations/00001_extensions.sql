-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Публичные номера заказов: sequence не блокируется транзакцией,
-- поэтому не создаёт точку сериализации при создании заказов.
CREATE SEQUENCE IF NOT EXISTS order_public_number_seq AS BIGINT START 1;

-- +goose Down
DROP SEQUENCE IF EXISTS order_public_number_seq;
DROP FUNCTION IF EXISTS set_updated_at();
DROP EXTENSION IF EXISTS pg_trgm;
DROP EXTENSION IF EXISTS pgcrypto;
