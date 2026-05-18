// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package gitpoller

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"fusion-platform.io/fusion-content/internal/config"
	"fusion-platform.io/fusion-content/internal/gitutil"
	"fusion-platform.io/fusion-content/internal/parser"
)

// OnUpdateFunc is called after a successful poll with the project name and parsed entries.
type OnUpdateFunc func(project string, entries []parser.Entry)

// Poller polls a set of git repositories and calls onUpdate after each successful read.
type Poller struct {
	repos        []config.RepoConfig
	interval     time.Duration
	cloneBaseDir string
	onUpdate     OnUpdateFunc
}

// New returns a Poller. cloneBaseDir must be writable (use an emptyDir volume in K8s).
func New(repos []config.RepoConfig, interval time.Duration, cloneBaseDir string, onUpdate OnUpdateFunc) *Poller {
	return &Poller{
		repos:        repos,
		interval:     interval,
		cloneBaseDir: cloneBaseDir,
		onUpdate:     onUpdate,
	}
}

// Start launches one goroutine per repo. It returns immediately; all goroutines
// respect ctx cancellation.
func (p *Poller) Start(ctx context.Context) {
	if err := os.MkdirAll(p.cloneBaseDir, 0755); err != nil {
		log.Printf("content: cannot create clone base dir %s: %v", p.cloneBaseDir, err)
	}
	for _, repo := range p.repos {
		go p.runRepo(ctx, repo)
	}
}

func (p *Poller) runRepo(ctx context.Context, repo config.RepoConfig) {
	p.pollRepo(repo)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.pollRepo(repo)
		}
	}
}

func (p *Poller) pollRepo(repo config.RepoConfig) {
	repoDir := filepath.Join(p.cloneBaseDir, gitutil.SanitizeName(repo.Name))

	if _, err := gitutil.EnsureRepo(repo.URL, repoDir, repo.Token); err != nil {
		log.Printf("content: sync %s: %v", repo.Name, err)
		return
	}

	data, err := os.ReadFile(filepath.Join(repoDir, repo.ChangelogPath))
	if err != nil {
		log.Printf("content: read %s/%s: %v", repo.Name, repo.ChangelogPath, err)
		return
	}

	entries := parser.Parse(data)
	p.onUpdate(repo.Name, entries)
	log.Printf("content: updated %s (%d entries)", repo.Name, len(entries))
}
