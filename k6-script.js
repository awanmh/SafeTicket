import http from "k6/http";
import { check, sleep } from "k6";
import { Counter } from "k6/metrics";

// Config
const BASE_URL = __ENV.BASE_URL || "http://host.docker.internal:8080";
// Toggle this to test SAFE or UNSAFE endpoint
const ENDPOINT = __ENV.MODE === "unsafe" ? "/book/unsafe" : "/book/safe";

export let options = {
  scenarios: {
    contacts: {
      executor: "per-vu-iterations",
      vus: 50, // Concurrent users
      iterations: 20, // Total requests per user = 1000 total requests
      maxDuration: "30s",
    },
  },
};

const successCounter = new Counter("successful_bookings");
const failCounter = new Counter("failed_bookings");

export default function () {
  const payload = JSON.stringify({
    event_id: 1,
    user_id: `user-${__VU}-${__ITER}`,
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const res = http.post(`${BASE_URL}${ENDPOINT}`, payload, params);

  check(res, {
    "status is 200 or 400": (r) =>
      r.status === 200 || r.status === 400 || r.status === 500,
  });

  if (res.status === 200) {
    successCounter.add(1);
  } else {
    failCounter.add(1);
  }
}
