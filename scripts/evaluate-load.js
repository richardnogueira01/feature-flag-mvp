import http from 'k6/http';
import { check } from 'k6';
import { Rate } from 'k6/metrics';

const target = Number(__ENV.TARGET_RPS || 1000000);
const baseURL = __ENV.BASE_URL || 'http://localhost:8080';
const flagKey = __ENV.FLAG_KEY || 'menu_itau_mobile';
const errors = new Rate('evaluate_errors');

export const options = {
  scenarios: {
    evaluate: {
      executor: 'constant-arrival-rate',
      rate: target,
      timeUnit: '1s',
      duration: __ENV.DURATION || '30s',
      gracefulStop: __ENV.GRACEFUL_STOP || '5s',
      preAllocatedVUs: Number(__ENV.PREALLOCATED_VUS || Math.min(target, 10000)),
      maxVUs: Number(__ENV.MAX_VUS || Math.max(target * 2, 20000)),
    },
  },
  thresholds: {
    evaluate_errors: ['rate<0.00001'],
    http_req_failed: ['rate<0.00001'],
    http_req_duration: ['p(95)<5000', 'p(99)<10000'],
  },
};

export default function () {
  const response = http.get(`${baseURL}/v1/evaluate/${encodeURIComponent(flagKey)}`, {
    tags: { endpoint: 'evaluate' },
  });
  const ok = check(response, {
    'evaluate returns 200': (r) => r.status === 200,
    'evaluate returns a value': (r) => r.status === 200 && r.body.length > 0,
  });
  errors.add(!ok);
}
