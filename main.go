package main

import (
	"context"
	"html/template"
	"log"
	"os"

	"bidprentjes-api/cloud"
	"bidprentjes-api/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()

	// Initialize handlers
	handler := handlers.NewHandler()

	// Get bucket name from environment variable
	bucketName := os.Getenv("STORAGE_BUCKET")
	if bucketName == "" {
		log.Printf("Warning: STORAGE_BUCKET environment variable not set, skipping initial data load")
	} else {
		// Try to load initial data from GCP
		storageClient, err := cloud.NewStorageClient(ctx, bucketName)
		if err != nil {
			log.Printf("Warning: Failed to initialize GCP storage client: %v", err)
		} else {
			defer storageClient.Close()

			// Try to download the file
			reader, err := storageClient.DownloadFile(ctx, "bidprentjes.csv")
			if err != nil {
				log.Printf("Warning: Failed to download bidprentjes.csv: %v", err)
			} else {
				// Use the existing upload logic to process the file
				count, err := handler.ProcessCSVUpload(reader)
				if err != nil {
					log.Printf("Warning: Failed to process CSV file: %v", err)
				} else {
					log.Printf("Successfully loaded %d records from GCP storage", count)
				}
			}
		}
	}

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

	// API routes
	r.GET("/api/bidprentjes", handler.ListBidprentjes)
	r.GET("/api/bidprentjes/:id", handler.GetBidprentje)
	r.POST("/api/bidprentjes", handler.CreateBidprentje)
	r.PUT("/api/bidprentjes/:id", handler.UpdateBidprentje)
	r.DELETE("/api/bidprentjes/:id", handler.DeleteBidprentje)
	r.POST("/api/bidprentjes/search", handler.SearchBidprentjes)
	r.POST("/api/bidprentjes/upload", handler.UploadCSV)

	// Keep only search and upload web endpoints
	r.GET("/search", handler.WebSearch)
	r.GET("/upload", handler.WebUpload)

	log.Fatal(r.Run(":8080"))
}
