package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
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

// ProcessCSVUpload handles the processing of uploaded CSV files
func (h *Handler) ProcessCSVUpload(reader io.Reader) (int, error) {
	startTime := time.Now()
	defer func() {
		log.Printf("Total upload time: %v", time.Since(startTime))
	}()

	csvReader := csv.NewReader(reader)
	// Skip header
	if _, err := csvReader.Read(); err != nil {
		log.Printf("Error reading CSV header: %v", err)
		return 0, fmt.Errorf("invalid CSV format: %v", err)
	}

	// Read all records first
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV records: %v", err)
		return 0, fmt.Errorf("error reading CSV: %v", err)
	}

	totalRecords := len(records)
	log.Printf("Processing %d records", totalRecords)

	// Optimize for Cloud Run: Use smaller chunks and fewer workers
	const chunkSize = 1000 // Smaller chunks for more frequent updates
	const numWorkers = 2   // Reduced workers for 1 vCPU
	chunks := (totalRecords + chunkSize - 1) / chunkSize

	// Pre-allocate batches to reduce memory allocations
	batchPool := sync.Pool{
		New: func() interface{} {
			return make([]*models.Bidprentje, 0, chunkSize)
		},
	}

	type chunkResult struct {
		chunkNum int
		batch    []*models.Bidprentje
		err      error
	}
	resultChan := make(chan chunkResult, 2) // Smaller buffer to reduce memory usage

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			for chunkNum := 0; chunkNum < chunks; chunkNum++ {
				// Each worker processes chunks based on their worker ID
				if chunkNum%numWorkers != workerId {
					continue
				}

				start := chunkNum * chunkSize
				end := start + chunkSize
				if end > totalRecords {
					end = totalRecords
				}

				// Get batch from pool
				batch := batchPool.Get().([]*models.Bidprentje)
				batch = batch[:0] // Reset slice but keep capacity

				// Process records in this chunk
				chunk := records[start:end]
				for _, record := range chunk {
					if len(record) != 9 {
						log.Printf("Worker %d: Invalid record length: got %d, want 9", workerId, len(record))
						continue
					}

					// Parse dates according to the format: YYYY-MM-DD
					geboortedatum, err := time.Parse("2006-01-02", record[4])
					if err != nil {
						log.Printf("Worker %d: Error parsing geboortedatum '%s': %v", workerId, record[4], err)
					}

					overlijdensdatum, err := time.Parse("2006-01-02", record[6])
					if err != nil {
						log.Printf("Worker %d: Error parsing overlijdensdatum '%s': %v", workerId, record[6], err)
					}

					scan := strings.ToLower(record[8]) == "true"

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
					}

					batch = append(batch, bidprentje)
				}

				// Send batch along with its chunk number
				resultChan <- chunkResult{
					chunkNum: chunkNum,
					batch:    batch,
				}

				if (chunkNum+1)%10 == 0 || chunkNum+1 == chunks {
					progress := ((chunkNum + 1) * 100) / chunks
					log.Printf("Worker %d progress: %d%% (%d/%d chunks)", workerId, progress, chunkNum+1, chunks)
				}
			}
		}(i)
	}

	// Wait for all workers to finish in a separate goroutine
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results immediately as they arrive
	processedChunks := make([]bool, chunks)
	nextChunkToProcess := 0
	var lastError error

	for result := range resultChan {
		if result.err != nil {
			lastError = result.err
			continue
		}

		// Mark this chunk as processed
		processedChunks[result.chunkNum] = true

		// Process any consecutive chunks that are ready
		for nextChunkToProcess < chunks && processedChunks[nextChunkToProcess] {
			// Get the batch for this chunk from the results
			if err := h.store.BatchCreate(result.batch); err != nil {
				log.Printf("Error storing batch %d: %v", nextChunkToProcess, err)
				lastError = err
			}

			// Return batch to pool
			batchPool.Put(result.batch)
			nextChunkToProcess++
		}
	}

	if lastError != nil {
		return 0, lastError
	}

	log.Printf("Successfully processed all %d records in %v", totalRecords, time.Since(startTime))
	return totalRecords, nil
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
