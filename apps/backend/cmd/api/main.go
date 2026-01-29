package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/LBaronceli/go-figure/internal/db"
	"github.com/LBaronceli/go-figure/internal/httpserver"
)

func main() {
	ctx := context.Background()

	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatalf("db init: %v", err)
	}
	defer pool.Close()

	srv := httpserver.NewServer(pool)

	handler := srv.Routes()

	addr := ":8080"
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		addr = v
	}

	log.Printf("api listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

