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

	type searchResult struct {
		bidprentje *models.Bidprentje
		score      int
	}

	query := strings.ToLower(params.Query)
	queryWords := strings.Fields(query)
	results := make([]searchResult, 0)

	// Search through all bidprentjes
	for _, b := range s.data {
		score := 0
		searchFields := []string{
			strings.ToLower(b.ID),
			strings.ToLower(b.Voornaam),
			strings.ToLower(b.Tussenvoegsel),
			strings.ToLower(b.Achternaam),
			strings.ToLower(b.Geboorteplaats),
			strings.ToLower(b.Overlijdensplaats),
			b.Geboortedatum.Format("02-01-2006"),
			b.Overlijdensdatum.Format("02-01-2006"),
			b.Geboortedatum.Format("01-2006"),
			b.Overlijdensdatum.Format("01-2006"),
			b.Geboortedatum.Format("2006"),
			b.Overlijdensdatum.Format("2006"),
		}

		// First check for exact ID match
		for _, word := range queryWords {
			if strings.ToLower(b.ID) == word {
				score += 100
				continue
			}
		}

		// Check each query word against each field
		for _, word := range queryWords {
			wordScore := 0
			for _, field := range searchFields {
				// Exact match (highest score)
				if field == word {
					wordScore = 100
					break
				}
				// Contains match (high score)
				if strings.Contains(field, word) {
					wordScore = max(wordScore, 75)
					continue
				}
				// Fuzzy match for longer words
				if len(word) > 3 {
					fieldWords := strings.Fields(field)
					for _, fieldWord := range fieldWords {
						distance := levenshteinDistance(word, fieldWord)
						// Score based on similarity
						if distance == 1 {
							wordScore = max(wordScore, 50)
						} else if distance == 2 {
							wordScore = max(wordScore, 25)
						}
					}
				}
			}
			score += wordScore
		}

		// Only include results that scored more than 25 points
		if score > 25 {
			results = append(results, searchResult{
				bidprentje: b,
				score:      score,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// Convert to items slice
	items := make([]models.Bidprentje, 0, len(results))
	for _, result := range results {
		items = append(items, *result.bidprentje)
	}

	// Calculate pagination
	totalCount := len(items)
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize

	// Ensure valid pagination bounds
	if start >= totalCount {
		start = 0
		params.Page = 1
	}
	if end > totalCount {
		end = totalCount
	}

	// Return paginated results
	return &models.PaginatedResponse{
		Items:      items[start:end],
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}
}

// levenshteinDistance calculates the minimum number of single-character edits required to change one string into another
func levenshteinDistance(s1, s2 string) int {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			if s1[i-1] == s2[j-1] {
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				matrix[i][j] = min(
					matrix[i-1][j]+1,   // deletion
					matrix[i][j-1]+1,   // insertion
					matrix[i-1][j-1]+1, // substitution
				)
			}
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(numbers ...int) int {
	result := numbers[0]
	for _, num := range numbers[1:] {
		if num < result {
			result = num
		}
	}
	return result
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
