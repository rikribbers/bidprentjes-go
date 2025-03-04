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

	"bidprentjes-api/cloud"
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

	// If we have a bucket name and no valid index was restored, try processing CSV
	if bucketName != "" && !store.HasValidIndex() {
		storageClient, err := cloud.NewStorageClient(ctx, bucketName)
		if err != nil {
			log.Printf("Warning: Failed to initialize GCP storage client: %v", err)
		} else {
			defer storageClient.Close()

			// Try to download the CSV file
			reader, err := storageClient.DownloadFile(ctx, "data/bidprentjes.csv")
			if err != nil {
				log.Printf("Warning: Failed to download bidprentjes.csv: %v", err)
			} else {
				// Process the CSV file
				count, err := handler.ProcessCSVUpload(reader)
				if err != nil {
					log.Printf("Warning: Failed to process CSV file: %v", err)
				} else {
					log.Printf("Successfully loaded %d records from GCP storage", count)

					// Move the processed CSV file
					if err := storageClient.MoveFile(ctx, "data/bidprentjes.csv", "data/processed/bidprentjes.csv."+time.Now().Format("20060102150405")); err != nil {
						log.Printf("Warning: Failed to move processed CSV file: %v", err)
					}

					// Create immediate backup of the index after processing
					log.Printf("Creating immediate backup of the index...")
					if err := store.BackupIndex(ctx); err != nil {
						log.Printf("Warning: Failed to create immediate index backup: %v", err)
					} else {
						log.Printf("Successfully created immediate index backup")
					}
				}
			}
		}
	}

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
	r.GET("/upload", handler.WebUpload)
	r.POST("/upload", handler.UploadCSV)

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
