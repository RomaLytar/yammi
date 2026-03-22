package metrics

import (
	"context"
	"fmt"
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

		// Recovery: panic → controlled gRPC error вместо crash
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

// ChainUnaryInterceptors объединяет несколько interceptors в один.
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			current := interceptors[i]
			next := chain
			chain = func(ctx context.Context, req any) (any, error) {
				return current(ctx, req, info, next)
			}
		}
		return chain(ctx, req)
	}
}

// RecoveryInterceptor — standalone recovery interceptor для сервисов без metrics.
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC recovered in %s: %v\n%s", info.FullMethod, r, debug.Stack())
				err = status.Errorf(codes.Internal, fmt.Sprintf("internal error: %v", r))
			}
		}()
		return handler(ctx, req)
	}
}
