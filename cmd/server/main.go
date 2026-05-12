// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"fusion-platform.io/fusion-content/internal/api"
	appconfig "fusion-platform.io/fusion-content/internal/config"
	"fusion-platform.io/fusion-content/internal/gitpoller"
	"fusion-platform.io/fusion-content/internal/parser"
	"fusion-platform.io/fusion-content/internal/store"
)

func main() {
	cfg := appconfig.Load()

	s := store.New()

	poller := gitpoller.New(cfg.Repos, cfg.PollInterval, cfg.CloneBaseDir, func(project string, entries []parser.Entry) {
		s.Update(project, entries)
	})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	poller.Start(ctx)

	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{Addr: addr, Handler: api.NewRouter(s, cfg)}

	go func() {
		log.Printf("starting fusion-content on %s (poll interval: %s, repos: %d)",
			addr, cfg.PollInterval, len(cfg.Repos))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
