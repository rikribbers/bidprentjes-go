package store

import (
	"sort"
	"strings"
	"sync"

	"bidprentjes-api/models"
)

type Store struct {
	data map[string]*models.Bidprentje
	trie *Trie
	mu   sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]*models.Bidprentje),
		trie: NewTrie(),
		mu:   sync.RWMutex{},
	}
}

func (s *Store) Create(b *models.Bidprentje) error {
	s.data[b.ID] = b
	s.indexBidprentje(b)
	return nil
}

func (s *Store) Get(id string) (*models.Bidprentje, bool) {
	b, exists := s.data[id]
	return b, exists
}

func (s *Store) Update(b *models.Bidprentje) error {
	if old, exists := s.data[b.ID]; exists {
		s.removeFromIndex(old)
	}
	s.data[b.ID] = b
	s.indexBidprentje(b)
	return nil
}

func (s *Store) Delete(id string) error {
	if b, exists := s.data[id]; exists {
		s.removeFromIndex(b)
		delete(s.data, id)
	}
	return nil
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

func (s *Store) indexBidprentje(b *models.Bidprentje) {
	// Index text fields
	s.trie.Insert(strings.ToLower(b.Voornaam), b.ID)
	s.trie.Insert(strings.ToLower(b.Tussenvoegsel), b.ID)
	s.trie.Insert(strings.ToLower(b.Achternaam), b.ID)
	s.trie.Insert(strings.ToLower(b.Geboorteplaats), b.ID)
	s.trie.Insert(strings.ToLower(b.Overlijdensplaats), b.ID)

	// Index dates
	s.trie.Insert(b.Geboortedatum.Format("2006"), b.ID)
	s.trie.Insert(b.Overlijdensdatum.Format("2006"), b.ID)
}

func (s *Store) removeFromIndex(b *models.Bidprentje) {
	// Remove text fields
	s.trie.Remove(strings.ToLower(b.Voornaam), b.ID)
	s.trie.Remove(strings.ToLower(b.Tussenvoegsel), b.ID)
	s.trie.Remove(strings.ToLower(b.Achternaam), b.ID)
	s.trie.Remove(strings.ToLower(b.Geboorteplaats), b.ID)
	s.trie.Remove(strings.ToLower(b.Overlijdensplaats), b.ID)

	// Remove dates
	s.trie.Remove(b.Geboortedatum.Format("2006"), b.ID)
	s.trie.Remove(b.Overlijdensdatum.Format("2006"), b.ID)
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

	query := strings.ToLower(params.Query)
	matches := make(map[string]bool)

	// Search through all bidprentjes
	for _, b := range s.data {
		// Check each field
		if strings.Contains(strings.ToLower(b.Voornaam), query) ||
			strings.Contains(strings.ToLower(b.Tussenvoegsel), query) ||
			strings.Contains(strings.ToLower(b.Achternaam), query) ||
			strings.Contains(strings.ToLower(b.Geboorteplaats), query) ||
			strings.Contains(strings.ToLower(b.Overlijdensplaats), query) ||
			strings.Contains(b.Geboortedatum.Format("2006"), query) ||
			strings.Contains(b.Overlijdensdatum.Format("2006"), query) {
			matches[b.ID] = true
		}
	}

	// Convert matches to slice
	items := make([]models.Bidprentje, 0, len(matches))
	for id := range matches {
		if b, exists := s.data[id]; exists {
			items = append(items, *b)
		}
	}

	// Sort items by CreatedAt descending
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	// Calculate pagination
	totalCount := len(items)
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize

	if start >= totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	return &models.PaginatedResponse{
		Items:      items[start:end],
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}
}
