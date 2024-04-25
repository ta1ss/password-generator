import http from "k6/http";
import { check, sleep } from "k6";

// Test configuration
export const options = {
  discardResponseBodies: true,
  thresholds: {
    http_req_failed: ['rate<0.01'], // http errors should be less than 1%
    http_req_duration: ['p(95)<200'], // 95% of requests should be below 200ms
  },
  scenarios: {
    load_typical: {
      executor: 'constant-arrival-rate',
      duration: '30s',
      preAllocatedVUs: 1000,
      maxVUs: 10000,

      rate: 1000, // 1000 RPS
      timeUnit: '1s',
    },
    load_peak: {
      executor: 'constant-arrival-rate',
      duration: '30s',
      preAllocatedVUs: 1000,
      maxVUs: 10000,
      
      rate: 20000, // 20000 RPS
      timeUnit: '1s',
    },
  },
};

// Simulated user behavior
export default function () {
  let res = http.get("http://localhost:8080");
  // Validate response status
  check(res, { "status was 200": (r) => r.status == 200 });
  sleep(1);
}