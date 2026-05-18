// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"fusion-platform.io/fusion-content/internal/api/dto"
	"fusion-platform.io/fusion-content/internal/helpstore"
)

// HelpHandler serves the help article API.
type HelpHandler struct {
	Store *helpstore.Store
}

// List handles GET /api/v1/help.
//
// Query params:
//
//	service  — filter by service name (exact match)
//	type     — filter by Diátaxis type (tutorial|how-to|reference|explanation)
//	tag      — filter to articles containing this tag
//	route    — filter to articles whose routes contain this path (exact match)
//	q        — full-text search query
//	page     — 1-based page number (default 1)
//	pageSize — entries per page (default 20, max 100)
func (h *HelpHandler) List(c *gin.Context) {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 20)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	summaries, total := h.Store.Query(helpstore.HelpQuery{
		Service:  c.Query("service"),
		Type:     c.Query("type"),
		Tag:      c.Query("tag"),
		Route:    c.Query("route"),
		Q:        c.Query("q"),
		Page:     page,
		PageSize: pageSize,
	})

	c.JSON(http.StatusOK, dto.HelpListResponse{
		Data: summaries,
		Pagination: dto.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// Get handles GET /api/v1/help/:service/:type/:slug.
func (h *HelpHandler) Get(c *gin.Context) {
	service := c.Param("service")
	docType := c.Param("type")
	slug := c.Param("slug")

	article, ok := h.Store.Get(service, docType, slug)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	c.JSON(http.StatusOK, dto.ArticleResponse{
		Service: article.Service,
		Type:    article.Type,
		Slug:    article.Slug,
		Title:   article.Title,
		Tags:    article.Tags,
		Routes:  article.Routes,
		Summary: article.Summary,
		Body:    article.Body,
	})
}
