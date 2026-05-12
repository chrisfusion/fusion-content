// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 fusion-platform contributors

package dto

import "fusion-platform.io/fusion-content/internal/store"

// Pagination describes the current page position in a paginated result set.
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

// ChangelogResponse is the JSON envelope returned by GET /api/v1/changelog.
type ChangelogResponse struct {
	Data       []store.DateGroup `json:"data"`
	Pagination Pagination        `json:"pagination"`
}
