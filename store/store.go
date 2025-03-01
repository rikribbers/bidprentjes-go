package store

import (
	"archive/tar"
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

	"bidprentjes-api/models"

	"cloud.google.com/go/storage"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/lang/nl"
	"github.com/blevesearch/bleve/v2/analysis/token/lowercase"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/unicode"
	"github.com/blevesearch/bleve/v2/search/query"
	"google.golang.org/api/option"
)

const (
	indexPath   = "/tmp/bidprentjes.bleve"
	bucketName  = "bidprentjes-go-storage"
	indexObject = "index/bidprentjes.bleve.tar.gz"
	csvObject   = "data/bidprentjes.csv"
)

type Store struct {
	data       map[string]*models.Bidprentje
	index      bleve.Index
	mu         sync.RWMutex
	gcsClient  *storage.Client
	syncTicker *time.Ticker
	done       chan bool
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

func NewStore() *Store {
	ctx := context.Background()

	// Create store instance with empty fields
	s := &Store{
		data: make(map[string]*models.Bidprentje),
		done: make(chan bool),
	}

	// Try to initialize GCS client, but continue without it if it fails
	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		log.Printf("Failed to create GCS client, continuing in local-only mode: %v", err)
	} else {
		s.gcsClient = client
	}

	// First try to open existing index
	if err := s.openExistingIndex(); err != nil {
		log.Printf("No existing index found or error opening: %v", err)
	}

	// Check for local CSV file first
	localCsvPath := "bidprentjes.csv"
	if _, err := os.Stat(localCsvPath); err == nil {
		log.Printf("Found local CSV file %s, recreating index...", localCsvPath)
		if err := s.recreateIndexFromCSV(localCsvPath); err != nil {
			log.Fatalf("Failed to process local CSV file: %v", err)
		}
		// Rename local CSV after processing
		if err := os.Rename(localCsvPath, localCsvPath+".processed"); err != nil {
			log.Printf("Error renaming local CSV file: %v", err)
		}
	} else if s.gcsClient != nil && s.index == nil {
		// Check for CSV in GCS only if we have a client and no index yet
		bucket := s.gcsClient.Bucket(bucketName)
		csvObj := bucket.Object(csvObject)

		// Check if CSV exists in GCS
		_, err = csvObj.Attrs(ctx)
		if err == nil {
			log.Printf("Found initialization file %s in GCS bucket, recreating index...", csvObject)

			// Download and process CSV file
			reader, err := csvObj.NewReader(ctx)
			if err != nil {
				log.Printf("Failed to read CSV from GCS: %v", err)
			} else {
				defer reader.Close()
				if err := s.recreateIndexFromReader(reader); err != nil {
					log.Printf("Failed to process GCS CSV file: %v", err)
				} else {
					// Move processed CSV to processed folder
					processedObj := bucket.Object("data/processed/" + filepath.Base(csvObject) + "." + time.Now().Format("20060102150405"))
					if _, err := processedObj.CopierFrom(csvObj).Run(ctx); err != nil {
						log.Printf("Error copying CSV to processed folder: %v", err)
					} else {
						// Delete original CSV only if copy was successful
						if err := csvObj.Delete(ctx); err != nil {
							log.Printf("Error deleting original CSV: %v", err)
						}
					}
				}
			}
		}
	}

	// If still no index exists, create one
	if s.index == nil {
		// Try to download existing index from GCS first if we have a client
		if s.gcsClient != nil {
			if err := s.downloadIndex(ctx); err != nil {
				log.Printf("Could not download index from GCS: %v", err)
			}
		}

		// If still no index, create a new one
		if s.index == nil {
			if err := s.createNewIndex(); err != nil {
				log.Fatalf("Failed to create new index: %v", err)
			}
		}
	}

	// Rebuild in-memory data from index
	if err := s.rebuildDataFromIndex(); err != nil {
		log.Printf("Error rebuilding data from index: %v", err)
	}

	// Start periodic sync only if we have a GCS client
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

// Helper function to recreate index from a CSV file
func (s *Store) recreateIndexFromCSV(csvPath string) error {
	if err := s.createNewIndex(); err != nil {
		return err
	}

	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	return s.recreateIndexFromReader(file)
}

// Helper function to recreate index from a reader
func (s *Store) recreateIndexFromReader(reader io.Reader) error {
	count, err := s.ProcessCSVUpload(reader)
	if err != nil {
		return fmt.Errorf("failed to process CSV: %v", err)
	}
	log.Printf("Successfully processed %d records", count)

	// Upload new index to GCS if we have a client
	if s.gcsClient != nil {
		ctx := context.Background()
		if err := s.uploadIndex(ctx); err != nil {
			log.Printf("Failed to upload initial index to GCS: %v", err)
		}
	}

	return nil
}

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

	log.Printf("Rebuilt in-memory data with %d records", len(s.data))
	return nil
}

func (s *Store) openOrCreateIndex() error {
	var err error
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		// Create a new index
		mapping := bleve.NewIndexMapping()
		s.index, err = bleve.New(indexPath, mapping)
	} else {
		// Open existing index
		s.index, err = bleve.Open(indexPath)
	}
	return err
}

func (s *Store) downloadIndex(ctx context.Context) error {
	bucket := s.gcsClient.Bucket(bucketName)
	obj := bucket.Object(indexObject)

	// Check if index exists in GCS
	_, err := obj.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return fmt.Errorf("no index found in GCS")
	}
	if err != nil {
		return fmt.Errorf("error checking index in GCS: %v", err)
	}

	// Create temporary directory
	if err := os.MkdirAll(filepath.Dir(indexPath), 0755); err != nil {
		return fmt.Errorf("failed to create index directory: %v", err)
	}

	// Download and extract index
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("failed to create reader: %v", err)
	}
	defer reader.Close()

	// Run tar command to extract
	if err := extractTarGz(reader, filepath.Dir(indexPath)); err != nil {
		return fmt.Errorf("failed to extract index: %v", err)
	}

	return nil
}

func (s *Store) uploadIndex(ctx context.Context) error {
	bucket := s.gcsClient.Bucket(bucketName)
	obj := bucket.Object(indexObject)

	// Create writer
	writer := obj.NewWriter(ctx)

	// Create tar.gz of the index directory
	if err := createTarGz(indexPath, writer); err != nil {
		writer.Close()
		return fmt.Errorf("failed to create tar.gz: %v", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
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
	startTime := time.Now()
	defer func() {
		log.Printf("Search completed in %v", time.Since(startTime))
	}()

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

	searchStartTime := time.Now()
	searchResults, err := s.index.Search(searchRequest)
	log.Printf("Bleve search completed in %v", time.Since(searchStartTime))

	if err != nil {
		log.Printf("Search error: %v", err)
		return &models.PaginatedResponse{
			Items:      []models.Bidprentje{},
			TotalCount: 0,
			Page:       params.Page,
			PageSize:   params.PageSize,
		}
	}

	// Convert results to Bidprentje objects
	items := make([]models.Bidprentje, 0, len(searchResults.Hits))
	for _, hit := range searchResults.Hits {
		if b, exists := s.data[hit.ID]; exists {
			items = append(items, *b)
		}
	}

	log.Printf("Found %d items matching search", len(items))
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
		// Swap slices
		previous, current = current, previous
	}

	if lastError != nil {
		return 0, lastError
	}

	log.Printf("Successfully processed all %d records in %v", totalRecords, time.Since(startTime))
	return totalRecords, nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
