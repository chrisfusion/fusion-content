# fusion-content

## Project
Content API service: polls configurable git repositories, serves changelog entries (Keep-a-Changelog format) and structured help articles (Diátaxis Markdown) via a paginated REST API. In-memory only — no database.

## Ecosystem
- **BFF**: `../fusion-bff` — proxies `/api/content/*` to this service; `content:changelog:read` permission required
- **Peer backends**: `../fusion-forge`, `../fusion-index`, `../fusion-flux` (weave operator)
- **Do not use fusion-flux as a convention reference** — it follows slightly different patterns

## Stack
- **Go 1.25**, **Gin**, no database (in-memory store), **go-git v5** for git polling
- Module: `fusion-platform.io/fusion-content`
- `internal/gitutil` — shared git clone/pull/auth utilities used by all pollers; use `gitutil.EnsureRepo` (not go-git directly) when adding a new content-type poller

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

## Help content (GET /api/v1/help)
- Articles live in a dedicated git repo, organised as `help/<service>/<type>/<slug>.md`
- Diátaxis types: `tutorial` | `how-to` | `reference` | `explanation` — type is derived from the directory name, not frontmatter
- YAML frontmatter fields: `title`, `tags []string`, `routes []string` (frontend paths for context-sensitive help), `summary`
- Full-text inverted index in `internal/helpstore`; search is AND-intersection over tokenised title+tags+summary+body
- Adding a new content type: create `internal/<type>/`, `internal/<type>store/`, `internal/<type>poller/` (uses `gitutil`), handler; wire in `router.go` and `main.go`

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

help:
  url: https://github.com/org/fusion-docs
  token: ""
  dir: "help"            # subdirectory within the repo; default "help"
```
Mounted from a K8s Secret at `REPOS_CONFIG_FILE` (default `/etc/fusion-content/repos.yaml`).
Git clone dir: `GIT_CLONE_DIR` (default `/tmp/repos`) — must be on a writable volume.
