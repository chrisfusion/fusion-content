---
title: "Getting started — 3. Query videos"
summary: "Learn how to list, filter, and retrieve individual video metadata entries from fusion-content."
tags:
  - video
  - quickstart
  - series:content-getting-started
routes:
  - /video
  - /video/browse
---

## What you will learn

- List all videos and filter by service
- Retrieve a single video by service and slug
- Understand the response structure for use in a UI

## Prerequisites

- Completed [part 2 — Query help articles](getting-started-02-query-help)
- A video repository configured in the repos Secret (`videos.url` set)
- Port-forward still running on `18082`

## Step 1 — List all videos

```sh
curl http://localhost:18082/api/v1/videos
```

Each item in `data` contains all metadata needed to render a video card:

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
    },
    {
      "service": "index",
      "slug": "index-overview",
      "title": "Fusion Index Overview",
      "summary": "Artifact indexing explained",
      "thumbnailUrl": "https://img.youtube.com/vi/YYYYYYYYYYY/hqdefault.jpg",
      "videoUrl": "https://youtube.com/watch?v=YYYYYYYYYYY",
      "tags": ["index", "overview"]
    }
  ],
  "pagination": { "page": 1, "pageSize": 20, "total": 8 }
}
```

Unlike the help list endpoint, video list responses include all fields —
there is no separate "full article" call needed because videos have no body.

## Step 2 — Filter by service

```sh
curl "http://localhost:18082/api/v1/videos?service=forge"
```

`service` is an exact match against the directory name in the video repo
(`videos/<service>/…`). Use the short identifier (`forge`, not `fusion-forge`).
Results are sorted alphabetically by service, then by slug within each service.

## Step 3 — Paginate

```sh
curl "http://localhost:18082/api/v1/videos?page=2&pageSize=5"
```

Defaults: `page=1`, `pageSize=20`. Maximum `pageSize` is 100. Filters and
pagination can be combined:

```sh
curl "http://localhost:18082/api/v1/videos?service=forge&page=1&pageSize=5"
```

## Step 4 — Fetch a single video

Use the service and slug to request one entry directly:

```sh
curl http://localhost:18082/api/v1/videos/forge/forge-overview
```

```json
{
  "service": "forge",
  "slug": "forge-overview",
  "title": "Fusion Forge Overview",
  "summary": "A walkthrough of venv builds and GitOps polling",
  "thumbnailUrl": "https://img.youtube.com/vi/XXXXXXXXXXX/hqdefault.jpg",
  "videoUrl": "https://youtube.com/watch?v=XXXXXXXXXXX",
  "tags": ["forge", "overview"]
}
```

Returns `404` if no video matches the service and slug combination.

## What you have learned

You can now:

- Retrieve all video metadata from fusion-content
- Filter by service identifier for context-aware display
- Look up an individual video directly by its identity pair

For the full parameter reference see [API query parameters](../reference/api-query-parameters).

---

**Series: Getting started with fusion-content**
[← 2. Query help articles](getting-started-02-query-help) · **3. Query videos**
