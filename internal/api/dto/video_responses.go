// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package dto

// VideoResponse is the JSON shape for a single video, used in both list and get responses.
type VideoResponse struct {
	Service      string   `json:"service"`
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Summary      string   `json:"summary"`
	ThumbnailURL string   `json:"thumbnailUrl"`
	VideoURL     string   `json:"videoUrl"`
	Tags         []string `json:"tags"`
}

// VideoListResponse is the JSON envelope returned by GET /api/v1/videos.
type VideoListResponse struct {
	Data       []VideoResponse `json:"data"`
	Pagination Pagination      `json:"pagination"`
}
