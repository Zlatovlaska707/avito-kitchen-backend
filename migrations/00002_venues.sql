-- +goose Up
CREATE TABLE venues (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug               TEXT        NOT NULL,
    name               TEXT        NOT NULL,
    description        TEXT,
    status             VARCHAR(16) NOT NULL DEFAULT 'draft',
    accepts_orders     BOOLEAN     NOT NULL DEFAULT false,
    city_id            INTEGER     NOT NULL,
    address_text       TEXT        NOT NULL,
    lat                NUMERIC(9,6) NOT NULL,
    lon                NUMERIC(9,6) NOT NULL,
    timezone           TEXT        NOT NULL DEFAULT 'Europe/Moscow',
    avg_prep_minutes   SMALLINT    NOT NULL DEFAULT 25,
    delivery_fee_minor BIGINT      NOT NULL DEFAULT 0,
    min_order_minor    BIGINT      NOT NULL DEFAULT 0,
    delivery_radius_m  INTEGER     NOT NULL DEFAULT 5000,
    rating             NUMERIC(2,1),
    pause_reason       TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT venues_slug_uniq        UNIQUE (slug),
    CONSTRAINT venues_slug_format      CHECK (slug ~ '^[a-z0-9-]{3,64}$'),
    CONSTRAINT venues_name_len         CHECK (char_length(name) BETWEEN 1 AND 200),
    CONSTRAINT venues_status_allowed   CHECK (status IN ('draft','active','suspended')),
    CONSTRAINT venues_lat_range        CHECK (lat BETWEEN -90 AND 90),
    CONSTRAINT venues_lon_range        CHECK (lon BETWEEN -180 AND 180),
    CONSTRAINT venues_prep_range       CHECK (avg_prep_minutes BETWEEN 1 AND 240),
    CONSTRAINT venues_fee_nonneg       CHECK (delivery_fee_minor >= 0),
    CONSTRAINT venues_min_order_nonneg CHECK (min_order_minor >= 0),
    CONSTRAINT venues_radius_range     CHECK (delivery_radius_m BETWEEN 100 AND 50000),
    CONSTRAINT venues_rating_range     CHECK (rating IS NULL OR rating BETWEEN 0 AND 5)
);

CREATE INDEX venues_city_active_idx ON venues (city_id) WHERE status = 'active';
CREATE INDEX venues_geo_idx         ON venues (lat, lon);

CREATE TRIGGER venues_set_updated_at BEFORE UPDATE ON venues
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE venue_working_hours (
    id               UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
    venue_id         UUID     NOT NULL REFERENCES venues(id) ON DELETE CASCADE,
    weekday          SMALLINT NOT NULL,
    opens_at         TIME     NOT NULL,
    closes_at        TIME     NOT NULL,
    crosses_midnight BOOLEAN  NOT NULL DEFAULT false,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT vwh_weekday_range CHECK (weekday BETWEEN 1 AND 7),
    CONSTRAINT vwh_uniq          UNIQUE (venue_id, weekday, opens_at),
    -- если интервал не переходит через полночь, закрытие строго позже открытия
    CONSTRAINT vwh_interval_valid CHECK (crosses_midnight OR closes_at > opens_at)
);

CREATE INDEX vwh_venue_weekday_idx ON venue_working_hours (venue_id, weekday);

CREATE TABLE venue_api_keys (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    venue_id     UUID        NOT NULL REFERENCES venues(id) ON DELETE CASCADE,
    key_prefix   CHAR(8)     NOT NULL,
    key_hash     TEXT        NOT NULL,
    scopes       TEXT[]      NOT NULL DEFAULT '{menu:write,orders:write}',
    is_active    BOOLEAN     NOT NULL DEFAULT true,
    last_used_at TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT vak_prefix_uniq  UNIQUE (key_prefix),
    CONSTRAINT vak_scopes_nonempty CHECK (array_length(scopes, 1) >= 1)
);

CREATE INDEX vak_venue_idx ON venue_api_keys (venue_id);

CREATE TRIGGER vak_set_updated_at BEFORE UPDATE ON venue_api_keys
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE venue_webhooks (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    venue_id    UUID        NOT NULL REFERENCES venues(id) ON DELETE CASCADE,
    url         TEXT        NOT NULL,
    secret      TEXT        NOT NULL,
    event_types TEXT[]      NOT NULL,
    is_active   BOOLEAN     NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT vw_url_uniq   UNIQUE (venue_id, url),
    CONSTRAINT vw_url_scheme CHECK (url ~ '^https?://'),
    CONSTRAINT vw_events_nonempty CHECK (array_length(event_types, 1) >= 1)
);

CREATE INDEX vw_venue_active_idx ON venue_webhooks (venue_id) WHERE is_active;

CREATE TRIGGER vw_set_updated_at BEFORE UPDATE ON venue_webhooks
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE IF EXISTS venue_webhooks;
DROP TABLE IF EXISTS venue_api_keys;
DROP TABLE IF EXISTS venue_working_hours;
DROP TABLE IF EXISTS venues;
