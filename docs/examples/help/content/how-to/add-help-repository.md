---
title: "Add a help article repository"
summary: "Configure fusion-content to serve structured help articles from a dedicated git repository."
tags:
  - configuration
  - help
  - git
routes:
  - /admin/content
  - /admin/content/help
---

## When to use this

fusion-content serves help articles separately from changelogs. The help
feature is disabled by default. Use this guide to enable it by pointing the
service at a repository that follows the `help/<service>/<type>/<slug>.md`
layout.

## Prerequisites

- A help repository organised as described in the
  [help article authoring guide](../../docs/help-tutorial.md)
- `kubectl` access to the `fusion` namespace

## Step 1 — Update the repos Secret

The help repository URL goes in the same `repos.yaml` Secret as the changelog
repos, under a top-level `help:` key.

Get the current config:

```sh
kubectl -n fusion get secret fusion-content-repos -o jsonpath='{.data.repos\.yaml}' \
  | base64 -d > /tmp/repos.yaml
```

Append the `help:` section:

```yaml
repos:
  - name: fusion-forge
    url: https://github.com/your-org/fusion-forge
    token: ""

help:
  url: https://github.com/your-org/fusion-docs
  token: ""        # omit or leave empty for public repos
  dir: "help"      # subdirectory that contains the article tree; default "help"
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
content: help repo configured: https://github.com/your-org/fusion-docs
content: help: loaded 23 article(s)
```

Query the API to confirm articles are visible:

```sh
kubectl port-forward -n fusion service/fusion-content 18082:8080
curl http://localhost:18082/api/v1/help
```

## Adjusting the poll interval

By default fusion-content polls every 60 seconds. Articles committed to the
help repository are live within one poll cycle. To change the interval, update
the `pollInterval` value in Helm and redeploy:

```yaml
server:
  config:
    pollInterval: "120s"
```

The same interval applies to both the changelog pollers and the help poller.

## Disabling the help feature

To disable without removing the Secret entry, set `url` to an empty string:

```yaml
help:
  url: ""
```

Restart the pod. The help poller will not start and `GET /api/v1/help` will
return an empty list.
