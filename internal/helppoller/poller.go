// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package helppoller

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fusion-platform.io/fusion-content/internal/config"
	"fusion-platform.io/fusion-content/internal/gitutil"
	"fusion-platform.io/fusion-content/internal/help"
)

// OnUpdateFunc is called with the full article list after each successful poll.
type OnUpdateFunc func(articles []help.Article)

// Poller polls a single help content git repository and calls onUpdate on every cycle.
type Poller struct {
	cfg          config.HelpConfig
	interval     time.Duration
	cloneBaseDir string
	onUpdate     OnUpdateFunc
}

// New returns a Poller for the given help config.
func New(cfg config.HelpConfig, interval time.Duration, cloneBaseDir string, onUpdate OnUpdateFunc) *Poller {
	return &Poller{
		cfg:          cfg,
		interval:     interval,
		cloneBaseDir: cloneBaseDir,
		onUpdate:     onUpdate,
	}
}

// Start launches the polling goroutine. It returns immediately and respects ctx cancellation.
func (p *Poller) Start(ctx context.Context) {
	if err := os.MkdirAll(p.cloneBaseDir, 0755); err != nil {
		log.Printf("content: help: cannot create clone dir %s: %v", p.cloneBaseDir, err)
	}
	go p.run(ctx)
}

func (p *Poller) run(ctx context.Context) {
	p.poll()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll()
		}
	}
}

func (p *Poller) poll() {
	repoDir := filepath.Join(p.cloneBaseDir, gitutil.SanitizeName("fusion-help"))

	if _, err := gitutil.EnsureRepo(p.cfg.URL, repoDir, p.cfg.Token); err != nil {
		log.Printf("content: help: sync repo: %v", err)
		return
	}

	helpRoot := filepath.Join(repoDir, p.cfg.Dir)
	var articles []help.Article

	err := filepath.WalkDir(helpRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, relErr := filepath.Rel(helpRoot, path)
		if relErr != nil {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			log.Printf("content: help: read %s: %v", rel, readErr)
			return nil
		}

		article, parseErr := help.ParseArticle(rel, data)
		if parseErr != nil {
			log.Printf("content: help: parse %s: %v", rel, parseErr)
			return nil
		}

		articles = append(articles, article)
		return nil
	})
	if err != nil {
		log.Printf("content: help: walk %s: %v", helpRoot, err)
	}

	p.onUpdate(articles)
	log.Printf("content: help: loaded %d article(s)", len(articles))
}
