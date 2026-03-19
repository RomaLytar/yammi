package infrastructure

import (
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	authpb "github.com/romanlovesweed/yammi/services/api-gateway/api/proto/v1"
	userpb "github.com/romanlovesweed/yammi/services/api-gateway/api/proto/v1/user"
)

type GRPCClients struct {
	authConn   *grpc.ClientConn
	AuthClient authpb.AuthServiceClient
	userConn   *grpc.ClientConn
	UserClient userpb.UserServiceClient
}

var defaultDialOpts = []grpc.DialOption{
	grpc.WithTransportCredentials(insecure.NewCredentials()),
	grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             3 * time.Second,
		PermitWithoutStream: true,
	}),
	grpc.WithConnectParams(grpc.ConnectParams{
		MinConnectTimeout: 5 * time.Second,
		Backoff:           backoff.DefaultConfig,
	}),
}

func NewGRPCClients(authAddr, userAddr string) (*GRPCClients, error) {
	authConn, err := grpc.NewClient("dns:///"+authAddr, defaultDialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}
	log.Printf("connected to auth service at %s", authAddr)

	userConn, err := grpc.NewClient("dns:///"+userAddr, defaultDialOpts...)
	if err != nil {
		authConn.Close()
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}
	log.Printf("connected to user service at %s", userAddr)

	return &GRPCClients{
		authConn:   authConn,
		AuthClient: authpb.NewAuthServiceClient(authConn),
		userConn:   userConn,
		UserClient: userpb.NewUserServiceClient(userConn),
	}, nil
}

func (c *GRPCClients) Close() {
	if c.authConn != nil {
		c.authConn.Close()
	}
	if c.userConn != nil {
		c.userConn.Close()
	}
}
