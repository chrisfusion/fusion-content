// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package gitutil

import (
	"errors"
	"log"
	"os"
	"strings"

	git "github.com/go-git/go-git/v5"
	gittransport "github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// SanitizeName converts a repository name to a filesystem-safe directory name.
func SanitizeName(name string) string {
	return strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, name)
}

// BuildAuth returns a go-git AuthMethod for the given token, or nil for public repos.
func BuildAuth(token string) gittransport.AuthMethod {
	if token == "" {
		return nil
	}
	return &githttp.BasicAuth{
		Username: "oauth2",
		Password: token,
	}
}

// EnsureRepo guarantees a clean, up-to-date shallow clone at repoDir.
// It opens an existing clone and pulls, or nukes and re-clones on any error.
func EnsureRepo(repoURL, repoDir, token string) (*git.Repository, error) {
	auth := BuildAuth(token)

	gitRepo, err := git.PlainOpen(repoDir)
	if err != nil {
		if !errors.Is(err, git.ErrRepositoryNotExists) {
			log.Printf("gitutil: open %s: %v — removing and re-cloning", repoDir, err)
			os.RemoveAll(repoDir)
		}
		return clone(repoURL, repoDir, auth)
	}

	if pullErr := pull(gitRepo, auth); pullErr != nil {
		log.Printf("gitutil: pull %s: %v — removing and re-cloning", repoDir, pullErr)
		os.RemoveAll(repoDir)
		return clone(repoURL, repoDir, auth)
	}

	return gitRepo, nil
}

func clone(repoURL, repoDir string, auth gittransport.AuthMethod) (*git.Repository, error) {
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return nil, err
	}
	r, err := git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:   repoURL,
		Auth:  auth,
		Depth: 1,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func pull(r *git.Repository, auth gittransport.AuthMethod) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	err = w.Pull(&git.PullOptions{Auth: auth, Force: true})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}
	return err
}
