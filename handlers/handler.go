package handlers

import (
	"bidprentjes-api/models"
	"bidprentjes-api/store"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	if err := c.ShouldBindJSON(&bidprentje); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bidprentje.ID = uuid.New().String()
	bidprentje.CreatedAt = time.Now()
	bidprentje.UpdatedAt = time.Now()

	h.store.Insert(&bidprentje)

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

	if _, exists := h.store.Get(id); !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bidprentje not found"})
		return
	}

	var updatedBidprentje models.Bidprentje
	if err := c.ShouldBindJSON(&updatedBidprentje); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedBidprentje.ID = id
	updatedBidprentje.UpdatedAt = time.Now()

	h.store.Update(&updatedBidprentje)

	c.JSON(http.StatusOK, updatedBidprentje)
}

func (h *Handler) DeleteBidprentje(c *gin.Context) {
	id := c.Param("id")

	if _, exists := h.store.Get(id); !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bidprentje not found"})
		return
	}

	h.store.Delete(id)

	c.JSON(http.StatusOK, gin.H{"message": "Bidprentje deleted successfully"})
}

func (h *Handler) ListBidprentjes(c *gin.Context) {
	var params models.ListParams
	if err := c.BindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if params.PageSize == 0 {
		params.PageSize = 25
	}
	if params.Page == 0 {
		params.Page = 1
	}

	response, err := h.store.ListPaginated(params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) SearchBidprentjes(c *gin.Context) {
	var params models.ListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if params.PageSize == 0 {
		params.PageSize = 25
	}
	if params.Page == 0 {
		params.Page = 1
	}

	response, err := h.store.SearchPaginated(params.Query, params.Page, params.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) WebIndex(c *gin.Context) {
	var params models.ListParams
	if err := c.BindQuery(&params); err != nil {
		c.Redirect(http.StatusFound, "/web")
		return
	}

	if params.PageSize == 0 {
		params.PageSize = 25
	}
	if params.Page == 0 {
		params.Page = 1
	}

	var response *models.PaginatedResponse
	var err error

	// If search query is present, use search instead of list
	if params.Query != "" {
		response, err = h.store.SearchPaginated(params.Query, params.Page, params.PageSize)
	} else {
		response, err = h.store.ListPaginated(params.Page, params.PageSize)
	}

	if err != nil {
		c.Redirect(http.StatusFound, "/web")
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"data":        response,
		"searchQuery": params.Query, // Pass the search query to the template
	})
}

func (h *Handler) WebCreate(c *gin.Context) {
	c.HTML(http.StatusOK, "create.html", nil)
}

func (h *Handler) WebEdit(c *gin.Context) {
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

func (h *Handler) WebUpload(c *gin.Context) {
	c.HTML(http.StatusOK, "upload.html", nil)
}
