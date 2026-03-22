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
)
