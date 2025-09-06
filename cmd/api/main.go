package main

import (
	"context"
	"log"

	"github.com/1107-adishjain/auth-system/config"
	"github.com/1107-adishjain/auth-system/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	srv := server.NewServer(cfg)

	if err := srv.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
