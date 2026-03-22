import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend, Rate, Gauge } from 'k6/metrics';
import exec from 'k6/execution';

// ─── Метрики ────────────────────────────────────────────────────────────

const latency = {
  createBoard:  new Trend('latency_create_board_ms'),
  createColumn: new Trend('latency_create_column_ms'),
  createCard:   new Trend('latency_create_card_ms'),
  moveCard:     new Trend('latency_move_card_ms'),
  getBoard:     new Trend('latency_get_board_ms'),
  listBoards:   new Trend('latency_list_boards_ms'),
  notifications: new Trend('latency_notifications_ms'),
};

const errorRate = new Rate('error_rate');
const rpsGauge = new Gauge('current_vus');

// ─── Конфигурация ───────────────────────────────────────────────────────

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TOTAL_USERS = parseInt(__ENV.USERS || '1000');
const MEMBERS = parseInt(__ENV.MEMBERS || '5');
const MAX_VU = parseInt(__ENV.MAX_VU || '3000');

// Step load: 500 → 1000 → 1500 → 2000 → 2500 → 3000 (по 1 мин на каждый)
export const options = {
  scenarios: {
    step_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 500 },   // warm up
        { duration: '1m',  target: 500 },   // step 1: 500
        { duration: '30s', target: 1000 },
        { duration: '1m',  target: 1000 },  // step 2: 1000
        { duration: '30s', target: 1500 },
        { duration: '1m',  target: 1500 },  // step 3: 1500
        { duration: '30s', target: 2000 },
        { duration: '1m',  target: 2000 },  // step 4: 2000
        { duration: '30s', target: 2500 },
        { duration: '1m',  target: 2500 },  // step 5: 2500
        { duration: '30s', target: MAX_VU },
        { duration: '1m',  target: MAX_VU }, // step 6: max
        { duration: '30s', target: 0 },      // cooldown
      ],
    },
  },
  thresholds: {
    'http_req_failed': ['rate<0.10'],       // до 10% ошибок допустимо (ищем предел)
    'http_req_duration': ['p(95)<5000'],    // p95 < 5s
  },
};

// ─── Setup ──────────────────────────────────────────────────────────────

export function setup() {
  console.log(`Ceiling test: registering ${TOTAL_USERS} users, ${MEMBERS} members/board, max ${MAX_VU} VU`);

  const users = [];
  const ts = Date.now();

  for (let i = 0; i < TOTAL_USERS; i++) {
    const res = http.post(
      `${BASE_URL}/api/v1/auth/register`,
      JSON.stringify({ email: `ceil-${i}-${ts}@y.io`, password: 'loadtest123456', name: `U${i}` }),
      { headers: { 'Content-Type': 'application/json' }, timeout: '30s' }
    );

    if (res.status === 201) {
      const body = res.json();
      users.push({ id: body.user_id, token: body.access_token });
    }

    if (i % 100 === 0 && i > 0) {
      sleep(0.3);
      console.log(`  registered ${i}/${TOTAL_USERS}`);
    }
  }

  console.log(`Setup done: ${users.length} users`);
  sleep(3);
  return { users };
}

// ─── Helpers ────────────────────────────────────────────────────────────

function headers(token) {
  return { headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` } };
}

function rnd(min, max) { return min + Math.random() * (max - min); }

function pickRandom(arr, excludeId, count) {
  const filtered = arr.filter(u => u.id !== excludeId);
  const shuffled = filtered.sort(() => Math.random() - 0.5);
  return shuffled.slice(0, Math.min(count, shuffled.length));
}

// ─── Main ───────────────────────────────────────────────────────────────

export default function(data) {
  const allUsers = data.users;
  if (!allUsers || allUsers.length === 0) return;

  const idx = exec.vu.idInTest % allUsers.length;
  const me = allUsers[idx];
  const h = headers(me.token);

  rpsGauge.add(exec.vu.idInTest);

  // Микс операций: 40% read, 40% write, 20% heavy
  const roll = Math.random();

  if (roll < 0.4) {
    // READ: list boards + get board
    let start = Date.now();
    const listRes = http.get(`${BASE_URL}/api/v1/boards?limit=10`, h);
    latency.listBoards.add(Date.now() - start);
    errorRate.add(listRes.status >= 400);

    if (listRes.status === 200) {
      const boards = listRes.json().boards || [];
      if (boards.length > 0) {
        const board = boards[Math.floor(Math.random() * boards.length)];
        start = Date.now();
        const getRes = http.get(`${BASE_URL}/api/v1/boards/${board.id}`, h);
        latency.getBoard.add(Date.now() - start);
        errorRate.add(getRes.status >= 400);
      }
    }

    // Notifications
    start = Date.now();
    const notifRes = http.get(`${BASE_URL}/api/v1/notifications/unread-count`, h);
    latency.notifications.add(Date.now() - start);
    errorRate.add(notifRes.status >= 400);

  } else if (roll < 0.8) {
    // WRITE: create board + column + card + move
    let start = Date.now();
    const boardRes = http.post(`${BASE_URL}/api/v1/boards`,
      JSON.stringify({ title: `B-${Date.now()}` }), h);
    latency.createBoard.add(Date.now() - start);
    errorRate.add(boardRes.status >= 400);

    if (boardRes.status !== 201) { sleep(rnd(0.5, 1)); return; }
    const board = boardRes.json().board;

    // Add members
    const members = pickRandom(allUsers, me.id, MEMBERS);
    for (const m of members) {
      http.post(`${BASE_URL}/api/v1/boards/${board.id}/members`,
        JSON.stringify({ user_id: m.id, role: 'member' }), h);
    }

    // Column
    start = Date.now();
    const colRes = http.post(`${BASE_URL}/api/v1/boards/${board.id}/columns`,
      JSON.stringify({ title: 'Todo', position: 0 }), h);
    latency.createColumn.add(Date.now() - start);
    errorRate.add(colRes.status >= 400);

    if (colRes.status !== 201) { sleep(rnd(0.5, 1)); return; }
    const col = colRes.json().column;

    // Card
    start = Date.now();
    const cardRes = http.post(`${BASE_URL}/api/v1/columns/${col.id}/cards`,
      JSON.stringify({ title: `C-${Date.now()}`, board_id: board.id, position: 'n' }), h);
    latency.createCard.add(Date.now() - start);
    errorRate.add(cardRes.status >= 400);

    sleep(rnd(0.2, 0.5));

  } else {
    // HEAVY: create + move cards
    let start = Date.now();
    const boardRes = http.post(`${BASE_URL}/api/v1/boards`,
      JSON.stringify({ title: `H-${Date.now()}` }), h);
    latency.createBoard.add(Date.now() - start);
    errorRate.add(boardRes.status >= 400);

    if (boardRes.status !== 201) { sleep(rnd(0.5, 1)); return; }
    const board = boardRes.json().board;

    const members = pickRandom(allUsers, me.id, MEMBERS);
    for (const m of members) {
      http.post(`${BASE_URL}/api/v1/boards/${board.id}/members`,
        JSON.stringify({ user_id: m.id, role: 'member' }), h);
    }

    const col1Res = http.post(`${BASE_URL}/api/v1/boards/${board.id}/columns`,
      JSON.stringify({ title: 'A', position: 0 }), h);
    const col2Res = http.post(`${BASE_URL}/api/v1/boards/${board.id}/columns`,
      JSON.stringify({ title: 'B', position: 1 }), h);

    if (col1Res.status !== 201 || col2Res.status !== 201) { sleep(rnd(0.5, 1)); return; }

    const col1 = col1Res.json().column;
    const col2 = col2Res.json().column;

    const cardRes = http.post(`${BASE_URL}/api/v1/columns/${col1.id}/cards`,
      JSON.stringify({ title: `M-${Date.now()}`, board_id: board.id, position: 'n' }), h);

    if (cardRes.status !== 201) { sleep(rnd(0.5, 1)); return; }
    const card = cardRes.json().card;

    // Move card back and forth
    for (let i = 0; i < 3; i++) {
      const from = i % 2 === 0 ? col1.id : col2.id;
      const to = i % 2 === 0 ? col2.id : col1.id;
      start = Date.now();
      http.put(`${BASE_URL}/api/v1/cards/${card.id}/move`,
        JSON.stringify({ board_id: board.id, from_column_id: from, to_column_id: to, position: 'n' }), h);
      latency.moveCard.add(Date.now() - start);
      sleep(rnd(0.1, 0.3));
    }
  }

  sleep(rnd(0.3, 1));
}

// ─── Summary ────────────────────────────────────────────────────────────

export function handleSummary(data) {
  const p95 = (m) => m ? Math.round(m.values['p(95)'] || 0) : 'N/A';
  const p99 = (m) => m ? Math.round(m.values['p(99)'] || 0) : 'N/A';

  const summary = `
╔══════════════════════════════════════════════════════════════════╗
║           CEILING TEST: step load → ${MAX_VU} VU                ║
╠══════════════════════════════════════════════════════════════════╣
║                                                                  ║
║  Members/board: ${MEMBERS}                                       ║
║                                                                  ║
║  ── Latency p95 / p99 ────────────────────────────────────────  ║
║  Create board:     ${String(p95(data.metrics.latency_create_board_ms)).padStart(6)} / ${String(p99(data.metrics.latency_create_board_ms)).padStart(6)} ms  ║
║  Create column:    ${String(p95(data.metrics.latency_create_column_ms)).padStart(6)} / ${String(p99(data.metrics.latency_create_column_ms)).padStart(6)} ms  ║
║  Create card:      ${String(p95(data.metrics.latency_create_card_ms)).padStart(6)} / ${String(p99(data.metrics.latency_create_card_ms)).padStart(6)} ms  ║
║  Move card:        ${String(p95(data.metrics.latency_move_card_ms)).padStart(6)} / ${String(p99(data.metrics.latency_move_card_ms)).padStart(6)} ms  ║
║  Get board:        ${String(p95(data.metrics.latency_get_board_ms)).padStart(6)} / ${String(p99(data.metrics.latency_get_board_ms)).padStart(6)} ms  ║
║  List boards:      ${String(p95(data.metrics.latency_list_boards_ms)).padStart(6)} / ${String(p99(data.metrics.latency_list_boards_ms)).padStart(6)} ms  ║
║  Unread count:     ${String(p95(data.metrics.latency_notifications_ms)).padStart(6)} / ${String(p99(data.metrics.latency_notifications_ms)).padStart(6)} ms  ║
║                                                                  ║
║  ── HTTP ─────────────────────────────────────────────────────  ║
║  Total requests:   ${String(Math.round(data.metrics.http_reqs.values.count)).padStart(8)}                              ║
║  RPS (avg):        ${String(Math.round(data.metrics.http_reqs.values.rate)).padStart(8)}                              ║
║  Failed:           ${String((data.metrics.http_req_failed.values.rate * 100).toFixed(1)).padStart(6)}%                              ║
║  Duration p95:     ${String(p95(data.metrics.http_req_duration)).padStart(6)} ms                              ║
║  Duration p99:     ${String(p99(data.metrics.http_req_duration)).padStart(6)} ms                              ║
║  Error rate:       ${String((data.metrics.error_rate.values.rate * 100).toFixed(1)).padStart(6)}%                              ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝`;

  console.log(summary);
  return {};
}
