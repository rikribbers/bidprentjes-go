package main

import (
	"html/template"
	"log"

	"bidprentjes-api/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Add template functions
	r.SetFuncMap(template.FuncMap{
		"subtract": func(a, b int) int { return a - b },
		"add":      func(a, b int) int { return a + b },
		"divide":   func(a, b int) int { return (a + b - 1) / b }, // Ceiling division for page count
		"sequence": func(n int) []int {
			seq := make([]int, n)
			for i := range seq {
				seq[i] = i
			}
			return seq
		},
	})

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// Initialize handlers
	handler := handlers.NewHandler()

	// API Routes
	r.POST("/bidprentjes", handler.CreateBidprentje)
	r.GET("/bidprentjes/:id", handler.GetBidprentje)
	r.PUT("/bidprentjes/:id", handler.UpdateBidprentje)
	r.DELETE("/bidprentjes/:id", handler.DeleteBidprentje)
	r.GET("/bidprentjes", handler.ListBidprentjes)
	r.POST("/bidprentjes/search", handler.SearchBidprentjes)

	// Web Routes
	r.GET("/web", handler.WebIndex)
	r.GET("/web/create", handler.WebCreate)
	r.GET("/web/edit/:id", handler.WebEdit)
	r.GET("/web/upload", handler.WebUpload)

	log.Fatal(r.Run(":8080"))
}
