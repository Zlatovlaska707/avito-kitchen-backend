-- +goose Up
CREATE TABLE orders (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    public_number      TEXT        NOT NULL,
    customer_id        UUID        NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    venue_id           UUID        NOT NULL REFERENCES venues(id)    ON DELETE RESTRICT,
    cart_id            UUID        REFERENCES carts(id) ON DELETE SET NULL,
    status             VARCHAR(24) NOT NULL,
    fulfillment_type   VARCHAR(16) NOT NULL,
    delivery_address   JSONB,
    contact_phone      TEXT        NOT NULL,
    items_total_minor  BIGINT      NOT NULL,
    delivery_fee_minor BIGINT      NOT NULL DEFAULT 0,
    discount_minor     BIGINT      NOT NULL DEFAULT 0,
    total_minor        BIGINT      NOT NULL,
    currency           CHAR(3)     NOT NULL DEFAULT 'RUB',
    payment_method     VARCHAR(24) NOT NULL,
    payment_status     VARCHAR(16) NOT NULL,
    comment            TEXT,
    external_order_id  TEXT,
    prep_minutes       SMALLINT,
    eta_at             TIMESTAMPTZ,
    placed_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    accepted_at        TIMESTAMPTZ,
    ready_at           TIMESTAMPTZ,
    delivered_at       TIMESTAMPTZ,
    closed_at          TIMESTAMPTZ,
    cancel_reason      TEXT,
    cancelled_by       VARCHAR(16),
    version            INTEGER     NOT NULL DEFAULT 1,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT orders_public_number_uniq UNIQUE (public_number),
    CONSTRAINT orders_status_allowed CHECK (status IN
        ('placed','accepted','ready','in_delivery','completed','cancelled','rejected')),
    CONSTRAINT orders_fulfillment_allowed CHECK (fulfillment_type IN ('delivery','pickup')),
    CONSTRAINT orders_payment_method_allowed CHECK (payment_method IN ('cash','card_online_mock')),
    CONSTRAINT orders_payment_status_allowed CHECK (payment_status IN
        ('not_required','pending','paid','failed')),
    CONSTRAINT orders_cancelled_by_allowed CHECK (cancelled_by IS NULL OR cancelled_by IN
        ('customer','venue','system','admin')),
    CONSTRAINT orders_currency_format CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT orders_items_total_nonneg CHECK (items_total_minor >= 0),
    CONSTRAINT orders_fee_nonneg        CHECK (delivery_fee_minor >= 0),
    CONSTRAINT orders_discount_nonneg   CHECK (discount_minor >= 0),
    -- главный денежный инвариант: защищён на уровне БД
    CONSTRAINT orders_total_consistent  CHECK
        (total_minor = items_total_minor + delivery_fee_minor - discount_minor),
    CONSTRAINT orders_total_nonneg      CHECK (total_minor >= 0),
    CONSTRAINT orders_delivery_has_address CHECK
        (fulfillment_type <> 'delivery' OR delivery_address IS NOT NULL),
    CONSTRAINT orders_phone_format   CHECK (contact_phone ~ '^\+7[0-9]{10}$'),
    CONSTRAINT orders_comment_len    CHECK (comment IS NULL OR char_length(comment) <= 500),
    CONSTRAINT orders_cancel_has_reason CHECK
        (status NOT IN ('cancelled','rejected') OR cancel_reason IS NOT NULL),
    CONSTRAINT orders_prep_range CHECK (prep_minutes IS NULL OR prep_minutes BETWEEN 1 AND 240)
);

CREATE INDEX orders_venue_queue_idx    ON orders (venue_id, status, created_at DESC);
CREATE INDEX orders_customer_feed_idx  ON orders (customer_id, created_at DESC);
CREATE INDEX orders_accept_timeout_idx ON orders (placed_at) WHERE status = 'placed';
CREATE UNIQUE INDEX orders_external_uniq_idx ON orders (venue_id, external_order_id)
    WHERE external_order_id IS NOT NULL;

CREATE TRIGGER orders_set_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE carts
    ADD CONSTRAINT carts_order_fk FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE SET NULL;

CREATE TABLE order_items (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id          UUID        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id      UUID        REFERENCES menu_items(id) ON DELETE SET NULL,
    item_external_id  TEXT        NOT NULL,
    name_snapshot     TEXT        NOT NULL,
    unit_price_minor  BIGINT      NOT NULL,
    quantity          SMALLINT    NOT NULL,
    options_snapshot  JSONB       NOT NULL DEFAULT '[]',
    line_total_minor  BIGINT      NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT oi_unit_price_nonneg CHECK (unit_price_minor >= 0),
    CONSTRAINT oi_quantity_range    CHECK (quantity BETWEEN 1 AND 99),
    CONSTRAINT oi_line_total_nonneg CHECK (line_total_minor >= 0),
    CONSTRAINT oi_options_array     CHECK (jsonb_typeof(options_snapshot) = 'array')
);

CREATE INDEX oi_order_idx ON order_items (order_id);

CREATE TABLE order_status_history (
    id          BIGSERIAL   PRIMARY KEY,
    order_id    UUID        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    from_status VARCHAR(24),
    to_status   VARCHAR(24) NOT NULL,
    actor_type  VARCHAR(16) NOT NULL,
    actor_id    TEXT,
    reason      TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT osh_actor_allowed CHECK (actor_type IN
        ('customer','venue','courier','system','admin'))
);

CREATE INDEX osh_order_idx ON order_status_history (order_id, created_at);

CREATE TABLE idempotency_keys (
    scope           VARCHAR(64) NOT NULL,
    key             TEXT        NOT NULL,
    request_hash    CHAR(64)    NOT NULL,
    response_status SMALLINT,
    response_body   JSONB,
    resource_id     UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL,

    PRIMARY KEY (scope, key),
    CONSTRAINT ik_key_len CHECK (char_length(key) BETWEEN 8 AND 200)
);

CREATE INDEX ik_expiry_idx ON idempotency_keys (expires_at);

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS order_status_history;
DROP TABLE IF EXISTS order_items;
ALTER TABLE carts DROP CONSTRAINT IF EXISTS carts_order_fk;
DROP TABLE IF EXISTS orders;
