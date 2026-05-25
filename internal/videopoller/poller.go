// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package videopoller

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
	"fusion-platform.io/fusion-content/internal/video"
)

// OnUpdateFunc is called with the full video list after each successful poll.
type OnUpdateFunc func(videos []video.Video)

// Poller polls a single video content git repository and calls onUpdate on every cycle.
type Poller struct {
	cfg          config.VideoConfig
	interval     time.Duration
	cloneBaseDir string
	onUpdate     OnUpdateFunc
}

// New returns a Poller for the given video config.
func New(cfg config.VideoConfig, interval time.Duration, cloneBaseDir string, onUpdate OnUpdateFunc) *Poller {
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
		log.Printf("content: videos: cannot create clone dir %s: %v", p.cloneBaseDir, err)
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
	repoDir := filepath.Join(p.cloneBaseDir, gitutil.SanitizeName("fusion-videos"))

	if _, err := gitutil.EnsureRepo(p.cfg.URL, repoDir, p.cfg.Token); err != nil {
		log.Printf("content: videos: sync repo: %v", err)
		return
	}

	videosRoot := filepath.Join(repoDir, p.cfg.Dir)
	var videos []video.Video

	err := filepath.WalkDir(videosRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, relErr := filepath.Rel(videosRoot, path)
		if relErr != nil {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			log.Printf("content: videos: read %s: %v", rel, readErr)
			return nil
		}

		v, parseErr := video.ParseVideo(rel, data)
		if parseErr != nil {
			log.Printf("content: videos: parse %s: %v", rel, parseErr)
			return nil
		}

		videos = append(videos, v)
		return nil
	})
	if err != nil {
		log.Printf("content: videos: walk %s: %v", videosRoot, err)
	}

	p.onUpdate(videos)
	log.Printf("content: videos: loaded %d video(s)", len(videos))
}
