// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package dto

import "fusion-platform.io/fusion-content/internal/help"

// HelpListResponse is the JSON envelope returned by GET /api/v1/help.
type HelpListResponse struct {
	Data       []help.ArticleSummary `json:"data"`
	Pagination Pagination            `json:"pagination"`
}

// ArticleResponse is the JSON envelope returned by GET /api/v1/help/:service/:type/:slug.
type ArticleResponse struct {
	Service string            `json:"service"`
	Type    help.DiátaxisType `json:"type"`
	Slug    string            `json:"slug"`
	Title   string            `json:"title"`
	Tags    []string          `json:"tags"`
	Routes  []string          `json:"routes"`
	Summary string            `json:"summary"`
	Body    string            `json:"body"`
}
