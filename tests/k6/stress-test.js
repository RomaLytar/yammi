import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// Custom metrics
const stressSuccessRate = new Rate('stress_success_rate');
const stressDuration = new Trend('stress_duration');
const stressErrors = new Counter('stress_errors');
const breakingPoint = new Counter('breaking_point_reached');

// Stress test configuration: постепенное увеличение до точки отказа
export const options = {
  stages: [
    { duration: '1m', target: 500 },    // Разогрев
    { duration: '2m', target: 1000 },   // Постепенное увеличение
    { duration: '3m', target: 2000 },
    { duration: '3m', target: 3000 },
    { duration: '3m', target: 4000 },
    { duration: '2m', target: 5000 },   // Пик
    { duration: '1m', target: 0 },      // Резкий спад
  ],
  thresholds: {
    http_req_duration: ['p(95)<1500'],
    http_req_failed: ['rate<0.15'],  // Допускаем до 15% ошибок при стресс-тестировании
    stress_success_rate: ['rate>0.80'], // 80% успешных запросов
  },
};

const BASE_URL = 'http://localhost:8080';
let authToken = null;
let errorCount = 0;
const ERROR_THRESHOLD = 0.2; // 20% ошибок = breaking point

export function setup() {
  console.log('=== Stress Test Starting ===');
  console.log('Цель: найти breaking point системы');
  console.log('Постепенное увеличение нагрузки до 5000 VUs');
  return { baseUrl: BASE_URL };
}

export default function(data) {
  // Авторизация
  if (!authToken) {
    const vuId = __VU;
    const timestamp = Date.now();
    const username = `stress_user_${vuId}_${timestamp}`;
    const email = `${username}@test.com`;
    const password = 'StressTest123!';

    const registerPayload = JSON.stringify({
      username: username,
      email: email,
      password: password,
    });

    const registerRes = http.post(`${BASE_URL}/api/v1/auth/register`, registerPayload, {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'stress_register' },
      timeout: '10s',
    });

    if (!check(registerRes, { 'registration successful': (r) => r.status === 201 })) {
      errorCount++;
      stressErrors.add(1);
      stressSuccessRate.add(0);
      sleep(1);
      return;
    }

    const loginPayload = JSON.stringify({
      email: email,
      password: password,
    });

    const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, loginPayload, {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'stress_login' },
      timeout: '10s',
    });

    if (!check(loginRes, { 'login successful': (r) => r.status === 200 })) {
      errorCount++;
      stressErrors.add(1);
      stressSuccessRate.add(0);
      sleep(1);
      return;
    }

    authToken = loginRes.json('access_token');
  }

  const authHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  };

  let operationSuccess = true;

  // Группа операций: создание доски и колонки
  group('Board Operations', function() {
    const startTime = Date.now();

    // Создание доски
    const boardPayload = JSON.stringify({
      name: `Stress Board ${__VU}-${__ITER}`,
      description: 'Stress test board',
    });

    const createBoardRes = http.post(`${BASE_URL}/api/v1/boards`, boardPayload, {
      headers: authHeaders,
      tags: { name: 'stress_create_board' },
      timeout: '15s',
    });

    const boardCreated = check(createBoardRes, {
      'board created': (r) => r.status === 201,
      'board_id received': (r) => r.json('board_id') !== undefined,
    });

    if (!boardCreated) {
      operationSuccess = false;
      errorCount++;
      stressErrors.add(1);
      console.error(`Board creation failed at VU=${__VU}: ${createBoardRes.status}`);
    } else {
      const boardId = createBoardRes.json('board_id');

      // Создание колонки
      const columnPayload = JSON.stringify({
        name: 'Stress Column',
        position: 0,
      });

      const createColumnRes = http.post(`${BASE_URL}/api/v1/boards/${boardId}/columns`, columnPayload, {
        headers: authHeaders,
        tags: { name: 'stress_create_column' },
        timeout: '10s',
      });

      const columnCreated = check(createColumnRes, {
        'column created': (r) => r.status === 201,
      });

      if (!columnCreated) {
        operationSuccess = false;
        errorCount++;
        stressErrors.add(1);
      }
    }

    const duration = Date.now() - startTime;
    stressDuration.add(duration);
  });

  // Группа операций: чтение
  group('Read Operations', function() {
    const listBoardsRes = http.get(`${BASE_URL}/api/v1/boards`, {
      headers: authHeaders,
      tags: { name: 'stress_list_boards' },
      timeout: '10s',
    });

    const listSuccess = check(listBoardsRes, {
      'boards listed': (r) => r.status === 200,
    });

    if (!listSuccess) {
      operationSuccess = false;
      errorCount++;
      stressErrors.add(1);
    }
  });

  stressSuccessRate.add(operationSuccess ? 1 : 0);

  // Проверка breaking point
  const currentErrorRate = errorCount / (__ITER + 1);
  if (currentErrorRate > ERROR_THRESHOLD) {
    breakingPoint.add(1);
    console.warn(`⚠️ Breaking point reached! VU=${__VU}, Error rate: ${(currentErrorRate * 100).toFixed(2)}%`);
  }

  sleep(0.1 + Math.random() * 0.3);
}

export function teardown(data) {
  console.log('=== Stress Test Summary ===');
  console.log(`Base URL: ${data.baseUrl}`);
  console.log('Цель: определить максимальную нагрузку до breaking point');
  console.log('Breaking point определяется как >20% ошибок');
  console.log('Анализируйте метрики для определения пределов системы');
}
