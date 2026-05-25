---
title: "Add a video repository"
summary: "Configure fusion-content to serve video metadata from a dedicated git repository."
tags:
  - configuration
  - video
  - git
routes:
  - /admin/content
  - /admin/content/video
---

## When to use this

fusion-content serves video metadata (titles, summaries, thumbnail URLs, and
video URLs) separately from changelogs and help articles. The video feature is
disabled by default. Use this guide to enable it by pointing the service at a
repository that follows the `videos/<service>/<slug>.md` layout.

## Prerequisites

- A video repository organised as `videos/<service>/<slug>.md`, where each
  file contains only YAML frontmatter (title, service, summary, thumbnailUrl,
  videoUrl, tags) and no body text
- `kubectl` access to the `fusion` namespace

## Step 1 — Update the repos Secret

The video repository URL goes in the same `repos.yaml` Secret as the changelog
repos and the help repo, under a top-level `videos:` key.

Get the current config:

```sh
kubectl -n fusion get secret fusion-content-repos -o jsonpath='{.data.repos\.yaml}' \
  | base64 -d > /tmp/repos.yaml
```

Append the `videos:` section:

```yaml
repos:
  - name: fusion-forge
    url: https://github.com/your-org/fusion-forge
    token: ""

help:
  url: https://github.com/your-org/fusion-docs
  token: ""
  dir: "help"

videos:
  url: https://github.com/your-org/fusion-videos
  token: ""        # omit or leave empty for public repos
  dir: "videos"    # subdirectory that contains the video tree; default "videos"
```

Re-apply:

```sh
kubectl -n fusion create secret generic fusion-content-repos \
  --from-file=repos.yaml=/tmp/repos.yaml \
  --dry-run=client -o yaml | kubectl apply -f -
```

## Step 2 — Restart the pod

```sh
kubectl -n fusion rollout restart deployment/fusion-content-server
```

## Step 3 — Verify

Check the logs:

```sh
kubectl -n fusion logs -l app=fusion-content-server --tail=20
```

Expected output:

```
content: loaded 1 repo(s) from /etc/fusion-content/repos.yaml
content: video repo configured: https://github.com/your-org/fusion-videos
content: videos: loaded 8 video(s)
```

Query the API to confirm videos are visible:

```sh
kubectl port-forward -n fusion service/fusion-content 18082:8080
curl http://localhost:18082/api/v1/videos
```

## Video file format

Each file in the repository must be at exactly `videos/<service>/<slug>.md`.
Only one directory level is allowed inside the `videos/` root — files in
subdirectories are skipped by the poller with a log warning.

The entire entry is expressed in YAML frontmatter; no body text is needed:

```markdown
---
title: "Fusion Forge Overview"
service: forge
summary: "A walkthrough of venv builds and GitOps polling"
thumbnailUrl: "https://img.youtube.com/vi/XXXXXXXXXXX/hqdefault.jpg"
videoUrl: "https://youtube.com/watch?v=XXXXXXXXXXX"
tags: ["forge", "overview"]
---
```

Valid `service` values must match the activity-rail context IDs used in
spectra: `data`, `weave`, `monitoring`, `forge`, `fusion-index`, `admin`.

## Adjusting the poll interval

The video poller shares the global `POLL_INTERVAL` with the changelog and help
pollers (default `60s`). A new video file committed to the repository is live
within one full poll cycle. To change the interval, update Helm values and
redeploy:

```yaml
server:
  config:
    pollInterval: "120s"
```

## Disabling the video feature

To disable without removing the Secret entry, set `url` to an empty string:

```yaml
videos:
  url: ""
```

Restart the pod. The video poller will not start and `GET /api/v1/videos` will
return an empty list.
