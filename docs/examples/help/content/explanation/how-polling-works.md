---
title: "How polling works"
summary: "Explains how fusion-content syncs changelog, help, and video content from git repositories into its in-memory store, and what to expect around consistency and latency."
tags:
  - internals
  - git
  - polling
  - changelog
  - help
  - video
routes:
  - /admin/content
---

## Overview

fusion-content has no database. All content lives in memory, rebuilt from git
on every poll cycle. This design keeps the service stateless and deployable
with a read-only root filesystem.

## The three polling pipelines

fusion-content runs three independent polling pipelines in parallel:

```
Changelog pipeline (one goroutine per repo):
  git clone/pull  →  parse CHANGELOG.md  →  changelog store

Help pipeline (one goroutine, one repo):
  git clone/pull  →  walk help/<service>/<type>/*.md  →  helpstore

Video pipeline (one goroutine, one repo):
  git clone/pull  →  walk videos/<service>/*.md  →  videostore
```

All three pipelines share `internal/gitutil.EnsureRepo` for clone/pull logic
but are otherwise independent. A failure in one does not affect the others.
The help and video pipelines only start when their respective `url` is
configured in the repos Secret.

## Startup behaviour

On startup, each goroutine performs one immediate poll before entering the
ticker loop. This means content is available within a few seconds of the pod
becoming ready, before the first readiness probe passes.

## Poll cycle

At each tick (controlled by `POLL_INTERVAL`, default `60s`):

1. `EnsureRepo` clones the repo if the local directory does not exist, or
   runs `git pull` (fast-forward only) if it does.
2. On success the content is re-parsed and the in-memory store is atomically
   replaced under a write lock.
3. On any git or parse error the existing store is left untouched and the
   error is logged. The next tick retries automatically.

## Atomicity and consistency

Each store update is an atomic swap under a `sync.RWMutex`. A reader always
sees either the previous complete corpus or the new complete corpus — never a
partial state. Concurrent reads during an update are not blocked; only the
swap itself acquires the write lock briefly.

The changelog, help, and video stores each have their own lock, so a slow
pull in one pipeline does not block reads on the others.

## Latency from commit to live

```
push commit  →  up to POLL_INTERVAL  →  git pull  →  store swap  →  API response
```

Maximum latency is one full poll interval. With the default of `60s`, a commit
is live within a minute. Set a shorter interval if faster propagation is
needed, at the cost of more frequent git operations.

## What happens if the remote is unreachable

The existing in-memory content continues to be served. fusion-content logs a
warning per cycle and retries on the next tick. There is no circuit-breaker or
exponential back-off — retries happen at every tick regardless of the failure
count.

## Memory footprint

All content for all configured repos lives in RAM. For typical platform
documentation volumes this is well within the 128 Mi container limit. The help
store also maintains an inverted index (token → article indices) which is
rebuilt on every update; its overhead is proportional to the total word count
across all articles. The video store has no index — it stores only frontmatter
fields, so its footprint is negligible even for large catalogues.

## Why not watch / webhooks

Polling requires no inbound network path into the cluster and no credential
management on the git host side. It also degrades gracefully — a poll failure
is silent to API consumers. Webhooks would offer lower latency but add
operational complexity (public endpoint, HMAC validation, retry logic) that is
not warranted for documentation content.
