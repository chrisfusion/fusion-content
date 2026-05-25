---
title: "API query parameters"
summary: "Complete reference for all query parameters accepted by the fusion-content changelog, help, and video endpoints."
tags:
  - api
  - reference
  - changelog
  - help
  - video
routes:
  - /admin/content
---

## Common parameters

Both endpoints support pagination via `page` and `pageSize`.

| Parameter  | Type    | Default | Max | Description                              |
|------------|---------|---------|-----|------------------------------------------|
| `page`     | integer | `1`     | —   | 1-based page number                      |
| `pageSize` | integer | `20`    | 100 | Number of items per page                 |

Values below 1 for `page`, or outside `[1, 100]` for `pageSize`, are silently
clamped to the default.

---

## GET /api/v1/changelog

Returns merged changelog entries grouped by release date, newest first.
`unreleased` always appears first regardless of sort order.

### Parameters

| Parameter | Type   | Description                                                                 |
|-----------|--------|-----------------------------------------------------------------------------|
| `project` | string | Exact match against the `name` field in the repos config. Optional.        |
| `date`    | string | Exact match. Use `unreleased` or `YYYY-MM-DD`. Optional.                   |
| `page`    | int    | See above.                                                                  |
| `pageSize`| int    | See above. Pagination is over date groups, not individual project entries.  |

### Response shape

```json
{
  "data": [
    {
      "date": "2026-05-01",
      "projects": [
        {
          "project": "fusion-forge",
          "version": "1.4.0",
          "changes": {
            "added":   ["…"],
            "changed": ["…"],
            "fixed":   ["…"],
            "removed": ["…"]
          }
        }
      ]
    }
  ],
  "pagination": { "page": 1, "pageSize": 20, "total": 12 }
}
```

Fields within `changes` are omitted when empty (not present in the JSON).

---

## GET /api/v1/help

Returns a paginated list of article metadata. The `body` field is not included
in list responses — use `GET /api/v1/help/:service/:type/:slug` for the body.

### Parameters

| Parameter  | Type   | Description                                                                           |
|------------|--------|---------------------------------------------------------------------------------------|
| `service`  | string | Exact match against the service directory name (e.g. `forge`). Optional.            |
| `type`     | string | Diátaxis type. One of: `tutorial`, `how-to`, `reference`, `explanation`. Optional.  |
| `tag`      | string | Articles whose `tags` list contains this value (exact match). Optional.              |
| `route`    | string | Articles whose `routes` list contains this exact path. Optional.                     |
| `q`        | string | Full-text search query. See search behaviour below. Optional.                        |
| `page`     | int    | See above.                                                                            |
| `pageSize` | int    | See above. Pagination is over individual articles.                                    |

Multiple filters are ANDed: only articles matching every supplied filter are returned.

### Full-text search behaviour (`q`)

- Indexed fields: `title`, `summary`, `body`, all `tags`
- Tokenisation: lowercase alphanumeric sequences; tokens shorter than 2 characters are dropped
- Stop words excluded: `a`, `an`, `the`, `is`, `in`, `of`, `for`, `to`, `and`, `or`, `with`, `it`, `on`, `at`, `by`
- Multiple words in `q` are ANDed: every token must appear somewhere in the article

### List response shape

```json
{
  "data": [
    {
      "service": "forge",
      "type": "how-to",
      "slug": "trigger-a-rebuild",
      "title": "Trigger a rebuild",
      "tags": ["build", "forge"],
      "routes": ["/forge/builds"],
      "summary": "Force fusion-forge to rebuild an existing artifact version."
    }
  ],
  "pagination": { "page": 1, "pageSize": 20, "total": 7 }
}
```

---

## GET /api/v1/help/:service/:type/:slug

Returns a single article including the Markdown body.

### Path parameters

| Parameter | Description                                                      |
|-----------|------------------------------------------------------------------|
| `service` | Service identifier (matches directory name in the help repo)     |
| `type`    | Diátaxis type: `tutorial`, `how-to`, `reference`, `explanation`  |
| `slug`    | File name without the `.md` extension                            |

### Response shape

```json
{
  "service": "forge",
  "type": "how-to",
  "slug": "trigger-a-rebuild",
  "title": "Trigger a rebuild",
  "tags": ["build", "forge"],
  "routes": ["/forge/builds"],
  "summary": "Force fusion-forge to rebuild an existing artifact version.",
  "body": "## Prerequisites\n\n..."
}
```

`body` is the raw Markdown text of the article, excluding the frontmatter block.
Returns `404` if no article matches the identity triple.

---

---

## GET /api/v1/videos

Returns a paginated list of video metadata entries. All fields are present in
list responses — there is no separate single-item endpoint needed for
additional detail, but the `/:service/:slug` form is available for direct
lookup.

### Parameters

| Parameter  | Type   | Description                                                                |
|------------|--------|----------------------------------------------------------------------------|
| `service`  | string | Exact match against the service directory name (e.g. `forge`). Optional. |
| `page`     | int    | See above.                                                                 |
| `pageSize` | int    | See above. Pagination is over individual video entries.                    |

### Response shape

```json
{
  "data": [
    {
      "service": "forge",
      "slug": "forge-overview",
      "title": "Fusion Forge Overview",
      "summary": "A walkthrough of venv builds and GitOps polling",
      "thumbnailUrl": "https://img.youtube.com/vi/XXXXXXXXXXX/hqdefault.jpg",
      "videoUrl": "https://youtube.com/watch?v=XXXXXXXXXXX",
      "tags": ["forge", "overview"]
    }
  ],
  "pagination": { "page": 1, "pageSize": 20, "total": 8 }
}
```

---

## GET /api/v1/videos/:service/:slug

Returns a single video entry. The response shape is identical to a single item
from the list endpoint.

### Path parameters

| Parameter | Description                                                     |
|-----------|-----------------------------------------------------------------|
| `service` | Service identifier (matches directory name in the video repo)   |
| `slug`    | File name without the `.md` extension                           |

Returns `404` if no video matches the service and slug combination.

---

## Health endpoints

These endpoints are always public — they do not require authentication.

| Endpoint          | Description                          |
|-------------------|--------------------------------------|
| `GET /q/health/live`  | Liveness probe — returns `{"status":"UP"}` |
| `GET /q/health/ready` | Readiness probe — returns `{"status":"UP"}` |
