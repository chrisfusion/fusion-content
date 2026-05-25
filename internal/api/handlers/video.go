// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"fusion-platform.io/fusion-content/internal/api/dto"
	"fusion-platform.io/fusion-content/internal/video"
	"fusion-platform.io/fusion-content/internal/videostore"
)

// VideoHandler serves the video metadata API.
type VideoHandler struct {
	Store *videostore.Store
}

// List handles GET /api/v1/videos.
//
// Query params:
//
//	service  — filter by service identifier (exact match)
//	page     — 1-based page number (default 1)
//	pageSize — entries per page (default 20, max 100)
func (h *VideoHandler) List(c *gin.Context) {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 20)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	videos, total := h.Store.Query(videostore.VideoQuery{
		Service:  c.Query("service"),
		Page:     page,
		PageSize: pageSize,
	})

	data := make([]dto.VideoResponse, len(videos))
	for i, v := range videos {
		data[i] = toVideoResponse(v)
	}

	c.JSON(http.StatusOK, dto.VideoListResponse{
		Data: data,
		Pagination: dto.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// Get handles GET /api/v1/videos/:service/:slug.
func (h *VideoHandler) Get(c *gin.Context) {
	service := c.Param("service")
	slug := c.Param("slug")

	v, ok := h.Store.Get(service, slug)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}

	c.JSON(http.StatusOK, toVideoResponse(v))
}

func toVideoResponse(v video.Video) dto.VideoResponse {
	return dto.VideoResponse{
		Service:      v.Service,
		Slug:         v.Slug,
		Title:        v.Title,
		Summary:      v.Summary,
		ThumbnailURL: v.ThumbnailURL,
		VideoURL:     v.VideoURL,
		Tags:         v.Tags,
	}
}
