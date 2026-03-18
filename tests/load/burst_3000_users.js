import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';

// Метрики
const regErrors = new Counter('reg_errors');
const regSuccess = new Counter('reg_success');
const profileCreated = new Counter('profile_created');
const profileNotReady = new Counter('profile_not_ready');
const profileLatency = new Trend('profile_latency_ms');
const regLatency = new Trend('reg_latency_ms');
const errorRate = new Rate('error_rate');

// 3000 уникальных пользователей за 1 минуту
export const options = {
  scenarios: {
    burst_registration: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 500 },
        { duration: '20s', target: 1500 },
        { duration: '20s', target: 3000 },
        { duration: '10s', target: 0 },
      ],
      gracefulRampDown: '10s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    'http_req_duration{type:register}': ['p(95)<5000'],
    error_rate: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const uid = `${__VU}-${__ITER}-${Date.now()}`;
  const email = `burst-${uid}@yammi.io`;

  // === РЕГИСТРАЦИЯ ===
  const t0 = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/auth/register`,
    JSON.stringify({
      email: email,
      password: 'bursttest12345',
      name: `Burst User ${uid}`,
    }),
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { type: 'register' },
      timeout: '30s',
    }
  );
  regLatency.add(Date.now() - t0);

  const ok = check(res, {
    'status 201': (r) => r.status === 201,
  });

  if (!ok) {
    regErrors.add(1);
    errorRate.add(1);

    const status = res.status;
    const body = res.body ? res.body.substring(0, 200) : 'empty';
    if (status === 0) {
      console.error(`[VU${__VU}] CONN_ERR: ${res.error}`);
    } else if (status === 409) {
      console.error(`[VU${__VU}] DUPLICATE`);
    } else if (status >= 500) {
      console.error(`[VU${__VU}] SERVER_${status}: ${body}`);
    } else {
      console.error(`[VU${__VU}] HTTP_${status}: ${body}`);
    }
    sleep(0.5);
    return;
  }

  regSuccess.add(1);
  errorRate.add(0);

  const userId = JSON.parse(res.body).user_id;

  // === ПРОВЕРКА ПРОФИЛЯ ===
  // Одна попытка после паузы (не DDoS-polling)
  // Даём NATS + consumer время обработать
  sleep(3);

  const profileRes = http.get(`${BASE_URL}/api/v1/users/${userId}`, {
    tags: { type: 'get_profile' },
  });

  if (profileRes.status === 200) {
    profileCreated.add(1);
    profileLatency.add(Date.now() - t0 - 3000); // минус sleep
  } else {
    // Вторая попытка через 5 секунд
    sleep(5);
    const retry = http.get(`${BASE_URL}/api/v1/users/${userId}`, {
      tags: { type: 'get_profile' },
    });
    if (retry.status === 200) {
      profileCreated.add(1);
      profileLatency.add(Date.now() - t0 - 8000);
    } else {
      profileNotReady.add(1);
    }
  }

  // Пауза (реальный пользователь не шлёт запросы в цикле)
  sleep(Math.random() * 1);
}

export function handleSummary(data) {
  const success = data.metrics.reg_success ? data.metrics.reg_success.values.count : 0;
  const errors = data.metrics.reg_errors ? data.metrics.reg_errors.values.count : 0;
  const total = success + errors;
  const profiles = data.metrics.profile_created ? data.metrics.profile_created.values.count : 0;
  const notReady = data.metrics.profile_not_ready ? data.metrics.profile_not_ready.values.count : 0;

  const regP50 = data.metrics.reg_latency_ms ? data.metrics.reg_latency_ms.values['p(50)'].toFixed(0) : '-';
  const regP95 = data.metrics.reg_latency_ms ? data.metrics.reg_latency_ms.values['p(95)'].toFixed(0) : '-';
  const regP99 = data.metrics.reg_latency_ms ? data.metrics.reg_latency_ms.values['p(99)'].toFixed(0) : '-';
  const regMax = data.metrics.reg_latency_ms ? data.metrics.reg_latency_ms.values['max'].toFixed(0) : '-';

  const profP50 = data.metrics.profile_latency_ms && data.metrics.profile_latency_ms.values['p(50)'] != null
    ? data.metrics.profile_latency_ms.values['p(50)'].toFixed(0) : '-';
  const profP95 = data.metrics.profile_latency_ms && data.metrics.profile_latency_ms.values['p(95)'] != null
    ? data.metrics.profile_latency_ms.values['p(95)'].toFixed(0) : '-';

  const httpReqs = data.metrics.http_reqs ? data.metrics.http_reqs.values.count : 0;
  const httpFailed = data.metrics.http_req_failed ? (data.metrics.http_req_failed.values.rate * 100).toFixed(2) : '0';

  const summary = `
╔════════════════════════════════════════════════════════════╗
║            BURST TEST: 3000 VUs / 1 MINUTE                 ║
╠════════════════════════════════════════════════════════════╣
║                                                            ║
║  РЕГИСТРАЦИЯ (Auth → DB + bcrypt + NATS publish)           ║
║  ──────────────────────────────────────────────            ║
║  Попыток:              ${String(total).padStart(7)}                           ║
║  Успешных (201):       ${String(success).padStart(7)}                           ║
║  Ошибок:               ${String(errors).padStart(7)}  (${(errors/Math.max(total,1)*100).toFixed(1)}%)                  ║
║  Latency p50:        ${regP50.padStart(7)} ms                          ║
║  Latency p95:        ${regP95.padStart(7)} ms                          ║
║  Latency p99:        ${regP99.padStart(7)} ms                          ║
║  Latency max:        ${regMax.padStart(7)} ms                          ║
║                                                            ║
║  NATS EVENT → PROFILE                                      ║
║  ──────────────────────────────────────────────            ║
║  Профилей создано:     ${String(profiles).padStart(7)}  / ${success}                    ║
║  Не дождались:         ${String(notReady).padStart(7)}                           ║
║  Profile latency p50:${profP50.padStart(7)} ms                          ║
║  Profile latency p95:${profP95.padStart(7)} ms                          ║
║                                                            ║
║  HTTP OVERVIEW                                             ║
║  ──────────────────────────────────────────────            ║
║  Total requests:       ${String(httpReqs).padStart(7)}                           ║
║  Failed rate:          ${httpFailed.padStart(6)}%                          ║
║                                                            ║
╚════════════════════════════════════════════════════════════╝
`;

  return { stdout: summary };
}
