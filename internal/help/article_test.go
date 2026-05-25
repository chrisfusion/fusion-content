// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package help

import (
	"testing"
)

func TestParseArticle_valid(t *testing.T) {
	// No blank line between closing fence and body — parser strips exactly one \n.
	src := []byte("---\ntitle: Getting Started\ntags: [quickstart, setup]\nroutes: [/forge]\nsummary: How to set up forge.\n---\nBody text here.\n")

	a, err := ParseArticle("forge/tutorial/getting-started.md", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Service != "forge" {
		t.Errorf("Service: got %q, want %q", a.Service, "forge")
	}
	if a.Type != TypeTutorial {
		t.Errorf("Type: got %q, want %q", a.Type, TypeTutorial)
	}
	if a.Slug != "getting-started" {
		t.Errorf("Slug: got %q, want %q", a.Slug, "getting-started")
	}
	if a.Title != "Getting Started" {
		t.Errorf("Title: got %q, want %q", a.Title, "Getting Started")
	}
	if len(a.Tags) != 2 || a.Tags[0] != "quickstart" {
		t.Errorf("Tags: got %v", a.Tags)
	}
	if len(a.Routes) != 1 || a.Routes[0] != "/forge" {
		t.Errorf("Routes: got %v", a.Routes)
	}
	if a.Summary != "How to set up forge." {
		t.Errorf("Summary: got %q", a.Summary)
	}
	if a.Body != "Body text here.\n" {
		t.Errorf("Body: got %q", a.Body)
	}
}

func TestParseArticle_invalidType(t *testing.T) {
	src := []byte("---\ntitle: X\n---\n\nbody\n")
	_, err := ParseArticle("forge/unknown/slug.md", src)
	if err == nil {
		t.Fatal("expected error for unknown Diátaxis type")
	}
}

func TestParseArticle_missingFrontmatter(t *testing.T) {
	_, err := ParseArticle("forge/tutorial/slug.md", []byte("no frontmatter"))
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
}

func TestParseArticle_badPath(t *testing.T) {
	src := []byte("---\ntitle: X\n---\n\nbody\n")
	_, err := ParseArticle("too-short.md", src)
	if err == nil {
		t.Fatal("expected error for path that does not match <service>/<type>/<slug>.md")
	}
}

func TestAsSummary(t *testing.T) {
	a := Article{
		Service: "index",
		Type:    TypeHowTo,
		Slug:    "upload",
		Title:   "Upload an artifact",
		Tags:    []string{"index"},
		Routes:  []string{"/index/upload"},
		Summary: "Step-by-step upload guide.",
		Body:    "long body",
	}
	s := a.AsSummary()
	if s.Service != a.Service || s.Slug != a.Slug || s.Title != a.Title {
		t.Errorf("AsSummary fields mismatch: %+v", s)
	}
}
