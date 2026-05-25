// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package helpstore

import (
	"testing"

	"fusion-platform.io/fusion-content/internal/help"
)

func makeArticles() []help.Article {
	return []help.Article{
		{
			Service: "forge",
			Type:    help.TypeTutorial,
			Slug:    "getting-started",
			Title:   "Getting Started with Forge",
			Tags:    []string{"quickstart", "forge"},
			Routes:  []string{"/forge"},
			Summary: "How to build your first venv.",
			Body:    "Create a build job and submit it.",
		},
		{
			Service: "index",
			Type:    help.TypeHowTo,
			Slug:    "upload-artifact",
			Title:   "Upload an Artifact",
			Tags:    []string{"index", "upload"},
			Routes:  []string{"/index/upload"},
			Summary: "Step-by-step upload guide.",
			Body:    "Use the REST API to push an artifact.",
		},
		{
			Service: "forge",
			Type:    help.TypeReference,
			Slug:    "api-reference",
			Title:   "Forge API Reference",
			Tags:    []string{"api", "forge"},
			Routes:  []string{"/forge/api"},
			Summary: "Complete API docs for forge.",
			Body:    "GET /api/v1/venvs returns a list of builds.",
		},
	}
}

func TestStore_QueryAll(t *testing.T) {
	s := New()
	s.Update(makeArticles())

	results, total := s.Query(HelpQuery{Page: 1, PageSize: 10})
	if total != 3 {
		t.Errorf("total: got %d, want 3", total)
	}
	if len(results) != 3 {
		t.Errorf("len(results): got %d, want 3", len(results))
	}
}

func TestStore_QueryByService(t *testing.T) {
	s := New()
	s.Update(makeArticles())

	results, total := s.Query(HelpQuery{Service: "forge", Page: 1, PageSize: 10})
	if total != 2 {
		t.Errorf("total: got %d, want 2", total)
	}
	for _, r := range results {
		if r.Service != "forge" {
			t.Errorf("unexpected service %q in results", r.Service)
		}
	}
}

func TestStore_QueryByType(t *testing.T) {
	s := New()
	s.Update(makeArticles())

	results, total := s.Query(HelpQuery{Type: help.TypeHowTo, Page: 1, PageSize: 10})
	if total != 1 {
		t.Errorf("total: got %d, want 1", total)
	}
	if len(results) > 0 && results[0].Type != help.TypeHowTo {
		t.Errorf("Type: got %q, want %q", results[0].Type, help.TypeHowTo)
	}
}

func TestStore_QueryByTag(t *testing.T) {
	s := New()
	s.Update(makeArticles())

	results, total := s.Query(HelpQuery{Tag: "forge", Page: 1, PageSize: 10})
	if total != 2 {
		t.Errorf("total: got %d, want 2", total)
	}
	_ = results
}

func TestStore_QueryByRoute(t *testing.T) {
	s := New()
	s.Update(makeArticles())

	results, total := s.Query(HelpQuery{Route: "/forge", Page: 1, PageSize: 10})
	if total != 1 {
		t.Errorf("total: got %d, want 1", total)
	}
	if len(results) > 0 && results[0].Slug != "getting-started" {
		t.Errorf("Slug: got %q, want getting-started", results[0].Slug)
	}
}

func TestStore_QueryFullText(t *testing.T) {
	s := New()
	s.Update(makeArticles())

	results, total := s.Query(HelpQuery{Q: "venv build", Page: 1, PageSize: 10})
	if total < 1 {
		t.Errorf("expected at least 1 result for full-text query, got %d", total)
	}
	_ = results
}

func TestStore_QueryFullTextNoMatch(t *testing.T) {
	s := New()
	s.Update(makeArticles())

	_, total := s.Query(HelpQuery{Q: "xyzzy", Page: 1, PageSize: 10})
	if total != 0 {
		t.Errorf("expected 0 results for unmatched query, got %d", total)
	}
}

func TestStore_QueryPagination(t *testing.T) {
	s := New()
	s.Update(makeArticles())

	page1, total := s.Query(HelpQuery{Page: 1, PageSize: 2})
	if total != 3 {
		t.Errorf("total: got %d, want 3", total)
	}
	if len(page1) != 2 {
		t.Errorf("page 1 len: got %d, want 2", len(page1))
	}

	page2, _ := s.Query(HelpQuery{Page: 2, PageSize: 2})
	if len(page2) != 1 {
		t.Errorf("page 2 len: got %d, want 1", len(page2))
	}

	// Pages should not overlap.
	if page1[0].Slug == page2[0].Slug {
		t.Errorf("pages overlap: both have slug %q", page1[0].Slug)
	}
}

func TestStore_QueryPageBeyondEnd(t *testing.T) {
	s := New()
	s.Update(makeArticles())

	results, total := s.Query(HelpQuery{Page: 99, PageSize: 10})
	if total != 3 {
		t.Errorf("total: got %d, want 3", total)
	}
	if len(results) != 0 {
		t.Errorf("expected empty page, got %d results", len(results))
	}
}

func TestStore_Get(t *testing.T) {
	s := New()
	s.Update(makeArticles())

	a, ok := s.Get("forge", help.TypeTutorial, "getting-started")
	if !ok {
		t.Fatal("expected article to be found")
	}
	if a.Title != "Getting Started with Forge" {
		t.Errorf("Title: got %q", a.Title)
	}
	if a.Body == "" {
		t.Error("Body should be populated in Get response")
	}
}

func TestStore_GetNotFound(t *testing.T) {
	s := New()
	s.Update(makeArticles())

	_, ok := s.Get("forge", help.TypeTutorial, "nonexistent")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestStore_UpdateReplacesCorpus(t *testing.T) {
	s := New()
	s.Update(makeArticles())

	// Replace with a single article.
	s.Update([]help.Article{
		{Service: "weave", Type: help.TypeExplanation, Slug: "concepts", Title: "Weave Concepts", Body: "DAG scheduling explained."},
	})

	_, total := s.Query(HelpQuery{Page: 1, PageSize: 10})
	if total != 1 {
		t.Errorf("total after update: got %d, want 1", total)
	}
	_, ok := s.Get("forge", help.TypeTutorial, "getting-started")
	if ok {
		t.Error("old article should be gone after Update")
	}
}

func TestStore_EmptyStore(t *testing.T) {
	s := New()
	results, total := s.Query(HelpQuery{Page: 1, PageSize: 10})
	if total != 0 || len(results) != 0 {
		t.Errorf("empty store: got total=%d results=%d", total, len(results))
	}
}
