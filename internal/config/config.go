// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package config

import (
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// RepoConfig describes a single git repository to poll.
type RepoConfig struct {
	Name          string `yaml:"name"`
	URL           string `yaml:"url"`
	Token         string `yaml:"token"`
	ChangelogPath string `yaml:"changelogPath"`
}

type reposFile struct {
	Repos []RepoConfig `yaml:"repos"`
}

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	Port           string
	PollInterval   time.Duration
	CloneBaseDir   string
	ReposConfigFile string
	AuthEnabled    bool
	AuthAudience   string
	AuthAllowedSAs []string
	Repos          []RepoConfig
}

// Load reads configuration from environment variables and the repos YAML file.
func Load() *Config {
	pollInterval, err := time.ParseDuration(getEnv("POLL_INTERVAL", "60s"))
	if err != nil {
		log.Printf("content: invalid POLL_INTERVAL, using 60s: %v", err)
		pollInterval = 60 * time.Second
	}

	reposConfigFile := getEnv("REPOS_CONFIG_FILE", "/etc/fusion-content/repos.yaml")

	cfg := &Config{
		Port:            getEnv("HTTP_PORT", "8080"),
		PollInterval:    pollInterval,
		CloneBaseDir:    getEnv("GIT_CLONE_DIR", "/tmp/repos"),
		ReposConfigFile: reposConfigFile,
		AuthEnabled:     getEnv("AUTH_ENABLED", "false") == "true",
		AuthAudience:    getEnv("AUTH_AUDIENCE", ""),
		AuthAllowedSAs:  splitCSV(getEnv("AUTH_ALLOWED_SA", "")),
	}

	cfg.Repos = loadRepos(reposConfigFile)
	return cfg
}

func loadRepos(path string) []RepoConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("content: cannot read repos config %s: %v", path, err)
		return nil
	}

	var rf reposFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		log.Fatalf("content: parse repos config %s: %v", path, err)
	}

	for i := range rf.Repos {
		if rf.Repos[i].ChangelogPath == "" {
			rf.Repos[i].ChangelogPath = "CHANGELOG.md"
		}
	}

	log.Printf("content: loaded %d repo(s) from %s", len(rf.Repos), path)
	return rf.Repos
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
