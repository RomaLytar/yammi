package infrastructure

import (
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	authpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1"
	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
	commentpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/comment"
	notificationpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/notification"
	userpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/user"
)

type GRPCClients struct {
	authConn           *grpc.ClientConn
	AuthClient         authpb.AuthServiceClient
	userConn           *grpc.ClientConn
	UserClient         userpb.UserServiceClient
	boardConn          *grpc.ClientConn
	BoardClient        boardpb.BoardServiceClient
	commentConn        *grpc.ClientConn
	CommentClient      commentpb.CommentServiceClient
	notificationConn   *grpc.ClientConn
	NotificationClient notificationpb.NotificationServiceClient
}

var defaultDialOpts = []grpc.DialOption{
	grpc.WithTransportCredentials(insecure.NewCredentials()),
	grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                5 * time.Minute,
		Timeout:             10 * time.Second,
		PermitWithoutStream: false,
	}),
	grpc.WithConnectParams(grpc.ConnectParams{
		MinConnectTimeout: 5 * time.Second,
		Backoff:           backoff.DefaultConfig,
	}),
}

func NewGRPCClients(authAddr, userAddr, boardAddr, commentAddr, notificationAddr string) (*GRPCClients, error) {
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

	boardConn, err := grpc.NewClient("dns:///"+boardAddr, defaultDialOpts...)
	if err != nil {
		authConn.Close()
		userConn.Close()
		return nil, fmt.Errorf("failed to connect to board service: %w", err)
	}
	log.Printf("connected to board service at %s", boardAddr)

	commentConn, err := grpc.NewClient("dns:///"+commentAddr, defaultDialOpts...)
	if err != nil {
		authConn.Close()
		userConn.Close()
		boardConn.Close()
		return nil, fmt.Errorf("failed to connect to comment service: %w", err)
	}
	log.Printf("connected to comment service at %s", commentAddr)

	notificationConn, err := grpc.NewClient("dns:///"+notificationAddr, defaultDialOpts...)
	if err != nil {
		authConn.Close()
		userConn.Close()
		boardConn.Close()
		commentConn.Close()
		return nil, fmt.Errorf("failed to connect to notification service: %w", err)
	}
	log.Printf("connected to notification service at %s", notificationAddr)

	return &GRPCClients{
		authConn:           authConn,
		AuthClient:         authpb.NewAuthServiceClient(authConn),
		userConn:           userConn,
		UserClient:         userpb.NewUserServiceClient(userConn),
		boardConn:          boardConn,
		BoardClient:        boardpb.NewBoardServiceClient(boardConn),
		commentConn:        commentConn,
		CommentClient:      commentpb.NewCommentServiceClient(commentConn),
		notificationConn:   notificationConn,
		NotificationClient: notificationpb.NewNotificationServiceClient(notificationConn),
	}, nil
}

func (c *GRPCClients) Close() {
	if c.authConn != nil {
		c.authConn.Close()
	}
	if c.userConn != nil {
		c.userConn.Close()
	}
	if c.boardConn != nil {
		c.boardConn.Close()
	}
	if c.commentConn != nil {
		c.commentConn.Close()
	}
	if c.notificationConn != nil {
		c.notificationConn.Close()
	}
}
