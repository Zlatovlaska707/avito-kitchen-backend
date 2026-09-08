// Shared constants for Avito Kitchen k6 scripts (seed data).
export const BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8080';

export const VENUE_SLUG = 'pizza-mario';
export const VENUE_ID = '11111111-1111-4111-8111-111111111101';
export const MENU_ITEM_ID = '33333333-3333-4333-8333-333333333301';
export const SIZE_MEDIUM = '66666666-6666-4666-8666-666666666302';
export const PARTNER_API_KEY = __ENV.PARTNER_API_KEY || 'ak_demopizz_xK9mN2pQ7vR4wL8hT3jF6sD1aB0cE5gY9uI2oP4';

// 1× Margherita Medium + delivery fee (seed: 49000 + 10000 + 19900).
export const ORDER_TOTAL_MINOR = 78900;

export const defaultThresholds = {
  http_req_failed: ['rate<0.01'],
  http_req_duration: ['p(95)<250'],
};

export const LOADTEST_CUSTOMER_POOL = 1000;

export function uniqueCustomerId() {
  // Spread VU/iter across pool to reduce concurrent collisions under load.
  const idx = (((__VU * 1000003 + __ITER * 7919) % LOADTEST_CUSTOMER_POOL) + 1);
  return `aaaaaaaa-aaaa-4aaa-8aaa-${String(idx).padStart(12, '0')}`;
}

export function uniqueIdempotencyKey(prefix) {
  return `${prefix}-${__VU}-${__ITER}-${Date.now()}`;
}
