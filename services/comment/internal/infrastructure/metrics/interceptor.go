package metrics

import (
	"context"
	"log"
	"path"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor возвращает gRPC interceptor для сбора метрик + panic recovery.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		method := path.Base(info.FullMethod)
		start := time.Now()

		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC recovered in %s: %v\n%s", method, r, debug.Stack())
				err = status.Errorf(codes.Internal, "internal error")
				GRPCRequests.WithLabelValues(method, "PANIC").Inc()
			}
		}()

		resp, err = handler(ctx, req)

		duration := time.Since(start).Seconds()
		code := status.Code(err).String()

		GRPCRequests.WithLabelValues(method, code).Inc()
		GRPCDuration.WithLabelValues(method).Observe(duration)

		return resp, err
	}
}
