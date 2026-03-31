package grpcauth

import (
	"context"
	"crypto/subtle"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const metadataKey = "x-internal-secret"

// ServerInterceptor returns a unary server interceptor that validates the shared secret.
// If secret is empty, validation is skipped (allows gradual rollout).
func ServerInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if secret == "" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get(metadataKey)
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing internal secret")
		}

		if subtle.ConstantTimeCompare([]byte(values[0]), []byte(secret)) != 1 {
			return nil, status.Error(codes.Unauthenticated, "invalid internal secret")
		}

		return handler(ctx, req)
	}
}

// ClientInterceptor returns a unary client interceptor that appends the shared secret to outgoing metadata.
// If secret is empty, it's a noop.
func ClientInterceptor(secret string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if secret != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, metadataKey, secret)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
