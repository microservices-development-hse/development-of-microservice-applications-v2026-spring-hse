package main

import (
	"log"
	"net/http"

	"github.com/microservices-development-hse/backend/internal/mock"
)

func main() {
	addr := ":8080"
	handler := mock.NewMux()

	log.Printf("mock backend listening on %s\n", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
