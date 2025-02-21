package main

import (
	"log"

	"bidprentjes-api/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Load HTML templates
	log.Println("Loading templates from templates/*")
	r.LoadHTMLGlob("templates/*.html")
	log.Println("Templates loaded successfully")

	// Initialize handlers
	handler := handlers.NewHandler()

	// API Routes
	r.POST("/api/bidprentjes", handler.CreateBidprentje)
	r.GET("/api/bidprentjes/:id", handler.GetBidprentje)
	r.PUT("/api/bidprentjes/:id", handler.UpdateBidprentje)
	r.DELETE("/api/bidprentjes/:id", handler.DeleteBidprentje)
	r.GET("/api/bidprentjes", handler.ListBidprentjes)
	r.POST("/api/search", handler.SearchBidprentjes)
	r.POST("/api/upload", handler.WebUpload)

	// Web Routes
	r.GET("/web", handler.WebIndex)
	r.GET("/web/create", handler.WebCreate)
	r.GET("/web/edit/:id", handler.WebEdit)
	r.GET("/search", handler.WebSearch)
	r.GET("/upload", handler.WebUpload)

	log.Fatal(r.Run(":8080"))
}
