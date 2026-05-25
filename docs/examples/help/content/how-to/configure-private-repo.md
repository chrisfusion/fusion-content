---
title: "Configure a private changelog repository"
summary: "Add a private GitHub repository as a changelog source using a personal access token stored in a Kubernetes Secret."
tags:
  - configuration
  - changelog
  - git
  - security
routes:
  - /admin/content
  - /admin/content/repos
---

## When to use this

Use this guide when the repository you want to track is private and requires
authentication. For public repositories, omit the `token` field entirely.

## Prerequisites

- `kubectl` access to the `fusion` namespace with permission to create/patch Secrets
- A GitHub personal access token (classic or fine-grained) with at least
  `Contents: Read` access on the target repository

## Option A — Edit the generated Secret directly

This is suitable for local development or one-off changes.

1. Decode and edit the current repos Secret:

   ```sh
   kubectl -n fusion get secret fusion-content-repos -o jsonpath='{.data.repos\.yaml}' \
     | base64 -d > /tmp/repos.yaml
   ```

2. Add the private repository entry under `repos:`:

   ```yaml
   repos:
     - name: fusion-forge
       url: https://github.com/your-org/fusion-forge
       token: ""
     - name: my-private-service
       url: https://github.com/your-org/my-private-service
       token: "ghp_xxxxxxxxxxxxxxxxxxxx"
       changelogPath: CHANGELOG.md
   ```

3. Re-apply the Secret:

   ```sh
   kubectl -n fusion create secret generic fusion-content-repos \
     --from-file=repos.yaml=/tmp/repos.yaml \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

4. Restart the pod to reload configuration (config is read at startup):

   ```sh
   kubectl -n fusion rollout restart deployment/fusion-content-server
   ```

## Option B — Use an existing Secret (ESO / production)

In production, manage the Secret externally (e.g. via External Secrets
Operator) and point the chart at it.

In your Helm values:

```yaml
repos:
  existingSecret: my-org-content-repos
```

The referenced Secret must have a `repos.yaml` key with the same format as
above. The chart will not generate its own Secret when `existingSecret` is set.

## Verifying

After the pod restarts, check the logs:

```sh
kubectl -n fusion logs -l app=fusion-content-server --tail=20
```

A successful load looks like:

```
content: loaded 2 repo(s) from /etc/fusion-content/repos.yaml
content: updated my-private-service (14 entries)
```

If authentication fails you will see:

```
content: sync my-private-service: authentication required
```

In that case verify the token has not expired and has Contents read access.

## Security notes

- Never commit tokens to the Helm values files. Use `repos.existingSecret` in
  production environments.
- The `values-local.yaml` file is listed in `.gitignore` for local overrides
  that include tokens.
- Tokens are transmitted over HTTPS only; the chart enforces `readOnlyRootFilesystem`
  on the pod so cloned repos stay in the emptyDir volume.
