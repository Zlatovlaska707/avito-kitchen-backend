-- +goose Up
CREATE TABLE customers (
    id         UUID        PRIMARY KEY,       -- приходит от клиента (X-Customer-Id)
    phone      TEXT,
    name       TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT customers_phone_uniq   UNIQUE (phone),
    CONSTRAINT customers_phone_format CHECK (phone IS NULL OR phone ~ '^\+7[0-9]{10}$'),
    CONSTRAINT customers_name_len     CHECK (name IS NULL OR char_length(name) <= 100)
);

CREATE TRIGGER customers_set_updated_at BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE carts (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID        NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    venue_id    UUID        NOT NULL REFERENCES venues(id) ON DELETE RESTRICT,
    status      VARCHAR(16) NOT NULL DEFAULT 'active',
    order_id    UUID,       -- FK добавляется в 00005 (таблица orders создаётся позже)
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT carts_status_allowed CHECK (status IN ('active','converted','abandoned')),
    CONSTRAINT carts_converted_has_order CHECK (status <> 'converted' OR order_id IS NOT NULL)
);

-- инвариант «одна активная корзина на покупателя» + индекс под получение корзины
CREATE UNIQUE INDEX carts_one_active_per_customer_idx ON carts (customer_id) WHERE status = 'active';
CREATE INDEX carts_expiry_idx ON carts (expires_at) WHERE status = 'active';

CREATE TRIGGER carts_set_updated_at BEFORE UPDATE ON carts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE cart_items (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id          UUID        NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    menu_item_id     UUID        NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    quantity         SMALLINT    NOT NULL,
    selected_options JSONB       NOT NULL DEFAULT '[]',
    options_hash     CHAR(32)    NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT ci_line_uniq      UNIQUE (cart_id, menu_item_id, options_hash),
    CONSTRAINT ci_quantity_range CHECK (quantity BETWEEN 1 AND 99),
    CONSTRAINT ci_options_array  CHECK (jsonb_typeof(selected_options) = 'array')
);

CREATE INDEX ci_cart_idx ON cart_items (cart_id);

CREATE TRIGGER ci_set_updated_at BEFORE UPDATE ON cart_items
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS carts;
DROP TABLE IF EXISTS customers;
