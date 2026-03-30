package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"bidprentjes-api/models"
	"bidprentjes-api/store"
	"bidprentjes-api/translations"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store *store.Store
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{
		store: store,
	}
}

// ProcessCSVUpload handles the processing of uploaded CSV files
func (h *Handler) ProcessCSVUpload(reader io.Reader) (int, error) {
	startTime := time.Now()
	defer func() {
		log.Printf("Total upload time: %v", time.Since(startTime))
	}()

	csvReader := csv.NewReader(reader)
	// Configure CSV reader to be more lenient with quotes
	csvReader.LazyQuotes = true       // Allow quotes inside fields
	csvReader.FieldsPerRecord = 9     // Expect exactly 9 fields per record
	csvReader.TrimLeadingSpace = true // Trim leading space from fields

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
					var geboortedatum, overlijdensdatum time.Time
					if record[4] != "" {
						var err error
						geboortedatum, err = time.Parse("2006-01-02", strings.TrimSpace(record[4]))
						if err != nil {
							log.Printf("Worker %d: Error parsing geboortedatum '%s': %v", workerId, record[4], err)
						}
					}

					if record[6] != "" {
						var err error
						overlijdensdatum, err = time.Parse("2006-01-02", strings.TrimSpace(record[6]))
						if err != nil {
							log.Printf("Worker %d: Error parsing overlijdensdatum '%s': %v", workerId, record[6], err)
						}
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
	})
}
