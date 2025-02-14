package store

import (
	"bidprentjes-api/models"
	"strings"
)

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	ids      []string
}

type Store struct {
	trie        *TrieNode
	bidprentjes map[string]*models.Bidprentje
}

func NewStore() *Store {
	return &Store{
		trie:        &TrieNode{children: make(map[rune]*TrieNode)},
		bidprentjes: make(map[string]*models.Bidprentje),
	}
}

func (s *Store) Insert(b *models.Bidprentje) {
	s.bidprentjes[b.ID] = b

	// Insert searchable fields into trie
	searchTerms := []string{
		strings.ToLower(b.Voornaam),
		strings.ToLower(b.Tussenvoegsel),
		strings.ToLower(b.Achternaam),
		strings.ToLower(b.Geboorteplaats),
		strings.ToLower(b.Overlijdensplaats),
	}

	for _, term := range searchTerms {
		s.insertWord(term, b.ID)
	}
}

func (s *Store) insertWord(word string, id string) {
	node := s.trie
	for _, ch := range word {
		if _, exists := node.children[ch]; !exists {
			node.children[ch] = &TrieNode{children: make(map[rune]*TrieNode)}
		}
		node = node.children[ch]
		node.ids = append(node.ids, id)
	}
	node.isEnd = true
}

func (s *Store) Get(id string) (*models.Bidprentje, bool) {
	b, exists := s.bidprentjes[id]
	return b, exists
}

func (s *Store) Update(b *models.Bidprentje) {
	// Remove old entries from trie
	if old, exists := s.bidprentjes[b.ID]; exists {
		s.remove(old)
	}

	// Insert updated bidprentje
	s.Insert(b)
}

func (s *Store) remove(b *models.Bidprentje) {
	// Remove from trie first
	searchTerms := []string{
		strings.ToLower(b.Voornaam),
		strings.ToLower(b.Tussenvoegsel),
		strings.ToLower(b.Achternaam),
		strings.ToLower(b.Geboorteplaats),
		strings.ToLower(b.Overlijdensplaats),
	}

	for _, term := range searchTerms {
		s.removeWord(term, b.ID)
	}

	// Then remove from map
	delete(s.bidprentjes, b.ID)
}

func (s *Store) removeWord(word string, id string) {
	node := s.trie
	var nodes []*TrieNode
	for _, ch := range word {
		if next, exists := node.children[ch]; exists {
			nodes = append(nodes, next)
			node = next
		} else {
			return
		}
	}

	// Remove ID from all nodes in the path
	for _, n := range nodes {
		newIds := make([]string, 0)
		for _, existingID := range n.ids {
			if existingID != id {
				newIds = append(newIds, existingID)
			}
		}
		n.ids = newIds
	}
}

func (s *Store) Delete(id string) {
	if b, exists := s.bidprentjes[id]; exists {
		s.remove(b)
	}
}

func (s *Store) List() []*models.Bidprentje {
	result := make([]*models.Bidprentje, 0, len(s.bidprentjes))
	for _, b := range s.bidprentjes {
		result = append(result, b)
	}
	return result
}

func (s *Store) Search(query string) []*models.Bidprentje {
	query = strings.ToLower(query)
	seen := make(map[string]bool)
	results := make([]*models.Bidprentje, 0)

	// Split query into words
	var year string
	var searchTerms []string
	words := strings.Fields(query)

	// Process each word
	for _, word := range words {
		if len(word) == 4 && word[0] >= '1' && word[0] <= '2' {
			// Looks like a year
			year = word
		} else {
			searchTerms = append(searchTerms, word)
		}
	}

	// Helper function to add result if not already seen
	addResult := func(id string) {
		if !seen[id] && s.bidprentjes[id] != nil {
			b := s.bidprentjes[id]

			// If year is specified, check if it matches birth or death year
			if year != "" {
				birthYear := b.Geboortedatum.Format("2006")
				deathYear := b.Overlijdensdatum.Format("2006")
				if birthYear != year && deathYear != year {
					return
				}
			}

			// Check if all search terms match
			allTermsMatch := true
			fields := []string{
				strings.ToLower(b.Voornaam),
				strings.ToLower(b.Tussenvoegsel),
				strings.ToLower(b.Achternaam),
				strings.ToLower(b.Geboorteplaats),
				strings.ToLower(b.Overlijdensplaats),
			}

			// For each search term, check if it matches any field
			for _, searchTerm := range searchTerms {
				termMatches := false
				for _, field := range fields {
					if levenshteinDistance(field, searchTerm) <= 2 {
						termMatches = true
						break
					}
				}
				if !termMatches {
					allTermsMatch = false
					break
				}
			}

			if allTermsMatch {
				seen[id] = true
				results = append(results, b)
			}
		}
	}

	// For each bidprentje, check if all terms match
	for id := range s.bidprentjes {
		addResult(id)
	}

	return results
}

// Helper function to calculate Levenshtein distance
func levenshteinDistance(s1, s2 string) int {
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

func (s *Store) ListPaginated(page, pageSize int) (*models.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}

	allItems := s.List()
	totalCount := len(allItems)

	start := (page - 1) * pageSize
	if start >= totalCount {
		return &models.PaginatedResponse{
			Items:      []*models.Bidprentje{},
			TotalCount: totalCount,
			Page:       page,
			PageSize:   pageSize,
		}, nil
	}

	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}

	return &models.PaginatedResponse{
		Items:      allItems[start:end],
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s *Store) SearchPaginated(query string, page, pageSize int) (*models.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}

	allResults := s.Search(query)
	totalCount := len(allResults)

	start := (page - 1) * pageSize
	if start >= totalCount {
		return &models.PaginatedResponse{
			Items:      []*models.Bidprentje{},
			TotalCount: totalCount,
			Page:       page,
			PageSize:   pageSize,
		}, nil
	}

	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}

	return &models.PaginatedResponse{
		Items:      allResults[start:end],
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}
