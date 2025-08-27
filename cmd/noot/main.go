package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/grantbirki/noot/internal/server"
)

func main() {
	// Optional .env loader
	server.LoadDotEnv()

	// Initialize logger after env variables are loaded
	server.InitLogger()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Printf("noot-api running at http://localhost:%s\n", port)
	if err := server.Run(ctx, port); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
