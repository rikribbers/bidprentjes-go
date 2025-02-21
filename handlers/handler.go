package handlers

import (
	"net/http"
	"time"

	"bidprentjes-api/models"
	"bidprentjes-api/store"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store *store.Store
}

func NewHandler() *Handler {
	return &Handler{
		store: store.NewStore(),
	}
}

func (h *Handler) CreateBidprentje(c *gin.Context) {
	var bidprentje models.Bidprentje
	if err := c.BindJSON(&bidprentje); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if bidprentje.ID == "" || bidprentje.Voornaam == "" || bidprentje.Achternaam == "" ||
		bidprentje.Geboorteplaats == "" || bidprentje.Overlijdensplaats == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Set timestamps
	bidprentje.CreatedAt = time.Now()
	bidprentje.UpdatedAt = time.Now()

	if err := h.store.Create(&bidprentje); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bidprentje"})
		return
	}

	c.JSON(http.StatusCreated, bidprentje)
}

func (h *Handler) GetBidprentje(c *gin.Context) {
	id := c.Param("id")
	bidprentje, exists := h.store.Get(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bidprentje not found"})
		return
	}

	c.JSON(http.StatusOK, bidprentje)
}

func (h *Handler) UpdateBidprentje(c *gin.Context) {
	id := c.Param("id")

	var bidprentje models.Bidprentje
	if err := c.BindJSON(&bidprentje); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bidprentje.ID = id
	bidprentje.UpdatedAt = time.Now()

	if err := h.store.Update(&bidprentje); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bidprentje"})
		return
	}

	c.JSON(http.StatusOK, bidprentje)
}

func (h *Handler) DeleteBidprentje(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bidprentje deleted successfully"})
}

func (h *Handler) ListBidprentjes(c *gin.Context) {
	var params models.SearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := h.store.List(params.Page, params.PageSize)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) SearchBidprentjes(c *gin.Context) {
	var params models.SearchParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := h.store.Search(params)
	c.JSON(http.StatusOK, response)
}
