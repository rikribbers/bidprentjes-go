package store

import (
	"strings"

	"bidprentjes-api/models"
)

type Store struct {
	data map[string]*models.Bidprentje
	trie *Trie
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]*models.Bidprentje),
		trie: NewTrie(),
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
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}

	total := len(s.data)
	start := (page - 1) * pageSize
	if start >= total {
		return &models.PaginatedResponse{
			Items:      []models.Bidprentje{},
			TotalCount: total,
			Page:       page,
			PageSize:   pageSize,
		}
	}

	items := make([]models.Bidprentje, 0, pageSize)
	count := 0
	for _, b := range s.data {
		if count >= start && len(items) < pageSize {
			items = append(items, *b)
		}
		count++
	}

	return &models.PaginatedResponse{
		Items:      items,
		TotalCount: total,
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
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 25
	}

	// Search with fuzzy matching
	results := s.trie.Search(params.Query, 2) // Allow up to 2 character differences

	// Convert results to slice for pagination
	items := make([]models.Bidprentje, 0, len(results))
	for id := range results {
		if b, exists := s.data[id]; exists {
			items = append(items, *b)
		}
	}

	// Apply pagination
	total := len(items)
	start := (params.Page - 1) * params.PageSize
	if start >= total {
		return &models.PaginatedResponse{
			Items:      []models.Bidprentje{},
			TotalCount: total,
			Page:       params.Page,
			PageSize:   params.PageSize,
		}
	}

	end := start + params.PageSize
	if end > total {
		end = total
	}

	return &models.PaginatedResponse{
		Items:      items[start:end],
		TotalCount: total,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}
}
