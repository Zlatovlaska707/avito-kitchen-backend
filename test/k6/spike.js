// S7: spike from 20 to 300 RPS — expect graceful 429, no 5xx.
// Run WITHOUT docker-compose.loadtest.yml to exercise default rate limits.
import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE_URL, VENUE_SLUG } from './config.js';

export const options = {
  scenarios: {
    spike: {
      executor: 'ramping-arrival-rate',
      startRate: 20,
      timeUnit: '1s',
      preAllocatedVUs: 50,
      maxVUs: 400,
      stages: [
        { duration: '10s', target: 20 },
        { duration: '10s', target: 300 },
        { duration: '20s', target: 300 },
        { duration: '10s', target: 50 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    'http_req_duration{status:429}': ['p(95)<50'],
    'http_req_duration{status:200}': ['p(95)<500'],
    checks: ['rate>0.99'],
  },
};

function readOk(res) {
  check(res, {
    'no 5xx': (r) => r.status < 500,
    '200 or 429': (r) => r.status === 200 || r.status === 429,
  });
}

export default function () {
  const path = Math.random() < 0.6
    ? `${BASE_URL}/api/v1/venues?city_id=1`
    : `${BASE_URL}/api/v1/venues/${VENUE_SLUG}/menu`;
  readOk(http.get(path));
  sleep(0.01);
}
