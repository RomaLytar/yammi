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
	port := os.Getenv("BOARD_GRPC_PORT")
	if port == "" {
		port = "50053"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("board-service: failed to listen: %v", err)
	}
	defer lis.Close()

	log.Printf("board-service started on :%s", port)
	fmt.Printf("board-service listening on :%s\n", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("board-service shutting down")
}
