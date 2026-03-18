package infrastructure

import (
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "github.com/romanlovesweed/yammi/services/api-gateway/api/proto/v1"
	userpb "github.com/romanlovesweed/yammi/services/api-gateway/api/proto/v1/user"
)

type GRPCClients struct {
	AuthConn   *grpc.ClientConn
	AuthClient authpb.AuthServiceClient
	UserConn   *grpc.ClientConn
	UserClient userpb.UserServiceClient
}

func NewGRPCClients(authAddr, userAddr string) (*GRPCClients, error) {
	authConn, err := grpc.NewClient(
		"dns:///"+authAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}
	log.Printf("connected to auth service at %s", authAddr)

	userConn, err := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		authConn.Close()
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}
	log.Printf("connected to user service at %s", userAddr)

	return &GRPCClients{
		AuthConn:   authConn,
		AuthClient: authpb.NewAuthServiceClient(authConn),
		UserConn:   userConn,
		UserClient: userpb.NewUserServiceClient(userConn),
	}, nil
}

func (c *GRPCClients) Close() {
	if c.AuthConn != nil {
		c.AuthConn.Close()
	}
	if c.UserConn != nil {
		c.UserConn.Close()
	}
}
