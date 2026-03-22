package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// GRPCRequests — счётчик gRPC запросов по методу и коду.
	GRPCRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "comment_grpc_requests_total",
		Help: "Total gRPC requests by method and status code",
	}, []string{"method", "code"})

	// GRPCDuration — гистограмма длительности gRPC запросов.
	GRPCDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "comment_grpc_request_duration_seconds",
		Help:    "gRPC request duration in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
	}, []string{"method"})

	// EventsPublished — счётчик опубликованных NATS событий.
	EventsPublished = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "comment_events_published_total",
		Help: "NATS events published by subject",
	}, []string{"subject"})

	// EventPublishErrors — ошибки публикации событий.
	EventPublishErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "comment_event_publish_errors_total",
		Help: "NATS event publish errors by subject",
	}, []string{"subject"})

	// CommentsCreated — счётчик созданных комментариев.
	CommentsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "comment_comments_created_total",
		Help: "Total comments created",
	})

	// RepliesCreated — счётчик созданных ответов.
	RepliesCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "comment_replies_created_total",
		Help: "Total replies created",
	})
)
