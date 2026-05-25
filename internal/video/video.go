// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package video

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// frontmatter is the YAML header block of a video article.
type frontmatter struct {
	Title        string   `yaml:"title"`
	Service      string   `yaml:"service"`
	Summary      string   `yaml:"summary"`
	ThumbnailURL string   `yaml:"thumbnailUrl"`
	VideoURL     string   `yaml:"videoUrl"`
	Tags         []string `yaml:"tags"`
}

// Video is a fully parsed video metadata entry.
type Video struct {
	Service      string
	Slug         string
	Title        string
	Summary      string
	ThumbnailURL string
	VideoURL     string
	Tags         []string
}

// ParseVideo parses a .md file and derives identity from relPath.
// relPath must match <service>/<slug>.md relative to the videos root.
func ParseVideo(relPath string, data []byte) (Video, error) {
	// Limit to 3 so a path with a subdirectory (e.g. forge/subdir/slug.md)
	// produces 3 parts and fails validation rather than silently storing an unreachable slug.
	parts := strings.SplitN(filepath.ToSlash(relPath), "/", 3)
	if len(parts) != 2 {
		return Video{}, fmt.Errorf("path does not match <service>/<slug>.md: %s", relPath)
	}
	service := parts[0]
	slug := strings.TrimSuffix(parts[1], ".md")

	fm, err := parseFrontmatter(data)
	if err != nil {
		return Video{}, fmt.Errorf("%s: %w", relPath, err)
	}

	return Video{
		Service:      service,
		Slug:         slug,
		Title:        fm.Title,
		Summary:      fm.Summary,
		ThumbnailURL: fm.ThumbnailURL,
		VideoURL:     fm.VideoURL,
		Tags:         fm.Tags,
	}, nil
}

func parseFrontmatter(data []byte) (frontmatter, error) {
	s := string(data)
	const fence = "---"

	if !strings.HasPrefix(s, fence+"\n") {
		return frontmatter{}, fmt.Errorf("missing frontmatter opening fence")
	}
	rest := s[len(fence)+1:]

	end := strings.Index(rest, "\n"+fence)
	if end < 0 {
		return frontmatter{}, fmt.Errorf("missing frontmatter closing fence")
	}

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(rest[:end]), &fm); err != nil {
		return frontmatter{}, fmt.Errorf("parse frontmatter: %w", err)
	}

	return fm, nil
}
