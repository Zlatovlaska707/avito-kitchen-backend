-- Avito Kitchen demo seed (idempotent).
-- API key: ak_demopizz_xK9mN2pQ7vR4wL8hT3jF6sD1aB0cE5gY9uI2oP4 (SEED_PARTNER_API_KEY)
-- SHA-256: 03204a224f24a349d13f98156843d0092a6d40e018ceb5bffc3ff671534aa36d
-- key_prefix: demopizz

-- ---------------------------------------------------------------------------
-- Venues
-- ---------------------------------------------------------------------------
INSERT INTO venues (
    id, slug, name, description, status, accepts_orders,
    city_id, address_text, lat, lon, timezone,
    avg_prep_minutes, delivery_fee_minor, min_order_minor, delivery_radius_m, rating
) VALUES
    (
        '11111111-1111-4111-8111-111111111101',
        'pizza-mario',
        'Mario Pizza',
        '24/7 pizzeria for demo and E2E tests',
        'active',
        true,
        1,
        'Moscow, Tverskaya st. 1',
        55.755800,
        37.617600,
        'Europe/Moscow',
        25,
        19900,
        50000,
        5000,
        4.8
    ),
    (
        '11111111-1111-4111-8111-111111111102',
        'sushi-tokyo',
        'Sushi Tokyo',
        'Sushi bar with schedule 10:00-23:00 and Fri/Sat overnight interval',
        'active',
        true,
        1,
        'Moscow, Arbat st. 10',
        55.752000,
        37.592000,
        'Europe/Moscow',
        30,
        24900,
        80000,
        4000,
        4.5
    ),
    (
        '11111111-1111-4111-8111-111111111103',
        'burger-closed',
        'Burger Closed',
        'Active venue with schedule closed during normal hours (venue_closed scenario)',
        'active',
        true,
        1,
        'Moscow, Myasnitskaya st. 5',
        55.748000,
        37.605000,
        'Europe/Moscow',
        20,
        14900,
        40000,
        3000,
        4.0
    )
ON CONFLICT (slug) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Working hours
-- ---------------------------------------------------------------------------

-- pizza-mario: 24/7 (00:00-23:59 every day)
INSERT INTO venue_working_hours (venue_id, weekday, opens_at, closes_at, crosses_midnight)
SELECT v.id, d.weekday, '00:00'::TIME, '23:59'::TIME, false
FROM venues v
CROSS JOIN generate_series(1, 7) AS d(weekday)
WHERE v.slug = 'pizza-mario'
ON CONFLICT (venue_id, weekday, opens_at) DO NOTHING;

-- sushi-tokyo: 10:00-23:00 (Mon-Thu, Sun)
INSERT INTO venue_working_hours (venue_id, weekday, opens_at, closes_at, crosses_midnight)
SELECT v.id, d.weekday, '10:00'::TIME, '23:00'::TIME, false
FROM venues v
CROSS JOIN unnest(ARRAY[1, 2, 3, 4, 7]::SMALLINT[]) AS d(weekday)
WHERE v.slug = 'sushi-tokyo'
ON CONFLICT (venue_id, weekday, opens_at) DO NOTHING;

-- sushi-tokyo: Fri/Sat overnight shift (10:00-02:00)
INSERT INTO venue_working_hours (venue_id, weekday, opens_at, closes_at, crosses_midnight)
SELECT v.id, d.weekday, '10:00'::TIME, '02:00'::TIME, true
FROM venues v
CROSS JOIN unnest(ARRAY[5, 6]::SMALLINT[]) AS d(weekday)
WHERE v.slug = 'sushi-tokyo'
ON CONFLICT (venue_id, weekday, opens_at) DO NOTHING;

-- burger-closed: narrow window 03:00-05:00 (closed during normal hours)
INSERT INTO venue_working_hours (venue_id, weekday, opens_at, closes_at, crosses_midnight)
SELECT v.id, d.weekday, '03:00'::TIME, '05:00'::TIME, false
FROM venues v
CROSS JOIN generate_series(1, 7) AS d(weekday)
WHERE v.slug = 'burger-closed'
ON CONFLICT (venue_id, weekday, opens_at) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Partner API key (pizza-mario)
-- ---------------------------------------------------------------------------
INSERT INTO venue_api_keys (id, venue_id, key_prefix, key_hash, scopes, is_active)
SELECT
    '44444444-4444-4444-8444-444444444401',
    v.id,
    'demopizz',
    '03204a224f24a349d13f98156843d0092a6d40e018ceb5bffc3ff671534aa36d',
    '{menu:write,orders:write}',
    true
FROM venues v
WHERE v.slug = 'pizza-mario'
ON CONFLICT (key_prefix) DO NOTHING;

-- ---------------------------------------------------------------------------
-- burger-closed menu (E2E venue_closed scenario)
-- ---------------------------------------------------------------------------
INSERT INTO menu_categories (id, venue_id, external_id, name, sort_order, is_active)
SELECT '22222222-2222-4222-8222-222222222301'::UUID, v.id, 'cat-burgers', 'Burgers', 1, true
FROM venues v
WHERE v.slug = 'burger-closed'
ON CONFLICT (venue_id, external_id) DO NOTHING;

INSERT INTO menu_items (
    id, venue_id, category_id, external_id, name, description,
    price_minor, weight_grams, sort_order, is_active, is_available
)
SELECT
    '33333333-3333-4333-8333-333333333401'::UUID,
    v.id,
    mc.id,
    'item-classic-burger',
    'Classic Burger',
    'Demo burger for venue_closed E2E',
    45000,
    300,
    1,
    true,
    true
FROM venues v
JOIN menu_categories mc ON mc.venue_id = v.id AND mc.external_id = 'cat-burgers'
WHERE v.slug = 'burger-closed'
ON CONFLICT (venue_id, external_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Demo customer (README examples: X-Customer-Id)
-- ---------------------------------------------------------------------------
INSERT INTO customers (id, phone, name)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    '+79001234567',
    'Demo Customer'
)
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- k6 load test customer pool (1000 ids; see test/k6/config.js)
-- ---------------------------------------------------------------------------
INSERT INTO customers (id)
SELECT ('aaaaaaaa-aaaa-4aaa-8aaa-' || lpad(gs::text, 12, '0'))::uuid
FROM generate_series(1, 1000) gs
ON CONFLICT (id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- pizza-mario menu: 3 categories, 12 items
-- ---------------------------------------------------------------------------
INSERT INTO menu_categories (id, venue_id, external_id, name, sort_order, is_active)
SELECT c.id, v.id, c.external_id, c.name, c.sort_order, true
FROM venues v
CROSS JOIN (VALUES
    ('22222222-2222-4222-8222-222222222201'::UUID, 'cat-pizzas',   'Pizzas',   1),
    ('22222222-2222-4222-8222-222222222202'::UUID, 'cat-drinks',   'Drinks',   2),
    ('22222222-2222-4222-8222-222222222203'::UUID, 'cat-desserts', 'Desserts', 3)
) AS c(id, external_id, name, sort_order)
WHERE v.slug = 'pizza-mario'
ON CONFLICT (venue_id, external_id) DO NOTHING;

INSERT INTO menu_items (
    id, venue_id, category_id, external_id, name, description,
    price_minor, weight_grams, sort_order, is_active, is_available
)
SELECT
    i.id,
    v.id,
    mc.id,
    i.external_id,
    i.name,
    i.description,
    i.price_minor,
    i.weight_grams,
    i.sort_order,
    true,
    i.is_available
FROM venues v
CROSS JOIN (VALUES
    ('33333333-3333-4333-8333-333333333301'::UUID, 'cat-pizzas', 'item-margherita',  'Margherita',      'Tomato, mozzarella, basil',              49000, 450, 1, true),
    ('33333333-3333-4333-8333-333333333302'::UUID, 'cat-pizzas', 'item-pepperoni',   'Pepperoni',       'Tomato, mozzarella, pepperoni',            59000, 480, 2, true),
    ('33333333-3333-4333-8333-333333333303'::UUID, 'cat-pizzas', 'item-four-cheese', 'Four Cheese',     'Mozzarella, gorgonzola, parmesan, brie', 65000, 470, 3, true),
    ('33333333-3333-4333-8333-333333333304'::UUID, 'cat-pizzas', 'item-truffle',     'Truffle',         'Truffle cream, mushrooms, parmesan',     89000, 460, 4, false),
    ('33333333-3333-4333-8333-333333333305'::UUID, 'cat-drinks', 'item-cola',        'Cola 0.5L',       'Carbonated drink',                       15000, 500, 1, true),
    ('33333333-3333-4333-8333-333333333306'::UUID, 'cat-drinks', 'item-lemonade',    'Lemonade 0.5L',   'Fresh lemonade',                         18000, 500, 2, true),
    ('33333333-3333-4333-8333-333333333307'::UUID, 'cat-drinks', 'item-water',       'Water 0.5L',      'Still water',                            10000, 500, 3, true),
    ('33333333-3333-4333-8333-333333333308'::UUID, 'cat-drinks', 'item-juice',       'Juice 0.3L',      'Orange juice',                           22000, 300, 4, true),
    ('33333333-3333-4333-8333-333333333309'::UUID, 'cat-desserts', 'item-tiramisu',  'Tiramisu',        'Classic Italian dessert',                32000, 150, 1, true),
    ('33333333-3333-4333-8333-333333333310'::UUID, 'cat-desserts', 'item-cheesecake','Cheesecake',      'NY-style cheesecake',                    35000, 160, 2, true),
    ('33333333-3333-4333-8333-333333333311'::UUID, 'cat-desserts', 'item-brownie',   'Brownie',         'Chocolate brownie with nuts',            28000, 120, 3, true),
    ('33333333-3333-4333-8333-333333333312'::UUID, 'cat-desserts', 'item-ice-cream', 'Ice Cream',       'Vanilla ice cream, 2 scoops',            25000, 100, 4, true)
) AS i(id, cat_external_id, external_id, name, description, price_minor, weight_grams, sort_order, is_available)
JOIN menu_categories mc ON mc.venue_id = v.id AND mc.external_id = i.cat_external_id
WHERE v.slug = 'pizza-mario'
ON CONFLICT (venue_id, external_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Modifiers on 3 pizza items: size (required) + extras (optional)
-- ---------------------------------------------------------------------------
INSERT INTO menu_item_option_groups (
    id, item_id, external_id, name, min_choices, max_choices, sort_order
)
SELECT
    g.id,
    mi.id,
    g.external_id,
    g.name,
    g.min_choices,
    g.max_choices,
    g.sort_order
FROM menu_items mi
JOIN venues v ON v.id = mi.venue_id AND v.slug = 'pizza-mario'
CROSS JOIN (VALUES
    ('55555555-5555-4555-8555-555555555301'::UUID, 'item-margherita',  'grp-size',   'Size',   1, 1, 1),
    ('55555555-5555-4555-8555-555555555302'::UUID, 'item-margherita',  'grp-extras', 'Extras', 0, 5, 2),
    ('55555555-5555-4555-8555-555555555303'::UUID, 'item-pepperoni',   'grp-size',   'Size',   1, 1, 1),
    ('55555555-5555-4555-8555-555555555304'::UUID, 'item-pepperoni',   'grp-extras', 'Extras', 0, 5, 2),
    ('55555555-5555-4555-8555-555555555305'::UUID, 'item-four-cheese', 'grp-size',   'Size',   1, 1, 1),
    ('55555555-5555-4555-8555-555555555306'::UUID, 'item-four-cheese', 'grp-extras', 'Extras', 0, 5, 2)
) AS g(id, item_external_id, external_id, name, min_choices, max_choices, sort_order)
WHERE mi.external_id = g.item_external_id
ON CONFLICT (item_id, external_id) DO NOTHING;

INSERT INTO menu_item_options (
    id, group_id, external_id, name, price_delta_minor, is_available, sort_order
)
SELECT
    o.id,
    g.id,
    o.external_id,
    o.name,
    o.price_delta_minor,
    true,
    o.sort_order
FROM menu_item_option_groups g
JOIN menu_items mi ON mi.id = g.item_id
JOIN venues v ON v.id = mi.venue_id AND v.slug = 'pizza-mario'
CROSS JOIN (VALUES
    ('66666666-6666-4666-8666-666666666301'::UUID, 'item-margherita', 'grp-size', 'opt-small',  'Small (25 cm)', 0,     1),
    ('66666666-6666-4666-8666-666666666302'::UUID, 'item-margherita', 'grp-size', 'opt-medium', 'Medium (30 cm)', 10000, 2),
    ('66666666-6666-4666-8666-666666666303'::UUID, 'item-margherita', 'grp-size', 'opt-large',  'Large (35 cm)',  20000, 3),
    ('66666666-6666-4666-8666-666666666304'::UUID, 'item-margherita', 'grp-extras', 'opt-extra-cheese', 'Extra cheese', 8000,  1),
    ('66666666-6666-4666-8666-666666666305'::UUID, 'item-margherita', 'grp-extras', 'opt-mushrooms',    'Mushrooms',    6000,  2),
    ('66666666-6666-4666-8666-666666666306'::UUID, 'item-margherita', 'grp-extras', 'opt-olives',       'Olives',       5000,  3),
    ('66666666-6666-4666-8666-666666666307'::UUID, 'item-margherita', 'grp-extras', 'opt-jalapeno',     'Jalapeno',     4000,  4),
    ('66666666-6666-4666-8666-666666666308'::UUID, 'item-margherita', 'grp-extras', 'opt-bacon',        'Bacon',        9000,  5),
    ('66666666-6666-4666-8666-666666666309'::UUID, 'item-pepperoni', 'grp-size', 'opt-small',  'Small (25 cm)', 0,     1),
    ('66666666-6666-4666-8666-666666666310'::UUID, 'item-pepperoni', 'grp-size', 'opt-medium', 'Medium (30 cm)', 10000, 2),
    ('66666666-6666-4666-8666-666666666311'::UUID, 'item-pepperoni', 'grp-size', 'opt-large',  'Large (35 cm)',  20000, 3),
    ('66666666-6666-4666-8666-666666666312'::UUID, 'item-pepperoni', 'grp-extras', 'opt-extra-cheese', 'Extra cheese', 8000,  1),
    ('66666666-6666-4666-8666-666666666313'::UUID, 'item-pepperoni', 'grp-extras', 'opt-mushrooms',    'Mushrooms',    6000,  2),
    ('66666666-6666-4666-8666-666666666314'::UUID, 'item-pepperoni', 'grp-extras', 'opt-olives',       'Olives',       5000,  3),
    ('66666666-6666-4666-8666-666666666315'::UUID, 'item-pepperoni', 'grp-extras', 'opt-jalapeno',     'Jalapeno',     4000,  4),
    ('66666666-6666-4666-8666-666666666316'::UUID, 'item-pepperoni', 'grp-extras', 'opt-bacon',        'Bacon',        9000,  5),
    ('66666666-6666-4666-8666-666666666317'::UUID, 'item-four-cheese', 'grp-size', 'opt-small',  'Small (25 cm)', 0,     1),
    ('66666666-6666-4666-8666-666666666318'::UUID, 'item-four-cheese', 'grp-size', 'opt-medium', 'Medium (30 cm)', 10000, 2),
    ('66666666-6666-4666-8666-666666666319'::UUID, 'item-four-cheese', 'grp-size', 'opt-large',  'Large (35 cm)',  20000, 3),
    ('66666666-6666-4666-8666-666666666320'::UUID, 'item-four-cheese', 'grp-extras', 'opt-extra-cheese', 'Extra cheese', 8000,  1),
    ('66666666-6666-4666-8666-666666666321'::UUID, 'item-four-cheese', 'grp-extras', 'opt-mushrooms',    'Mushrooms',    6000,  2),
    ('66666666-6666-4666-8666-666666666322'::UUID, 'item-four-cheese', 'grp-extras', 'opt-olives',       'Olives',       5000,  3),
    ('66666666-6666-4666-8666-666666666323'::UUID, 'item-four-cheese', 'grp-extras', 'opt-jalapeno',     'Jalapeno',     4000,  4),
    ('66666666-6666-4666-8666-666666666324'::UUID, 'item-four-cheese', 'grp-extras', 'opt-bacon',        'Bacon',        9000,  5)
) AS o(id, item_external_id, group_external_id, external_id, name, price_delta_minor, sort_order)
WHERE mi.external_id = o.item_external_id
  AND g.external_id = o.group_external_id
ON CONFLICT (group_id, external_id) DO NOTHING;
