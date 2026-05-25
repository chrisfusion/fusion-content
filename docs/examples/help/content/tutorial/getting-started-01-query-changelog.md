---
title: "Getting started — 1. Query the changelog"
summary: "Learn how to fetch merged changelog entries from fusion-content and filter them by project or date."
tags:
  - changelog
  - quickstart
  - series:content-getting-started
routes:
  - /changelog
  - /changelog/browse
---

## What you will learn

By the end of this tutorial you will be able to:

- Fetch all changelog entries across every tracked project
- Filter by project name and by release date
- Understand the response structure so you can render entries in a UI

## Prerequisites

- `kubectl` access to the `fusion` namespace
- fusion-content deployed and at least one repository configured in the repos Secret

## Step 1 — Port-forward the service

```sh
kubectl port-forward -n fusion service/fusion-content 18082:8080
```

Leave this running in a separate terminal. All `curl` commands below target `http://localhost:18082`.

## Step 2 — Fetch all entries

```sh
curl http://localhost:18082/api/v1/changelog
```

The response groups entries by release date across all configured projects:

```json
{
  "data": [
    {
      "date": "unreleased",
      "projects": [
        {
          "project": "fusion-forge",
          "version": "Unreleased",
          "changes": {
            "added": ["Support for custom base images"]
          }
        }
      ]
    },
    {
      "date": "2026-05-01",
      "projects": [
        {
          "project": "fusion-forge",
          "version": "1.4.0",
          "changes": {
            "added": ["Parallel build support"],
            "fixed": ["Artifact upload retry on 503"]
          }
        },
        {
          "project": "fusion-index",
          "version": "1.2.1",
          "changes": {
            "fixed": ["Tag listing pagination off-by-one"]
          }
        }
      ]
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 12
  }
}
```

Date groups are ordered newest first. `unreleased` always appears at the top.
Pagination is over date groups, not individual project entries.

## Step 3 — Filter by project

```sh
curl "http://localhost:18082/api/v1/changelog?project=fusion-forge"
```

`project` is an exact match against the `name` field from the repos config.
Only date groups that contain an entry for the specified project are returned,
and within each group only that project's entry is included.

## Step 4 — Filter by date

```sh
# All entries released on a specific date
curl "http://localhost:18082/api/v1/changelog?date=2026-05-01"

# Only unreleased entries
curl "http://localhost:18082/api/v1/changelog?date=unreleased"
```

`date` is an exact match. The only special value is `unreleased`; all other
values must be `YYYY-MM-DD`.

## Step 5 — Combine filters and paginate

```sh
curl "http://localhost:18082/api/v1/changelog?project=fusion-forge&page=2&pageSize=5"
```

Defaults: `page=1`, `pageSize=20`. Maximum `pageSize` is 100.

## Next step

Continue to [part 2 — Query help articles](getting-started-02-query-help) to
learn how to fetch structured documentation through the same service.

---

**Series: Getting started with fusion-content**
**1. Query the changelog** · [2. Query help articles →](getting-started-02-query-help)
