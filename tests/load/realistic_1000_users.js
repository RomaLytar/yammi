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
const MEMBERS_PER_BOARD = parseInt(__ENV.MEMBERS || '2');

export const options = {
  setupTimeout: '180s',
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
  const ok = check(res, {
    'replyComment: 201': (r) => r.status === 201,
  });

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
  const ok = check(res, {
    'createChecklistItem: 201': (r) => r.status === 201,
  });

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
  const ok = check(res, {
    'createAutomation: 201': (r) => r.status === 201,
  });

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
  const ok = check(res, {
    'createCustomField: 201': (r) => r.status === 201,
  });

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
  const ok = check(res, {
    'createUserLabel: 201': (r) => r.status === 201,
  });

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

function createCardTemplate(h, boardId, name, title, priority, taskType) {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/api/v1/boards/${boardId}/card-templates`,
    JSON.stringify({
      name,
      title: title || name,
      description: 'Template card',
      priority: priority || 'medium',
      task_type: taskType || 'task',
    }),
    h
  );
  latency.createTemplate.add(Date.now() - start);

  const body = res.json();
  const ok = check(res, { 'createCardTemplate: 201': (r) => r.status === 201 });
  if (!ok) { trackError(errors.template, 'createCardTemplate', res); return null; }
  errorRate.add(false);
  return body.template;
}

function listBoardTemplates(h) {
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/v1/board-templates`, h);
  latency.createTemplate.add(Date.now() - start); // reuse trend for read too

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

function measureDelivery(ownerH, memberH, boardID) {
  if (!boardID) return;

  markAllRead(memberH);
  sleep(0.2);
  const beforeCount = getUnreadCount(memberH);
  if (beforeCount < 0) return;

  const actionTime = Date.now();
  const col = createColumn(ownerH, boardID, `measure-${Date.now()}`);
  if (!col) return;

  const maxPollMs = 10000;
  const pollInterval = 200;

  for (let elapsed = 0; elapsed < maxPollMs; elapsed += pollInterval) {
    sleep(pollInterval / 1000);
    const count = getUnreadCount(memberH);
    if (count > beforeCount) {
      latency.notifDelivery.add(Date.now() - actionTime);
      deleteColumn(ownerH, col.id, boardID);
      return;
    }
  }

  latency.notifDelivery.add(maxPollMs);
  deleteColumn(ownerH, col.id, boardID);
}

// ─── Генерация due_date (через 1-14 дней от сейчас) ────────────────────

function randomDueDate() {
  const d = new Date();
  d.setDate(d.getDate() + Math.floor(Math.random() * 14) + 1);
  return d.toISOString();
}

// ═══════════════════════════════════════════════════════════════════════════
// ─── СЦЕНАРИИ ПОЛЬЗОВАТЕЛЕЙ ─────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════

// ─── Worker (60%): полный цикл работы с доской ─────────────────────────

function workerScenario(me, allUsers, h) {
  // 1. Создание доски
  const board = createBoard(h, `Доска ${me.name}`);
  if (!board) return;
  sleep(randomBetween(0.3, 0.8));

  // 2. Добавление участников
  const memberCount = MEMBERS_PER_BOARD > 2 ? MEMBERS_PER_BOARD : (Math.random() < 0.5 ? 1 : 2);
  const members = pickRandomUsers(allUsers, me.id, memberCount);
  for (const m of members) {
    addMember(h, board.id, m.id, 'member');
    sleep(randomBetween(0.1, 0.2));
  }

  // 3. Создание 5-7 меток на доске
  const labelCount = Math.floor(randomBetween(5, 8));
  const boardLabels = [];
  for (let i = 0; i < labelCount && i < LABEL_PRESETS.length; i++) {
    const lbl = createLabel(h, board.id, LABEL_PRESETS[i].name, LABEL_PRESETS[i].color);
    if (lbl) boardLabels.push(lbl);
    sleep(randomBetween(0.05, 0.15));
  }

  // 4. ~5% создают автоматизации (1-2 правила)
  if (Math.random() < 0.05) {
    createAutomation(
      h, board.id,
      'Done → Low Priority',
      'card_moved_to_column', { column_name: 'Done' },
      'set_priority', { priority: 'low' }
    );
    sleep(randomBetween(0.1, 0.3));

    if (Math.random() < 0.5) {
      createAutomation(
        h, board.id,
        'New card → add Task label',
        'card_created', {},
        'add_label', { label_name: 'Task' }
      );
      sleep(randomBetween(0.1, 0.3));
    }
  }

  // 5. ~10% обновляют настройки доски
  if (Math.random() < 0.1) {
    updateBoardSettings(h, board.id, true);
    sleep(randomBetween(0.1, 0.2));
  }

  // 6. Создание колонок
  const col1 = createColumn(h, board.id, 'To Do');
  sleep(randomBetween(0.2, 0.5));
  const col2 = createColumn(h, board.id, 'In Progress');
  sleep(randomBetween(0.2, 0.5));
  const col3 = createColumn(h, board.id, 'Done');
  sleep(randomBetween(0.2, 0.5));
  if (!col1 || !col2) return;

  // 7. Создание карт с metadata
  const cardCount = Math.floor(randomBetween(2, 4));
  const cards = [];
  for (let i = 0; i < cardCount; i++) {
    const card = createCard(h, col1.id, board.id, `Задача ${i + 1}`, i, {
      priority: pickRandom(PRIORITIES),
      task_type: pickRandom(TASK_TYPES),
      due_date: Math.random() < 0.6 ? randomDueDate() : undefined,
    });
    if (card) cards.push(card);
    sleep(randomBetween(0.2, 0.5));
  }

  // 8. Привязать метки к картам (~60% карт получают 1-3 метки)
  for (const card of cards) {
    if (Math.random() < 0.6 && boardLabels.length > 0) {
      const labelsToPick = Math.floor(randomBetween(1, Math.min(4, boardLabels.length + 1)));
      const shuffled = boardLabels.slice().sort(() => Math.random() - 0.5);
      for (let i = 0; i < labelsToPick; i++) {
        attachLabel(h, board.id, card.id, shuffled[i].id);
        sleep(randomBetween(0.05, 0.1));
      }
    }
  }

  // 9. Создать чеклист на ~40% карт (3-5 items)
  for (const card of cards) {
    if (Math.random() < 0.4) {
      const cl = createChecklist(h, board.id, card.id, 'Задачи', 0);
      if (cl) {
        const itemCount = Math.floor(randomBetween(3, 6));
        const items = [];
        for (let i = 0; i < itemCount; i++) {
          const item = createChecklistItem(h, board.id, cl.id, `Пункт ${i + 1}`, i);
          if (item) items.push(item);
          sleep(randomBetween(0.05, 0.1));
        }
        // Toggle 2-3 items (~30%)
        if (Math.random() < 0.3) {
          for (let i = 0; i < Math.min(3, items.length); i++) {
            toggleChecklistItem(h, board.id, items[i].id);
            sleep(randomBetween(0.05, 0.1));
          }
        }
      }
      sleep(randomBetween(0.1, 0.3));
    }
  }

  // 10. Написать комментарий к карте (~40%)
  const commentsMap = {}; // cardId → comment
  for (const card of cards) {
    if (Math.random() < 0.4) {
      const comment = createComment(h, card.id, board.id, `Комментарий к задаче ${card.title}`);
      if (comment) commentsMap[card.id] = comment;
      sleep(randomBetween(0.1, 0.3));
    }
  }

  // 11. Назначить карту на участника (~50%)
  for (const card of cards) {
    if (Math.random() < 0.5 && members.length > 0) {
      assignCard(h, card.id, pickRandom(members).id, board.id);
      sleep(randomBetween(0.1, 0.2));
    }
  }

  // 12. Перемещение карт
  if (cards.length > 0) {
    moveCard(h, cards[0], col2.id, board.id);
    cards[0].column_id = col2.id;
    sleep(randomBetween(0.3, 0.8));
  }

  // 13. Замер реальной latency доставки нотификации
  if (Math.random() < 0.2 && members.length > 0) {
    const memberH = auth(members[0].token);
    measureDelivery(h, memberH, board.id);
  } else {
    checkNotifications(h);
    markAllRead(h);
  }
  sleep(randomBetween(0.3, 0.8));

  // 14. ~50% удаляют доску
  if (Math.random() < 0.5) {
    deleteBoard(h, board.id);
  }
}

// ─── Collaborator (15%): работа на чужой доске ──────────────────────────

function collaboratorScenario(me, allUsers, h) {
  // Найти доску где мы участник
  const boards = listBoards(h);
  sleep(randomBetween(0.5, 1));

  if (boards.length === 0) {
    // Нет досок — создадим мини-доску и поработаем
    const board = createBoard(h, `Collab-${me.name}`);
    if (!board) return;

    const col = createColumn(h, board.id, 'Tasks');
    if (!col) return;

    const card = createCard(h, col.id, board.id, 'Collab task', 0, {
      priority: 'medium', task_type: 'task',
    });
    if (card) {
      createComment(h, card.id, board.id, 'Начинаю работу над задачей');
    }
    sleep(randomBetween(0.5, 1));
    return;
  }

  const boardInfo = getBoard(h, boards[0].id);
  if (!boardInfo || !boardInfo.board) return;
  const board = boardInfo.board;
  sleep(randomBetween(0.3, 0.8));

  // Получить колонки и карточки (softGet — доска может быть удалена)
  const colData = softGet(h, `${BASE_URL}/api/v1/boards/${board.id}/columns`);
  const columns = colData ? (colData.columns || []) : [];
  if (columns.length === 0) return;
  sleep(randomBetween(0.2, 0.5));

  const cardData = softGet(h, `${BASE_URL}/api/v1/columns/${columns[0].id}/cards`);
  const cards = cardData ? (cardData.cards || []) : [];
  sleep(randomBetween(0.2, 0.5));

  if (cards.length > 0) {
    const card = cards[0];

    // Прочитать комментарии и ответить на них
    const commData = softGet(h, `${BASE_URL}/api/v1/cards/${card.id}/comments`, latency.listComments);
    const comments = commData ? (commData.comments || []) : [];
    sleep(randomBetween(0.2, 0.5));

    if (comments.length > 0) {
      replyComment(h, card.id, board.id, comments[0].id, `Ответ от ${me.name}: согласен, работаю`);
      sleep(randomBetween(0.2, 0.5));
    } else {
      createComment(h, card.id, board.id, `${me.name}: взял задачу в работу`);
      sleep(randomBetween(0.2, 0.5));
    }

    // Прочитать чеклисты и toggle items
    const clData = softGet(h, `${BASE_URL}/api/v1/boards/${board.id}/cards/${card.id}/checklists`, latency.listChecklists);
    const checklists = clData ? (clData.checklists || []) : [];
    sleep(randomBetween(0.1, 0.3));
    if (checklists.length > 0 && checklists[0].items && checklists[0].items.length > 0) {
      const unchecked = checklists[0].items.filter(i => !i.is_checked);
      if (unchecked.length > 0) {
        toggleChecklistItem(h, board.id, unchecked[0].id);
        sleep(randomBetween(0.1, 0.3));
      }
    }

    // Прочитать метки карточки
    softGet(h, `${BASE_URL}/api/v1/boards/${board.id}/cards/${card.id}/labels`, latency.getCardLabels);
    sleep(randomBetween(0.1, 0.3));

    // Переместить карточку если есть вторая колонка
    if (columns.length > 1) {
      moveCard(h, card, columns[1].id, board.id);
      sleep(randomBetween(0.3, 0.8));
    }
  }

  // Получить available labels
  softGet(h, `${BASE_URL}/api/v1/boards/${board.id}/available-labels`, latency.availableLabels);
  sleep(randomBetween(0.1, 0.3));

  // Нотификации
  checkNotifications(h);
  markAllRead(h);
  sleep(randomBetween(0.3, 0.8));
}

// ─── Reader (15%): чтение данных доски ──────────────────────────────────

// Мягкие GET-хелперы для reader/collaborator: 403/404 — нормальная гонка, не ошибка
function softGet(h, url, latencyTrend) {
  const start = Date.now();
  const res = http.get(url, h);
  if (latencyTrend) latencyTrend.add(Date.now() - start);
  if (res.status === 200) { errorRate.add(false); return res.json(); }
  // 403/404 — доска удалена или участник удалён другим VU — не считаем ошибкой
  return null;
}

function readerScenario(me, allUsers, h) {
  const boards = listBoards(h);
  sleep(randomBetween(0.5, 1));

  if (boards.length > 0) {
    const boardInfo = getBoard(h, boards[0].id);
    sleep(randomBetween(0.5, 1));

    if (boardInfo && boardInfo.board) {
      const boardId = boardInfo.board.id;

      // Прочитать метки доски
      softGet(h, `${BASE_URL}/api/v1/boards/${boardId}/labels`, latency.listLabels);
      sleep(randomBetween(0.3, 0.8));

      // Прочитать available labels (board + global)
      softGet(h, `${BASE_URL}/api/v1/boards/${boardId}/available-labels`, latency.availableLabels);
      sleep(randomBetween(0.3, 0.8));

      // Прочитать автоматизации
      softGet(h, `${BASE_URL}/api/v1/boards/${boardId}/automations`, latency.listAutomations);
      sleep(randomBetween(0.3, 0.8));

      // Прочитать custom fields
      softGet(h, `${BASE_URL}/api/v1/boards/${boardId}/custom-fields`, latency.listCustomFields);
      sleep(randomBetween(0.3, 0.8));

      // Прочитать настройки доски
      softGet(h, `${BASE_URL}/api/v1/boards/${boardId}/settings`, latency.getSettings);
      sleep(randomBetween(0.3, 0.5));

      // Получить колонки → карточки → детали
      const colData = softGet(h, `${BASE_URL}/api/v1/boards/${boardId}/columns`);
      const columns = colData ? (colData.columns || []) : [];
      sleep(randomBetween(0.3, 0.5));

      if (columns.length > 0) {
        const cardData = softGet(h, `${BASE_URL}/api/v1/columns/${columns[0].id}/cards`);
        const cards = cardData ? (cardData.cards || []) : [];
        sleep(randomBetween(0.3, 0.5));

        if (cards.length > 0) {
          const card = cards[0];

          // Прочитать метки карточки
          softGet(h, `${BASE_URL}/api/v1/boards/${boardId}/cards/${card.id}/labels`, latency.getCardLabels);
          sleep(randomBetween(0.2, 0.5));

          // Прочитать чеклисты
          softGet(h, `${BASE_URL}/api/v1/boards/${boardId}/cards/${card.id}/checklists`, latency.listChecklists);
          sleep(randomBetween(0.2, 0.5));

          // Прочитать комментарии
          softGet(h, `${BASE_URL}/api/v1/cards/${card.id}/comments`, latency.listComments);
          sleep(randomBetween(0.2, 0.5));

          // Прочитать custom field values
          softGet(h, `${BASE_URL}/api/v1/boards/${boardId}/cards/${card.id}/custom-fields`, latency.getCardFields);
          sleep(randomBetween(0.2, 0.5));

          // Прочитать дочерние карточки
          softGet(h, `${BASE_URL}/api/v1/boards/${boardId}/cards/${card.id}/children`, latency.getChildren);
          sleep(randomBetween(0.2, 0.5));
        }
      }
    }
  }

  // Нотификации
  checkNotifications(h);
  sleep(randomBetween(0.3, 0.5));
  getUnreadCount(h);
  sleep(randomBetween(0.5, 1));
}

// ─── Heavy / Admin (7%): полная настройка + масса операций ──────────────

function heavyUserScenario(me, allUsers, h) {
  const boards = [];
  for (let b = 0; b < 2; b++) {
    const board = createBoard(h, `Heavy Board ${b + 1} — ${me.name}`);
    if (board) boards.push(board);
    sleep(randomBetween(0.2, 0.5));
  }
  if (boards.length === 0) return;

  for (const board of boards) {
    // Участники
    const heavyMemberCount = MEMBERS_PER_BOARD > 3 ? MEMBERS_PER_BOARD : 3;
    const members = pickRandomUsers(allUsers, me.id, heavyMemberCount);
    for (const m of members) {
      addMember(h, board.id, m.id, 'member');
      sleep(randomBetween(0.05, 0.15));
    }

    // ── Полная настройка доски (owner) ──

    // Все 7 меток
    const boardLabels = [];
    for (const preset of LABEL_PRESETS) {
      const lbl = createLabel(h, board.id, preset.name, preset.color);
      if (lbl) boardLabels.push(lbl);
      sleep(randomBetween(0.05, 0.1));
    }

    // 2 автоматизации
    createAutomation(
      h, board.id,
      'Move to Done → set Low',
      'card_moved_to_column', { column_name: 'Done' },
      'set_priority', { priority: 'low' }
    );
    sleep(randomBetween(0.1, 0.2));

    createAutomation(
      h, board.id,
      'New card → assign member',
      'card_created', {},
      'assign_member', { member_id: members.length > 0 ? members[0].id : me.id }
    );
    sleep(randomBetween(0.1, 0.2));

    // 2 custom fields
    const textField = createCustomField(h, board.id, 'Sprint', 'text', { position: 0 });
    sleep(randomBetween(0.05, 0.1));
    const dropdownField = createCustomField(h, board.id, 'Severity', 'dropdown', {
      position: 1,
      options: ['Critical', 'Major', 'Minor', 'Trivial'],
    });
    sleep(randomBetween(0.05, 0.1));

    // Обновить настройки
    updateBoardSettings(h, board.id, true);
    sleep(randomBetween(0.1, 0.2));

    // ── Колонки ──
    const cols = [];
    for (const title of ['Backlog', 'In Progress', 'Review', 'Done']) {
      const col = createColumn(h, board.id, title);
      if (col) cols.push(col);
      sleep(randomBetween(0.1, 0.3));
    }
    if (cols.length < 3) continue;

    // ── 5 карт с полным набором ──
    const cards = [];
    for (let i = 0; i < 5; i++) {
      const card = createCard(h, cols[0].id, board.id, `Heavy Task ${i + 1}`, i, {
        priority: PRIORITIES[i % PRIORITIES.length],
        task_type: TASK_TYPES[i % TASK_TYPES.length],
        due_date: randomDueDate(),
      });
      if (card) cards.push(card);
      sleep(randomBetween(0.1, 0.3));
    }

    // Привязать метки ко всем картам (2-3 метки на карту)
    for (const card of cards) {
      const shuffled = boardLabels.slice().sort(() => Math.random() - 0.5);
      const count = Math.floor(randomBetween(2, 4));
      for (let i = 0; i < count && i < shuffled.length; i++) {
        attachLabel(h, board.id, card.id, shuffled[i].id);
        sleep(randomBetween(0.03, 0.08));
      }
    }

    // Чеклисты на первых 3 картах (4-5 items каждый)
    const allItems = [];
    for (let i = 0; i < Math.min(3, cards.length); i++) {
      const cl = createChecklist(h, board.id, cards[i].id, `Чеклист ${i + 1}`, 0);
      if (cl) {
        for (let j = 0; j < Math.floor(randomBetween(4, 6)); j++) {
          const item = createChecklistItem(h, board.id, cl.id, `Шаг ${j + 1}`, j);
          if (item) allItems.push({ boardId: board.id, item });
          sleep(randomBetween(0.03, 0.08));
        }
      }
      sleep(randomBetween(0.1, 0.2));
    }

    // Toggle половину items
    for (let i = 0; i < Math.floor(allItems.length / 2); i++) {
      toggleChecklistItem(h, allItems[i].boardId, allItems[i].item.id);
      sleep(randomBetween(0.03, 0.08));
    }

    // Комментарии: owner пишет, member отвечает
    for (let i = 0; i < Math.min(3, cards.length); i++) {
      const comment = createComment(h, cards[i].id, board.id, `Описание задачи ${i + 1}: нужно сделать...`);
      sleep(randomBetween(0.1, 0.2));

      // Участник отвечает на комментарий
      if (comment && members.length > 0) {
        const memberH = auth(members[0].token);
        replyComment(memberH, cards[i].id, board.id, comment.id, 'Принял, начинаю работу');
        sleep(randomBetween(0.1, 0.2));

        // Второй участник тоже отвечает
        if (members.length > 1) {
          const memberH2 = auth(members[1].token);
          replyComment(memberH2, cards[i].id, board.id, comment.id, 'Могу помочь, если нужно');
          sleep(randomBetween(0.1, 0.2));
        }
      }
    }

    // Назначить карты на участников
    for (let i = 0; i < cards.length && i < members.length; i++) {
      assignCard(h, cards[i].id, members[i].id, board.id);
      sleep(randomBetween(0.05, 0.15));
    }

    // Set custom field values
    if (textField) {
      for (const card of cards) {
        setFieldValue(h, board.id, card.id, textField.id, { value_text: 'Sprint 12' });
        sleep(randomBetween(0.03, 0.08));
      }
    }
    if (dropdownField) {
      for (const card of cards) {
        setFieldValue(h, board.id, card.id, dropdownField.id, {
          value_text: pickRandom(['Critical', 'Major', 'Minor', 'Trivial']),
        });
        sleep(randomBetween(0.03, 0.08));
      }
    }

    // Card links: parent-child между первой и остальными
    if (cards.length >= 3) {
      linkCards(h, board.id, cards[0].id, cards[1].id);
      sleep(randomBetween(0.05, 0.1));
      linkCards(h, board.id, cards[0].id, cards[2].id);
      sleep(randomBetween(0.05, 0.1));
      getChildren(h, board.id, cards[0].id);
      sleep(randomBetween(0.1, 0.2));
    }

    // ── Перемещения ──
    for (let i = 0; i < Math.min(3, cards.length); i++) {
      moveCard(h, cards[i], cols[1].id, board.id);
      cards[i].column_id = cols[1].id;
      sleep(randomBetween(0.2, 0.5));
    }

    if (cards.length > 0) {
      moveCard(h, cards[0], cols[cols.length - 1].id, board.id);
      cards[0].column_id = cols[cols.length - 1].id;
      sleep(randomBetween(0.2, 0.5));
    }

    // Обновить пару карт
    for (let i = 0; i < Math.min(2, cards.length); i++) {
      updateCard(h, cards[i], board.id, `Updated Heavy Task ${i + 1}`);
      sleep(randomBetween(0.2, 0.3));
    }

    // Удалить последнюю карту
    if (cards.length > 2) {
      deleteCard(h, cards[cards.length - 1].id, board.id);
      sleep(randomBetween(0.2, 0.3));
    }

    // ── Чтение: member проверяет всё ──
    if (members.length > 0) {
      const memberH = auth(members[0].token);

      // Замер latency доставки нотификаций
      measureDelivery(h, memberH, board.id);

      // Member читает available labels, automations, settings
      getAvailableLabels(memberH, board.id);
      sleep(randomBetween(0.1, 0.2));
      listAutomations(memberH, board.id);
      sleep(randomBetween(0.1, 0.2));
      getBoardSettings(memberH, board.id);
      sleep(randomBetween(0.1, 0.2));
    }

    // Удалить одного участника
    if (members.length > 0) {
      removeMember(h, board.id, members[members.length - 1].id);
      sleep(randomBetween(0.2, 0.3));
    }
  }

  // ~30% удаляют первую доску
  if (Math.random() < 0.3 && boards.length > 0) {
    deleteBoard(h, boards[0].id);
  }
}

// ─── Setup-once (3%): шаблоны и глобальные метки ────────────────────────

function setupOnceScenario(me, allUsers, h) {
  // 1. Создать 2-3 глобальные метки пользователя (с уникальным суффиксом чтобы избежать 409)
  const userLabels = [];
  const suffix = Date.now() % 100000;
  const globalPresets = [
    { name: `Мой приоритет ${suffix}`, color: '#E53E3E' },
    { name: `Следить ${suffix}`,       color: '#3182CE' },
    { name: `Важное ${suffix}`,        color: '#D69E2E' },
  ];
  const count = Math.floor(randomBetween(2, 4));
  for (let i = 0; i < count; i++) {
    const lbl = createUserLabel(h, globalPresets[i].name, globalPresets[i].color);
    if (lbl) userLabels.push(lbl);
    sleep(randomBetween(0.1, 0.3));
  }

  // 2. Прочитать свои глобальные метки
  listUserLabels(h);
  sleep(randomBetween(0.3, 0.5));

  // 3. Создать board template
  const tmpl = createBoardTemplate(h, `Scrum Board Template ${suffix}`, 'Стандартный scrum', [
    { title: 'Backlog', position: 0 },
    { title: 'Sprint', position: 1 },
    { title: 'In Progress', position: 2 },
    { title: 'Review', position: 3 },
    { title: 'Done', position: 4 },
  ], [
    { name: 'Bug', color: '#E53E3E' },
    { name: 'Feature', color: '#38A169' },
    { name: 'Tech Debt', color: '#718096' },
  ]);
  sleep(randomBetween(0.3, 0.8));

  // 4. Создать доску из шаблона
  if (tmpl) {
    const board = createBoardFromTemplate(h, tmpl.id, `${me.name}'s Scrum Board`);
    sleep(randomBetween(0.3, 0.8));

    if (board) {
      // Проверить что доска создалась — прочитать
      getBoard(h, board.id);
      sleep(randomBetween(0.2, 0.5));

      // Прочитать метки созданной доски
      listLabels(h, board.id);
      sleep(randomBetween(0.2, 0.5));

      // Available labels = board labels + user global labels
      getAvailableLabels(h, board.id);
      sleep(randomBetween(0.2, 0.5));

      // Создать card template на этой доске
      createCardTemplate(h, board.id, 'Bug Report', 'Новый баг', 'high', 'bug');
      sleep(randomBetween(0.2, 0.5));
    }
  }

  // 5. Проверить список шаблонов
  listBoardTemplates(h);
  sleep(randomBetween(0.3, 0.5));

  // Нотификации
  checkNotifications(h);
  sleep(randomBetween(0.3, 0.5));
}

// ═══════════════════════════════════════════════════════════════════════════
// ─── Main ───────────────────────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════

export default function (data) {
  const allUsers = data.users;
  if (!allUsers || allUsers.length === 0) {
    console.error('No users from setup!');
    return;
  }

  const iterGlobal = exec.scenario.iterationInTest;
  const userIdx = ((exec.vu.idInTest - 1) + iterGlobal) % allUsers.length;
  const me = allUsers[userIdx];
  const h = auth(me.token);

  // Распределение: 60% worker, 15% collaborator, 15% reader, 7% heavy, 3% setup-once
  const roll = iterGlobal % 100;

  if (roll < 60) {
    workerScenario(me, allUsers, h);
  } else if (roll < 75) {
    collaboratorScenario(me, allUsers, h);
  } else if (roll < 90) {
    readerScenario(me, allUsers, h);
  } else if (roll < 97) {
    heavyUserScenario(me, allUsers, h);
  } else {
    setupOnceScenario(me, allUsers, h);
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
║           НАГРУЗОЧНЫЙ ТЕСТ: 1000 пользователей (FULL)                ║
╠════════════════════════════════════════════════════════════════════════╣
║                                                                        ║
║  Распределение:                                                        ║
║    60% workers / 15% collaborators / 15% readers / 7% heavy / 3% setup ║
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
