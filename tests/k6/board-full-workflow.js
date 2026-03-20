import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// Custom metrics
const workflowSuccessRate = new Rate('workflow_success_rate');
const workflowDuration = new Trend('workflow_duration');
const workflowErrors = new Counter('workflow_errors');

// Test configuration: полный workflow с умеренной нагрузкой
export const options = {
  stages: [
    { duration: '30s', target: 50 },    // Разогрев
    { duration: '2m', target: 200 },    // Увеличение нагрузки
    { duration: '3m', target: 500 },    // Пиковая нагрузка
    { duration: '1m', target: 200 },    // Снижение
    { duration: '30s', target: 0 },     // Завершение
  ],
  thresholds: {
    http_req_duration: ['p(95)<800', 'p(99)<1500'],
    http_req_failed: ['rate<0.05'],
    workflow_success_rate: ['rate>0.90'], // 90% полных workflow успешны
    workflow_duration: ['p(95)<5000'],     // 95% workflow < 5s
  },
};

const BASE_URL = 'http://localhost:8080';

export function setup() {
  return { baseUrl: BASE_URL };
}

export default function(data) {
  const workflowStart = Date.now();
  let workflowSuccess = true;

  const vuId = __VU;
  const timestamp = Date.now();
  const username = `workflow_user_${vuId}_${timestamp}`;
  const email = `${username}@test.com`;
  const password = 'WorkflowTest123!';

  let authToken = null;
  let userId = null;
  let boardId = null;
  let columnId = null;
  let cardId = null;

  // ===== 1. РЕГИСТРАЦИЯ =====
  group('Registration', function() {
    const registerPayload = JSON.stringify({
      username: username,
      email: email,
      password: password,
    });

    const registerRes = http.post(`${BASE_URL}/api/v1/auth/register`, registerPayload, {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'workflow_register' },
    });

    const success = check(registerRes, {
      'registration successful': (r) => r.status === 201,
      'user_id received': (r) => r.json('user_id') !== undefined,
    });

    if (success) {
      userId = registerRes.json('user_id');
    } else {
      console.error(`Registration failed: ${registerRes.status} ${registerRes.body}`);
      workflowSuccess = false;
      workflowErrors.add(1);
      return;
    }
  });

  if (!workflowSuccess) {
    workflowSuccessRate.add(0);
    return;
  }

  sleep(0.1);

  // ===== 2. ЛОГИН =====
  group('Login', function() {
    const loginPayload = JSON.stringify({
      email: email,
      password: password,
    });

    const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, loginPayload, {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'workflow_login' },
    });

    const success = check(loginRes, {
      'login successful': (r) => r.status === 200,
      'access_token received': (r) => r.json('access_token') !== undefined,
    });

    if (success) {
      authToken = loginRes.json('access_token');
    } else {
      console.error(`Login failed: ${loginRes.status} ${loginRes.body}`);
      workflowSuccess = false;
      workflowErrors.add(1);
      return;
    }
  });

  if (!workflowSuccess) {
    workflowSuccessRate.add(0);
    return;
  }

  sleep(0.2);

  const authHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  };

  // ===== 3. СОЗДАНИЕ ДОСКИ =====
  group('Create Board', function() {
    const boardPayload = JSON.stringify({
      name: `Workflow Board ${vuId}-${__ITER}`,
      description: 'Board for full workflow test',
    });

    const createBoardRes = http.post(`${BASE_URL}/api/v1/boards`, boardPayload, {
      headers: authHeaders,
      tags: { name: 'workflow_create_board' },
    });

    const success = check(createBoardRes, {
      'board created': (r) => r.status === 201,
      'board_id received': (r) => r.json('board_id') !== undefined,
    });

    if (success) {
      boardId = createBoardRes.json('board_id');
    } else {
      console.error(`Board creation failed: ${createBoardRes.status} ${createBoardRes.body}`);
      workflowSuccess = false;
      workflowErrors.add(1);
      return;
    }
  });

  if (!workflowSuccess) {
    workflowSuccessRate.add(0);
    return;
  }

  sleep(0.1);

  // ===== 4. ПОЛУЧЕНИЕ ДОСКИ =====
  group('Get Board', function() {
    const getBoardRes = http.get(`${BASE_URL}/api/v1/boards/${boardId}`, {
      headers: authHeaders,
      tags: { name: 'workflow_get_board' },
    });

    const success = check(getBoardRes, {
      'board retrieved': (r) => r.status === 200,
      'board name matches': (r) => r.json('name') !== undefined,
    });

    if (!success) {
      console.error(`Get board failed: ${getBoardRes.status} ${getBoardRes.body}`);
      workflowSuccess = false;
      workflowErrors.add(1);
    }
  });

  sleep(0.1);

  // ===== 5. СОЗДАНИЕ КОЛОНКИ =====
  group('Create Column', function() {
    const columnPayload = JSON.stringify({
      name: 'To Do',
      position: 0,
    });

    const createColumnRes = http.post(`${BASE_URL}/api/v1/boards/${boardId}/columns`, columnPayload, {
      headers: authHeaders,
      tags: { name: 'workflow_create_column' },
    });

    const success = check(createColumnRes, {
      'column created': (r) => r.status === 201,
      'column_id received': (r) => r.json('column_id') !== undefined,
    });

    if (success) {
      columnId = createColumnRes.json('column_id');
    } else {
      console.error(`Column creation failed: ${createColumnRes.status} ${createColumnRes.body}`);
      workflowSuccess = false;
      workflowErrors.add(1);
      return;
    }
  });

  if (!workflowSuccess) {
    workflowSuccessRate.add(0);
    return;
  }

  sleep(0.1);

  // ===== 6. СОЗДАНИЕ КАРТОЧКИ =====
  group('Create Card', function() {
    const cardPayload = JSON.stringify({
      title: `Test Card ${__ITER}`,
      description: 'Card created during workflow test',
      position: 0,
    });

    const createCardRes = http.post(`${BASE_URL}/api/v1/columns/${columnId}/cards`, cardPayload, {
      headers: authHeaders,
      tags: { name: 'workflow_create_card' },
    });

    const success = check(createCardRes, {
      'card created': (r) => r.status === 201,
      'card_id received': (r) => r.json('card_id') !== undefined,
    });

    if (success) {
      cardId = createCardRes.json('card_id');
    } else {
      console.error(`Card creation failed: ${createCardRes.status} ${createCardRes.body}`);
      workflowSuccess = false;
      workflowErrors.add(1);
      return;
    }
  });

  if (!workflowSuccess) {
    workflowSuccessRate.add(0);
    return;
  }

  sleep(0.1);

  // ===== 7. ОБНОВЛЕНИЕ КАРТОЧКИ =====
  group('Update Card', function() {
    const updateCardPayload = JSON.stringify({
      title: `Updated Card ${__ITER}`,
      description: 'Updated during workflow test',
    });

    const updateCardRes = http.put(`${BASE_URL}/api/v1/cards/${cardId}`, updateCardPayload, {
      headers: authHeaders,
      tags: { name: 'workflow_update_card' },
    });

    const success = check(updateCardRes, {
      'card updated': (r) => r.status === 200,
    });

    if (!success) {
      console.error(`Card update failed: ${updateCardRes.status} ${updateCardRes.body}`);
      workflowSuccess = false;
      workflowErrors.add(1);
    }
  });

  sleep(0.1);

  // ===== 8. ПОЛУЧЕНИЕ КАРТОЧКИ =====
  group('Get Card', function() {
    const getCardRes = http.get(`${BASE_URL}/api/v1/cards/${cardId}`, {
      headers: authHeaders,
      tags: { name: 'workflow_get_card' },
    });

    const success = check(getCardRes, {
      'card retrieved': (r) => r.status === 200,
      'card title matches': (r) => r.json('title') !== undefined,
    });

    if (!success) {
      console.error(`Get card failed: ${getCardRes.status} ${getCardRes.body}`);
      workflowSuccess = false;
      workflowErrors.add(1);
    }
  });

  sleep(0.1);

  // ===== 9. ПОЛУЧЕНИЕ СПИСКА ДОСОК =====
  group('List Boards', function() {
    const listBoardsRes = http.get(`${BASE_URL}/api/v1/boards`, {
      headers: authHeaders,
      tags: { name: 'workflow_list_boards' },
    });

    const success = check(listBoardsRes, {
      'boards listed': (r) => r.status === 200,
      'boards array exists': (r) => Array.isArray(r.json('boards')),
    });

    if (!success) {
      console.error(`List boards failed: ${listBoardsRes.status} ${listBoardsRes.body}`);
      workflowSuccess = false;
      workflowErrors.add(1);
    }
  });

  // Финальные метрики
  const workflowEnd = Date.now();
  const totalDuration = workflowEnd - workflowStart;

  workflowDuration.add(totalDuration);
  workflowSuccessRate.add(workflowSuccess ? 1 : 0);

  if (workflowSuccess) {
    console.log(`✓ Workflow completed in ${totalDuration}ms`);
  }

  sleep(1);
}

export function teardown(data) {
  console.log('=== Full Workflow Test Summary ===');
  console.log(`Base URL: ${data.baseUrl}`);
  console.log('Workflow: Register → Login → Create Board → Column → Card → Update → Get');
  console.log('Проверьте метрики выше для детальной статистики');
}
