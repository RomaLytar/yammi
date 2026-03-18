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
	port := os.Getenv("NOTIFICATION_GRPC_PORT")
	if port == "" {
		port = "50055"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("notification-service: failed to listen: %v", err)
	}
	defer lis.Close()

	log.Printf("notification-service started on :%s", port)
	fmt.Printf("notification-service listening on :%s\n", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("notification-service shutting down")
}
