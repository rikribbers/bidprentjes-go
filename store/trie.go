package store

import (
	"strings"
)

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	ids      map[string]bool
}

type Trie struct {
	root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
			ids:      make(map[string]bool),
		},
	}
}

func (t *Trie) Insert(word string, id string) {
	if word == "" {
		return
	}

	node := t.root
	for _, ch := range word {
		if _, exists := node.children[ch]; !exists {
			node.children[ch] = &TrieNode{
				children: make(map[rune]*TrieNode),
				ids:      make(map[string]bool),
			}
		}
		node = node.children[ch]
		node.ids[id] = true
	}
	node.isEnd = true
}

func (t *Trie) Remove(word string, id string) {
	if word == "" {
		return
	}

	var removeFromNode func(node *TrieNode, word []rune, index int)
	removeFromNode = func(node *TrieNode, word []rune, index int) {
		if index == len(word) {
			delete(node.ids, id)
			return
		}

		ch := word[index]
		if child, exists := node.children[ch]; exists {
			removeFromNode(child, word, index+1)
			delete(child.ids, id)

			// Clean up empty nodes
			if len(child.ids) == 0 && len(child.children) == 0 {
				delete(node.children, ch)
			}
		}
	}

	removeFromNode(t.root, []rune(word), 0)
}

func (t *Trie) Search(query string, maxDistance int) map[string]bool {
	results := make(map[string]bool)
	if query == "" {
		return results
	}

	query = strings.ToLower(query)
	var search func(node *TrieNode, prefix string, distance int)
	search = func(node *TrieNode, prefix string, distance int) {
		if distance > maxDistance {
			return
		}

		// Add IDs if we're within distance
		if len(prefix) >= len(query) {
			if distance <= maxDistance {
				for id := range node.ids {
					results[id] = true
				}
			}
		}

		// Continue searching
		if len(prefix) < len(query) {
			ch := rune(query[len(prefix)])
			// Match
			if child, exists := node.children[ch]; exists {
				search(child, prefix+string(ch), distance)
			}
			// Substitution
			for r, child := range node.children {
				if r != ch {
					search(child, prefix+string(r), distance+1)
				}
			}
			// Deletion
			search(node, prefix+string(ch), distance+1)
		} else {
			// Insertion
			for r, child := range node.children {
				search(child, prefix+string(r), distance+1)
			}
		}
	}

	search(t.root, "", 0)
	return results
}
