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
  label:        new Counter('errors_label'),
  comment:      new Counter('errors_comment'),
  checklist:    new Counter('errors_checklist'),
  automation:   new Counter('errors_automation'),
  customField:  new Counter('errors_custom_field'),
  cardLink:     new Counter('errors_card_link'),
  settings:     new Counter('errors_settings'),
  userLabel:    new Counter('errors_user_label'),
  template:     new Counter('errors_template'),
};

const latency = {
  register:           new Trend('latency_register_ms'),
  createBoard:        new Trend('latency_create_board_ms'),
  addMember:          new Trend('latency_add_member_ms'),
  createColumn:       new Trend('latency_create_column_ms'),
  createCard:         new Trend('latency_create_card_ms'),
  moveCard:           new Trend('latency_move_card_ms'),
  updateCard:         new Trend('latency_update_card_ms'),
  deleteCard:         new Trend('latency_delete_card_ms'),
  deleteColumn:       new Trend('latency_delete_column_ms'),
  deleteMember:       new Trend('latency_delete_member_ms'),
  deleteBoard:        new Trend('latency_delete_board_ms'),
  listBoards:         new Trend('latency_list_boards_ms'),
  getBoard:           new Trend('latency_get_board_ms'),
  notifications:      new Trend('latency_notifications_ms'),
  notifDelivery:      new Trend('latency_notif_delivery_ms'),
  // Labels
  createLabel:        new Trend('latency_create_label_ms'),
  attachLabel:        new Trend('latency_attach_label_ms'),
  detachLabel:        new Trend('latency_detach_label_ms'),
  listLabels:         new Trend('latency_list_labels_ms'),
  getCardLabels:      new Trend('latency_get_card_labels_ms'),
  availableLabels:    new Trend('latency_available_labels_ms'),
  // Comments
  createComment:      new Trend('latency_create_comment_ms'),
  replyComment:       new Trend('latency_reply_comment_ms'),
  listComments:       new Trend('latency_list_comments_ms'),
  // Checklists
  createChecklist:    new Trend('latency_create_checklist_ms'),
  createChecklistItem:new Trend('latency_create_checklist_item_ms'),
  toggleItem:         new Trend('latency_toggle_item_ms'),
  listChecklists:     new Trend('latency_list_checklists_ms'),
  // Automation
  createAutomation:   new Trend('latency_create_automation_ms'),
  listAutomations:    new Trend('latency_list_automations_ms'),
  // Custom Fields
  createCustomField:  new Trend('latency_create_custom_field_ms'),
  setFieldValue:      new Trend('latency_set_field_value_ms'),
  listCustomFields:   new Trend('latency_list_custom_fields_ms'),
  getCardFields:      new Trend('latency_get_card_fields_ms'),
  // Card Links
  linkCards:          new Trend('latency_link_cards_ms'),
  getChildren:        new Trend('latency_get_children_ms'),
  // Board Settings
  getSettings:        new Trend('latency_get_settings_ms'),
  updateSettings:     new Trend('latency_update_settings_ms'),
  // User Labels
  createUserLabel:    new Trend('latency_create_user_label_ms'),
  listUserLabels:     new Trend('latency_list_user_labels_ms'),
  // Templates
  createTemplate:     new Trend('latency_create_template_ms'),
  createFromTemplate: new Trend('latency_create_from_template_ms'),
  // Assign
  assignCard:         new Trend('latency_assign_card_ms'),
};

const errorRate = new Rate('error_rate');

// ─── Конфигурация ───────────────────────────────────────────────────────

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TOTAL_USERS = parseInt(__ENV.USERS || '1000');
const BOARD_CREATORS = parseInt(__ENV.BOARD_CREATORS || '50');
const MIN_BOARDS = parseInt(__ENV.MIN_BOARDS || '1');
const MAX_BOARDS = parseInt(__ENV.MAX_BOARDS || '3');
const MIN_COLUMNS = parseInt(__ENV.MIN_COLUMNS || '1');
const MAX_COLUMNS = parseInt(__ENV.MAX_COLUMNS || '8');
const MIN_CARDS = parseInt(__ENV.MIN_CARDS || '0');
const MAX_CARDS = parseInt(__ENV.MAX_CARDS || '15');
const MEMBERS_PER_BOARD = parseInt(__ENV.MEMBERS || '5');

export const options = {
  setupTimeout: '600s',
  scenarios: {
    realistic_usage: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 300 },
        { duration: '30s', target: 1000 },
        { duration: '1m',  target: 1000 },
        { duration: '30s', target: 0 },
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
    'latency_create_label_ms':      ['p(95)<500'],
    'latency_create_comment_ms':    ['p(95)<500'],
    'latency_create_checklist_ms':  ['p(95)<500'],
    'latency_create_automation_ms': ['p(95)<1000'],
    'latency_notifications_ms':     ['p(95)<500'],
    'latency_notif_delivery_ms':    ['p(95)<5000'],
    'error_rate':                   ['rate<0.05'],
  },
};

// ─── Setup: регистрация пользователей + создание тяжёлых досок ──────────

export function setup() {
  // 1. Регистрация пользователей
  console.log(`Регистрируем ${TOTAL_USERS} пользователей...`);
  const users = [];
  const ts = Date.now();

  for (let i = 0; i < TOTAL_USERS; i++) {
    const email = `load-${i}-${ts}@yammi.io`;
    const name = `User ${i}`;

    const res = http.post(
      `${BASE_URL}/api/v1/auth/register`,
      JSON.stringify({ email, password: 'LoadTest123456', name }),
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

  if (users.length === 0) {
    console.error('Не удалось зарегистрировать пользователей!');
    return { users: [], boards: [] };
  }

  // 2. Каждый пользователь создаёт 1-3 доски с 1-8 колонками и 0-15 карточками
  const creators = Math.min(BOARD_CREATORS, users.length);
  console.log(`\n${creators} юзеров создают по ${MIN_BOARDS}-${MAX_BOARDS} досок (${MIN_COLUMNS}-${MAX_COLUMNS} кол, ${MIN_CARDS}-${MAX_CARDS} карт/кол)...`);

  const boardsData = [];
  const columnNames = ['Backlog', 'To Do', 'In Progress', 'Review', 'QA', 'Staging', 'Done', 'Archive'];
  const labelPresets = [
    { name: 'Bug', color: '#E53E3E' }, { name: 'Feature', color: '#38A169' },
    { name: 'Task', color: '#3182CE' }, { name: 'Improvement', color: '#D69E2E' },
    { name: 'Urgent', color: '#E53E3E' }, { name: 'Low', color: '#718096' },
    { name: 'Backend', color: '#805AD5' },
  ];
  const priorities = ['low', 'medium', 'high', 'critical'];
  const taskTypes = ['bug', 'feature', 'task', 'improvement'];
  let boardCounter = 0;

  for (let u = 0; u < creators; u++) {
    const owner = users[u];
    const h = authHeaders(owner.token);
    const numBoards = MIN_BOARDS + Math.floor(Math.random() * (MAX_BOARDS - MIN_BOARDS + 1));

    for (let b = 0; b < numBoards; b++) {
      boardCounter++;

      // Создать доску
      const boardRes = http.post(
        `${BASE_URL}/api/v1/boards`,
        JSON.stringify({ title: `${owner.name} — Board ${b + 1}`, description: `Доска ${b + 1} пользователя ${owner.name}` }),
        h
      );
      if (boardRes.status !== 201) {
        console.warn(`Setup: board ${boardCounter} failed: ${boardRes.status}`);
        continue;
      }
      const board = boardRes.json().board;

      // Добавить участников
      const memberUsers = [];
      for (let m = 0; m < MEMBERS_PER_BOARD; m++) {
        const memberIdx = (boardCounter * MEMBERS_PER_BOARD + m + 1) % users.length;
        if (users[memberIdx].id === owner.id) continue;
        const mRes = http.post(
          `${BASE_URL}/api/v1/boards/${board.id}/members`,
          JSON.stringify({ user_id: users[memberIdx].id, role: 'member' }),
          h
        );
        if (mRes.status >= 200 && mRes.status < 300) {
          memberUsers.push(users[memberIdx]);
        }
      }

      // Создать метки (3-7 случайных)
      const boardLabels = [];
      const numLabels = 3 + Math.floor(Math.random() * 5);
      const shuffledLabels = labelPresets.slice().sort(() => Math.random() - 0.5);
      for (let i = 0; i < numLabels && i < shuffledLabels.length; i++) {
        const lRes = http.post(
          `${BASE_URL}/api/v1/boards/${board.id}/labels`,
          JSON.stringify({ name: shuffledLabels[i].name, color: shuffledLabels[i].color }),
          h
        );
        if (lRes.status === 201) boardLabels.push(lRes.json().label);
      }

      // Создать колонки (1-8)
      const numColumns = MIN_COLUMNS + Math.floor(Math.random() * (MAX_COLUMNS - MIN_COLUMNS + 1));
      const columnIds = [];

      for (let c = 0; c < numColumns; c++) {
        const colTitle = c < columnNames.length ? columnNames[c] : `Column ${c + 1}`;
        const cRes = http.post(
          `${BASE_URL}/api/v1/boards/${board.id}/columns`,
          JSON.stringify({ title: colTitle }),
          h
        );
        if (cRes.status === 201) {
          columnIds.push(cRes.json().column.id);
        }
      }

      if (columnIds.length === 0) {
        console.warn(`Setup: board ${boardCounter} — no columns created`);
        continue;
      }

      // Создать карточки в каждой колонке (0-15)
      const allCardIds = [];
      let totalCards = 0;

      for (let c = 0; c < columnIds.length; c++) {
        const numCards = MIN_CARDS + Math.floor(Math.random() * (MAX_CARDS - MIN_CARDS + 1));

        for (let i = 0; i < numCards; i++) {
          const payload = {
            title: `Task ${c + 1}-${i + 1}`,
            description: `Задача в колонке ${c + 1}, позиция ${i + 1}`,
            board_id: board.id,
            position: lexorank(i),
            priority: priorities[Math.floor(Math.random() * priorities.length)],
            task_type: taskTypes[Math.floor(Math.random() * taskTypes.length)],
          };
          if (Math.random() < 0.6) {
            const d = new Date();
            d.setDate(d.getDate() + Math.floor(Math.random() * 14) + 1);
            payload.due_date = d.toISOString();
          }

          const cardRes = http.post(
            `${BASE_URL}/api/v1/columns/${columnIds[c]}/cards`,
            JSON.stringify(payload),
            h
          );
          if (cardRes.status === 201) {
            const card = cardRes.json().card;
            allCardIds.push({ id: card.id, column_id: columnIds[c], version: card.version || 1 });
            totalCards++;

            // Привязать метку (~40%)
            if (Math.random() < 0.4 && boardLabels.length > 0) {
              const lbl = boardLabels[Math.floor(Math.random() * boardLabels.length)];
              http.post(
                `${BASE_URL}/api/v1/boards/${board.id}/cards/${card.id}/labels`,
                JSON.stringify({ label_id: lbl.id }),
                h
              );
            }
          }
        }
      }

      boardsData.push({
        id: board.id,
        ownerId: owner.id,
        ownerToken: owner.token,
        memberIds: memberUsers.map(u => u.id),
        memberTokens: memberUsers.map(u => u.token),
        columnIds: columnIds,
        cardIds: allCardIds,
        labelIds: boardLabels.map(l => l.id),
      });

      if (boardCounter % 10 === 0) {
        console.log(`  ...досок: ${boardCounter}, последняя: ${columnIds.length} кол, ${totalCards} карт`);
        sleep(0.3);
      }
    }
  }

  console.log(`\nSetup готов: ${users.length} юзеров, ${boardsData.length} досок`);
  const totalCardsCreated = boardsData.reduce((sum, b) => sum + b.cardIds.length, 0);
  const totalColumnsCreated = boardsData.reduce((sum, b) => sum + b.columnIds.length, 0);
  console.log(`Всего: ${totalColumnsCreated} колонок, ${totalCardsCreated} карточек`);

  return { users, boards: boardsData };
}

// ─── Teardown ───────────────────────────────────────────────────────────

export function teardown(data) {
  console.log('Тест завершён. Данные оставлены в БД.');
}

// ─── Helpers ────────────────────────────────────────────────────────────

function authHeaders(token) {
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

function pickRandom(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
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

function lexorank(index) {
  const chars = 'abcdefghijklmnopqrstuvwxyz';
  if (index < 26) return chars[index];
  return chars[Math.floor(index / 26) % 26] + chars[index % 26];
}

const LABEL_PRESETS = [
  { name: 'Bug',         color: '#E53E3E' },
  { name: 'Feature',     color: '#38A169' },
  { name: 'Task',        color: '#3182CE' },
  { name: 'Improvement', color: '#D69E2E' },
  { name: 'Urgent',      color: '#E53E3E' },
  { name: 'Low',         color: '#718096' },
  { name: 'Backend',     color: '#805AD5' },
];

const PRIORITIES = ['low', 'medium', 'high', 'critical'];
const TASK_TYPES = ['bug', 'feature', 'task', 'improvement'];

// ─── Soft GET (403/404 — нормальная гонка) ──────────────────────────────

function softGet(h, url, latencyTrend) {
  const start = Date.now();
  const res = http.get(url, h);
  if (latencyTrend) latencyTrend.add(Date.now() - start);
  if (res.status === 200) { errorRate.add(false); return res.json(); }
  return null;
}

// ─── Board CRUD helpers ─────────────────────────────────────────────────

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

  const ok = check(res, { 'addMember: 2xx': (r) => r.status >= 200 && r.status < 300 });
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

function createCard(h, columnId, boardId, title, posIndex, opts) {
  const payload = {
    title,
    description: (opts && opts.description) || '',
    board_id: boardId,
    position: lexorank(posIndex || 0),
  };
  if (opts && opts.priority)  payload.priority = opts.priority;
  if (opts && opts.task_type) payload.task_type = opts.task_type;
  if (opts && opts.due_date)  payload.due_date = opts.due_date;

  const start = Date.now();
  const res = http.post(`${BASE_URL}/api/v1/columns/${columnId}/cards`, JSON.stringify(payload), h);
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

function moveCard(h, cardId, fromColumnId, targetColumnId, boardId, version) {
  const start = Date.now();
  const res = http.put(
    `${BASE_URL}/api/v1/cards/${cardId}/move`,
    JSON.stringify({
      board_id: boardId,
      from_column_id: fromColumnId,
      to_column_id: targetColumnId,
      position: lexorank(Math.floor(Math.random() * 26)),
      version: version || 1,
    }),
    h
  );
  latency.moveCard.add(Date.now() - start);

  const ok = check(res, { 'moveCard: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.card, 'moveCard', res); return false; }
  errorRate.add(false);
  return true;
}

function updateCard(h, cardId, boardId, title, version) {
  const start = Date.now();
  const res = http.put(
    `${BASE_URL}/api/v1/cards/${cardId}`,
    JSON.stringify({ title, description: 'Обновлено в load test', board_id: boardId, version: version || 1 }),
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

// ─── Label helpers ──────────────────────────────────────────────────────

function createLabel(h, boardId, name, color) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/boards/${boardId}/labels`,
    JSON.stringify({ name, color }),
    h
  );
  latency.createLabel.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, {
    'createLabel: 201': (r) => r.status === 201,
    'createLabel: has id': () => body.label && body.label.id !== '',
  });

  if (!ok) { trackError(errors.label, 'createLabel', res); return null; }
  errorRate.add(false);
  return body.label;
}

function listLabels(h, boardId) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/boards/${boardId}/labels`, h);
  latency.listLabels.add(Date.now() - start);

  const ok = check(res, { 'listLabels: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.label, 'listLabels', res); return []; }
  errorRate.add(false);
  return res.json().labels || [];
}

function attachLabel(h, boardId, cardId, labelId) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/boards/${boardId}/cards/${cardId}/labels`,
    JSON.stringify({ label_id: labelId }),
    h
  );
  latency.attachLabel.add(Date.now() - start);

  const ok = check(res, { 'attachLabel: 2xx': (r) => r.status >= 200 && r.status < 300 });
  if (!ok) { trackError(errors.label, 'attachLabel', res); return false; }
  errorRate.add(false);
  return true;
}

function detachLabel(h, boardId, cardId, labelId) {
  const start = Date.now();
  const res = http.del(`${BASE_URL}/api/v1/boards/${boardId}/cards/${cardId}/labels/${labelId}`, null, h);
  latency.detachLabel.add(Date.now() - start);

  const ok = check(res, { 'detachLabel: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.label, 'detachLabel', res); return false; }
  errorRate.add(false);
  return true;
}

function getCardLabels(h, boardId, cardId) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/boards/${boardId}/cards/${cardId}/labels`, h);
  latency.getCardLabels.add(Date.now() - start);

  const ok = check(res, { 'getCardLabels: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.label, 'getCardLabels', res); return []; }
  errorRate.add(false);
  return res.json().labels || [];
}

function getAvailableLabels(h, boardId) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/boards/${boardId}/available-labels`, h);
  latency.availableLabels.add(Date.now() - start);

  const ok = check(res, { 'availableLabels: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.label, 'availableLabels', res); return []; }
  errorRate.add(false);
  return res.json().labels || [];
}

// ─── Comment helpers ────────────────────────────────────────────────────

function createComment(h, cardId, boardId, content) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/cards/${cardId}/comments`,
    JSON.stringify({ board_id: boardId, content }),
    h
  );
  latency.createComment.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, {
    'createComment: 201': (r) => r.status === 201,
    'createComment: has id': () => body.comment && body.comment.id !== '',
  });

  if (!ok) { trackError(errors.comment, 'createComment', res); return null; }
  errorRate.add(false);
  return body.comment;
}

function replyComment(h, cardId, boardId, parentId, content) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/cards/${cardId}/comments`,
    JSON.stringify({ board_id: boardId, content, parent_id: parentId }),
    h
  );
  latency.replyComment.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, { 'replyComment: 201': (r) => r.status === 201 });
  if (!ok) { trackError(errors.comment, 'replyComment', res); return null; }
  errorRate.add(false);
  return body.comment;
}

function listComments(h, cardId) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/cards/${cardId}/comments`, h);
  latency.listComments.add(Date.now() - start);

  const ok = check(res, { 'listComments: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.comment, 'listComments', res); return []; }
  errorRate.add(false);
  return res.json().comments || [];
}

// ─── Checklist helpers ──────────────────────────────────────────────────

function createChecklist(h, boardId, cardId, title, position) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/boards/${boardId}/cards/${cardId}/checklists`,
    JSON.stringify({ title, position: position || 0 }),
    h
  );
  latency.createChecklist.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, {
    'createChecklist: 201': (r) => r.status === 201,
    'createChecklist: has id': () => body.checklist && body.checklist.id !== '',
  });

  if (!ok) { trackError(errors.checklist, 'createChecklist', res); return null; }
  errorRate.add(false);
  return body.checklist;
}

function createChecklistItem(h, boardId, checklistId, title, position) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/boards/${boardId}/checklists/${checklistId}/items`,
    JSON.stringify({ title, position: position || 0 }),
    h
  );
  latency.createChecklistItem.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, { 'createChecklistItem: 201': (r) => r.status === 201 });
  if (!ok) { trackError(errors.checklist, 'createChecklistItem', res); return null; }
  errorRate.add(false);
  return body.item;
}

function toggleChecklistItem(h, boardId, itemId) {
  const start = Date.now();
  const res = http.put(
    `${BASE_URL}/api/v1/boards/${boardId}/checklist-items/${itemId}/toggle`,
    null,
    h
  );
  latency.toggleItem.add(Date.now() - start);

  const ok = check(res, { 'toggleItem: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.checklist, 'toggleItem', res); return false; }
  errorRate.add(false);
  return true;
}

function listChecklists(h, boardId, cardId) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/boards/${boardId}/cards/${cardId}/checklists`, h);
  latency.listChecklists.add(Date.now() - start);

  const ok = check(res, { 'listChecklists: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.checklist, 'listChecklists', res); return []; }
  errorRate.add(false);
  return res.json().checklists || [];
}

// ─── Automation helpers ─────────────────────────────────────────────────

function createAutomation(h, boardId, name, triggerType, triggerConfig, actionType, actionConfig) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/boards/${boardId}/automations`,
    JSON.stringify({
      name,
      trigger_type: triggerType,
      trigger_config: triggerConfig,
      action_type: actionType,
      action_config: actionConfig,
    }),
    h
  );
  latency.createAutomation.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, { 'createAutomation: 201': (r) => r.status === 201 });
  if (!ok) { trackError(errors.automation, 'createAutomation', res); return null; }
  errorRate.add(false);
  return body.rule;
}

function listAutomations(h, boardId) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/boards/${boardId}/automations`, h);
  latency.listAutomations.add(Date.now() - start);

  const ok = check(res, { 'listAutomations: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.automation, 'listAutomations', res); return []; }
  errorRate.add(false);
  return res.json().rules || [];
}

// ─── Custom Field helpers ───────────────────────────────────────────────

function createCustomField(h, boardId, name, fieldType, opts) {
  const payload = { name, field_type: fieldType };
  if (opts && opts.options)  payload.options = opts.options;
  if (opts && opts.position) payload.position = opts.position;
  if (opts && opts.required) payload.required = opts.required;

  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/boards/${boardId}/custom-fields`,
    JSON.stringify(payload),
    h
  );
  latency.createCustomField.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, { 'createCustomField: 201': (r) => r.status === 201 });
  if (!ok) { trackError(errors.customField, 'createCustomField', res); return null; }
  errorRate.add(false);
  return body.field;
}

function setFieldValue(h, boardId, cardId, fieldId, value) {
  const start = Date.now();
  const res = http.put(
    `${BASE_URL}/api/v1/boards/${boardId}/cards/${cardId}/custom-fields/${fieldId}`,
    JSON.stringify(value),
    h
  );
  latency.setFieldValue.add(Date.now() - start);

  const ok = check(res, { 'setFieldValue: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.customField, 'setFieldValue', res); return false; }
  errorRate.add(false);
  return true;
}

function listCustomFields(h, boardId) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/boards/${boardId}/custom-fields`, h);
  latency.listCustomFields.add(Date.now() - start);

  const ok = check(res, { 'listCustomFields: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.customField, 'listCustomFields', res); return []; }
  errorRate.add(false);
  return res.json().fields || [];
}

function getCardFieldValues(h, boardId, cardId) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/boards/${boardId}/cards/${cardId}/custom-fields`, h);
  latency.getCardFields.add(Date.now() - start);

  const ok = check(res, { 'getCardFields: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.customField, 'getCardFields', res); return []; }
  errorRate.add(false);
  return res.json().values || [];
}

// ─── Card Link helpers ──────────────────────────────────────────────────

function linkCards(h, boardId, parentCardId, childCardId) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/boards/${boardId}/cards/${parentCardId}/links`,
    JSON.stringify({ child_id: childCardId }),
    h
  );
  latency.linkCards.add(Date.now() - start);

  const ok = check(res, { 'linkCards: 201': (r) => r.status === 201 });
  if (!ok) { trackError(errors.cardLink, 'linkCards', res); return false; }
  errorRate.add(false);
  return true;
}

function getChildren(h, boardId, cardId) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/boards/${boardId}/cards/${cardId}/children`, h);
  latency.getChildren.add(Date.now() - start);

  const ok = check(res, { 'getChildren: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.cardLink, 'getChildren', res); return []; }
  errorRate.add(false);
  return res.json().links || [];
}

// ─── Board Settings helpers ─────────────────────────────────────────────

function getBoardSettings(h, boardId) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/boards/${boardId}/settings`, h);
  latency.getSettings.add(Date.now() - start);

  const ok = check(res, { 'getSettings: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.settings, 'getSettings', res); return null; }
  errorRate.add(false);
  return res.json();
}

function updateBoardSettings(h, boardId, useBoardLabelsOnly) {
  const start = Date.now();
  const res = http.put(
    `${BASE_URL}/api/v1/boards/${boardId}/settings`,
    JSON.stringify({ use_board_labels_only: useBoardLabelsOnly }),
    h
  );
  latency.updateSettings.add(Date.now() - start);

  const ok = check(res, { 'updateSettings: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.settings, 'updateSettings', res); return false; }
  errorRate.add(false);
  return true;
}

// ─── User Label helpers ─────────────────────────────────────────────────

function createUserLabel(h, name, color) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/user-labels`,
    JSON.stringify({ name, color }),
    h
  );
  latency.createUserLabel.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, { 'createUserLabel: 201': (r) => r.status === 201 });
  if (!ok) { trackError(errors.userLabel, 'createUserLabel', res); return null; }
  errorRate.add(false);
  return body.label;
}

function listUserLabels(h) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/user-labels`, h);
  latency.listUserLabels.add(Date.now() - start);

  const ok = check(res, { 'listUserLabels: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.userLabel, 'listUserLabels', res); return []; }
  errorRate.add(false);
  return res.json().labels || [];
}

// ─── Assign helpers ─────────────────────────────────────────────────────

function assignCard(h, cardId, assigneeId, boardId) {
  const start = Date.now();
  const res = http.put(
    `${BASE_URL}/api/v1/cards/${cardId}/assign`,
    JSON.stringify({ assignee_id: assigneeId, board_id: boardId }),
    h
  );
  latency.assignCard.add(Date.now() - start);

  const ok = check(res, { 'assignCard: 200': (r) => r.status === 200 });
  if (!ok) { trackError(errors.card, 'assignCard', res); return false; }
  errorRate.add(false);
  return true;
}

// ─── Template helpers ───────────────────────────────────────────────────

function createBoardTemplate(h, name, description, columnsData, labelsData) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/board-templates`,
    JSON.stringify({
      name,
      description: description || '',
      columns_data: columnsData,
      labels_data: labelsData,
    }),
    h
  );
  latency.createTemplate.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, { 'createBoardTemplate: 201': (r) => r.status === 201 });
  if (!ok) { trackError(errors.template, 'createBoardTemplate', res); return null; }
  errorRate.add(false);
  return body.template;
}

function createBoardFromTemplate(h, templateId, title) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/boards/from-template`,
    JSON.stringify({ template_id: templateId, title }),
    h
  );
  latency.createFromTemplate.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, { 'createBoardFromTemplate: 201': (r) => r.status === 201 });
  if (!ok) { trackError(errors.template, 'createBoardFromTemplate', res); return null; }
  errorRate.add(false);
  return body.board;
}

function listBoardTemplates(h) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/board-templates`, h);
  latency.createTemplate.add(Date.now() - start);

  if (res.status !== 200) return [];
  errorRate.add(false);
  return res.json().templates || [];
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

function randomDueDate() {
  const d = new Date();
  d.setDate(d.getDate() + Math.floor(Math.random() * 14) + 1);
  return d.toISOString();
}

// ═══════════════════════════════════════════════════════════════════════════
// ─── СЦЕНАРИИ ПОЛЬЗОВАТЕЛЕЙ ─────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════

// Выбрать случайную доску где юзер — owner или member
function pickBoard(boards, userId) {
  const myBoards = boards.filter(b => b.ownerId === userId || b.memberIds.includes(userId));
  if (myBoards.length > 0) return pickRandom(myBoards);
  return pickRandom(boards); // fallback — случайная (доступ проверит сервер)
}

function getTokenForBoard(board, userId, userToken) {
  if (board.ownerId === userId) return userToken;
  const idx = board.memberIds.indexOf(userId);
  if (idx >= 0) return board.memberTokens[idx];
  return board.ownerToken; // fallback
}

// ─── Worker (55%): работа с существующей доской ─────────────────────────

function workerScenario(me, allUsers, boards, h) {
  const board = pickBoard(boards, me.id);
  const token = getTokenForBoard(board, me.id, me.token);
  const bh = authHeaders(token);

  // Посмотреть доску
  getBoard(bh, board.id);
  sleep(randomBetween(0.2, 0.5));

  // Создать карточку в случайной колонке
  if (board.columnIds.length > 0) {
    const colId = pickRandom(board.columnIds);
    const card = createCard(bh, colId, board.id, `Новая задача ${Date.now() % 10000}`, Math.floor(Math.random() * 26), {
      priority: pickRandom(PRIORITIES),
      task_type: pickRandom(TASK_TYPES),
      due_date: Math.random() < 0.6 ? randomDueDate() : undefined,
    });
    sleep(randomBetween(0.2, 0.5));

    if (card) {
      // Привязать метку
      if (board.labelIds.length > 0 && Math.random() < 0.6) {
        attachLabel(bh, board.id, card.id, pickRandom(board.labelIds));
        sleep(randomBetween(0.05, 0.15));
      }

      // Написать комментарий (~50%)
      if (Math.random() < 0.5) {
        createComment(bh, card.id, board.id, `Работаю над: ${card.title}`);
        sleep(randomBetween(0.1, 0.3));
      }

      // Создать чеклист (~30%)
      if (Math.random() < 0.3) {
        const cl = createChecklist(bh, board.id, card.id, 'TODO', 0);
        if (cl) {
          for (let i = 0; i < Math.floor(randomBetween(3, 6)); i++) {
            createChecklistItem(bh, board.id, cl.id, `Шаг ${i + 1}`, i);
            sleep(randomBetween(0.03, 0.08));
          }
        }
        sleep(randomBetween(0.1, 0.2));
      }
    }
  }

  // Переместить существующую карточку (~40%)
  if (Math.random() < 0.4 && board.cardIds.length > 0 && board.columnIds.length > 1) {
    const card = pickRandom(board.cardIds);
    const targetCol = pickRandom(board.columnIds.filter(c => c !== card.column_id));
    if (targetCol) {
      moveCard(bh, card.id, card.column_id, targetCol, board.id, card.version);
      sleep(randomBetween(0.2, 0.5));
    }
  }

  // Обновить существующую карточку (~30%)
  if (Math.random() < 0.3 && board.cardIds.length > 0) {
    const card = pickRandom(board.cardIds);
    updateCard(bh, card.id, board.id, `Updated ${Date.now() % 10000}`, card.version);
    sleep(randomBetween(0.2, 0.4));
  }

  // Назначить карточку (~20%)
  if (Math.random() < 0.2 && board.cardIds.length > 0 && board.memberIds.length > 0) {
    const card = pickRandom(board.cardIds);
    assignCard(bh, card.id, pickRandom(board.memberIds), board.id);
    sleep(randomBetween(0.1, 0.2));
  }

  // Нотификации
  checkNotifications(bh);
  markAllRead(bh);
  sleep(randomBetween(0.2, 0.5));
}

// ─── Collaborator (15%): работа на чужих досках ─────────────────────────

function collaboratorScenario(me, allUsers, boards, h) {
  const board = pickBoard(boards, me.id);
  const token = getTokenForBoard(board, me.id, me.token);
  const bh = authHeaders(token);

  // Получить доску
  const boardInfo = getBoard(bh, board.id);
  if (!boardInfo) return;
  sleep(randomBetween(0.3, 0.5));

  // Прочитать метки
  listLabels(bh, board.id);
  sleep(randomBetween(0.1, 0.3));

  // Посмотреть available labels
  getAvailableLabels(bh, board.id);
  sleep(randomBetween(0.1, 0.3));

  // Работа с карточками
  if (board.cardIds.length > 0) {
    const card = pickRandom(board.cardIds);

    // Прочитать комментарии и ответить
    const commData = softGet(bh, `${BASE_URL}/api/v1/cards/${card.id}/comments`, latency.listComments);
    const comments = commData ? (commData.comments || []) : [];
    sleep(randomBetween(0.2, 0.5));

    if (comments.length > 0) {
      replyComment(bh, card.id, board.id, comments[0].id, `Ответ от ${me.name}`);
    } else {
      createComment(bh, card.id, board.id, `${me.name}: начинаю работу`);
    }
    sleep(randomBetween(0.2, 0.5));

    // Прочитать чеклисты и toggle
    const clData = softGet(bh, `${BASE_URL}/api/v1/boards/${board.id}/cards/${card.id}/checklists`, latency.listChecklists);
    const checklists = clData ? (clData.checklists || []) : [];
    if (checklists.length > 0 && checklists[0].items && checklists[0].items.length > 0) {
      const unchecked = checklists[0].items.filter(i => !i.is_checked);
      if (unchecked.length > 0) {
        toggleChecklistItem(bh, board.id, unchecked[0].id);
      }
    }
    sleep(randomBetween(0.1, 0.3));

    // Прочитать метки карточки
    softGet(bh, `${BASE_URL}/api/v1/boards/${board.id}/cards/${card.id}/labels`, latency.getCardLabels);
    sleep(randomBetween(0.1, 0.3));

    // Переместить карточку
    if (board.columnIds.length > 1) {
      const targetCol = pickRandom(board.columnIds.filter(c => c !== card.column_id));
      if (targetCol) {
        moveCard(bh, card.id, card.column_id, targetCol, board.id, card.version);
      }
    }
    sleep(randomBetween(0.2, 0.5));
  }

  // Нотификации
  checkNotifications(bh);
  markAllRead(bh);
  sleep(randomBetween(0.2, 0.5));
}

// ─── Reader (15%): чтение данных досок ──────────────────────────────────

function readerScenario(me, allUsers, boards, h) {
  const board = pickBoard(boards, me.id);
  const token = getTokenForBoard(board, me.id, me.token);
  const bh = authHeaders(token);

  // Список досок
  listBoards(bh);
  sleep(randomBetween(0.3, 0.5));

  // Полная загрузка доски
  getBoard(bh, board.id);
  sleep(randomBetween(0.3, 0.5));

  // Метки
  softGet(bh, `${BASE_URL}/api/v1/boards/${board.id}/labels`, latency.listLabels);
  sleep(randomBetween(0.2, 0.4));

  // Available labels
  softGet(bh, `${BASE_URL}/api/v1/boards/${board.id}/available-labels`, latency.availableLabels);
  sleep(randomBetween(0.2, 0.4));

  // Автоматизации
  softGet(bh, `${BASE_URL}/api/v1/boards/${board.id}/automations`, latency.listAutomations);
  sleep(randomBetween(0.2, 0.4));

  // Custom fields
  softGet(bh, `${BASE_URL}/api/v1/boards/${board.id}/custom-fields`, latency.listCustomFields);
  sleep(randomBetween(0.2, 0.4));

  // Настройки доски
  softGet(bh, `${BASE_URL}/api/v1/boards/${board.id}/settings`, latency.getSettings);
  sleep(randomBetween(0.2, 0.3));

  // Детали карточек
  if (board.cardIds.length > 0) {
    const card = pickRandom(board.cardIds);

    softGet(bh, `${BASE_URL}/api/v1/boards/${board.id}/cards/${card.id}/labels`, latency.getCardLabels);
    sleep(randomBetween(0.1, 0.3));

    softGet(bh, `${BASE_URL}/api/v1/boards/${board.id}/cards/${card.id}/checklists`, latency.listChecklists);
    sleep(randomBetween(0.1, 0.3));

    softGet(bh, `${BASE_URL}/api/v1/cards/${card.id}/comments`, latency.listComments);
    sleep(randomBetween(0.1, 0.3));

    softGet(bh, `${BASE_URL}/api/v1/boards/${board.id}/cards/${card.id}/custom-fields`, latency.getCardFields);
    sleep(randomBetween(0.1, 0.3));

    softGet(bh, `${BASE_URL}/api/v1/boards/${board.id}/cards/${card.id}/children`, latency.getChildren);
    sleep(randomBetween(0.1, 0.3));
  }

  // Нотификации
  checkNotifications(bh);
  sleep(randomBetween(0.2, 0.4));
  getUnreadCount(bh);
  sleep(randomBetween(0.3, 0.5));
}

// ─── Heavy Admin (7%): массовые операции на доске ───────────────────────

function heavyUserScenario(me, allUsers, boards, h) {
  // Берём доску где мы owner
  const ownerBoards = boards.filter(b => b.ownerId === me.id);
  const board = ownerBoards.length > 0 ? pickRandom(ownerBoards) : pickRandom(boards);
  const bh = authHeaders(board.ownerToken);

  // Создать автоматизации
  createAutomation(
    bh, board.id,
    `Auto ${Date.now() % 10000}: Done → Low`,
    'card_moved_to_column', { column_name: 'Done' },
    'set_priority', { priority: 'low' }
  );
  sleep(randomBetween(0.1, 0.3));

  // Создать custom fields
  const textField = createCustomField(bh, board.id, `Sprint-${Date.now() % 10000}`, 'text', { position: 0 });
  sleep(randomBetween(0.1, 0.2));
  const dropdownField = createCustomField(bh, board.id, `Severity-${Date.now() % 10000}`, 'dropdown', {
    position: 1,
    options: ['Critical', 'Major', 'Minor', 'Trivial'],
  });
  sleep(randomBetween(0.1, 0.2));

  // Обновить настройки
  updateBoardSettings(bh, board.id, Math.random() < 0.5);
  sleep(randomBetween(0.1, 0.2));

  // Массовое создание карточек (5-10 штук)
  const cards = [];
  if (board.columnIds.length > 0) {
    const colId = pickRandom(board.columnIds);
    const count = Math.floor(randomBetween(5, 11));
    for (let i = 0; i < count; i++) {
      const card = createCard(bh, colId, board.id, `Heavy Task ${Date.now() % 10000}-${i}`, i, {
        priority: PRIORITIES[i % PRIORITIES.length],
        task_type: TASK_TYPES[i % TASK_TYPES.length],
        due_date: randomDueDate(),
      });
      if (card) cards.push(card);
      sleep(randomBetween(0.05, 0.15));
    }
  }

  // Привязать метки ко всем новым картам
  for (const card of cards) {
    if (board.labelIds.length > 0) {
      const count = Math.floor(randomBetween(2, 4));
      const shuffled = board.labelIds.slice().sort(() => Math.random() - 0.5);
      for (let i = 0; i < count && i < shuffled.length; i++) {
        attachLabel(bh, board.id, card.id, shuffled[i]);
        sleep(randomBetween(0.03, 0.08));
      }
    }
  }

  // Чеклисты на новых картах
  for (let i = 0; i < Math.min(3, cards.length); i++) {
    const cl = createChecklist(bh, board.id, cards[i].id, `Чеклист ${i + 1}`, 0);
    if (cl) {
      const items = [];
      for (let j = 0; j < Math.floor(randomBetween(4, 7)); j++) {
        const item = createChecklistItem(bh, board.id, cl.id, `Шаг ${j + 1}`, j);
        if (item) items.push(item);
        sleep(randomBetween(0.03, 0.08));
      }
      // Toggle половину
      for (let j = 0; j < Math.floor(items.length / 2); j++) {
        toggleChecklistItem(bh, board.id, items[j].id);
        sleep(randomBetween(0.03, 0.08));
      }
    }
    sleep(randomBetween(0.1, 0.2));
  }

  // Комментарии
  for (let i = 0; i < Math.min(3, cards.length); i++) {
    createComment(bh, cards[i].id, board.id, `Описание задачи ${i + 1}`);
    sleep(randomBetween(0.1, 0.2));
  }

  // Set custom field values
  if (textField) {
    for (const card of cards) {
      setFieldValue(bh, board.id, card.id, textField.id, { value_text: 'Sprint 12' });
      sleep(randomBetween(0.03, 0.08));
    }
  }
  if (dropdownField) {
    for (const card of cards) {
      setFieldValue(bh, board.id, card.id, dropdownField.id, {
        value_text: pickRandom(['Critical', 'Major', 'Minor', 'Trivial']),
      });
      sleep(randomBetween(0.03, 0.08));
    }
  }

  // Card links
  if (cards.length >= 3) {
    linkCards(bh, board.id, cards[0].id, cards[1].id);
    sleep(randomBetween(0.05, 0.1));
    linkCards(bh, board.id, cards[0].id, cards[2].id);
    sleep(randomBetween(0.05, 0.1));
    getChildren(bh, board.id, cards[0].id);
    sleep(randomBetween(0.1, 0.2));
  }

  // Перемещения
  if (board.columnIds.length > 1) {
    for (let i = 0; i < Math.min(3, cards.length); i++) {
      const targetCol = pickRandom(board.columnIds);
      moveCard(bh, cards[i].id, board.columnIds[0], targetCol, board.id, cards[i].version || 1);
      sleep(randomBetween(0.1, 0.3));
    }
  }

  // Чтение member'ом
  if (board.memberTokens.length > 0) {
    const mh = authHeaders(board.memberTokens[0]);
    getAvailableLabels(mh, board.id);
    sleep(randomBetween(0.1, 0.2));
    listAutomations(mh, board.id);
    sleep(randomBetween(0.1, 0.2));
    getBoardSettings(mh, board.id);
    sleep(randomBetween(0.1, 0.2));
  }
}

// ─── Setup-once (8%): шаблоны и глобальные метки ────────────────────────

function setupOnceScenario(me, allUsers, boards, h) {
  const suffix = Date.now() % 100000;

  // Глобальные метки
  const globalPresets = [
    { name: `Мой приоритет ${suffix}`, color: '#E53E3E' },
    { name: `Следить ${suffix}`,       color: '#3182CE' },
    { name: `Важное ${suffix}`,        color: '#D69E2E' },
  ];
  for (let i = 0; i < Math.floor(randomBetween(2, 4)); i++) {
    createUserLabel(h, globalPresets[i].name, globalPresets[i].color);
    sleep(randomBetween(0.1, 0.3));
  }

  listUserLabels(h);
  sleep(randomBetween(0.2, 0.4));

  // Создать board template
  const tmpl = createBoardTemplate(h, `Template ${suffix}`, 'Шаблон из теста', [
    { title: 'Backlog', position: 0 },
    { title: 'Sprint', position: 1 },
    { title: 'In Progress', position: 2 },
    { title: 'Review', position: 3 },
    { title: 'Done', position: 4 },
  ], [
    { name: 'Bug', color: '#E53E3E' },
    { name: 'Feature', color: '#38A169' },
  ]);
  sleep(randomBetween(0.3, 0.5));

  if (tmpl) {
    const newBoard = createBoardFromTemplate(h, tmpl.id, `${me.name}'s Board from template`);
    sleep(randomBetween(0.3, 0.5));

    if (newBoard) {
      getBoard(h, newBoard.id);
      sleep(randomBetween(0.2, 0.3));
      listLabels(h, newBoard.id);
      sleep(randomBetween(0.2, 0.3));
      getAvailableLabels(h, newBoard.id);
      sleep(randomBetween(0.2, 0.3));
    }
  }

  listBoardTemplates(h);
  sleep(randomBetween(0.2, 0.4));

  checkNotifications(h);
  sleep(randomBetween(0.2, 0.4));
}

// ═══════════════════════════════════════════════════════════════════════════
// ─── Main ───────────────────────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════

export default function (data) {
  const allUsers = data.users;
  const boards = data.boards;

  if (!allUsers || allUsers.length === 0 || !boards || boards.length === 0) {
    console.error('No users or boards from setup!');
    return;
  }

  const iterGlobal = exec.scenario.iterationInTest;
  const userIdx = ((exec.vu.idInTest - 1) + iterGlobal) % allUsers.length;
  const me = allUsers[userIdx];
  const h = authHeaders(me.token);

  // Распределение: 55% worker, 15% collaborator, 15% reader, 7% heavy, 8% setup-once
  const roll = iterGlobal % 100;

  if (roll < 55) {
    workerScenario(me, allUsers, boards, h);
  } else if (roll < 70) {
    collaboratorScenario(me, allUsers, boards, h);
  } else if (roll < 85) {
    readerScenario(me, allUsers, boards, h);
  } else if (roll < 92) {
    heavyUserScenario(me, allUsers, boards, h);
  } else {
    setupOnceScenario(me, allUsers, boards, h);
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// ─── Summary ────────────────────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════

export function handleSummary(data) {
  function val(name, field) {
    const m = data.metrics[name];
    if (!m) return 'N/A';
    if (field === 'count') return m.values.count || 0;
    if (field === 'rate') return (m.values.rate * 100).toFixed(1) + '%';
    return m.values[field] ? m.values[field].toFixed(0) : 'N/A';
  }

  function pad(v) { return String(v).padStart(7); }

  const summary = `
╔════════════════════════════════════════════════════════════════════════╗
║  НАГРУЗОЧНЫЙ ТЕСТ: 1000 юзеров, 1-3 доски/юзер (1-8 кол, 0-15 карт) ║
╠════════════════════════════════════════════════════════════════════════╣
║                                                                        ║
║  Распределение:                                                        ║
║    55% workers / 15% collab / 15% readers / 7% heavy / 8% setup       ║
║                                                                        ║
║  ── Core Latency (p95) ──────────────────────────────────────────────  ║
║  Создание доски:       ${pad(val('latency_create_board_ms', 'p(95)'))} ms                               ║
║  Добавление участника: ${pad(val('latency_add_member_ms', 'p(95)'))} ms                               ║
║  Создание колонки:     ${pad(val('latency_create_column_ms', 'p(95)'))} ms                               ║
║  Создание карточки:    ${pad(val('latency_create_card_ms', 'p(95)'))} ms                               ║
║  Перемещение карточки: ${pad(val('latency_move_card_ms', 'p(95)'))} ms                               ║
║  Назначение карты:     ${pad(val('latency_assign_card_ms', 'p(95)'))} ms                               ║
║                                                                        ║
║  ── Labels (p95) ────────────────────────────────────────────────────  ║
║  Создание метки:       ${pad(val('latency_create_label_ms', 'p(95)'))} ms                               ║
║  Привязка метки:       ${pad(val('latency_attach_label_ms', 'p(95)'))} ms                               ║
║  Available labels:     ${pad(val('latency_available_labels_ms', 'p(95)'))} ms                               ║
║                                                                        ║
║  ── Comments (p95) ──────────────────────────────────────────────────  ║
║  Создание коммент.:    ${pad(val('latency_create_comment_ms', 'p(95)'))} ms                               ║
║  Ответ на коммент.:    ${pad(val('latency_reply_comment_ms', 'p(95)'))} ms                               ║
║  Список комментариев:  ${pad(val('latency_list_comments_ms', 'p(95)'))} ms                               ║
║                                                                        ║
║  ── Checklists (p95) ────────────────────────────────────────────────  ║
║  Создание чеклиста:    ${pad(val('latency_create_checklist_ms', 'p(95)'))} ms                               ║
║  Toggle item:          ${pad(val('latency_toggle_item_ms', 'p(95)'))} ms                               ║
║                                                                        ║
║  ── Automation & Fields (p95) ───────────────────────────────────────  ║
║  Создание автоматиз.:  ${pad(val('latency_create_automation_ms', 'p(95)'))} ms                               ║
║  Создание custom field:${pad(val('latency_create_custom_field_ms', 'p(95)'))} ms                               ║
║  Set field value:      ${pad(val('latency_set_field_value_ms', 'p(95)'))} ms                               ║
║  Card links:           ${pad(val('latency_link_cards_ms', 'p(95)'))} ms                               ║
║                                                                        ║
║  ── Templates (p95) ─────────────────────────────────────────────────  ║
║  Создание шаблона:     ${pad(val('latency_create_template_ms', 'p(95)'))} ms                               ║
║  Из шаблона:           ${pad(val('latency_create_from_template_ms', 'p(95)'))} ms                               ║
║  User labels:          ${pad(val('latency_create_user_label_ms', 'p(95)'))} ms                               ║
║                                                                        ║
║  ── Notifications (p95) ─────────────────────────────────────────────  ║
║  Нотификации:          ${pad(val('latency_notifications_ms', 'p(95)'))} ms                               ║
║  Notif delivery:       ${pad(val('latency_notif_delivery_ms', 'p(95)'))} ms                               ║
║  Board settings:       ${pad(val('latency_get_settings_ms', 'p(95)'))} ms                               ║
║                                                                        ║
║  ── Ошибки ──────────────────────────────────────────────────────────  ║
║  Error rate:           ${pad(val('error_rate', 'rate'))}                                  ║
║  Board:                ${pad(val('errors_board', 'count'))}    Label:         ${pad(val('errors_label', 'count'))}        ║
║  Card:                 ${pad(val('errors_card', 'count'))}    Comment:       ${pad(val('errors_comment', 'count'))}        ║
║  Member:               ${pad(val('errors_member', 'count'))}    Checklist:     ${pad(val('errors_checklist', 'count'))}        ║
║  Notification:         ${pad(val('errors_notification', 'count'))}    Automation:    ${pad(val('errors_automation', 'count'))}        ║
║  Custom Field:         ${pad(val('errors_custom_field', 'count'))}    Template:      ${pad(val('errors_template', 'count'))}        ║
║  Card Link:            ${pad(val('errors_card_link', 'count'))}    Settings:      ${pad(val('errors_settings', 'count'))}        ║
║  User Label:           ${pad(val('errors_user_label', 'count'))}                                        ║
║                                                                        ║
║  ── HTTP ────────────────────────────────────────────────────────────  ║
║  Total requests:       ${pad(data.metrics.http_reqs ? data.metrics.http_reqs.values.count : 0)}                                  ║
║  Failed requests:      ${pad(val('http_req_failed', 'rate'))}                                  ║
║  Duration p95:         ${pad(val('http_req_duration', 'p(95)'))} ms                               ║
║                                                                        ║
╚════════════════════════════════════════════════════════════════════════╝
`;

  return { stdout: summary };
}
