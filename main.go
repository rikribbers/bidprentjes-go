package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bidprentjes-api/handlers"
	"bidprentjes-api/store"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get bucket name from environment variable
	bucketName := os.Getenv("STORAGE_BUCKET")
	if bucketName == "" {
		log.Printf("Warning: STORAGE_BUCKET environment variable not set, running in local-only mode")
	}

	// Initialize store with bucket name
	store := store.NewStore(ctx, bucketName)
	defer store.Close()

	// Initialize handlers with store
	handler := handlers.NewHandler(store)

	// Create Gin router
	r := gin.Default()

	// Add template functions
	r.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
		"divide": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"sequence": func(n int) []int {
			seq := make([]int, n)
			for i := range seq {
				seq[i] = i
			}
			return seq
		},
	})

	// Load HTML templates
	log.Println("Loading templates from templates/*")
	r.LoadHTMLGlob("templates/*.html")
	log.Println("Templates loaded successfully")

	// Keep only search and upload web endpoints
	r.GET("/search", handler.WebSearch)

	// Create a server with timeouts
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
		// Set timeouts to prevent hanging connections
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a timeout context for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown the server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Cancel the main context to trigger cleanup in other goroutines
	cancel()

	log.Println("Server exiting")
}
