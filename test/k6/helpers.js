import http from 'k6/http';
import { check } from 'k6';
import {
  BASE_URL,
  VENUE_SLUG,
  VENUE_ID,
  MENU_ITEM_ID,
  SIZE_MEDIUM,
  PARTNER_API_KEY,
  ORDER_TOTAL_MINOR,
  uniqueCustomerId,
  uniqueIdempotencyKey,
} from './config.js';

const jsonHeaders = { 'Content-Type': 'application/json' };

export function getVenues() {
  const res = http.get(`${BASE_URL}/api/v1/venues?city_id=1`);
  check(res, {
    'venues status 200': (r) => r.status === 200,
  });
  return res;
}

export function getMenu() {
  const res = http.get(`${BASE_URL}/api/v1/venues/${VENUE_SLUG}/menu`);
  check(res, {
    'menu status 200': (r) => r.status === 200,
  });
  return res;
}

export function addToCart(customerId) {
  const body = JSON.stringify({
    venue_id: VENUE_ID,
    menu_item_id: MENU_ITEM_ID,
    quantity: 1,
    option_ids: [SIZE_MEDIUM],
  });
  const res = http.post(`${BASE_URL}/api/v1/cart/items`, body, {
    headers: { ...jsonHeaders, 'X-Customer-Id': customerId },
  });
  check(res, {
    'cart add status 2xx': (r) => r.status >= 200 && r.status < 300,
  });
  return res;
}

export function getCart(customerId) {
  const res = http.get(`${BASE_URL}/api/v1/cart`, {
    headers: { 'X-Customer-Id': customerId },
  });
  check(res, {
    'cart get status 200': (r) => r.status === 200,
  });
  return res;
}

export function createOrder(customerId) {
  const addRes = addToCart(customerId);
  if (addRes.status < 200 || addRes.status >= 300) {
    return addRes;
  }

  let totalMinor = ORDER_TOTAL_MINOR;
  try {
    const payload = addRes.json();
    if (payload?.cart?.total_minor != null) {
      totalMinor = payload.cart.total_minor;
    }
  } catch (_) {
    // fallback to default total
  }

  const body = JSON.stringify({
    fulfillment_type: 'delivery',
    address: { text: 'Load test address', lat: 55.76, lon: 37.62 },
    contact_phone: '+79001234567',
    payment_method: 'cash',
    expected_total_minor: totalMinor,
  });
  const res = http.post(`${BASE_URL}/api/v1/orders`, body, {
    headers: {
      ...jsonHeaders,
      'X-Customer-Id': customerId,
      'Idempotency-Key': uniqueIdempotencyKey('k6-order'),
    },
  });
  check(res, {
    'order status 201': (r) => r.status === 201,
  });
  return res;
}

export function cartFlow() {
  const customerId = uniqueCustomerId();
  addToCart(customerId);
  getCart(customerId);
}

export function orderFlow() {
  createOrder(uniqueCustomerId());
}

export function partnerOrdersList() {
  const res = http.get(`${BASE_URL}/partner/v1/orders?status=placed&limit=20`, {
    headers: { Authorization: `Bearer ${PARTNER_API_KEY}` },
  });
  check(res, {
    'partner orders status 200': (r) => r.status === 200,
  });
  return res;
}
