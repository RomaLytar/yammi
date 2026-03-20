import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// Custom metrics
const boardCreationDuration = new Trend('board_creation_duration');
const boardCreationSuccessRate = new Rate('board_creation_success_rate');
const boardCreationErrors = new Counter('board_creation_errors');

// Test configuration: 10k concurrent users, 20k boards/min target
export const options = {
  stages: [
    { duration: '30s', target: 2000 },   // Разогрев: 0 -> 2000 пользователей
    { duration: '1m', target: 5000 },    // Увеличение: 2000 -> 5000
    { duration: '2m', target: 10000 },   // Пиковая нагрузка: 10000 пользователей
    { duration: '1m', target: 10000 },   // Удержание пика
    { duration: '30s', target: 0 },      // Плавное снижение
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% запросов < 500ms, 99% < 1s
    http_req_failed: ['rate<0.05'],                  // Менее 5% ошибок
    board_creation_success_rate: ['rate>0.95'],      // Более 95% успешных создаётся
    board_creation_duration: ['p(95)<300'],          // 95% создания досок < 300ms
  },
};

const BASE_URL = 'http://localhost:8080';
let authToken = null;

// Setup: создаём тестового пользователя и получаем токен (выполняется один раз для каждого VU)
export function setup() {
  // Регистрация не нужна для каждого VU - используем общего тестового пользователя
  // или создаём пользователя для каждого VU в default function
  return { baseUrl: BASE_URL };
}

export default function(data) {
  // Каждый VU создаёт своего пользователя и логинится (или используем пул пользователей)
  // Для простоты: создаём уникального пользователя для каждого VU
  if (!authToken) {
    const vuId = __VU;
    const timestamp = Date.now();
    const username = `loadtest_user_${vuId}_${timestamp}`;
    const email = `${username}@test.com`;
    const password = 'TestPassword123!';

    // Регистрация
    const registerPayload = JSON.stringify({
      username: username,
      email: email,
      password: password,
    });

    const registerRes = http.post(`${BASE_URL}/api/v1/auth/register`, registerPayload, {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'register' },
    });

    if (!check(registerRes, { 'registration successful': (r) => r.status === 201 })) {
      console.error(`Registration failed for ${username}: ${registerRes.status} ${registerRes.body}`);
      boardCreationErrors.add(1);
      sleep(1);
      return;
    }

    // Логин
    const loginPayload = JSON.stringify({
      email: email,
      password: password,
    });

    const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, loginPayload, {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'login' },
    });

    const loginSuccess = check(loginRes, {
      'login successful': (r) => r.status === 200,
      'token received': (r) => r.json('access_token') !== undefined,
    });

    if (!loginSuccess) {
      console.error(`Login failed for ${username}: ${loginRes.status} ${loginRes.body}`);
      boardCreationErrors.add(1);
      sleep(1);
      return;
    }

    authToken = loginRes.json('access_token');
  }

  // Создание доски
  const boardName = `Load Test Board ${__VU}-${__ITER}`;
  const boardPayload = JSON.stringify({
    name: boardName,
    description: `Board created during load test by VU ${__VU}, iteration ${__ITER}`,
  });

  const startTime = Date.now();
  const createBoardRes = http.post(`${BASE_URL}/api/v1/boards`, boardPayload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authToken}`,
    },
    tags: { name: 'create_board' },
  });
  const duration = Date.now() - startTime;

  const success = check(createBoardRes, {
    'board created': (r) => r.status === 201,
    'board has id': (r) => r.json('board_id') !== undefined,
    'response time OK': () => duration < 500,
  });

  boardCreationDuration.add(duration);
  boardCreationSuccessRate.add(success);

  if (!success) {
    boardCreationErrors.add(1);
    console.error(`Board creation failed: ${createBoardRes.status} ${createBoardRes.body}`);
  }

  // Имитация реального поведения: небольшая пауза между запросами
  sleep(0.1 + Math.random() * 0.2); // 100-300ms
}

export function teardown(data) {
  console.log('=== Load Test Summary ===');
  console.log(`Base URL: ${data.baseUrl}`);
  console.log('Настройки: 10k concurrent users, цель 20k boards/min');
  console.log('Проверьте метрики выше для детальной статистики');
}
