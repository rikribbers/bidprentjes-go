package handlers

import (
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"bidprentjes-api/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) WebIndex(c *gin.Context) {
	log.Printf("WebIndex handler called")

	// Get page from query params, default to 1
	page := 1
	pageSize := 25
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}

	// Get data from store
	response := h.store.List(page, pageSize)
	log.Printf("Found %d items in store", len(response.Items))

	c.HTML(http.StatusOK, "index.html", gin.H{
		"data": response,
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
	var params models.SearchParams
	if err := c.BindQuery(&params); err != nil {
		params.Page = 1
		params.PageSize = 25
	}

	// Get query from URL
	params.Query = c.DefaultQuery("query", "")
	log.Printf("Search query: %s, page: %d", params.Query, params.Page)

	// Get search results
	response := h.store.Search(params)
	log.Printf("Found %d items matching search", len(response.Items))

	c.HTML(http.StatusOK, "search.html", gin.H{
		"data":        response,
		"searchQuery": params.Query,
	})
}

func (h *Handler) WebUpload(c *gin.Context) {
	log.Printf("WebUpload handler called")
	c.HTML(http.StatusOK, "upload.html", nil)
}

func (h *Handler) UploadCSV(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip header
	if _, err := reader.Read(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format"})
		return
	}

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if len(record) != 8 {
			continue
		}

		geboortedatum, _ := time.Parse("2006-01-02", record[3])
		overlijdensdatum, _ := time.Parse("2006-01-02", record[5])
		scan := record[7] == "true"

		bidprentje := &models.Bidprentje{
			ID:                uuid.New().String(),
			Voornaam:          record[0],
			Tussenvoegsel:     record[1],
			Achternaam:        record[2],
			Geboortedatum:     geboortedatum,
			Geboorteplaats:    record[4],
			Overlijdensdatum:  overlijdensdatum,
			Overlijdensplaats: record[6],
			Scan:              scan,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		if err := h.store.Create(bidprentje); err != nil {
			continue
		}
		count++
	}

	c.JSON(http.StatusOK, gin.H{"message": "Upload complete", "count": count})
}
