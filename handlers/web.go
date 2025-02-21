package handlers

import (
	"log"
	"net/http"

	"bidprentjes-api/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) WebIndex(c *gin.Context) {
	log.Printf("WebIndex handler called")
	c.HTML(http.StatusOK, "index.html", gin.H{
		"data": &models.PaginatedResponse{
			Items:      []models.Bidprentje{},
			TotalCount: 0,
			Page:       1,
			PageSize:   25,
		},
	})
}

func (h *Handler) WebCreate(c *gin.Context) {
	log.Printf("WebCreate handler called")
	c.HTML(http.StatusOK, "create.html", nil)
}

func (h *Handler) WebEdit(c *gin.Context) {
	log.Printf("WebEdit handler called")
	c.HTML(http.StatusOK, "edit.html", nil)
}

func (h *Handler) WebSearch(c *gin.Context) {
	log.Printf("WebSearch handler called")
	c.HTML(http.StatusOK, "search.html", gin.H{
		"data": &models.PaginatedResponse{
			Items:      []models.Bidprentje{},
			TotalCount: 0,
			Page:       1,
			PageSize:   25,
		},
	})
}

func (h *Handler) WebUpload(c *gin.Context) {
	log.Printf("WebUpload handler called")
	c.HTML(http.StatusOK, "upload.html", nil)
}
