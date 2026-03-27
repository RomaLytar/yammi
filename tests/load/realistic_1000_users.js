import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';
import exec from 'k6/execution';

// ─── Метрики ────────────────────────────────────────────────────────────

const errors = {
  register:     new Counter('errors_register'),
  board:        new Counter('errors_board'),
  column:       new Counter('errors_column'),
  card:         new Counter('errors_card'),
  member:       new Counter('errors_member'),
  notification: new Counter('errors_notification'),
};

const latency = {
  register:      new Trend('latency_register_ms'),
  createBoard:   new Trend('latency_create_board_ms'),
  addMember:     new Trend('latency_add_member_ms'),
  createColumn:  new Trend('latency_create_column_ms'),
  createCard:    new Trend('latency_create_card_ms'),
  moveCard:      new Trend('latency_move_card_ms'),
  updateCard:    new Trend('latency_update_card_ms'),
  deleteCard:    new Trend('latency_delete_card_ms'),
  deleteColumn:  new Trend('latency_delete_column_ms'),
  deleteMember:  new Trend('latency_delete_member_ms'),
  deleteBoard:   new Trend('latency_delete_board_ms'),
  listBoards:    new Trend('latency_list_boards_ms'),
  getBoard:      new Trend('latency_get_board_ms'),
  notifications: new Trend('latency_notifications_ms'),
  notifDelivery: new Trend('latency_notif_delivery_ms'),
};

const errorRate = new Rate('error_rate');

// ─── Конфигурация ───────────────────────────────────────────────────────

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TOTAL_USERS = parseInt(__ENV.USERS || '1000');
const MEMBERS_PER_BOARD = parseInt(__ENV.MEMBERS || '2');  // 2 default, 10-50 для big board test

export const options = {
  setupTimeout: '180s',
  scenarios: {
    realistic_usage: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 300 },    // мягкий старт
        { duration: '30s', target: 1000 },   // рост до пика
        { duration: '1m',  target: 1000 },   // удержание пика
        { duration: '30s', target: 0 },      // cooldown
      ],
    },
  },
  thresholds: {
    'http_req_failed':              ['rate<0.05'],
    'http_req_duration':            ['p(95)<1500'],
    'latency_create_board_ms':      ['p(95)<500'],
    'latency_create_card_ms':       ['p(95)<500'],
    'latency_move_card_ms':         ['p(95)<500'],
    'latency_add_member_ms':        ['p(95)<500'],
    'latency_notifications_ms':     ['p(95)<500'],
    'latency_notif_delivery_ms':    ['p(95)<5000'],
    'error_rate':                   ['rate<0.05'],
  },
};

// ─── Setup: регистрация всех пользователей ──────────────────────────────

export function setup() {
  console.log(`Регистрируем ${TOTAL_USERS} пользователей...`);

  const users = [];
  const ts = Date.now();

  for (let i = 0; i < TOTAL_USERS; i++) {
    const email = `load-${i}-${ts}@yammi.io`;
    const name = `User ${i}`;

    const res = http.post(
      `${BASE_URL}/api/v1/auth/register`,
      JSON.stringify({ email, password: 'loadtest123456', name }),
      { headers: { 'Content-Type': 'application/json' }, timeout: '30s' }
    );

    if (res.status === 201) {
      const body = res.json();
      users.push({ id: body.user_id, token: body.access_token, email, name });
    } else {
      console.warn(`Setup: register failed for user ${i}: ${res.status}`);
    }

    if (i % 50 === 0 && i > 0) {
      sleep(0.5);
      console.log(`  ...зарегистрировано ${i}/${TOTAL_USERS}`);
    }
  }

  console.log(`Зарегистрировано ${users.length}/${TOTAL_USERS}. Ждём создание профилей...`);
  sleep(5);

  return { users };
}

// ─── Teardown ───────────────────────────────────────────────────────────
// После теста запустите очистку: ./tests/load/cleanup.sh

export function teardown(data) {
  console.log('Тест завершён. Для очистки БД/Redis запустите: ./tests/load/cleanup.sh');
}

// ─── Helpers ────────────────────────────────────────────────────────────

function auth(token) {
  return {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    timeout: '15s',
  };
}

function randomBetween(min, max) {
  return min + Math.random() * (max - min);
}

function pickRandomUsers(allUsers, excludeId, count) {
  const candidates = allUsers.filter(u => u.id !== excludeId);
  const result = [];
  for (let i = 0; i < count && candidates.length > 0; i++) {
    const idx = Math.floor(Math.random() * candidates.length);
    result.push(candidates.splice(idx, 1)[0]);
  }
  return result;
}

function trackError(counter, endpoint, res) {
  counter.add(1);
  errorRate.add(true);
  console.warn(`${endpoint}: ${res.status} ${res.body ? res.body.substring(0, 100) : ''}`);
}

// ─── Board helpers ──────────────────────────────────────────────────────

function createBoard(h, title) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/boards`,
    JSON.stringify({ title, description: 'Доска создана в load test' }),
    h
  );
  latency.createBoard.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, {
    'createBoard: 201': (r) => r.status === 201,
    'createBoard: has id': () => body.board && body.board.id !== '',
  });

  if (!ok) { trackError(errors.board, 'createBoard', res); return null; }
  errorRate.add(false);
  return body.board;
}

function addMember(h, boardId, userId, role) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/boards/${boardId}/members`,
    JSON.stringify({ user_id: userId, role: role || 'member' }),
    h
  );
  latency.addMember.add(Date.now() - start);

  const ok = check(res, {
    'addMember: 2xx': (r) => r.status >= 200 && r.status < 300,
  });

  if (!ok) { trackError(errors.member, 'addMember', res); return false; }
  errorRate.add(false);
  return true;
}

function createColumn(h, boardId, title) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/boards/${boardId}/columns`,
    JSON.stringify({ title }),
    h
  );
  latency.createColumn.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, {
    'createColumn: 201': (r) => r.status === 201,
    'createColumn: has id': () => body.column && body.column.id !== '',
  });

  if (!ok) { trackError(errors.column, 'createColumn', res); return null; }
  errorRate.add(false);
  return body.column;
}

// Простой lexorank: генерируем позиции "a", "b", "c" ... по индексу
function lexorank(index) {
  const chars = 'abcdefghijklmnopqrstuvwxyz';
  if (index < 26) return chars[index];
  return chars[Math.floor(index / 26) % 26] + chars[index % 26];
}

function createCard(h, columnId, boardId, title, posIndex) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/columns/${columnId}/cards`,
    JSON.stringify({ title, description: '', board_id: boardId, position: lexorank(posIndex || 0) }),
    h
  );
  latency.createCard.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, {
    'createCard: 201': (r) => r.status === 201,
    'createCard: has id': () => body.card && body.card.id !== '',
  });

  if (!ok) { trackError(errors.card, 'createCard', res); return null; }
  errorRate.add(false);
  return body.card;
}

function moveCard(h, card, targetColumnId, boardId) {
  const start = Date.now();
  const res = http.put(
    `${BASE_URL}/api/v1/cards/${card.id}/move`,
    JSON.stringify({
      board_id: boardId,
      from_column_id: card.column_id,
      to_column_id: targetColumnId,
      position: lexorank(Math.floor(Math.random() * 26)),
      version: card.version || 1,
    }),
    h
  );
  latency.moveCard.add(Date.now() - start);

  const ok = check(res, { 'moveCard: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.card, 'moveCard', res); return false; }
  errorRate.add(false);
  return true;
}

function updateCard(h, card, boardId, title) {
  const start = Date.now();
  const res = http.put(
    `${BASE_URL}/api/v1/cards/${card.id}`,
    JSON.stringify({ title, description: 'Обновлено в load test', board_id: boardId, version: card.version || 1 }),
    h
  );
  latency.updateCard.add(Date.now() - start);

  const ok = check(res, { 'updateCard: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.card, 'updateCard', res); return false; }
  errorRate.add(false);
  return true;
}

function deleteCard(h, cardId, boardId) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/cards/delete`,
    JSON.stringify({ card_ids: [cardId], board_id: boardId }),
    h
  );
  latency.deleteCard.add(Date.now() - start);

  const ok = check(res, { 'deleteCard: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.card, 'deleteCard', res); return false; }
  errorRate.add(false);
  return true;
}

function deleteColumn(h, columnId, boardId) {
  const start = Date.now();
  const res = http.del(
    `${BASE_URL}/api/v1/columns/${columnId}`,
    JSON.stringify({ board_id: boardId }),
    h
  );
  latency.deleteColumn.add(Date.now() - start);

  const ok = check(res, { 'deleteColumn: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.column, 'deleteColumn', res); return false; }
  errorRate.add(false);
  return true;
}

function removeMember(h, boardId, userId) {
  const start = Date.now();
  const res = http.del(`${BASE_URL}/api/v1/boards/${boardId}/members/${userId}`, null, h);
  latency.deleteMember.add(Date.now() - start);

  const ok = check(res, { 'removeMember: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.member, 'removeMember', res); return false; }
  errorRate.add(false);
  return true;
}

function deleteBoard(h, boardId) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/boards/delete`,
    JSON.stringify({ board_ids: [boardId] }),
    h
  );
  latency.deleteBoard.add(Date.now() - start);

  const ok = check(res, { 'deleteBoard: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.board, 'deleteBoard', res); return false; }
  errorRate.add(false);
  return true;
}

function listBoards(h) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/boards?limit=10`, h);
  latency.listBoards.add(Date.now() - start);

  const ok = check(res, { 'listBoards: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.board, 'listBoards', res); return []; }
  errorRate.add(false);
  return res.json().boards || [];
}

function getBoard(h, boardId) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/boards/${boardId}`, h);
  latency.getBoard.add(Date.now() - start);

  if (res.status !== 200) return null;
  errorRate.add(false);
  return res.json();
}

// ─── Notification helpers ───────────────────────────────────────────────

function checkNotifications(h) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/notifications?limit=20`, h);
  latency.notifications.add(Date.now() - start);

  const ok = check(res, { 'notifications: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.notification, 'notifications', res); return null; }
  errorRate.add(false);
  return res.json();
}

function markAllRead(h) {
  const res = http.post(`${BASE_URL}/api/v1/notifications/read-all`, null, h);
  check(res, { 'markAllRead: 200': (r) => r.status === 200 });
}

function getUnreadCount(h) {
  const res = http.get(`${BASE_URL}/api/v1/notifications/unread-count`, h);
  if (res.status === 200) return res.json().count;
  return -1;
}

// ─── Сценарии пользователей ─────────────────────────────────────────────

function workerScenario(me, allUsers, h) {
  const board = createBoard(h, `Доска ${me.name}`);
  if (!board) return;
  sleep(randomBetween(0.5, 1.5));

  const memberCount = MEMBERS_PER_BOARD > 2 ? MEMBERS_PER_BOARD : (Math.random() < 0.5 ? 1 : 2);
  const members = pickRandomUsers(allUsers, me.id, memberCount);
  for (const m of members) {
    addMember(h, board.id, m.id, 'member');
    sleep(randomBetween(0.1, 0.3));
  }

  const col1 = createColumn(h, board.id, 'To Do');
  sleep(randomBetween(0.3, 0.8));
  const col2 = createColumn(h, board.id, 'In Progress');
  sleep(randomBetween(0.3, 0.8));
  if (!col1 || !col2) return;

  const cardCount = Math.random() < 0.5 ? 2 : 3;
  const cards = [];
  for (let i = 0; i < cardCount; i++) {
    const card = createCard(h, col1.id, board.id, `Задача ${i + 1}`, i);
    if (card) cards.push(card);
    sleep(randomBetween(0.3, 0.8));
  }

  if (cards.length > 0) {
    moveCard(h, cards[0], col2.id, board.id);
    sleep(randomBetween(0.5, 1.5));
  }

  checkNotifications(h);
  markAllRead(h);
  sleep(randomBetween(0.5, 1));

  if (Math.random() < 0.5) {
    deleteBoard(h, board.id);
  }
}

function readerScenario(me, allUsers, h) {
  const boards = listBoards(h);
  sleep(randomBetween(1, 2));

  if (boards.length > 0) {
    getBoard(h, boards[0].id);
    sleep(randomBetween(1, 2));
  }

  checkNotifications(h);
  sleep(randomBetween(0.5, 1));
  getUnreadCount(h);
  sleep(randomBetween(1, 3));
}

function heavyUserScenario(me, allUsers, h) {
  const boards = [];
  for (let b = 0; b < 2; b++) {
    const board = createBoard(h, `Heavy Board ${b + 1} — ${me.name}`);
    if (board) boards.push(board);
    sleep(randomBetween(0.3, 0.8));
  }
  if (boards.length === 0) return;

  for (const board of boards) {
    const heavyMemberCount = MEMBERS_PER_BOARD > 3 ? MEMBERS_PER_BOARD : 3;
    const members = pickRandomUsers(allUsers, me.id, heavyMemberCount);
    for (const m of members) {
      addMember(h, board.id, m.id, 'member');
      sleep(randomBetween(0.1, 0.3));
    }

    const cols = [];
    for (const title of ['Backlog', 'In Progress', 'Done']) {
      const col = createColumn(h, board.id, title);
      if (col) cols.push(col);
      sleep(randomBetween(0.2, 0.5));
    }
    if (cols.length < 3) continue;

    const cards = [];
    for (let i = 0; i < 5; i++) {
      const card = createCard(h, cols[0].id, board.id, `Task ${i + 1}`, i);
      if (card) cards.push(card);
      sleep(randomBetween(0.2, 0.5));
    }

    for (let i = 0; i < Math.min(3, cards.length); i++) {
      moveCard(h, cards[i], cols[1].id, board.id);
      // После move обновляем column_id для следующих операций
      cards[i].column_id = cols[1].id;
      sleep(randomBetween(0.3, 0.8));
    }

    if (cards.length > 0) {
      moveCard(h, cards[0], cols[2].id, board.id);
      cards[0].column_id = cols[2].id;
      sleep(randomBetween(0.3, 0.8));
    }

    for (let i = 0; i < Math.min(2, cards.length); i++) {
      updateCard(h, cards[i], board.id, `Updated Task ${i + 1}`);
      sleep(randomBetween(0.3, 0.5));
    }

    if (cards.length > 2) {
      deleteCard(h, cards[cards.length - 1].id, board.id);
      sleep(randomBetween(0.3, 0.5));
    }

    deleteColumn(h, cols[2].id, board.id);
    sleep(randomBetween(0.3, 0.5));

    if (members.length > 0) {
      removeMember(h, board.id, members[0].id);
      sleep(randomBetween(0.3, 0.5));
    }
  }

  // Notification delivery latency: находим самую свежую нотификацию,
  // сравниваем её created_at с текущим временем
  sleep(2);
  const notifs = checkNotifications(h);
  if (notifs && notifs.notifications && notifs.notifications.length > 0) {
    let maxCreatedAt = 0;
    for (const n of notifs.notifications) {
      if (n.created_at) {
        const t = new Date(n.created_at).getTime();
        if (t > maxCreatedAt) maxCreatedAt = t;
      }
    }
    if (maxCreatedAt > 0) {
      const deliveryMs = Date.now() - maxCreatedAt;
      if (deliveryMs > 0 && deliveryMs < 60000) {
        latency.notifDelivery.add(deliveryMs);
      }
    }
  }
  markAllRead(h);

  if (Math.random() < 0.3 && boards.length > 0) {
    deleteBoard(h, boards[0].id);
  }
}

// ─── Main ───────────────────────────────────────────────────────────────

export default function (data) {
  const allUsers = data.users;
  if (!allUsers || allUsers.length === 0) {
    console.error('No users from setup!');
    return;
  }

  // Гибрид: VU + iteration — минимизирует коллизии при параллельном доступе
  const iterGlobal = exec.scenario.iterationInTest;
  const userIdx = ((exec.vu.idInTest - 1) + iterGlobal) % allUsers.length;
  const me = allUsers[userIdx];
  const h = auth(me.token);

  // Тип: 70% worker, 20% reader, 10% heavy
  const roll = iterGlobal % 10;

  if (roll < 7) {
    workerScenario(me, allUsers, h);
  } else if (roll < 9) {
    readerScenario(me, allUsers, h);
  } else {
    heavyUserScenario(me, allUsers, h);
  }
}

// ─── Summary ────────────────────────────────────────────────────────────

export function handleSummary(data) {
  function val(name, field) {
    const m = data.metrics[name];
    if (!m) return 'N/A';
    if (field === 'count') return m.values.count || 0;
    if (field === 'rate') return (m.values.rate * 100).toFixed(1) + '%';
    return m.values[field] ? m.values[field].toFixed(0) : 'N/A';
  }

  const summary = `
╔══════════════════════════════════════════════════════════════════╗
║           НАГРУЗОЧНЫЙ ТЕСТ: 1000 пользователей                 ║
╠══════════════════════════════════════════════════════════════════╣
║                                                                  ║
║  Распределение:  70% workers / 20% readers / 10% heavy users   ║
║                                                                  ║
║  ── Latency (p95) ──────────────────────────────────────────    ║
║  Создание доски:       ${String(val('latency_create_board_ms', 'p(95)')).padStart(6)} ms                          ║
║  Добавление участника: ${String(val('latency_add_member_ms', 'p(95)')).padStart(6)} ms                          ║
║  Создание колонки:     ${String(val('latency_create_column_ms', 'p(95)')).padStart(6)} ms                          ║
║  Создание карточки:    ${String(val('latency_create_card_ms', 'p(95)')).padStart(6)} ms                          ║
║  Перемещение карточки: ${String(val('latency_move_card_ms', 'p(95)')).padStart(6)} ms                          ║
║  Нотификации:          ${String(val('latency_notifications_ms', 'p(95)')).padStart(6)} ms                          ║
║  Notif delivery:       ${String(val('latency_notif_delivery_ms', 'p(95)')).padStart(6)} ms                          ║
║                                                                  ║
║  ── Ошибки ─────────────────────────────────────────────────    ║
║  Error rate:           ${String(val('error_rate', 'rate')).padStart(6)}                                ║
║  Board errors:         ${String(val('errors_board', 'count')).padStart(6)}                                ║
║  Card errors:          ${String(val('errors_card', 'count')).padStart(6)}                                ║
║  Member errors:        ${String(val('errors_member', 'count')).padStart(6)}                                ║
║  Notification errors:  ${String(val('errors_notification', 'count')).padStart(6)}                                ║
║                                                                  ║
║  ── HTTP ───────────────────────────────────────────────────    ║
║  Total requests:       ${String(data.metrics.http_reqs ? data.metrics.http_reqs.values.count : 0).padStart(6)}                                ║
║  Failed requests:      ${String(val('http_req_failed', 'rate')).padStart(6)}                                ║
║  Duration p95:         ${String(val('http_req_duration', 'p(95)')).padStart(6)} ms                          ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
`;

  return { stdout: summary };
}
