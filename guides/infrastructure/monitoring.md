# Мониторинг — Prometheus + Grafana

> Метрики сервисов, scrape targets, Grafana дашборды.

---

## Обзор

```
Go Services (порт 2112)     NATS Exporter (порт 7777)
        │                            │
        ▼                            ▼
┌──────────────────────────────────────────┐
│            Prometheus (:9090)            │
│         scrape каждые 5 секунд          │
└──────────────────┬───────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────┐
│            Grafana (:3033)               │
│         4 дашборда, auto-refresh 5s     │
└──────────────────────────────────────────┘
```

**Доступ к Grafana:** `http://localhost:3033` (admin/admin, anonymous access включён)

---

## Prometheus Targets

Конфиг: `deployments/monitoring/prometheus/prometheus.yml`

| Job | Target | Что скрейпит |
|-----|--------|--------------|
| `nats` | `nats-exporter:7777` | NATS JetStream метрики |
| `notification-service` | `notification:2112` | Notification Service метрики |

### Как добавить новый сервис

1. Добавить Prometheus client в `go.mod`:
   ```
   github.com/prometheus/client_golang v1.21.1
   ```

2. Создать HTTP `/metrics` endpoint:
   ```go
   import "github.com/prometheus/client_golang/prometheus/promhttp"

   mux := http.NewServeMux()
   mux.Handle("/metrics", promhttp.Handler())
   http.ListenAndServe(":2112", mux)
   ```

3. Добавить scrape target в `prometheus.yml`:
   ```yaml
   - job_name: 'my-service'
     static_configs:
       - targets: ['my-service:2112']
   ```

4. Перезапустить Prometheus:
   ```bash
   docker compose restart prometheus
   ```

---

## Grafana Provisioning

### Datasource

Конфиг: `deployments/monitoring/grafana/provisioning/datasources/datasource.yml`

```yaml
datasources:
  - name: Prometheus
    uid: prometheus          # ← Важно: этот UID используется во всех дашбордах
    type: prometheus
    url: http://prometheus:9090
    isDefault: true
```

> ⚠️ **UID обязателен.** Пустой `"uid": ""` в дашбордах **не работает** — Grafana не резолвит пустой UID в default datasource. Всегда указывать `"uid": "prometheus"`.

### Дашборды

Конфиг: `deployments/monitoring/grafana/provisioning/dashboards/dashboard.yml`

Дашборды загружаются из: `deployments/monitoring/grafana/dashboards/*.json`

| Файл | Дашборд | Описание |
|------|---------|----------|
| `notification-service.json` | Notification Service | Уведомления, NATS events, gRPC, latency, ошибки |
| `board-service.json` | Board Service | Операции, latency, cache, ошибки |
| `nats-jetstream.json` | NATS JetStream | Streams, consumers, messages |
| `nats-user-deleted.json` | NATS User Deleted | Специфичный мониторинг |

---

## Notification Service — Метрики

### Counters

| Метрика | Labels | Описание |
|---------|--------|----------|
| `notification_created_total` | `type` | Уведомления созданные в БД |
| `notification_skipped_total` | — | Пропущенные (настройки отключены) |
| `notification_events_consumed_total` | `subject` | NATS события успешно обработанные |
| `notification_event_errors_total` | `subject` | Ошибки обработки событий |
| `notification_event_retries_total` | `subject` | Ретраи (NAK + delay) |
| `notification_events_dlq_total` | `subject` | Отправлены в DLQ |
| `notification_grpc_requests_total` | `method`, `code` | gRPC запросы (OK, NotFound, ...) |

### Histograms

| Метрика | Labels | Buckets | Описание |
|---------|--------|---------|----------|
| `notification_event_processing_duration_seconds` | `subject` | 1ms — 1s | Время обработки NATS событий |
| `notification_grpc_request_duration_seconds` | `method` | 1ms — 1s | Время gRPC запросов |

---

## Правила создания дашбордов

### Datasource UID

Во **всех** панелях:
```json
"datasource": { "type": "prometheus", "uid": "prometheus" }
```

❌ **Не использовать** `"uid": ""` — не работает.

### Stat панели (счётчики)

Для отображения текущего значения counter/gauge:
```json
"reduceOptions": { "calcs": ["lastNotNull"] }
```

❌ **Не использовать** `"calcs": ["sum"]` — суммирует все data points за time range, число растёт бесконечно.

### Timeseries (rate)

Для counters всегда использовать `rate()`:
```
rate(notification_created_total[1m])
```

### Auto-refresh

```json
{
  "refresh": "5s",
  "liveNow": true,
  "timepicker": {
    "refresh_intervals": ["5s", "10s", "30s", "1m", "5m"]
  }
}
```

---

## Troubleshooting

### Prometheus не скрейпит target

```bash
# Проверить targets
curl http://localhost:9090/api/v1/targets

# Если target отсутствует — перезапустить Prometheus
docker compose restart prometheus
```

### Grafana показывает "No data"

1. Проверить datasource UID в JSON: должен быть `"uid": "prometheus"`, не `""`
2. Проверить что target UP в Prometheus: `http://localhost:9090/targets`
3. Проверить что метрики приходят: `docker compose exec <service> wget -qO- http://localhost:2112/metrics`
4. Перезапустить Grafana: `docker compose restart grafana`

### Числа в stat-панелях растут бесконечно

Заменить `"calcs": ["sum"]` на `"calcs": ["lastNotNull"]` в `reduceOptions`.
