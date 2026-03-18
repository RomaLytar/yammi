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
	port := os.Getenv("USER_GRPC_PORT")
	if port == "" {
		port = "50052"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("user-service: failed to listen: %v", err)
	}
	defer lis.Close()

	log.Printf("user-service started on :%s", port)
	fmt.Printf("user-service listening on :%s\n", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("user-service shutting down")
}
