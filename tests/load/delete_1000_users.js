import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';
import exec from 'k6/execution';

// Метрики
const delErrors = new Counter('del_errors');
const delSuccess = new Counter('del_success');
const profileGone = new Counter('profile_confirmed_gone');
const profileStillExists = new Counter('profile_still_exists');
const authGone = new Counter('auth_confirmed_gone');
const delLatency = new Trend('del_latency_ms');
const profileDelLatency = new Trend('profile_del_latency_ms');
const errorRate = new Rate('error_rate');

// 1000 удалений за ~1 минуту с плавным разгоном
// ~75 + ~350 + ~500 + ~150 ≈ 1075 итераций
export const options = {
  scenarios: {
    delete_users: {
      executor: 'ramping-arrival-rate',
      startRate: 5,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 200,
      stages: [
        { duration: '10s', target: 10 },   // мягкий разгон
        { duration: '20s', target: 25 },   // наращиваем
        { duration: '20s', target: 25 },   // держим пик
        { duration: '10s', target: 5 },    // плавное снижение
      ],
    },
  },
  thresholds: {
    'http_req_failed{type:delete}': ['rate<0.05'],  // только delete-запросы
    'http_req_duration{type:delete}': ['p(95)<3000'],
    error_rate: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TOTAL_USERS = 1000;

// === SETUP: регистрируем 1000 юзеров для последующего удаления ===
export function setup() {
  console.log(`=== SETUP: Регистрация ${TOTAL_USERS} пользователей ===`);

  const users = [];
  const ts = Date.now();

  for (let i = 0; i < TOTAL_USERS; i++) {
    const email = `deltest-${i}-${ts}@yammi.io`;

    const res = http.post(
      `${BASE_URL}/api/v1/auth/register`,
      JSON.stringify({
        email: email,
        password: 'deltest12345',
        name: `DelTest User ${i}`,
      }),
      {
        headers: { 'Content-Type': 'application/json' },
        timeout: '30s',
      }
    );

    if (res.status === 201) {
      const body = JSON.parse(res.body);
      users.push({ user_id: body.user_id, email: email, access_token: body.access_token });
    } else {
      console.error(`Setup: register ${i} failed: ${res.status} ${res.body}`);
    }

    if (i > 0 && i % 100 === 0) {
      console.log(`Setup: зарегистрировано ${i}/${TOTAL_USERS}`);
      sleep(0.3);
    }
  }

  // Ждём NATS — профили должны создаться в User Service
  console.log(`Setup: зарегистрировано ${users.length}, ждём создания профилей (10s)...`);
  sleep(10);

  // Верифицируем что профили созданы
  let confirmed = 0;
  for (let i = 0; i < Math.min(users.length, 10); i++) {
    const r = http.get(`${BASE_URL}/api/v1/users/${users[i].user_id}`, {
      headers: { 'Authorization': `Bearer ${users[i].access_token}` },
    });
    if (r.status === 200) confirmed++;
  }
  console.log(`Setup: профили-сэмпл ${confirmed}/10 подтверждены`);

  return { users };
}

// === ОСНОВНОЙ СЦЕНАРИЙ: удаление + верификация ===
export default function (data) {
  const idx = exec.scenario.iterationInTest;

  if (idx >= data.users.length) {
    return; // все юзеры уже обработаны
  }

  const user = data.users[idx];
  const authHeader = { 'Authorization': `Bearer ${user.access_token}` };

  // === 1. DELETE /api/v1/users/{id} ===
  const t0 = Date.now();
  const res = http.del(
    `${BASE_URL}/api/v1/users/${user.user_id}`,
    null,
    {
      headers: authHeader,
      tags: { type: 'delete' },
      timeout: '30s',
    }
  );
  delLatency.add(Date.now() - t0);

  const ok = check(res, {
    'delete status 200': (r) => r.status === 200,
    'delete body status=deleted': (r) => {
      try { return JSON.parse(r.body).status === 'deleted'; }
      catch { return false; }
    },
  });

  if (!ok) {
    delErrors.add(1);
    errorRate.add(1);

    const status = res.status;
    const body = res.body ? res.body.substring(0, 200) : 'empty';
    if (status === 0) {
      console.error(`[VU${__VU}] CONN_ERR: ${res.error}`);
    } else if (status === 404) {
      console.error(`[VU${__VU}] NOT_FOUND: ${user.user_id}`);
    } else if (status >= 500) {
      console.error(`[VU${__VU}] SERVER_${status}: ${body}`);
    } else {
      console.error(`[VU${__VU}] HTTP_${status}: ${body}`);
    }
    sleep(0.5);
    return;
  }

  delSuccess.add(1);
  errorRate.add(0);

  // === 2. ВЕРИФИКАЦИЯ: Auth Service — повторный DELETE должен вернуть 404 ===
  // JWT stateless — токен ещё валиден после удаления, OwnerOnly пропустит
  sleep(0.5);
  const authCheck = http.del(
    `${BASE_URL}/api/v1/users/${user.user_id}`,
    null,
    { headers: authHeader, tags: { type: 'verify_auth' } }
  );

  check(authCheck, {
    'auth: re-delete returns 404': (r) => r.status === 404,
  });
  if (authCheck.status === 404) {
    authGone.add(1);
  }

  // === 3. ВЕРИФИКАЦИЯ: User Service — профиль должен исчезнуть (NATS event) ===
  sleep(3); // даём NATS + consumer обработать событие

  const profileCheck = http.get(`${BASE_URL}/api/v1/users/${user.user_id}`, {
    headers: authHeader,
    tags: { type: 'verify_profile' },
  });

  if (profileCheck.status === 404) {
    profileGone.add(1);
    profileDelLatency.add(Date.now() - t0 - 3500); // минус sleeps
  } else {
    // Вторая попытка через 5 секунд
    sleep(5);
    const retry = http.get(`${BASE_URL}/api/v1/users/${user.user_id}`, {
      headers: authHeader,
      tags: { type: 'verify_profile' },
    });
    if (retry.status === 404) {
      profileGone.add(1);
      profileDelLatency.add(Date.now() - t0 - 8500);
    } else {
      profileStillExists.add(1);
      console.warn(`[VU${__VU}] PROFILE STILL EXISTS: ${user.user_id} (status ${retry.status})`);
    }
  }

  sleep(Math.random() * 0.5);
}

export function handleSummary(data) {
  const success = data.metrics.del_success ? data.metrics.del_success.values.count : 0;
  const errors = data.metrics.del_errors ? data.metrics.del_errors.values.count : 0;
  const total = success + errors;
  const authConfirmed = data.metrics.auth_confirmed_gone ? data.metrics.auth_confirmed_gone.values.count : 0;
  const profGone = data.metrics.profile_confirmed_gone ? data.metrics.profile_confirmed_gone.values.count : 0;
  const profExists = data.metrics.profile_still_exists ? data.metrics.profile_still_exists.values.count : 0;

  const safe = (metric, key) => {
    if (!metric || !metric.values) return '-';
    // k6 может использовать разные форматы ключей для перцентилей
    const val = metric.values[key];
    if (val == null) return '-';
    return val.toFixed(0);
  };

  const delMed = safe(data.metrics.del_latency_ms, 'med');
  const delP90 = safe(data.metrics.del_latency_ms, 'p(90)');
  const delP95 = safe(data.metrics.del_latency_ms, 'p(95)');
  const delMax = safe(data.metrics.del_latency_ms, 'max');

  const profMed = safe(data.metrics.profile_del_latency_ms, 'med');
  const profP95 = safe(data.metrics.profile_del_latency_ms, 'p(95)');

  const httpReqs = data.metrics.http_reqs ? data.metrics.http_reqs.values.count : 0;
  const httpFailed = data.metrics.http_req_failed
    ? (data.metrics.http_req_failed.values.rate * 100).toFixed(2) : '0';

  const summary = `
╔════════════════════════════════════════════════════════════╗
║          DELETE TEST: 1000 USERS / 1 MINUTE                ║
╠════════════════════════════════════════════════════════════╣
║                                                            ║
║  УДАЛЕНИЕ (Gateway → Auth → DB + NATS publish)             ║
║  ──────────────────────────────────────────────            ║
║  Попыток:              ${String(total).padStart(7)}                           ║
║  Успешных (200):       ${String(success).padStart(7)}                           ║
║  Ошибок:               ${String(errors).padStart(7)}  (${(errors/Math.max(total,1)*100).toFixed(1)}%)                  ║
║  Latency med:        ${delMed.padStart(7)} ms                          ║
║  Latency p90:        ${delP90.padStart(7)} ms                          ║
║  Latency p95:        ${delP95.padStart(7)} ms                          ║
║  Latency max:        ${delMax.padStart(7)} ms                          ║
║                                                            ║
║  ВЕРИФИКАЦИЯ: AUTH SERVICE                                 ║
║  ──────────────────────────────────────────────            ║
║  Re-delete → 404:     ${String(authConfirmed).padStart(7)}  / ${success}                    ║
║                                                            ║
║  ВЕРИФИКАЦИЯ: USER SERVICE (NATS event → profile deleted)  ║
║  ──────────────────────────────────────────────            ║
║  Профилей удалено:    ${String(profGone).padStart(7)}  / ${success}                    ║
║  Профиль остался:     ${String(profExists).padStart(7)}                           ║
║  Profile del med:   ${profMed.padStart(7)} ms                          ║
║  Profile del p95:   ${profP95.padStart(7)} ms                          ║
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
