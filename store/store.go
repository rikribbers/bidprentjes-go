package store

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"bidprentjes-api/models"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/lang/nl"
	"github.com/blevesearch/bleve/v2/analysis/token/lowercase"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/unicode"
	"github.com/blevesearch/bleve/v2/search/query"
)

type Store struct {
	data  map[string]*models.Bidprentje
	index bleve.Index
	mu    sync.RWMutex
}

func NewStore() *Store {
	// Create a custom analyzer for Dutch names
	indexMapping := bleve.NewIndexMapping()
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
		log.Fatal(err)
	}

	// Create document mapping
	docMapping := bleve.NewDocumentMapping()

	// Add field mappings with appropriate analyzers
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = "bidprentje"

	dateFieldMapping := bleve.NewDateTimeFieldMapping()
	boolFieldMapping := bleve.NewBooleanFieldMapping()

	// Configure field mappings
	docMapping.AddFieldMappingsAt("ID", textFieldMapping)
	docMapping.AddFieldMappingsAt("Voornaam", textFieldMapping)
	docMapping.AddFieldMappingsAt("Tussenvoegsel", textFieldMapping)
	docMapping.AddFieldMappingsAt("Achternaam", textFieldMapping)
	docMapping.AddFieldMappingsAt("Geboorteplaats", textFieldMapping)
	docMapping.AddFieldMappingsAt("Overlijdensplaats", textFieldMapping)
	docMapping.AddFieldMappingsAt("Geboortedatum", dateFieldMapping)
	docMapping.AddFieldMappingsAt("Overlijdensdatum", dateFieldMapping)
	docMapping.AddFieldMappingsAt("Scan", boolFieldMapping)

	indexMapping.DefaultMapping = docMapping

	// Create in-memory index
	index, err := bleve.NewMemOnly(indexMapping)
	if err != nil {
		log.Fatal(err)
	}

	return &Store{
		data:  make(map[string]*models.Bidprentje),
		index: index,
		mu:    sync.RWMutex{},
	}
}

func (s *Store) Create(b *models.Bidprentje) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[b.ID] = b
	return s.index.Index(b.ID, b)
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
	return s.index.Index(b.ID, b)
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

	// Sort items by CreatedAt descending (newest first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

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
	queryStr := strings.ToLower(params.Query)

	// Create individual field queries
	var queries []query.Query

	// Add exact match queries with higher boost
	exactFields := []struct {
		field string
		boost float64
	}{
		{"ID", 10.0},
		{"Voornaam", 5.0},
		{"Achternaam", 5.0},
		{"Tussenvoegsel", 3.0},
		{"Geboorteplaats", 3.0},
		{"Overlijdensplaats", 3.0},
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
		{"Voornaam", 2.0},
		{"Achternaam", 2.0},
		{"Geboorteplaats", 1.0},
		{"Overlijdensplaats", 1.0},
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

	// Pre-allocate map space
	s.mu.Lock()
	for _, b := range bidprentjes {
		s.data[b.ID] = b
	}
	s.mu.Unlock()

	// Create and fill batch
	batch := s.index.NewBatch()
	for _, b := range bidprentjes {
		if err := batch.Index(b.ID, b); err != nil {
			return fmt.Errorf("error adding to batch: %v", err)
		}
	}

	// Execute batch with a write lock
	s.mu.Lock()
	err := s.index.Batch(batch)
	s.mu.Unlock()

	if err != nil {
		return fmt.Errorf("error executing batch: %v", err)
	}

	return nil
}
