# fusion-content

## Project
Changelog aggregation service: polls configurable git repositories on a configurable interval, parses CHANGELOG.md files (Keep-a-Changelog format), merges by date, and serves the result via a paginated REST API.

## Ecosystem
- **BFF**: `../fusion-bff` — proxies `/api/content/*` to this service; `content:changelog:read` permission required
- **Peer backends**: `../fusion-forge`, `../fusion-index`, `../fusion-flux` (weave operator)
- **Do not use fusion-flux as a convention reference** — it follows slightly different patterns

## Stack
- **Go 1.25**, **Gin**, no database (in-memory store), **go-git v5** for git polling
- Module: `fusion-platform.io/fusion-content`

## Build
- `make build` — `go build ./...`
- `make tidy` — `go mod tidy`
- `make docker-build` — builds inside minikube (`eval $(minikube docker-env)` handled by Makefile)
- `make run` — `go run ./cmd/server/`

## Local minikube deploy
- `make docker-build && helm upgrade --install fusion-content deployment/ -n fusion -f deployment/values-local.yaml`
- `deployment/values-local.yaml` — local overrides; not committed with real tokens
- Changelog sources for local dev: public repos `github.com/chrisfusion/fusion-{forge,index,spectra,bff}` (no token needed)
- Do **not** try to serve host-filesystem repos via git daemon — use the public GitHub mirrors instead
- Port-forward for testing: `kubectl port-forward -n fusion service/fusion-content 18082:8080`
- `fusion` namespace already exists — `createNamespace: false` (the default) is correct

## Conventions (follow fusion-forge, not fusion-flux)
- Health probes at `/q/health/live` and `/q/health/ready` (Quarkus path convention)
- CORS: `GET,POST,PUT,DELETE,OPTIONS` — match platform standard even if only GET is exposed today
- SPDX header on every `.go` file: `// SPDX-License-Identifier: GPL-3.0-or-later`
- Auth middleware: K8s SA TokenReview — copy pattern from `../fusion-forge/internal/api/middleware/auth.go`
- SA token re-read per request (kubelet rotates projected tokens — never cache it)

## Helm chart (deployment/)
- Every component section must expose: `podSecurityContext`, `containerSecurityContext`, `deploymentLabels`, `deploymentAnnotations`, `podLabels`, `podAnnotations`
- `readOnlyRootFilesystem: true` requires an `emptyDir` volume mounted at `/tmp` — git clone writes there
- Auth toggle: `auth.enabled` is the single source of truth; `configmap.yaml` reads from `auth.*`, not from `server.config.auth*`
- Repos config is a K8s **Secret** (not ConfigMap) — it contains tokens; supports `repos.existingSecret` for ESO

## CHANGELOG
- Entry required before every commit — Keep-a-Changelog format
- Version heading format: `## [x.y.z] — YYYY-MM-DD` (em dash `—`, not hyphen)
- Add entries under `## [Unreleased]`; create a new `## [x.y.z]` section when releasing

## Repos config file format
```yaml
repos:
  - name: fusion-forge
    url: https://github.com/org/fusion-forge
    token: ""            # omit for public repos; HTTP personal access token for private
    changelogPath: CHANGELOG.md   # optional, this is the default
```
Mounted from a K8s Secret at `REPOS_CONFIG_FILE` (default `/etc/fusion-content/repos.yaml`).
Git clone dir: `GIT_CLONE_DIR` (default `/tmp/repos`) — must be on a writable volume.
