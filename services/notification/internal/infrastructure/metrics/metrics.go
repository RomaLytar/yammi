package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// EventsConsumed — счётчик обработанных NATS событий по subject.
	EventsConsumed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notification_events_consumed_total",
		Help: "Total NATS events consumed by subject",
	}, []string{"subject"})

	// EventErrors — счётчик ошибок обработки событий.
	EventErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notification_event_errors_total",
		Help: "Event processing errors by subject",
	}, []string{"subject"})

	// EventsDLQ — счётчик событий отправленных в DLQ.
	EventsDLQ = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notification_events_dlq_total",
		Help: "Events sent to dead letter queue by subject",
	}, []string{"subject"})

	// EventProcessingDuration — гистограмма длительности обработки NATS событий.
	EventProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "notification_event_processing_duration_seconds",
		Help:    "NATS event processing duration in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	}, []string{"subject"})

	// NotificationsCreated — счётчик созданных уведомлений по типу.
	NotificationsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notification_created_total",
		Help: "Total notifications created by type",
	}, []string{"type"})

	// NotificationsSkipped — счётчик пропущенных уведомлений (отключены настройками).
	NotificationsSkipped = promauto.NewCounter(prometheus.CounterOpts{
		Name: "notification_skipped_total",
		Help: "Notifications skipped due to disabled settings",
	})

	// GRPCRequests — счётчик gRPC запросов по методу и коду.
	GRPCRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notification_grpc_requests_total",
		Help: "Total gRPC requests by method and status code",
	}, []string{"method", "code"})

	// GRPCDuration — гистограмма длительности gRPC запросов.
	GRPCDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "notification_grpc_request_duration_seconds",
		Help:    "gRPC request duration in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	}, []string{"method"})

	// EventRetries — счётчик ретраев обработки событий.
	EventRetries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notification_event_retries_total",
		Help: "Event processing retries by subject",
	}, []string{"subject"})

	// BoardEventsCreated — счётчик board events (1 на событие, без fan-out).
	BoardEventsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notification_board_events_created_total",
		Help: "Board events created (1 per event, replaces N notifications fan-out)",
	}, []string{"type"})

	// RedisIncrements — счётчик Redis INCR операций для unread counters.
	RedisIncrements = promauto.NewCounter(prometheus.CounterOpts{
		Name: "notification_redis_increments_total",
		Help: "Redis INCR operations for unread counters",
	})

	// RedisLatency — гистограмма латенси Redis операций.
	RedisLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "notification_redis_latency_seconds",
		Help:    "Redis operation latency in seconds",
		Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
	})

	// DBWaitDuration — время ожидания DB connection из пула.
	DBWaitDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "notification_db_wait_seconds",
		Help:    "Time waiting for DB connection from pool",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	})

	// Goroutines — текущее количество горутин (gauge).
	Goroutines = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "notification_goroutines",
		Help: "Current number of goroutines",
	})

	// MembersPerEvent — сколько участников обрабатывается на 1 board event.
	MembersPerEvent = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "notification_members_per_event",
		Help:    "Number of board members notified per event",
		Buckets: []float64{1, 2, 5, 10, 20, 50, 100},
	})
)
