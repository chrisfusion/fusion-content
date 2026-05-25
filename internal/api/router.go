// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"fusion-platform.io/fusion-content/internal/api/handlers"
	"fusion-platform.io/fusion-content/internal/api/middleware"
	"fusion-platform.io/fusion-content/internal/config"
	"fusion-platform.io/fusion-content/internal/helpstore"
	"fusion-platform.io/fusion-content/internal/store"
	"fusion-platform.io/fusion-content/internal/videostore"
)

// NewRouter wires up the Gin router with all routes and middleware.
func NewRouter(s *store.Store, hs *helpstore.Store, vs *videostore.Store, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// Health probes (compatible with Quarkus Smallrye Health path convention).
	r.GET("/q/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})
	r.GET("/q/health/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	ch := &handlers.ChangelogHandler{Store: s}
	hh := &handlers.HelpHandler{Store: hs}
	vh := &handlers.VideoHandler{Store: vs}

	v1 := r.Group("/api/v1")
	v1.Use(middleware.NewAuthMiddleware(cfg))
	v1.GET("/changelog", ch.List)
	v1.GET("/help", hh.List)
	v1.GET("/help/:service/:type/:slug", hh.Get)
	v1.GET("/videos", vh.List)
	v1.GET("/videos/:service/:slug", vh.Get)

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "accept,authorization,content-type,x-requested-with")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
