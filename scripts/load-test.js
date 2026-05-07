import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
  stages: [
    { duration: '30s', target: 100 },
    { duration: '1m', target: 200 },
    { duration: '2m', target: 200 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // Relaxed threshold for absolute stress test
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const uniqueId = uuidv4();
  const email = `u_${uniqueId}@example.com`;
  const password = 'password123';

  // 1. Registration Flow
  let regRes = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify({
    name: 'Stress Test User',
    email: email,
    password: password,
    password_confirmation: password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const regCheck = check(regRes, {
    'registration status is 201': (r) => r.status === 201,
  });

  // Only proceed if registration was successful
  if (regCheck) {
    // 2. Login Flow
    let loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
      email: email,
      password: password,
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    const loginCheck = check(loginRes, {
      'login status is 200': (r) => r.status === 200,
    });

    // 3. Authenticated Request (Profile)
    if (loginCheck) {
      const token = loginRes.json('data.access_token');
      
      let profileRes = http.get(`${BASE_URL}/api/v1/profile`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      check(profileRes, {
        'profile status is 200': (r) => r.status === 200,
      });
    }
  }

  // 4. Intentional Errors (Reduced frequency)
  if (__ITER % 50 === 0) {
    http.get(`${BASE_URL}/api/v1/non-existent-endpoint`);
  }

  // MINIMAL sleep to maximize RPS
  sleep(0.01); 
}
