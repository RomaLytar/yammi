import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// Custom metrics
const spikeSuccessRate = new Rate('spike_success_rate');
const spikeDuration = new Trend('spike_duration');
const spikeErrors = new Counter('spike_errors');
const recoveryTime = new Trend('recovery_time');

// Spike test configuration: резкое увеличение нагрузки
export const options = {
  stages: [
    { duration: '10s', target: 100 },    // Нормальная нагрузка
    { duration: '10s', target: 100 },    // Удержание
    { duration: '10s', target: 5000 },   // SPIKE! Резкий скачок до 5k пользователей
    { duration: '30s', target: 5000 },   // Удержание пика
    { duration: '10s', target: 100 },    // Резкое снижение
    { duration: '30s', target: 100 },    // Восстановление и стабилизация
    { duration: '10s', target: 0 },      // Завершение
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000', 'p(99)<5000'], // Мягче для spike test
    http_req_failed: ['rate<0.10'],                   // Допускаем до 10% ошибок во время spike
    spike_success_rate: ['rate>0.85'],                // 85% успешных запросов
  },
};

const BASE_URL = 'http://localhost:8080';
let authToken = null;
let spikeStartTime = null;
let spikeRecovered = false;

export function setup() {
  console.log('=== Spike Test Starting ===');
  console.log('Сценарий: 100 → 5000 пользователей за 10 секунд');
  return { baseUrl: BASE_URL };
}

export default function(data) {
  // Определяем, находимся ли мы в spike фазе
  const currentVUs = __VU;
  const isSpike = currentVUs > 1000;

  if (isSpike && !spikeStartTime) {
    spikeStartTime = Date.now();
  }

  // Аутентификация для каждого VU
  if (!authToken) {
    const vuId = __VU;
    const timestamp = Date.now();
    const username = `spike_user_${vuId}_${timestamp}`;
    const email = `${username}@test.com`;
    const password = 'SpikeTest123!';

    // Регистрация
    const registerPayload = JSON.stringify({
      username: username,
      email: email,
      password: password,
    });

    const registerRes = http.post(`${BASE_URL}/api/v1/auth/register`, registerPayload, {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'spike_register', spike: isSpike },
    });

    if (!check(registerRes, { 'registration successful': (r) => r.status === 201 })) {
      console.error(`Registration failed during ${isSpike ? 'SPIKE' : 'normal'}: ${registerRes.status}`);
      spikeErrors.add(1);
      spikeSuccessRate.add(0);
      sleep(0.5);
      return;
    }

    // Логин
    const loginPayload = JSON.stringify({
      email: email,
      password: password,
    });

    const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, loginPayload, {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'spike_login', spike: isSpike },
    });

    if (!check(loginRes, { 'login successful': (r) => r.status === 200 })) {
      console.error(`Login failed during ${isSpike ? 'SPIKE' : 'normal'}: ${loginRes.status}`);
      spikeErrors.add(1);
      spikeSuccessRate.add(0);
      sleep(0.5);
      return;
    }

    authToken = loginRes.json('access_token');
  }

  // Основная операция: создание доски (самая ресурсоёмкая)
  const startTime = Date.now();
  const boardPayload = JSON.stringify({
    name: `Spike Board ${__VU}-${__ITER}`,
    description: `Board created during ${isSpike ? 'SPIKE' : 'normal load'}`,
  });

  const createBoardRes = http.post(`${BASE_URL}/api/v1/boards`, boardPayload, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authToken}`,
    },
    tags: { name: 'spike_create_board', spike: isSpike },
  });

  const duration = Date.now() - startTime;
  spikeDuration.add(duration);

  const success = check(createBoardRes, {
    'board created': (r) => r.status === 201,
    'response time acceptable': () => duration < 3000, // 3s во время spike допустимо
  });

  spikeSuccessRate.add(success ? 1 : 0);

  if (!success) {
    spikeErrors.add(1);
    if (isSpike) {
      console.error(`SPIKE FAILURE: Board creation failed (${duration}ms): ${createBoardRes.status}`);
    }
  } else {
    // Проверяем восстановление после spike
    if (!isSpike && spikeStartTime && !spikeRecovered) {
      const recoveryDuration = Date.now() - spikeStartTime;
      recoveryTime.add(recoveryDuration);
      spikeRecovered = true;
      console.log(`✓ System recovered after ${recoveryDuration}ms from spike start`);
    }
  }

  // Дополнительная проверка: чтение доски
  if (success && createBoardRes.json('board_id')) {
    const boardId = createBoardRes.json('board_id');
    const getBoardRes = http.get(`${BASE_URL}/api/v1/boards/${boardId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
      },
      tags: { name: 'spike_get_board', spike: isSpike },
    });

    check(getBoardRes, {
      'board retrieved': (r) => r.status === 200,
    });
  }

  // Минимальная пауза (имитация реальных пользователей)
  sleep(0.05 + Math.random() * 0.1); // 50-150ms
}

export function teardown(data) {
  console.log('=== Spike Test Summary ===');
  console.log(`Base URL: ${data.baseUrl}`);
  console.log('Сценарий: 100 → 5000 пользователей за 10 секунд');
  console.log('Цель: проверить устойчивость системы к резким скачкам нагрузки');
  console.log('Проверьте метрики recovery_time для времени восстановления');
}
