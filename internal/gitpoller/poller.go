// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package gitpoller

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	gittransport "github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"fusion-platform.io/fusion-content/internal/config"
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
	repoDir := filepath.Join(p.cloneBaseDir, sanitizeName(repo.Name))
	auth := buildAuth(repo)

	gitRepo, err := git.PlainOpen(repoDir)
	if err != nil {
		if !errors.Is(err, git.ErrRepositoryNotExists) {
			// Unexpected error (corrupt .git, permission issue) — nuke and re-clone.
			log.Printf("content: open %s: %v — removing and re-cloning", repo.Name, err)
			os.RemoveAll(repoDir)
		}
		gitRepo = p.cloneRepo(repo, repoDir, auth)
		if gitRepo == nil {
			return
		}
	} else {
		if pullErr := p.pullRepo(repo, gitRepo, auth); pullErr != nil {
			// Non-recoverable pull error — nuke and re-clone immediately.
			log.Printf("content: pull %s: %v — removing and re-cloning", repo.Name, pullErr)
			os.RemoveAll(repoDir)
			gitRepo = p.cloneRepo(repo, repoDir, auth)
			if gitRepo == nil {
				return
			}
		}
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

func (p *Poller) cloneRepo(repo config.RepoConfig, repoDir string, auth gittransport.AuthMethod) *git.Repository {
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		log.Printf("content: mkdir %s: %v", repo.Name, err)
		return nil
	}
	gitRepo, err := git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:   repo.URL,
		Auth:  auth,
		Depth: 1,
	})
	if err != nil {
		log.Printf("content: clone %s: %v", repo.Name, err)
		return nil
	}
	log.Printf("content: cloned %s", repo.Name)
	return gitRepo
}

func (p *Poller) pullRepo(repo config.RepoConfig, gitRepo *git.Repository, auth gittransport.AuthMethod) error {
	w, err := gitRepo.Worktree()
	if err != nil {
		return err
	}
	err = w.Pull(&git.PullOptions{Auth: auth, Force: true})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}
	return err
}

func buildAuth(repo config.RepoConfig) gittransport.AuthMethod {
	if repo.Token == "" {
		return nil
	}
	return &githttp.BasicAuth{
		Username: "oauth2",
		Password: repo.Token,
	}
}

func sanitizeName(name string) string {
	return strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, name)
}
