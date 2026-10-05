package handlers

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"bidprentjes-api/models"
	"bidprentjes-api/store"
	"bidprentjes-api/translations"

	"github.com/gin-gonic/gin"
)

// scanGUIDPattern restricts the :guid route parameter to a well-formed UUID,
// preventing path traversal or server-side request forgery when building the
// upstream CDN URL.
var scanGUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

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

// DownloadScan proxies a single scan image from the CDN and forces the
// browser to download it (via Content-Disposition), rather than merely
// opening it. This proxy is necessary because the CDN does not send CORS
// headers, so a plain cross-origin <a download> link is ignored by browsers.
func (h *Handler) DownloadScan(c *gin.Context) {
	guid := c.Param("guid")
	if !scanGUIDPattern.MatchString(guid) {
		c.String(http.StatusBadRequest, "invalid scan id")
		return
	}

	if h.cdnBaseURL == "" {
		c.String(http.StatusServiceUnavailable, "CDN is not configured")
		return
	}

	url := fmt.Sprintf("%s/%s.jpg", h.cdnBaseURL, guid)

	resp, err := http.Get(url)
	if err != nil {
		c.String(http.StatusBadGateway, "failed to fetch scan")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.Status(resp.StatusCode)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.jpg"`, guid))
	c.Header("Content-Type", "image/jpeg")
	if resp.ContentLength > 0 {
		c.Header("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	}
	c.Status(http.StatusOK)

	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		// Client likely disconnected mid-stream; nothing else to do.
		return
	}
}
