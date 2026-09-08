// S1 + S2: read-heavy catalog endpoints.
import { sleep } from 'k6';
import { defaultThresholds } from './config.js';
import { getVenues, getMenu } from './helpers.js';

export const options = {
  scenarios: {
    s1_venues: {
      executor: 'constant-arrival-rate',
      rate: 150,
      timeUnit: '1s',
      duration: '1m',
      preAllocatedVUs: 50,
      maxVUs: 200,
      exec: 'venues',
    },
    s2_menu: {
      executor: 'constant-arrival-rate',
      rate: 100,
      timeUnit: '1s',
      duration: '1m',
      preAllocatedVUs: 40,
      maxVUs: 150,
      exec: 'menu',
      startTime: '5s',
    },
  },
  thresholds: {
    ...defaultThresholds,
    'http_req_duration{scenario:s1_venues}': ['p(95)<80'],
    'http_req_duration{scenario:s2_menu}': ['p(95)<120'],
  },
};

export function venues() {
  getVenues();
  sleep(0.01);
}

export function menu() {
  getMenu();
  sleep(0.01);
}
