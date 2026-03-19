import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';

// Кастомные метрики
const registrationErrors = new Counter('registration_errors');
const profileCreated = new Counter('profile_created');
const profileNotReady = new Counter('profile_not_ready_yet');
const profileLatency = new Trend('profile_creation_latency_ms');

// 3000 юзеров с плавным разгоном (ramping-arrival-rate)
// ~175 + ~825 + ~1400 + ~375 ≈ 2800–3200 регистраций за 1 минуту
export const options = {
  scenarios: {
    mass_registration: {
      executor: 'ramping-arrival-rate',
      startRate: 5,
      timeUnit: '1s',
      preAllocatedVUs: 200,
      maxVUs: 300,
      stages: [
        { duration: '10s', target: 35 },   // мягкий разгон
        { duration: '15s', target: 75 },   // наращиваем темп
        { duration: '25s', target: 75 },   // держим пиковую нагрузку
        { duration: '10s', target: 5 },    // плавное снижение
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],           // <1% ошибок
    http_req_duration: ['p(95)<600'],        // 95-й перцентиль < 600мс
    registration_errors: ['count<30'],        // < 1% фейлов
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const uniqueId = `${__VU}-${__ITER}-${Date.now()}`;
  const email = `loadtest-${uniqueId}@yammi.io`;

  // 1. Регистрация
  const registerRes = http.post(
    `${BASE_URL}/api/v1/auth/register`,
    JSON.stringify({
      email: email,
      password: 'loadtest123456',
      name: `Load Test User ${uniqueId}`,
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  const registerOk = check(registerRes, {
    'register status 201': (r) => r.status === 201,
    'register has user_id': (r) => {
      try { return JSON.parse(r.body).user_id !== ''; }
      catch { return false; }
    },
  });

  if (!registerOk) {
    registrationErrors.add(1);
    console.error(`Registration failed: ${registerRes.status} ${registerRes.body}`);
    return;
  }

  const regBody = JSON.parse(registerRes.body);
  const userId = regBody.user_id;
  const accessToken = regBody.access_token;
  const authHeaders = { headers: { 'Authorization': `Bearer ${accessToken}` } };

  // 2. Ждём пока NATS доставит событие и User Service создаст профиль
  //    Пробуем несколько раз с небольшой задержкой
  const startTime = Date.now();
  let profileFound = false;

  for (let attempt = 0; attempt < 10; attempt++) {
    sleep(0.3); // 300ms между попытками

    const profileRes = http.get(`${BASE_URL}/api/v1/users/${userId}`, authHeaders);

    if (profileRes.status === 200) {
      const profile = JSON.parse(profileRes.body);

      check(profile, {
        'profile email matches': (p) => p.email === email,
        'profile has name': (p) => p.name !== '',
      });

      profileCreated.add(1);
      profileLatency.add(Date.now() - startTime);
      profileFound = true;
      break;
    }
  }

  if (!profileFound) {
    profileNotReady.add(1);
    console.warn(`Profile not created within 3s for user ${userId}`);
  }
}

export function handleSummary(data) {
  const created = data.metrics.profile_created ? data.metrics.profile_created.values.count : 0;
  const notReady = data.metrics.profile_not_ready_yet ? data.metrics.profile_not_ready_yet.values.count : 0;
  const errors = data.metrics.registration_errors ? data.metrics.registration_errors.values.count : 0;
  const total = created + notReady + errors;
  const p95 = data.metrics.profile_creation_latency_ms
    ? data.metrics.profile_creation_latency_ms.values['p(95)'].toFixed(0)
    : 'N/A';

  const summary = `
╔══════════════════════════════════════════════╗
║        РЕЗУЛЬТАТЫ НАГРУЗОЧНОГО ТЕСТА         ║
╠══════════════════════════════════════════════╣
║  Всего регистраций:        ${String(total).padStart(6)}            ║
║  Успешных регистраций:     ${String(total - errors).padStart(6)}            ║
║  Профилей создано (NATS):  ${String(created).padStart(6)}            ║
║  Профилей не дождались:    ${String(notReady).padStart(6)}            ║
║  Ошибок регистрации:       ${String(errors).padStart(6)}            ║
║  Задержка профиля p95:     ${String(p95).padStart(6)} ms        ║
╚══════════════════════════════════════════════╝
`;

  return {
    stdout: summary + '\n' + textSummary(data, { indent: '  ', enableColors: true }),
  };
}

// k6 встроенная функция для текстового summary
function textSummary(data, opts) {
  // Используем встроенный вывод k6
  return '';
}
