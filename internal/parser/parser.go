// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package parser

import (
	"bufio"
	"bytes"
	"log"
	"regexp"
	"strings"
)

var (
	// Matches: ## [1.2.3] - date  /  ## [1.2.3](url) - date  /  ## [1.2.3][ref] - date
	// Handles plain, parenthesised-link, and bracket-ref variants of the Keep-a-Changelog format.
	versionRe    = regexp.MustCompile(`^## \[([^\]]+)\](?:\([^)]*\)|\[[^\]]*\])?\s*[-–—]+\s*(\d{4}-\d{2}-\d{2})`)
	unreleasedRe = regexp.MustCompile(`^## \[Unreleased\]`)
	sectionRe    = regexp.MustCompile(`^### (.+)`)
	itemRe       = regexp.MustCompile(`^[*-]\s+(.+)`)
)

// Entry represents one versioned block from a CHANGELOG.md.
type Entry struct {
	Date    string // "unreleased" or "YYYY-MM-DD"
	Version string // "Unreleased" or semver string
	Added   []string
	Changed []string
	Fixed   []string
	Removed []string
}

func (e *Entry) hasChanges() bool {
	return len(e.Added)+len(e.Changed)+len(e.Fixed)+len(e.Removed) > 0
}

// Parse parses the content of a CHANGELOG.md file and returns all entries
// that contain at least one change item.
func Parse(data []byte) []Entry {
	var entries []Entry
	var current *Entry
	var section string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1<<20), 1<<20) // 1 MB max line — handles embedded code blocks

	for scanner.Scan() {
		line := scanner.Text()

		if m := versionRe.FindStringSubmatch(line); m != nil {
			if current != nil && current.hasChanges() {
				entries = append(entries, *current)
			}
			current = &Entry{Version: m[1], Date: m[2]}
			section = ""
			continue
		}

		if unreleasedRe.MatchString(line) {
			if current != nil && current.hasChanges() {
				entries = append(entries, *current)
			}
			current = &Entry{Version: "Unreleased", Date: "unreleased"}
			section = ""
			continue
		}

		if current == nil {
			continue
		}

		if m := sectionRe.FindStringSubmatch(line); m != nil {
			section = strings.ToLower(strings.TrimSpace(m[1]))
			continue
		}

		if m := itemRe.FindStringSubmatch(line); m != nil {
			item := strings.TrimSpace(m[1])
			switch section {
			case "added":
				current.Added = append(current.Added, item)
			case "changed":
				current.Changed = append(current.Changed, item)
			case "fixed":
				current.Fixed = append(current.Fixed, item)
			case "removed":
				current.Removed = append(current.Removed, item)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("content: scanner error while parsing changelog: %v", err)
	}

	if current != nil && current.hasChanges() {
		entries = append(entries, *current)
	}

	return entries
}
