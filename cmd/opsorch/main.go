package main

import (
	"context"
	"log"
	"os"

	"github.com/opsorch/opsorch-core/api"
)

func main() {
	ctx := context.Background()

	srv, err := api.NewServerFromEnv(ctx)
	if err != nil {
		log.Fatalf("failed to init server: %v", err)
	}

	addr := os.Getenv("OPSORCH_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("opsorch core api listening on %s", addr)
	if err := srv.ListenAndServe(addr); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
