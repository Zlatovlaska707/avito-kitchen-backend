// S6: mixed realistic traffic (85% reads, 10% cart, 5% orders).
import { sleep } from 'k6';
import { defaultThresholds } from './config.js';
import { getVenues, getMenu, cartFlow, orderFlow, partnerOrdersList } from './helpers.js';

export const options = {
  scenarios: {
    mixed: {
      executor: 'ramping-arrival-rate',
      startRate: 20,
      timeUnit: '1s',
      preAllocatedVUs: 80,
      maxVUs: 300,
      stages: [
        { duration: '30s', target: 100 },
        { duration: '1m', target: 150 },
        { duration: '30s', target: 100 },
      ],
    },
    partner_poll: {
      executor: 'constant-arrival-rate',
      rate: 20,
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 5,
      maxVUs: 20,
      exec: 'partnerPoll',
    },
  },
  thresholds: defaultThresholds,
};

export default function () {
  const roll = Math.random();
  if (roll < 0.425) {
    getVenues();
  } else if (roll < 0.85) {
    getMenu();
  } else if (roll < 0.95) {
    cartFlow();
  } else {
    orderFlow();
  }
  sleep(0.02);
}

export function partnerPoll() {
  partnerOrdersList();
  sleep(0.2);
}
