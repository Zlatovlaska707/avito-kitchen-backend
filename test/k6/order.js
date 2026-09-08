// S3 + S4: cart mutations and order creation.
import { sleep } from 'k6';
import { defaultThresholds } from './config.js';
import { cartFlow, orderFlow } from './helpers.js';

export const options = {
  scenarios: {
    s3_cart: {
      executor: 'constant-arrival-rate',
      rate: 30,
      timeUnit: '1s',
      duration: '1m',
      preAllocatedVUs: 20,
      maxVUs: 80,
      exec: 'cart',
    },
    s4_orders: {
      executor: 'constant-arrival-rate',
      rate: 10,
      timeUnit: '1s',
      duration: '1m',
      preAllocatedVUs: 10,
      maxVUs: 40,
      exec: 'orders',
      startTime: '5s',
    },
  },
  thresholds: {
    ...defaultThresholds,
    'http_req_duration{scenario:s3_cart}': ['p(95)<100'],
    'http_req_duration{scenario:s4_orders}': ['p(95)<250'],
  },
};

export function cart() {
  cartFlow();
  sleep(0.05);
}

export function orders() {
  orderFlow();
  sleep(0.1);
}
