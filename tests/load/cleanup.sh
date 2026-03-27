#!/bin/bash
# Полная очистка всех данных после нагрузочного теста.
# Запуск: ./tests/load/cleanup.sh

set -e

echo "=== Очистка всех БД ==="

docker compose exec -T pgbouncer psql 'host=postgres user=yammi password=yammi dbname=yammi_auth' \
  -c 'TRUNCATE users, refresh_tokens CASCADE;'

docker compose exec -T pgbouncer psql 'host=postgres user=yammi password=yammi dbname=yammi_board' \
  -c 'TRUNCATE boards, columns, cards, board_members, user_labels, board_templates CASCADE;'

docker compose exec -T pgbouncer psql 'host=postgres user=yammi password=yammi dbname=yammi_comment' \
  -c 'TRUNCATE comments CASCADE;' 2>/dev/null || true

docker compose exec -T pgbouncer psql 'host=postgres user=yammi password=yammi dbname=yammi_notification' \
  -c 'TRUNCATE notifications, board_events, board_members, user_board_cursors, board_names, user_names, card_names, column_names, notification_settings CASCADE;'

docker compose exec -T pgbouncer psql 'host=postgres user=yammi password=yammi dbname=yammi_user' \
  -c 'TRUNCATE profiles CASCADE;' 2>/dev/null || true

echo "=== Flush Redis ==="
docker compose exec -T redis redis-cli -a "${REDIS_PASSWORD:-yammi_redis}" FLUSHALL 2>/dev/null || \
  docker compose exec -T redis redis-cli FLUSHALL 2>/dev/null || true

echo "=== VACUUM ANALYZE ==="
for db in yammi_auth yammi_board yammi_notification yammi_user; do
  docker compose exec -T pgbouncer psql "host=postgres user=yammi password=yammi dbname=$db" \
    -c 'VACUUM ANALYZE;' 2>/dev/null || true
done

echo "=== Проверка ==="
echo -n "Auth users: "
docker compose exec -T pgbouncer psql 'host=postgres user=yammi password=yammi dbname=yammi_auth' -t -c 'SELECT count(*) FROM users;'
echo -n "Boards: "
docker compose exec -T pgbouncer psql 'host=postgres user=yammi password=yammi dbname=yammi_board' -t -c 'SELECT count(*) FROM boards;'
echo -n "Notifications: "
docker compose exec -T pgbouncer psql 'host=postgres user=yammi password=yammi dbname=yammi_notification' -t -c 'SELECT count(*) FROM notifications;'
echo -n "Board events: "
docker compose exec -T pgbouncer psql 'host=postgres user=yammi password=yammi dbname=yammi_notification' -t -c 'SELECT count(*) FROM board_events;'
echo -n "Redis keys: "
docker compose exec -T redis redis-cli -a "${REDIS_PASSWORD:-yammi_redis}" DBSIZE 2>/dev/null || \
  docker compose exec -T redis redis-cli DBSIZE 2>/dev/null || echo "N/A"

echo "=== Очистка завершена ==="
