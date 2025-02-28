package handlers

import (
	"log"
	"net/http"

	"bidprentjes-api/models"
	"bidprentjes-api/translations"

	"github.com/gin-gonic/gin"
)

func (h *Handler) WebIndex(c *gin.Context) {
	log.Printf("WebIndex handler called")

	// Get language preference, default to Dutch
	lang := c.DefaultQuery("lang", "nl")

	// Get translations
	t := translations.GetTranslation(lang)

	// Get search and pagination params
	var params models.SearchParams
	if err := c.BindQuery(&params); err != nil {
		params.Page = 1
		params.PageSize = 25
	}

	// Ensure default values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 25
	}

	// Get query from URL and ensure it's properly set in params
	params.Query = c.DefaultQuery("query", "")
	log.Printf("Search query: %s, page: %d", params.Query, params.Page)

	var response *models.PaginatedResponse
	if params.Query != "" {
		// If there's a search query, use search
		response = h.store.Search(params)
		log.Printf("Found %d items matching search", response.TotalCount)
	} else {
		// Otherwise show all items
		response = h.store.List(params.Page, params.PageSize)
		log.Printf("Found %d total items in store", response.TotalCount)
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"data":        response,
		"searchQuery": params.Query,
		"t":           t,
		"lang":        lang,
		"languages":   translations.SupportedLanguages,
	})
}

func (h *Handler) WebCreate(c *gin.Context) {
	log.Printf("WebCreate handler called")

	// Get language preference, default to Dutch
	lang := c.DefaultQuery("lang", "nl")

	// Get translations
	t := translations.GetTranslation(lang)

	c.HTML(http.StatusOK, "create.html", gin.H{
		"t":         t,
		"lang":      lang,
		"languages": translations.SupportedLanguages,
	})
}

func (h *Handler) WebEdit(c *gin.Context) {
	log.Printf("WebEdit handler called")
	id := c.Param("id")

	// Get language preference, default to Dutch
	lang := c.DefaultQuery("lang", "nl")

	// Get translations
	t := translations.GetTranslation(lang)

	bidprentje, exists := h.store.Get(id)
	if !exists {
		c.Redirect(http.StatusFound, "/web")
		return
	}

	c.HTML(http.StatusOK, "edit.html", gin.H{
		"bidprentje": bidprentje,
		"t":          t,
		"lang":       lang,
		"languages":  translations.SupportedLanguages,
	})
}

func (h *Handler) WebSearch(c *gin.Context) {
	log.Printf("WebSearch handler called")

	// Get language preference, default to Dutch
	lang := c.DefaultQuery("lang", "nl")

	// Get translations
	t := translations.GetTranslation(lang)

	var params models.SearchParams
	if err := c.BindQuery(&params); err != nil {
		params.Page = 1
		params.PageSize = 25
	}

	// Ensure default values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 25
	}

	// Get query from URL
	params.Query = c.DefaultQuery("query", "")
	log.Printf("Search query: %s, page: %d", params.Query, params.Page)

	// Get search results
	response := h.store.Search(params)
	log.Printf("Found %d items matching search", response.TotalCount)

	c.HTML(http.StatusOK, "search.html", gin.H{
		"data":        response,
		"searchQuery": params.Query,
		"t":           t,
		"lang":        lang,
		"languages":   translations.SupportedLanguages,
	})
}

func (h *Handler) WebUpload(c *gin.Context) {
	log.Printf("WebUpload handler called")

	// Get language preference, default to Dutch
	lang := c.DefaultQuery("lang", "nl")

	// Get translations
	t := translations.GetTranslation(lang)

	c.HTML(http.StatusOK, "upload.html", gin.H{
		"t":         t,
		"lang":      lang,
		"languages": translations.SupportedLanguages,
	})
}

func (h *Handler) UploadCSV(c *gin.Context) {
	if h.readOnly {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		return
	}

	log.Printf("UploadCSV handler called")

	// Set headers to prevent caching
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("Error getting form file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	log.Printf("Processing file: %s, size: %d bytes", fileHeader.Filename, fileHeader.Size)

	count, err := h.ProcessCSVUpload(file)
	if err != nil {
		log.Printf("Error processing CSV: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Upload complete, sending response with count: %d", count)
	c.JSON(http.StatusOK, gin.H{
		"message": "Upload complete",
		"count":   count,
	})
}
