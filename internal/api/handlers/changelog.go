// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"fusion-platform.io/fusion-content/internal/api/dto"
	"fusion-platform.io/fusion-content/internal/store"
)

// ChangelogHandler serves the merged changelog API.
type ChangelogHandler struct {
	Store *store.Store
}

// List handles GET /api/v1/changelog.
//
// Query params:
//
//	page     — 1-based page number (default 1)
//	pageSize — entries per page (default 20, max 100)
//	date     — filter to a single date ("unreleased" or "YYYY-MM-DD")
//	project  — filter to a single project name
func (h *ChangelogHandler) List(c *gin.Context) {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 20)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	project := c.Query("project")
	date := c.Query("date")

	groups, total := h.Store.Query(project, date, page, pageSize)

	if groups == nil {
		groups = []store.DateGroup{}
	}

	c.JSON(http.StatusOK, dto.ChangelogResponse{
		Data: groups,
		Pagination: dto.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

func queryInt(c *gin.Context, key string, def int) int {
	raw := c.Query(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}
