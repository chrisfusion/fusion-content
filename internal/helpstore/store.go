// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package helpstore

import (
	"regexp"
	"sort"
	"strings"
	"sync"

	"fusion-platform.io/fusion-content/internal/help"
)

// HelpQuery holds all filter and pagination parameters for the help store.
type HelpQuery struct {
	Service  string
	Type     string
	Tag      string
	Route    string
	Q        string
	Page     int
	PageSize int
}

// Store is a thread-safe in-memory help article store with a full-text inverted index.
type Store struct {
	mu       sync.RWMutex
	articles []help.Article     // sorted by service+type+slug
	index    map[string][]int   // token → sorted article indices
}

// New returns an empty Store.
func New() *Store {
	return &Store{index: make(map[string][]int)}
}

// Update atomically replaces the full article corpus and rebuilds the index.
func (s *Store) Update(articles []help.Article) {
	sorted := make([]help.Article, len(articles))
	copy(sorted, articles)
	sort.Slice(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		if a.Service != b.Service {
			return a.Service < b.Service
		}
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		return a.Slug < b.Slug
	})

	idx := buildIndex(sorted)

	s.mu.Lock()
	s.articles = sorted
	s.index = idx
	s.mu.Unlock()
}

// Query returns a filtered, paginated list of article summaries and the total filtered count.
func (s *Store) Query(q HelpQuery) ([]help.ArticleSummary, int) {
	s.mu.RLock()
	articles := s.articles
	index := s.index
	s.mu.RUnlock()

	// Full-text search: compute matching index set first.
	var matchSet map[int]struct{}
	if q.Q != "" {
		hits := searchIndex(index, q.Q)
		matchSet = make(map[int]struct{}, len(hits))
		for _, i := range hits {
			matchSet[i] = struct{}{}
		}
	}

	var filtered []help.ArticleSummary
	for i, a := range articles {
		if matchSet != nil {
			if _, ok := matchSet[i]; !ok {
				continue
			}
		}
		if q.Service != "" && a.Service != q.Service {
			continue
		}
		if q.Type != "" && a.Type != q.Type {
			continue
		}
		if q.Tag != "" && !containsString(a.Tags, q.Tag) {
			continue
		}
		if q.Route != "" && !containsString(a.Routes, q.Route) {
			continue
		}
		filtered = append(filtered, a.AsSummary())
	}

	total := len(filtered)
	if total == 0 {
		return []help.ArticleSummary{}, 0
	}

	start := (q.Page - 1) * q.PageSize
	if start >= total {
		return []help.ArticleSummary{}, total
	}
	end := start + q.PageSize
	if end > total {
		end = total
	}
	return filtered[start:end], total
}

// Get returns a single article by its identity triple.
func (s *Store) Get(service, docType, slug string) (help.Article, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.articles {
		if a.Service == service && a.Type == docType && a.Slug == slug {
			return a, true
		}
	}
	return help.Article{}, false
}

// tokenRe matches alphanumeric tokens.
var tokenRe = regexp.MustCompile(`[a-z0-9]+`)

// stopWords are excluded from the index.
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "is": true, "in": true,
	"of": true, "for": true, "to": true, "and": true, "or": true,
	"with": true, "it": true, "on": true, "at": true, "by": true,
}

func tokenize(text string) []string {
	tokens := tokenRe.FindAllString(strings.ToLower(text), -1)
	seen := make(map[string]struct{}, len(tokens))
	out := tokens[:0]
	for _, t := range tokens {
		if len(t) < 2 || stopWords[t] {
			continue
		}
		if _, dup := seen[t]; !dup {
			seen[t] = struct{}{}
			out = append(out, t)
		}
	}
	return out
}

func buildIndex(articles []help.Article) map[string][]int {
	idx := make(map[string][]int)
	for i, a := range articles {
		text := strings.Join(append(
			[]string{a.Title, a.Summary, a.Body},
			a.Tags...,
		), " ")
		for _, tok := range tokenize(text) {
			idx[tok] = append(idx[tok], i)
		}
	}
	return idx
}

func searchIndex(index map[string][]int, q string) []int {
	tokens := tokenize(q)
	if len(tokens) == 0 {
		return nil
	}
	result := index[tokens[0]]
	if result == nil {
		return nil
	}
	// Copy to avoid mutating the index.
	result = append([]int(nil), result...)
	for _, tok := range tokens[1:] {
		result = intersect(result, index[tok])
		if len(result) == 0 {
			return nil
		}
	}
	return result
}

// intersect merges two sorted int slices — O(m+n).
func intersect(a, b []int) []int {
	out := a[:0]
	j := 0
	for _, v := range a {
		for j < len(b) && b[j] < v {
			j++
		}
		if j < len(b) && b[j] == v {
			out = append(out, v)
		}
	}
	return out
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
