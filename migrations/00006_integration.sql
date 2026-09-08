-- +goose Up
CREATE TABLE deliveries (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id      UUID        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status        VARCHAR(16) NOT NULL DEFAULT 'pending',
    courier_name  TEXT,
    courier_phone TEXT,
    eta_minutes   SMALLINT,
    assigned_at   TIMESTAMPTZ,
    picked_up_at  TIMESTAMPTZ,
    delivered_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT deliveries_order_uniq UNIQUE (order_id),
    CONSTRAINT deliveries_status_allowed CHECK (status IN
        ('pending','assigned','picked_up','delivered','failed')),
    CONSTRAINT deliveries_eta_pos CHECK (eta_minutes IS NULL OR eta_minutes > 0)
);

CREATE INDEX deliveries_status_idx ON deliveries (status)
    WHERE status IN ('pending','assigned','picked_up');

CREATE TRIGGER deliveries_set_updated_at BEFORE UPDATE ON deliveries
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE outbox_events (
    id              BIGSERIAL   PRIMARY KEY,
    aggregate_type  VARCHAR(32) NOT NULL,
    aggregate_id    UUID        NOT NULL,
    venue_id        UUID,
    event_type      VARCHAR(64) NOT NULL,
    payload         JSONB       NOT NULL,
    status          VARCHAR(16) NOT NULL DEFAULT 'pending',
    attempts        SMALLINT    NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT oe_status_allowed CHECK (status IN ('pending','published','failed','skipped')),
    CONSTRAINT oe_attempts_nonneg CHECK (attempts >= 0)
);

-- главный индекс воркера: размер = размеру необработанной очереди
CREATE INDEX oe_pending_idx ON outbox_events (next_attempt_at, id) WHERE status = 'pending';
CREATE INDEX oe_aggregate_idx ON outbox_events (aggregate_type, aggregate_id);

CREATE TABLE webhook_deliveries (
    id               BIGSERIAL   PRIMARY KEY,
    outbox_event_id  BIGINT      NOT NULL REFERENCES outbox_events(id) ON DELETE CASCADE,
    webhook_id       UUID        REFERENCES venue_webhooks(id) ON DELETE SET NULL,
    attempt          SMALLINT    NOT NULL,
    http_status      SMALLINT,
    duration_ms      INTEGER,
    error            TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT wd_attempt_pos CHECK (attempt >= 1)
);

CREATE INDEX wd_event_idx ON webhook_deliveries (outbox_event_id, attempt);

-- +goose Down
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS deliveries;
