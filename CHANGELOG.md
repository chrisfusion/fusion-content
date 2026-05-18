# Changelog

All notable changes to fusion-content are documented here.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [Unreleased]

### Added
- fusion-bff integration: proxied at `/api/content/*`; `content:changelog:read` permission required (granted to admin, engineer, and viewer roles)
- Help content API: `GET /api/v1/help` (list/filter/full-text search) and `GET /api/v1/help/:service/:type/:slug` (full article with body); articles served from a dedicated git repo organised as `help/<service>/<type>/<slug>.md` (Diátaxis taxonomy: tutorial, how-to, reference, explanation) with YAML frontmatter (title, tags, routes, summary)
- In-memory full-text inverted index for help articles: tokenised AND-intersection search across title, tags, summary, and body; rebuilt atomically on every poll cycle
- Route-aware help filtering: `?route=<path>` returns articles whose `routes` frontmatter field contains the given frontend path
- Shared git utility package (`internal/gitutil`): `SanitizeName`, `BuildAuth`, `EnsureRepo` (clone-or-pull-or-nuke); used by both the changelog and help pollers
- Help repo configured via new `help:` section in the existing `repos.yaml` Secret (url, token, dir); feature is disabled when `help.url` is absent

---

## [0.1.0] — 2026-05-12

### Added
- Go 1.25 / Gin service scaffold: `cmd/server/main.go`, `internal/config/config.go`, `Dockerfile` (golang:1.25-alpine builder + distroless:nonroot runtime), `Makefile`
- Keep-a-Changelog parser (`internal/parser/parser.go`): handles plain, parenthesised-link, and bracket-ref version headings; normalises section names; 1 MB scanner buffer guards against long lines
- Git poller (`internal/gitpoller/poller.go`): one goroutine per repo, configurable poll interval (`POLL_INTERVAL`, default `60s`), shallow clone (`--depth 1`), HTTP token auth for private repos; corrupt or partial clone directories are detected and re-cloned automatically
- Thread-safe in-memory store (`internal/store/store.go`): merges entries from all repos by date; multiple versions of the same project on the same day each produce a distinct entry; sorted newest-first with `unreleased` always at the top
- REST endpoint `GET /api/v1/changelog` with page/pageSize pagination and optional `date` + `project` query filters
- K8s SA TokenReview auth middleware (`internal/api/middleware/auth.go`): toggled via `AUTH_ENABLED`; own SA token re-read per request to survive kubelet rotation; `AUTH_ALLOWED_SA` allowlist support
- Health probes at `/q/health/live` and `/q/health/ready` (Quarkus Smallrye path convention)
- Graceful HTTP shutdown: SIGINT/SIGTERM stops the poller goroutines and drains in-flight requests within a 15 s deadline
- Repos configuration via a YAML file (path: `REPOS_CONFIG_FILE`, default `/etc/fusion-content/repos.yaml`) with per-repo `name`, `url`, `token`, and `changelogPath` fields
- Helm chart (`deployment/`): Deployment, Service, ConfigMap, repos Secret (with `existingSecret` escape hatch for ESO), optional Ingress, ServiceAccount, conditional ClusterRole + ClusterRoleBinding for TokenReview; all security contexts, pod/deployment labels and annotations fully configurable in `values.yaml`
- `.gitignore` covering Go build artefacts, popular editors (IntelliJ, VS Code, Emacs, Vim, Neovim, Sublime Text, Atom, TextMate, Eclipse, Notepad++), Helm packaging artefacts, local values overrides, TLS keys, and `.claude/`
