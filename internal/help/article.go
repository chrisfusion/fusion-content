// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package help

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DiátaxisType is one of the four Diátaxis documentation categories.
type DiátaxisType = string

const (
	TypeTutorial    DiátaxisType = "tutorial"
	TypeHowTo       DiátaxisType = "how-to"
	TypeReference   DiátaxisType = "reference"
	TypeExplanation DiátaxisType = "explanation"
)

var validTypes = map[DiátaxisType]bool{
	TypeTutorial:    true,
	TypeHowTo:       true,
	TypeReference:   true,
	TypeExplanation: true,
}

// frontmatter is the YAML header block of a help article.
type frontmatter struct {
	Title   string   `yaml:"title"`
	Tags    []string `yaml:"tags"`
	Routes  []string `yaml:"routes"`
	Summary string   `yaml:"summary"`
}

// Article is a fully parsed help article.
type Article struct {
	Service string
	Type    DiátaxisType
	Slug    string

	Title   string
	Tags    []string
	Routes  []string
	Summary string
	Body    string
}

// ArticleSummary is Article without the body, used in list responses.
type ArticleSummary struct {
	Service string       `json:"service"`
	Type    DiátaxisType `json:"type"`
	Slug    string       `json:"slug"`
	Title   string       `json:"title"`
	Tags    []string     `json:"tags"`
	Routes  []string     `json:"routes"`
	Summary string       `json:"summary"`
}

// ParseArticle parses a .md file and derives identity from relPath.
// relPath must match <service>/<type>/<slug>.md relative to the help root.
func ParseArticle(relPath string, data []byte) (Article, error) {
	parts := strings.SplitN(filepath.ToSlash(relPath), "/", 3)
	if len(parts) != 3 {
		return Article{}, fmt.Errorf("path does not match <service>/<type>/<slug>.md: %s", relPath)
	}
	service := parts[0]
	docType := parts[1]
	slug := strings.TrimSuffix(parts[2], ".md")

	if !validTypes[docType] {
		return Article{}, fmt.Errorf("unknown Diátaxis type %q in %s", docType, relPath)
	}

	fm, body, err := parseFrontmatter(data)
	if err != nil {
		return Article{}, fmt.Errorf("%s: %w", relPath, err)
	}

	return Article{
		Service: service,
		Type:    docType,
		Slug:    slug,
		Title:   fm.Title,
		Tags:    fm.Tags,
		Routes:  fm.Routes,
		Summary: fm.Summary,
		Body:    body,
	}, nil
}

// AsSummary returns an ArticleSummary (without body) for the article.
func (a Article) AsSummary() ArticleSummary {
	return ArticleSummary{
		Service: a.Service,
		Type:    a.Type,
		Slug:    a.Slug,
		Title:   a.Title,
		Tags:    a.Tags,
		Routes:  a.Routes,
		Summary: a.Summary,
	}
}

func parseFrontmatter(data []byte) (frontmatter, string, error) {
	s := string(data)
	const fence = "---"

	if !strings.HasPrefix(s, fence+"\n") {
		return frontmatter{}, "", fmt.Errorf("missing frontmatter opening fence")
	}
	rest := s[len(fence)+1:]

	end := strings.Index(rest, "\n"+fence)
	if end < 0 {
		return frontmatter{}, "", fmt.Errorf("missing frontmatter closing fence")
	}

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(rest[:end]), &fm); err != nil {
		return frontmatter{}, "", fmt.Errorf("parse frontmatter: %w", err)
	}

	body := strings.TrimPrefix(rest[end+len(fence)+1:], "\n")
	return fm, body, nil
}
