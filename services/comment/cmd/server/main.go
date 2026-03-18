package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	port := os.Getenv("COMMENT_GRPC_PORT")
	if port == "" {
		port = "50054"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("comment-service: failed to listen: %v", err)
	}
	defer lis.Close()

	log.Printf("comment-service started on :%s", port)
	fmt.Printf("comment-service listening on :%s\n", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("comment-service shutting down")
}
