package store

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bidprentjes-api/cloud"
	"bidprentjes-api/models"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/lang/nl"
	"github.com/blevesearch/bleve/v2/analysis/token/lowercase"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/unicode"
	"github.com/blevesearch/bleve/v2/search/query"
)

const (
	indexPath   = "/tmp/bidprentjes.bleve"
	indexObject = "index/bidprentjes.bleve.tar.gz"
	csvObject   = "data/bidprentjes.csv"
)

type Store struct {
	data          map[string]*models.Bidprentje
	index         bleve.Index
	mu            sync.RWMutex
	gcsClient     *cloud.StorageClient
	bucketName    string
	syncTicker    *time.Ticker
	done          chan bool
	hasValidIndex bool
}

// BleveDocument represents a document in the Bleve index
type BleveDocument struct {
	ID                string `json:"id"`
	Voornaam          string `json:"voornaam"`
	Tussenvoegsel     string `json:"tussenvoegsel"`
	Achternaam        string `json:"achternaam"`
	Geboortedatum     string `json:"geboortedatum"`
	Geboorteplaats    string `json:"geboorteplaats"`
	Overlijdensdatum  string `json:"overlijdensdatum"`
	Overlijdensplaats string `json:"overlijdensplaats"`
	Scan              bool   `json:"scan"`
}

func NewStore(ctx context.Context, bucketName string) *Store {
	// Create store instance with empty fields
	s := &Store{
		data:          make(map[string]*models.Bidprentje),
		bucketName:    bucketName,
		done:          make(chan bool),
		hasValidIndex: false,
	}

	// Try to initialize GCS client
	if bucketName != "" {
		client, err := cloud.NewStorageClient(ctx, bucketName)
		if err != nil {
			log.Printf("Failed to create GCS client, continuing in local-only mode: %v", err)
		} else {
			s.gcsClient = client
		}
	}

	// First try to restore index from GCP backup
	if s.gcsClient != nil {
		if err := s.downloadIndex(ctx); err != nil {
			log.Printf("Could not download index from GCP: %v", err)
		} else {
			log.Printf("Successfully restored index from GCP backup")
			if err := s.openExistingIndex(); err != nil {
				log.Printf("Error opening restored index: %v", err)
			} else {
				// Rebuild in-memory data from restored index
				if err := s.rebuildDataFromIndex(); err != nil {
					log.Printf("Error rebuilding data from restored index: %v", err)
				} else {
					// Start periodic sync
					s.startPeriodicSync()
					s.hasValidIndex = true
					return s
				}
			}
		}
	}

	// If no index exists or restore failed, check for local CSV
	if err := s.createNewIndex(); err != nil {
		log.Fatalf("Failed to create new index: %v", err)
	}

	// Start periodic sync if we have GCS client
	if s.gcsClient != nil {
		s.startPeriodicSync()
	}

	return s
}

// Helper function to create a new index with proper mapping
func (s *Store) createNewIndex() error {
	// Remove existing index if it exists
	if err := os.RemoveAll(indexPath); err != nil {
		log.Printf("Error removing existing index: %v", err)
	}

	// Create new index with proper mapping
	indexMapping := bleve.NewIndexMapping()

	// Add custom analyzer for Dutch names
	err := indexMapping.AddCustomAnalyzer("bidprentje",
		map[string]interface{}{
			"type":      custom.Name,
			"tokenizer": unicode.Name,
			"token_filters": []string{
				lowercase.Name,
				nl.StopName,
			},
		})
	if err != nil {
		return fmt.Errorf("failed to create analyzer: %v", err)
	}

	// Create document mapping
	docMapping := bleve.NewDocumentMapping()

	// Create field mappings with storage enabled
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Store = true
	textFieldMapping.Index = true
	textFieldMapping.Analyzer = "bidprentje"

	boolFieldMapping := bleve.NewBooleanFieldMapping()
	boolFieldMapping.Store = true
	boolFieldMapping.Index = true

	// Configure field mappings
	docMapping.AddFieldMappingsAt("_id", textFieldMapping)
	docMapping.AddFieldMappingsAt("id", textFieldMapping)
	docMapping.AddFieldMappingsAt("voornaam", textFieldMapping)
	docMapping.AddFieldMappingsAt("tussenvoegsel", textFieldMapping)
	docMapping.AddFieldMappingsAt("achternaam", textFieldMapping)
	docMapping.AddFieldMappingsAt("geboorteplaats", textFieldMapping)
	docMapping.AddFieldMappingsAt("overlijdensplaats", textFieldMapping)
	docMapping.AddFieldMappingsAt("geboortedatum", textFieldMapping)
	docMapping.AddFieldMappingsAt("overlijdensdatum", textFieldMapping)
	docMapping.AddFieldMappingsAt("scan", boolFieldMapping)

	indexMapping.DefaultMapping = docMapping
	indexMapping.DefaultAnalyzer = "bidprentje"

	// Create new index
	index, err := bleve.New(indexPath, indexMapping)
	if err != nil {
		return fmt.Errorf("failed to create index: %v", err)
	}
	s.index = index
	return nil
}

// Helper function to open existing index
func (s *Store) openExistingIndex() error {
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return fmt.Errorf("index does not exist")
	}

	index, err := bleve.Open(indexPath)
	if err != nil {
		return fmt.Errorf("failed to open index: %v", err)
	}

	s.index = index
	return nil
}

// Helper function to rebuild in-memory data from index
func (s *Store) rebuildDataFromIndex() error {
	// Create a search request that matches all documents
	matchAll := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(matchAll)
	searchRequest.Size = 100000 // Adjust this number based on your expected maximum documents
	searchRequest.Fields = []string{"*"}

	results, err := s.index.Search(searchRequest)
	if err != nil {
		return fmt.Errorf("failed to search index: %v", err)
	}

	// Clear existing data
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]*models.Bidprentje)

	// Rebuild data from search results
	for _, hit := range results.Hits {
		b := &models.Bidprentje{
			ID:                hit.ID,
			Voornaam:          hit.Fields["voornaam"].(string),
			Tussenvoegsel:     hit.Fields["tussenvoegsel"].(string),
			Achternaam:        hit.Fields["achternaam"].(string),
			Geboorteplaats:    hit.Fields["geboorteplaats"].(string),
			Overlijdensplaats: hit.Fields["overlijdensplaats"].(string),
			Scan:              hit.Fields["scan"].(bool),
		}

		// Parse dates
		if geboortedatum, ok := hit.Fields["geboortedatum"].(string); ok && geboortedatum != "" {
			if parsed, err := time.Parse("2006-01-02", geboortedatum); err == nil {
				b.Geboortedatum = parsed
			}
		}
		if overlijdensdatum, ok := hit.Fields["overlijdensdatum"].(string); ok && overlijdensdatum != "" {
			if parsed, err := time.Parse("2006-01-02", overlijdensdatum); err == nil {
				b.Overlijdensdatum = parsed
			}
		}

		s.data[hit.ID] = b
	}

	return nil
}

func (s *Store) startPeriodicSync() {
	s.syncTicker = time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-s.syncTicker.C:
				ctx := context.Background()
				if err := s.uploadIndex(ctx); err != nil {
					log.Printf("Failed to sync index to GCS: %v", err)
				} else {
					log.Printf("Successfully synced index to GCS")
				}
			case <-s.done:
				s.syncTicker.Stop()
				return
			}
		}
	}()
}

func (s *Store) Close() error {
	if s.syncTicker != nil {
		s.done <- true
		close(s.done)
	}

	// Final sync to GCS if we have a client
	if s.gcsClient != nil {
		ctx := context.Background()
		if err := s.uploadIndex(ctx); err != nil {
			log.Printf("Failed final index sync to GCS: %v", err)
		}
	}

	if err := s.index.Close(); err != nil {
		return fmt.Errorf("failed to close index: %v", err)
	}

	if s.gcsClient != nil {
		return s.gcsClient.Close()
	}
	return nil
}

// Helper functions for tar.gz operations
func createTarGz(src string, writer io.Writer) error {
	gzWriter := gzip.NewWriter(writer)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk through the source directory
	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return fmt.Errorf("failed to create tar header: %v", err)
		}

		// Update header name to be relative to src directory
		relPath, err := filepath.Rel(src, file)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %v", err)
		}

		// If it's a regular file, write the contents
		if !fi.Mode().IsRegular() {
			return nil
		}

		// Open and copy file contents
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
		}
		defer f.Close()

		if _, err := io.Copy(tarWriter, f); err != nil {
			return fmt.Errorf("failed to copy file contents: %v", err)
		}

		return nil
	})
}

func extractTarGz(reader io.Reader, dst string) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %v", err)
		}

		// Get the target path
		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeReg:
			// Create containing directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %v", err)
			}

			// Create the file
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			defer f.Close()

			// Copy contents
			if _, err := io.Copy(f, tarReader); err != nil {
				return fmt.Errorf("failed to copy file contents: %v", err)
			}
		}
	}

	return nil
}

func (s *Store) Create(b *models.Bidprentje) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[b.ID] = b

	// Create Bleve document
	doc := BleveDocument{
		ID:                b.ID,
		Voornaam:          b.Voornaam,
		Tussenvoegsel:     b.Tussenvoegsel,
		Achternaam:        b.Achternaam,
		Geboortedatum:     b.Geboortedatum.Format("2006-01-02"),
		Geboorteplaats:    b.Geboorteplaats,
		Overlijdensdatum:  b.Overlijdensdatum.Format("2006-01-02"),
		Overlijdensplaats: b.Overlijdensplaats,
		Scan:              b.Scan,
	}

	return s.index.Index(b.ID, doc)
}

func (s *Store) Get(id string) (*models.Bidprentje, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, exists := s.data[id]
	return b, exists
}

func (s *Store) Update(b *models.Bidprentje) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[b.ID] = b

	// Create Bleve document
	doc := BleveDocument{
		ID:                b.ID,
		Voornaam:          b.Voornaam,
		Tussenvoegsel:     b.Tussenvoegsel,
		Achternaam:        b.Achternaam,
		Geboortedatum:     b.Geboortedatum.Format("2006-01-02"),
		Geboorteplaats:    b.Geboorteplaats,
		Overlijdensdatum:  b.Overlijdensdatum.Format("2006-01-02"),
		Overlijdensplaats: b.Overlijdensplaats,
		Scan:              b.Scan,
	}

	return s.index.Index(b.ID, doc)
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, id)
	return s.index.Delete(id)
}

func (s *Store) List(page, pageSize int) *models.PaginatedResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert map to slice
	items := make([]models.Bidprentje, 0, len(s.data))
	for _, item := range s.data {
		items = append(items, *item)
	}

	// Calculate pagination
	totalCount := len(items)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	return &models.PaginatedResponse{
		Items:      items[start:end],
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}
}

func (s *Store) Search(params models.SearchParams) *models.PaginatedResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if params.Query == "" {
		return &models.PaginatedResponse{
			Items:      []models.Bidprentje{},
			TotalCount: 0,
			Page:       params.Page,
			PageSize:   params.PageSize,
		}
	}

	// Create a multi-field query that searches across all text fields
	queryStr := strings.TrimSpace(params.Query)

	// Create individual field queries
	var queries []query.Query

	// Add exact match queries with higher boost
	exactFields := []struct {
		field string
		boost float64
	}{
		{"id", 10.0},
		{"voornaam", 5.0},
		{"achternaam", 5.0},
		{"tussenvoegsel", 3.0},
		{"geboorteplaats", 3.0},
		{"overlijdensplaats", 3.0},
	}

	for _, f := range exactFields {
		q := query.NewMatchQuery(queryStr)
		q.SetField(f.field)
		q.SetBoost(f.boost)
		queries = append(queries, q)
	}

	// Add fuzzy match queries with lower boost
	fuzzyFields := []struct {
		field string
		boost float64
	}{
		{"voornaam", 2.0},
		{"achternaam", 2.0},
		{"geboorteplaats", 1.0},
		{"overlijdensplaats", 1.0},
	}

	for _, f := range fuzzyFields {
		q := query.NewFuzzyQuery(queryStr)
		q.SetField(f.field)
		q.SetBoost(f.boost)
		q.SetFuzziness(1)
		queries = append(queries, q)
	}

	// Combine all queries with OR
	searchQuery := query.NewDisjunctionQuery(queries)

	searchRequest := bleve.NewSearchRequest(searchQuery)
	searchRequest.Size = params.PageSize
	searchRequest.From = (params.Page - 1) * params.PageSize
	searchRequest.SortBy([]string{"-_score"}) // Sort by score descending
	searchRequest.Fields = []string{"*"}      // Request all stored fields

	startTime := time.Now()
	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		log.Printf("Search error: %v", err)
		return &models.PaginatedResponse{
			Items:      []models.Bidprentje{},
			TotalCount: 0,
			Page:       params.Page,
			PageSize:   params.PageSize,
		}
	}
	log.Printf("Found %d items in %v", searchResults.Total, time.Since(startTime))

	// Convert results to Bidprentje objects
	items := make([]models.Bidprentje, 0, len(searchResults.Hits))
	for _, hit := range searchResults.Hits {
		if b, exists := s.data[hit.ID]; exists {
			items = append(items, *b)
		}
	}

	return &models.PaginatedResponse{
		Items:      items,
		TotalCount: int(searchResults.Total),
		Page:       params.Page,
		PageSize:   params.PageSize,
	}
}

// BatchCreate adds multiple bidprentjes in a single batch operation
func (s *Store) BatchCreate(bidprentjes []*models.Bidprentje) error {
	if len(bidprentjes) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	batch := s.index.NewBatch()
	for _, b := range bidprentjes {
		s.data[b.ID] = b

		doc := BleveDocument{
			ID:                b.ID,
			Voornaam:          b.Voornaam,
			Tussenvoegsel:     b.Tussenvoegsel,
			Achternaam:        b.Achternaam,
			Geboortedatum:     b.Geboortedatum.Format("2006-01-02"),
			Geboorteplaats:    b.Geboorteplaats,
			Overlijdensdatum:  b.Overlijdensdatum.Format("2006-01-02"),
			Overlijdensplaats: b.Overlijdensplaats,
			Scan:              b.Scan,
		}

		if err := batch.Index(b.ID, doc); err != nil {
			return fmt.Errorf("failed to add document to batch: %v", err)
		}
	}

	return s.index.Batch(batch)
}

// ProcessCSVUpload processes a CSV file and adds its contents to the index
func (s *Store) ProcessCSVUpload(reader io.Reader) (int, error) {
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

	// Process records in chunks
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

					// Parse dates and convert to RFC3339 format for Bleve compatibility
					geboortedatum, err := time.Parse("2006-01-02", strings.TrimSpace(record[4]))
					if err != nil {
						log.Printf("Worker %d: Error parsing geboortedatum '%s': %v", workerId, record[4], err)
						// Set to zero time if parsing fails
						geboortedatum = time.Time{}
					}

					overlijdensdatum, err := time.Parse("2006-01-02", strings.TrimSpace(record[6]))
					if err != nil {
						log.Printf("Worker %d: Error parsing overlijdensdatum '%s': %v", workerId, record[6], err)
						// Set to zero time if parsing fails
						overlijdensdatum = time.Time{}
					}

					scan := strings.ToLower(record[8]) == "true"

					// Only create record if at least one date is valid
					if !geboortedatum.IsZero() || !overlijdensdatum.IsZero() {
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
					} else {
						log.Printf("Worker %d: Skipping record with invalid dates: %v", workerId, record)
					}
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
			if err := s.BatchCreate(result.batch); err != nil {
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

	s.mu.Lock()
	s.hasValidIndex = true
	s.mu.Unlock()

	log.Printf("Successfully processed all %d records in %v", totalRecords, time.Since(startTime))
	return totalRecords, nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// HasGCPConnectivity returns true if the store has a connection to GCP
func (s *Store) HasGCPConnectivity() bool {
	return s.gcsClient != nil
}

// downloadIndex downloads and extracts the index backup from GCP
func (s *Store) downloadIndex(ctx context.Context) error {
	reader, err := s.gcsClient.DownloadFile(ctx, indexObject)
	if err != nil {
		return fmt.Errorf("failed to download index: %v", err)
	}

	// Create temporary directory
	if err := os.MkdirAll(filepath.Dir(indexPath), 0755); err != nil {
		return fmt.Errorf("failed to create index directory: %v", err)
	}

	// Extract the tar.gz
	if err := extractTarGz(reader, filepath.Dir(indexPath)); err != nil {
		return fmt.Errorf("failed to extract index: %v", err)
	}

	return nil
}

// uploadIndex creates a tar.gz of the index and uploads it to GCP
func (s *Store) uploadIndex(ctx context.Context) error {
	// Create a buffer to write the tar.gz to
	var buf bytes.Buffer

	// Create tar.gz of the index directory
	if err := createTarGz(indexPath, &buf); err != nil {
		return fmt.Errorf("failed to create tar.gz: %v", err)
	}

	// Create a reader from the buffer
	reader := bytes.NewReader(buf.Bytes())

	// Upload the tar.gz to GCP
	if err := s.gcsClient.UploadFile(ctx, indexObject, reader); err != nil {
		return fmt.Errorf("failed to upload index: %v", err)
	}

	return nil
}

// BackupIndex creates an immediate backup of the index to GCP
func (s *Store) BackupIndex(ctx context.Context) error {
	if s.gcsClient == nil {
		return fmt.Errorf("no GCP connectivity available")
	}
	return s.uploadIndex(ctx)
}

// HasValidIndex returns true if we have successfully restored or created an index with data
func (s *Store) HasValidIndex() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hasValidIndex
}
