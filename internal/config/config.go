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

// RepoConfig describes a single git repository to poll for changelogs.
type RepoConfig struct {
	Name          string `yaml:"name"`
	URL           string `yaml:"url"`
	Token         string `yaml:"token"`
	ChangelogPath string `yaml:"changelogPath"`
}

// HelpConfig describes the single git repository used for help content.
type HelpConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
	Dir   string `yaml:"dir"` // subdirectory within the repo; default "help"
}

// Enabled returns true when a help repo URL is configured.
func (h HelpConfig) Enabled() bool { return h.URL != "" }

// VideoConfig describes the single git repository used for video content.
type VideoConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
	Dir   string `yaml:"dir"` // subdirectory within the repo; default "videos"
}

// Enabled returns true when a video repo URL is configured.
func (v VideoConfig) Enabled() bool { return v.URL != "" }

type reposFile struct {
	Repos  []RepoConfig `yaml:"repos"`
	Help   HelpConfig   `yaml:"help"`
	Videos VideoConfig  `yaml:"videos"`
}

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	Port            string
	PollInterval    time.Duration
	CloneBaseDir    string
	ReposConfigFile string
	AuthEnabled     bool
	AuthAudience    string
	AuthAllowedSAs  []string
	Repos           []RepoConfig
	Help            HelpConfig
	Videos          VideoConfig
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

	repos, help, videos := loadReposFile(reposConfigFile)
	cfg.Repos = repos
	cfg.Help = help
	cfg.Videos = videos
	return cfg
}

func loadReposFile(path string) ([]RepoConfig, HelpConfig, VideoConfig) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("content: cannot read repos config %s: %v", path, err)
		return nil, HelpConfig{}, VideoConfig{}
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
	if rf.Help.Dir == "" {
		rf.Help.Dir = "help"
	}
	if rf.Videos.Dir == "" {
		rf.Videos.Dir = "videos"
	}

	log.Printf("content: loaded %d repo(s) from %s", len(rf.Repos), path)
	if rf.Help.Enabled() {
		log.Printf("content: help repo configured: %s", rf.Help.URL)
	}
	if rf.Videos.Enabled() {
		log.Printf("content: video repo configured: %s", rf.Videos.URL)
	}
	return rf.Repos, rf.Help, rf.Videos
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
