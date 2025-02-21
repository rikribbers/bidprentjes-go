package handlers

import (
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"time"

	"bidprentjes-api/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) WebIndex(c *gin.Context) {
	log.Printf("WebIndex handler called")

	// Get search and pagination params
	var params models.SearchParams
	if err := c.BindQuery(&params); err != nil {
		params.Page = 1
		params.PageSize = 25
	}

	// Get query from URL
	params.Query = c.DefaultQuery("query", "")
	log.Printf("Search query: %s, page: %d", params.Query, params.Page)

	var response *models.PaginatedResponse
	if params.Query != "" {
		// If there's a search query, use search
		response = h.store.Search(params)
		log.Printf("Found %d items matching search", len(response.Items))
	} else {
		// Otherwise show all items
		response = h.store.List(params.Page, params.PageSize)
		log.Printf("Found %d items in store", len(response.Items))
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"data":        response,
		"searchQuery": params.Query,
	})
}

func (h *Handler) WebCreate(c *gin.Context) {
	log.Printf("WebCreate handler called")
	c.HTML(http.StatusOK, "create.html", nil)
}

func (h *Handler) WebEdit(c *gin.Context) {
	log.Printf("WebEdit handler called")
	id := c.Param("id")

	bidprentje, exists := h.store.Get(id)
	if !exists {
		c.Redirect(http.StatusFound, "/web")
		return
	}

	c.HTML(http.StatusOK, "edit.html", gin.H{
		"bidprentje": bidprentje,
	})
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

	c.JSON(http.StatusOK, gin.H{"message": "Upload complete", "count": count})
}
