// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package store

import (
	"sort"
	"sync"

	"fusion-platform.io/fusion-content/internal/parser"
)

// Changes holds the categorised items for a single project version.
type Changes struct {
	Added   []string `json:"added,omitempty"`
	Changed []string `json:"changed,omitempty"`
	Fixed   []string `json:"fixed,omitempty"`
	Removed []string `json:"removed,omitempty"`
}

// ProjectEntry is one project's contribution to a DateGroup.
type ProjectEntry struct {
	Project string  `json:"project"`
	Version string  `json:"version"`
	Changes Changes `json:"changes"`
}

// DateGroup groups all project entries that share the same date.
type DateGroup struct {
	Date     string         `json:"date"`
	Projects []ProjectEntry `json:"projects"`
}

// Store is a thread-safe in-memory view of merged changelog entries.
type Store struct {
	mu        sync.RWMutex
	byProject map[string][]parser.Entry
	merged    []DateGroup
}

// New returns an empty Store.
func New() *Store {
	return &Store{byProject: make(map[string][]parser.Entry)}
}

// Update replaces the entries for the given project and rebuilds the merged view.
func (s *Store) Update(project string, entries []parser.Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byProject[project] = entries
	s.merged = s.rebuild()
}

// Query returns a paginated, optionally-filtered slice of DateGroups.
// Pagination is over date groups (not individual project entries).
// Returns the filtered slice and the total number of matching date groups.
func (s *Store) Query(project, date string, page, pageSize int) ([]DateGroup, int) {
	s.mu.RLock()
	merged := s.merged
	s.mu.RUnlock()

	var filtered []DateGroup
	for _, g := range merged {
		if date != "" && g.Date != date {
			continue
		}
		var projects []ProjectEntry
		for _, p := range g.Projects {
			if project != "" && p.Project != project {
				continue
			}
			projects = append(projects, p)
		}
		if len(projects) > 0 {
			filtered = append(filtered, DateGroup{Date: g.Date, Projects: projects})
		}
	}

	total := len(filtered)
	start := (page - 1) * pageSize
	if start >= total {
		return []DateGroup{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return filtered[start:end], total
}

// rebuild merges byProject into a sorted []DateGroup.
// Must be called with the write lock held.
//
// A project may have multiple versions on the same day (e.g. patch releases);
// each version produces its own ProjectEntry within the date group, ordered
// alphabetically by project name then by original CHANGELOG order (newest version first).
func (s *Store) rebuild() []DateGroup {
	dateMap := make(map[string][]ProjectEntry)

	for project, entries := range s.byProject {
		for _, e := range entries {
			dateMap[e.Date] = append(dateMap[e.Date], ProjectEntry{
				Project: project,
				Version: e.Version,
				Changes: Changes{
					Added:   e.Added,
					Changed: e.Changed,
					Fixed:   e.Fixed,
					Removed: e.Removed,
				},
			})
		}
	}

	groups := make([]DateGroup, 0, len(dateMap))
	for date, projects := range dateMap {
		// Stable sort: sort by project name; entries for the same project keep
		// their original order (newest version first, as they appear in CHANGELOG).
		sort.SliceStable(projects, func(i, j int) bool {
			return projects[i].Project < projects[j].Project
		})
		groups = append(groups, DateGroup{Date: date, Projects: projects})
	}

	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Date == "unreleased" {
			return true
		}
		if groups[j].Date == "unreleased" {
			return false
		}
		return groups[i].Date > groups[j].Date
	})

	return groups
}
