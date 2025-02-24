package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bidprentjes-api/models"
	"bidprentjes-api/store"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store    *store.Store
	readOnly bool
}

func NewHandler() *Handler {
	return &Handler{
		store: store.NewStore(),
	}
}

func (h *Handler) SetReadOnly(readonly bool) {
	h.readOnly = readonly
}

func (h *Handler) CreateBidprentje(c *gin.Context) {
	if h.readOnly {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		return
	}

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
	if h.readOnly {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		return
	}

	id := c.Param("id")
	bidprentje, exists := h.store.Get(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bidprentje not found"})
		return
	}

	c.JSON(http.StatusOK, bidprentje)
}

func (h *Handler) UpdateBidprentje(c *gin.Context) {
	if h.readOnly {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		return
	}

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
	if h.readOnly {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		return
	}

	id := c.Param("id")
	if err := h.store.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bidprentje deleted successfully"})
}

func (h *Handler) ListBidprentjes(c *gin.Context) {
	if h.readOnly {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		return
	}

	var params models.SearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := h.store.List(params.Page, params.PageSize)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) SearchBidprentjes(c *gin.Context) {
	if h.readOnly {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		return
	}

	var params models.SearchParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := h.store.Search(params)
	c.JSON(http.StatusOK, response)
}

func (h *Handler) ProcessCSVUpload(reader io.Reader) (int, error) {
	csvReader := csv.NewReader(reader)
	// Skip header
	if _, err := csvReader.Read(); err != nil {
		return 0, fmt.Errorf("invalid CSV format: %v", err)
	}

	count := 0
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if len(record) != 9 {
			continue
		}

		geboortedatum, _ := time.Parse("2006-01-02", record[4])
		overlijdensdatum, _ := time.Parse("2006-01-02", record[6])
		scan := record[8] == "true"

		bidprentje := &models.Bidprentje{
			ID:                record[0],
			Voornaam:          record[1],
			Tussenvoegsel:     record[2],
			Achternaam:        record[3],
			Geboortedatum:     geboortedatum,
			Geboorteplaats:    record[5],
			Overlijdensdatum:  overlijdensdatum,
			Overlijdensplaats: record[7],
			Scan:              scan,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		if err := h.store.Create(bidprentje); err != nil {
			continue
		}
		count++
	}

	return count, nil
}

func (h *Handler) WebHandler(c *gin.Context) {
	if h.readOnly {
		c.String(http.StatusNotFound, "Not Found")
		return
	}

	// Call the appropriate handler based on the path
	switch c.Request.URL.Path {
	case "/web":
		h.WebIndex(c)
	case "/web/create":
		h.WebCreate(c)
	case "/upload":
		h.WebUpload(c)
	default:
		if strings.HasPrefix(c.Request.URL.Path, "/web/edit/") {
			h.WebEdit(c)
		} else {
			c.String(http.StatusNotFound, "Not Found")
		}
	}
}
