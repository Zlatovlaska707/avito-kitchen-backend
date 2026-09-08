-- +goose Up
CREATE TABLE menu_categories (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    venue_id    UUID        NOT NULL REFERENCES venues(id) ON DELETE CASCADE,
    external_id TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    sort_order  INTEGER     NOT NULL DEFAULT 0,
    is_active   BOOLEAN     NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT mc_external_uniq UNIQUE (venue_id, external_id),
    CONSTRAINT mc_name_len      CHECK (char_length(name) BETWEEN 1 AND 200),
    CONSTRAINT mc_external_len  CHECK (char_length(external_id) BETWEEN 1 AND 128)
);

CREATE INDEX mc_venue_sort_idx ON menu_categories (venue_id, sort_order);

CREATE TRIGGER mc_set_updated_at BEFORE UPDATE ON menu_categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE menu_items (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    venue_id          UUID        NOT NULL REFERENCES venues(id) ON DELETE CASCADE,
    category_id       UUID        NOT NULL REFERENCES menu_categories(id) ON DELETE RESTRICT,
    external_id       TEXT        NOT NULL,
    name              TEXT        NOT NULL,
    description       TEXT,
    price_minor       BIGINT      NOT NULL,
    weight_grams      INTEGER,
    image_url         TEXT,
    sort_order        INTEGER     NOT NULL DEFAULT 0,
    is_active         BOOLEAN     NOT NULL DEFAULT true,
    is_available      BOOLEAN     NOT NULL DEFAULT true,
    unavailable_until TIMESTAMPTZ,
    revision          INTEGER     NOT NULL DEFAULT 1,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT mi_external_uniq UNIQUE (venue_id, external_id),
    CONSTRAINT mi_name_len      CHECK (char_length(name) BETWEEN 1 AND 200),
    CONSTRAINT mi_desc_len      CHECK (description IS NULL OR char_length(description) <= 2000),
    CONSTRAINT mi_price_nonneg  CHECK (price_minor >= 0),
    CONSTRAINT mi_price_sane    CHECK (price_minor <= 10000000),  -- 100 000 ₽, защита от опечатки в интеграции
    CONSTRAINT mi_weight_pos    CHECK (weight_grams IS NULL OR weight_grams > 0)
);

CREATE INDEX mi_venue_category_sort_idx ON menu_items (venue_id, category_id, sort_order);
CREATE INDEX mi_venue_sellable_idx      ON menu_items (venue_id) WHERE is_active AND is_available;
CREATE INDEX mi_name_trgm_idx           ON menu_items USING gin (name gin_trgm_ops);
CREATE INDEX mi_unavailable_until_idx   ON menu_items (unavailable_until)
    WHERE unavailable_until IS NOT NULL;

CREATE TRIGGER mi_set_updated_at BEFORE UPDATE ON menu_items
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE menu_item_option_groups (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id     UUID        NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    external_id TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    min_choices SMALLINT    NOT NULL DEFAULT 0,
    max_choices SMALLINT    NOT NULL DEFAULT 1,
    sort_order  INTEGER     NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT miog_external_uniq UNIQUE (item_id, external_id),
    CONSTRAINT miog_min_nonneg    CHECK (min_choices >= 0),
    CONSTRAINT miog_max_ge_min    CHECK (max_choices >= min_choices),
    CONSTRAINT miog_max_limit     CHECK (max_choices BETWEEN 1 AND 20)
);

CREATE INDEX miog_item_idx ON menu_item_option_groups (item_id, sort_order);

CREATE TRIGGER miog_set_updated_at BEFORE UPDATE ON menu_item_option_groups
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE menu_item_options (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id          UUID        NOT NULL REFERENCES menu_item_option_groups(id) ON DELETE CASCADE,
    external_id       TEXT        NOT NULL,
    name              TEXT        NOT NULL,
    price_delta_minor BIGINT      NOT NULL DEFAULT 0,
    is_available      BOOLEAN     NOT NULL DEFAULT true,
    sort_order        INTEGER     NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT mio_external_uniq UNIQUE (group_id, external_id),
    CONSTRAINT mio_delta_nonneg  CHECK (price_delta_minor >= 0)
);

CREATE INDEX mio_group_idx ON menu_item_options (group_id, sort_order);

CREATE TRIGGER mio_set_updated_at BEFORE UPDATE ON menu_item_options
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE IF EXISTS menu_item_options;
DROP TABLE IF EXISTS menu_item_option_groups;
DROP TABLE IF EXISTS menu_items;
DROP TABLE IF EXISTS menu_categories;
