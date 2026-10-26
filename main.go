package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	handler "essite/api"
)

//go:embed public/*
var publicFS embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3300"
	}

	subFS, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatalf("Gagal memuat embedded filesystem: %v", err)
	}
	fileServer := http.FileServer(http.FS(subFS))

	mux := http.NewServeMux()
	mux.Handle("/", fileServer)
	mux.HandleFunc("/api/capture", handler.Handler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("[essite] Server berdiri mandiri di http://localhost:%s", port)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[essite] Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[essite] Menghentikan server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	log.Println("[essite] Server berhenti.")
}
