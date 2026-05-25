---
title: "Getting started — 2. Query help articles"
summary: "Learn how to search and retrieve structured help articles from fusion-content using service, type, tag, route, and full-text filters."
tags:
  - help
  - search
  - quickstart
  - series:content-getting-started
routes:
  - /help
  - /help/browse
---

## What you will learn

- List and filter help articles by service, Diátaxis type, tag, and route
- Run a full-text search query
- Fetch the full body of a single article

## Prerequisites

- Completed [part 1 — Query the changelog](getting-started-01-query-changelog)
- A help repository configured in the repos Secret (`help.url` set)
- Port-forward still running on `18082`

## Step 1 — List all articles

```sh
curl http://localhost:18082/api/v1/help
```

The list endpoint returns article metadata and summary but **not** the body.

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
  "pagination": { "page": 1, "pageSize": 20, "total": 1 }
}
```

## Step 2 — Filter by service

```sh
curl "http://localhost:18082/api/v1/help?service=forge"
```

`service` is an exact match against the directory name in the help repo
(`help/<service>/…`). Use the short service identifier, not the full name
(e.g. `forge`, not `fusion-forge`).

## Step 3 — Filter by Diátaxis type

```sh
curl "http://localhost:18082/api/v1/help?service=forge&type=tutorial"
```

Valid values: `tutorial`, `how-to`, `reference`, `explanation`.

## Step 4 — Filter by tag

```sh
curl "http://localhost:18082/api/v1/help?tag=quickstart"
```

A single tag value; returns all articles whose `tags` list contains it.
You cannot pass multiple tags in one query — filter client-side if needed.

## Step 5 — Filter by route

```sh
curl "http://localhost:18082/api/v1/help?route=/forge/builds/new"
```

Returns articles whose `routes` list contains the exact path. This is the
mechanism used for context-sensitive help: pass the current frontend URL and
receive articles relevant to that view.

## Step 6 — Full-text search

```sh
curl "http://localhost:18082/api/v1/help?q=rebuild+artifact"
```

The search is AND semantics: every word in `q` must appear somewhere in the
article's title, summary, tags, or body. Short words (< 2 characters) and
common stop words are ignored.

You can combine `q` with any of the other filters:

```sh
curl "http://localhost:18082/api/v1/help?service=forge&type=how-to&q=rebuild"
```

## Step 7 — Fetch the full body

The list endpoint omits the body to keep responses small. To get the Markdown
body, call the single-article endpoint:

```sh
curl http://localhost:18082/api/v1/help/forge/how-to/trigger-a-rebuild
```

```json
{
  "service": "forge",
  "type": "how-to",
  "slug": "trigger-a-rebuild",
  "title": "Trigger a rebuild",
  "tags": ["build", "forge"],
  "routes": ["/forge/builds"],
  "summary": "Force fusion-forge to rebuild an existing artifact version.",
  "body": "## Prerequisites\n\n- You need the artifact name..."
}
```

The `body` field contains the raw Markdown text. Render it client-side.

## What you have learned

You can now:

- List and filter articles by every available dimension
- Implement context-sensitive help using the `route` filter
- Retrieve the full Markdown body for rendering

For the full parameter reference see [Help API query parameters](../reference/api-query-parameters).

## Next step

Continue to [part 3 — Query videos](getting-started-03-query-videos) to
learn how to list and filter video metadata through the same service.

---

**Series: Getting started with fusion-content**
[← 1. Query the changelog](getting-started-01-query-changelog) · **2. Query help articles** · [3. Query videos →](getting-started-03-query-videos)
