package handlers

import (
	"net/http"
	"strconv"

	"bidprentjes-api/models"
	"bidprentjes-api/store"
	"bidprentjes-api/translations"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store      *store.Store
	cdnBaseURL string
}

func NewHandler(store *store.Store, cdnBaseURL string) *Handler {
	return &Handler{
		store:      store,
		cdnBaseURL: cdnBaseURL,
	}
}

func (h *Handler) WebSearch(c *gin.Context) {

	query := c.Query("query")
	lang := c.DefaultQuery("lang", "nl") // Default to Dutch
	exactMatch := c.Query("exact_match") == "on"

	// Parse page and pageSize from query parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	var response *models.PaginatedResponse
	if query != "" {
		response = h.store.Search(models.SearchParams{
			Query:      query,
			Page:       page,
			PageSize:   pageSize,
			ExactMatch: exactMatch,
		})
	} else {
		response = h.store.List(page, pageSize)
	}

	t := translations.GetTranslation(lang)
	languages := translations.SupportedLanguages

	c.HTML(http.StatusOK, "search.html", gin.H{
		"data":        response,
		"searchQuery": query,
		"lang":        lang,
		"languages":   languages,
		"t":           t,
		"title":       t.Search,
		"description": t.SearchHelp,
		"exactMatch":  exactMatch,
		"cdnBaseURL":  h.cdnBaseURL,
	})
}
