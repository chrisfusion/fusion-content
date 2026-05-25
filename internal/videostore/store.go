// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package videostore

import (
	"sort"
	"sync"

	"fusion-platform.io/fusion-content/internal/video"
)

// VideoQuery holds filter and pagination parameters for the video store.
type VideoQuery struct {
	Service  string
	Page     int
	PageSize int
}

// Store is a thread-safe in-memory video store.
type Store struct {
	mu     sync.RWMutex
	videos []video.Video // sorted by service+slug
}

// New returns an empty Store.
func New() *Store {
	return &Store{}
}

// Update atomically replaces the full video corpus.
func (s *Store) Update(videos []video.Video) {
	sorted := make([]video.Video, len(videos))
	copy(sorted, videos)
	sort.Slice(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		if a.Service != b.Service {
			return a.Service < b.Service
		}
		return a.Slug < b.Slug
	})

	s.mu.Lock()
	s.videos = sorted
	s.mu.Unlock()
}

// Query returns a filtered, paginated list of videos and the total filtered count.
func (s *Store) Query(q VideoQuery) ([]video.Video, int) {
	s.mu.RLock()
	videos := s.videos
	s.mu.RUnlock()

	var filtered []video.Video
	for _, v := range videos {
		if q.Service != "" && v.Service != q.Service {
			continue
		}
		filtered = append(filtered, v)
	}

	total := len(filtered)
	if total == 0 {
		return []video.Video{}, 0
	}

	start := (q.Page - 1) * q.PageSize
	if start >= total {
		return []video.Video{}, total
	}
	end := start + q.PageSize
	if end > total {
		end = total
	}
	return filtered[start:end], total
}

// Get returns a single video by service and slug.
func (s *Store) Get(service, slug string) (video.Video, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.videos {
		if v.Service == service && v.Slug == slug {
			return v, true
		}
	}
	return video.Video{}, false
}
