// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package videostore

import (
	"testing"

	"fusion-platform.io/fusion-content/internal/video"
)

func makeVideos() []video.Video {
	return []video.Video{
		{
			Service:      "forge",
			Slug:         "forge-overview",
			Title:        "Fusion Forge Overview",
			Summary:      "A walkthrough of venv builds",
			ThumbnailURL: "https://img.youtube.com/vi/AAA/hqdefault.jpg",
			VideoURL:     "https://youtube.com/watch?v=AAA",
			Tags:         []string{"forge", "overview"},
		},
		{
			Service:      "index",
			Slug:         "index-overview",
			Title:        "Fusion Index Overview",
			Summary:      "Artifact indexing explained",
			ThumbnailURL: "https://img.youtube.com/vi/BBB/hqdefault.jpg",
			VideoURL:     "https://youtube.com/watch?v=BBB",
			Tags:         []string{"index", "overview"},
		},
		{
			Service:      "forge",
			Slug:         "forge-git-builds",
			Title:        "Git-triggered Builds",
			Summary:      "How GitOps polling triggers builds",
			ThumbnailURL: "https://img.youtube.com/vi/CCC/hqdefault.jpg",
			VideoURL:     "https://youtube.com/watch?v=CCC",
			Tags:         []string{"forge", "git"},
		},
	}
}

func TestStore_QueryAll(t *testing.T) {
	s := New()
	s.Update(makeVideos())

	results, total := s.Query(VideoQuery{Page: 1, PageSize: 10})
	if total != 3 {
		t.Errorf("total: got %d, want 3", total)
	}
	if len(results) != 3 {
		t.Errorf("len(results): got %d, want 3", len(results))
	}
}

func TestStore_QueryByService(t *testing.T) {
	s := New()
	s.Update(makeVideos())

	results, total := s.Query(VideoQuery{Service: "forge", Page: 1, PageSize: 10})
	if total != 2 {
		t.Errorf("total: got %d, want 2", total)
	}
	for _, r := range results {
		if r.Service != "forge" {
			t.Errorf("unexpected service %q in results", r.Service)
		}
	}
}

func TestStore_QueryPagination(t *testing.T) {
	s := New()
	s.Update(makeVideos())

	page1, total := s.Query(VideoQuery{Page: 1, PageSize: 2})
	if total != 3 {
		t.Errorf("total: got %d, want 3", total)
	}
	if len(page1) != 2 {
		t.Errorf("page 1 len: got %d, want 2", len(page1))
	}

	page2, _ := s.Query(VideoQuery{Page: 2, PageSize: 2})
	if len(page2) != 1 {
		t.Errorf("page 2 len: got %d, want 1", len(page2))
	}

	if page1[0].Slug == page2[0].Slug {
		t.Errorf("pages overlap: both have slug %q", page1[0].Slug)
	}
}

func TestStore_QueryPageBeyondEnd(t *testing.T) {
	s := New()
	s.Update(makeVideos())

	results, total := s.Query(VideoQuery{Page: 99, PageSize: 10})
	if total != 3 {
		t.Errorf("total: got %d, want 3", total)
	}
	if len(results) != 0 {
		t.Errorf("expected empty page, got %d results", len(results))
	}
}

func TestStore_Get(t *testing.T) {
	s := New()
	s.Update(makeVideos())

	v, ok := s.Get("forge", "forge-overview")
	if !ok {
		t.Fatal("expected video to be found")
	}
	if v.Title != "Fusion Forge Overview" {
		t.Errorf("Title: got %q", v.Title)
	}
}

func TestStore_GetNotFound(t *testing.T) {
	s := New()
	s.Update(makeVideos())

	_, ok := s.Get("forge", "nonexistent")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestStore_UpdateReplacesCorpus(t *testing.T) {
	s := New()
	s.Update(makeVideos())

	s.Update([]video.Video{
		{Service: "weave", Slug: "weave-chains", Title: "Weave Chains"},
	})

	_, total := s.Query(VideoQuery{Page: 1, PageSize: 10})
	if total != 1 {
		t.Errorf("total after update: got %d, want 1", total)
	}
	_, ok := s.Get("forge", "forge-overview")
	if ok {
		t.Error("old video should be gone after Update")
	}
}

func TestStore_EmptyStore(t *testing.T) {
	s := New()
	results, total := s.Query(VideoQuery{Page: 1, PageSize: 10})
	if total != 0 || len(results) != 0 {
		t.Errorf("empty store: got total=%d results=%d", total, len(results))
	}
}
