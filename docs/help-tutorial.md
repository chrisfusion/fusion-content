# Writing and deploying help articles for fusion-content

This guide covers everything you need to write a single article or a full
series, get it into the help repository, and have it served through the
fusion-content API.

---

## How the pipeline works

```
fusion-docs repo (GitHub)
  └── help/
        └── <service>/<type>/<slug>.md

          ↓  fusion-content polls on a configurable interval (default 60 s)

  in-memory store  →  GET /api/v1/help
                   →  GET /api/v1/help/:service/:type/:slug
```

fusion-content clones the help repository once, then pulls at every poll
interval. Any commit you push is live within one poll cycle — no deployment
step is required.

---

## Repository layout

Every article is a Markdown file placed at exactly this path inside the
`help/` directory of the help repository:

```
help/<service>/<type>/<slug>.md
```

| Segment   | Meaning                                                        |
|-----------|----------------------------------------------------------------|
| `service` | The platform service the article belongs to (e.g. `forge`)    |
| `type`    | One of the four Diátaxis categories (see below)               |
| `slug`    | URL-safe identifier, lowercase, hyphens for spaces            |

**Example tree for three services:**

```
help/
  forge/
    tutorial/
      create-your-first-build.md
      build-from-a-custom-base.md
    how-to/
      trigger-a-rebuild.md
      override-venv-packages.md
    reference/
      build-api.md
      environment-variables.md
    explanation/
      how-builds-are-isolated.md
  index/
    reference/
      artifact-schema.md
    how-to/
      upload-artifact.md
  bff/
    explanation/
      auth-flow.md
```

---

## Diátaxis types — choosing the right one

| Type          | Directory name  | Answers the question                   | Oriented toward |
|---------------|-----------------|----------------------------------------|-----------------|
| Tutorial      | `tutorial`      | "How do I learn this from scratch?"    | Learning        |
| How-to guide  | `how-to`        | "How do I achieve this specific goal?" | Doing           |
| Reference     | `reference`     | "What does this setting/field do?"     | Information     |
| Explanation   | `explanation`   | "Why does it work this way?"           | Understanding   |

**Rule of thumb:** if you are writing step-by-step instructions for a
beginner, it is a tutorial. If you are documenting what parameters an API
accepts, it is a reference. If you are answering "how do I…?" for an
experienced user, it is a how-to. If you are explaining a design decision or
concept, it is an explanation.

---

## Article file format

Every file must start with a YAML frontmatter block fenced by `---`.

```markdown
---
title: "Create your first build"
summary: "Walk through submitting a Python venv build request end-to-end."
tags:
  - build
  - quickstart
  - forge
routes:
  - /forge/builds
  - /forge/builds/new
---

Your article body goes here. Plain Markdown — headings, lists, code blocks,
links. No special extensions required.
```

### Frontmatter fields

| Field     | Required | Type            | Notes                                                                    |
|-----------|----------|-----------------|--------------------------------------------------------------------------|
| `title`   | yes      | string          | Displayed in search results and the article header                       |
| `summary` | yes      | string          | One or two sentences. Indexed for full-text search; shown in list views  |
| `tags`    | no       | list of strings | Freeform labels; used for tag-filter queries                             |
| `routes`  | no       | list of strings | Frontend URL paths where this article should surface as contextual help  |

**Notes:**
- `type` is **not** a frontmatter field — it is derived from the directory name.
  Do not add a `type:` key; it will be ignored.
- `title` and `summary` are both indexed for full-text search alongside the
  body. Write a meaningful summary — it matters for discoverability.
- `routes` must be exact path strings matching your frontend router (e.g.
  `/forge/builds/new`). The API consumer filters by exact route match.

---

## Writing a single article

1. **Pick the type.** Use the table above.
2. **Choose a slug.** Lowercase, hyphens only, descriptive.
   `trigger-a-rebuild` is good; `doc1` is not.
3. **Create the file** at `help/<service>/<type>/<slug>.md`.
4. **Write frontmatter** — at minimum `title` and `summary`.
5. **Write the body.** Standard Markdown. Keep each article focused on one
   thing; split rather than combine.
6. **Commit and push** to the help repository's default branch.
   fusion-content picks it up on the next poll.

### Minimal example

`help/forge/how-to/trigger-a-rebuild.md`

```markdown
---
title: "Trigger a rebuild"
summary: "Force fusion-forge to rebuild an existing artifact version."
tags:
  - build
  - forge
routes:
  - /forge/builds
---

## Prerequisites

- You need the artifact name and version you want to rebuild.
- Your account must have the `forge:build:write` permission.

## Steps

1. Navigate to **Builds** in the sidebar.
2. Find the artifact row and click **Rebuild**.
3. Confirm in the dialog — the new build job appears at the top of the list.

Alternatively, call the API directly:

```http
POST /api/v1/builds
Authorization: Bearer <token>
Content-Type: application/json

{
  "artifact": "my-service",
  "version": "1.2.0"
}
```

The response includes a `jobId` you can poll for status.
```

---

## Writing a series

A series is a set of articles that form a narrative sequence — typically
tutorials. fusion-content has no native "series" object; you model a series
through tags and slug naming conventions.

### Naming convention

Prefix slugs with a shared series name and a step number:

```
help/forge/tutorial/
  build-basics-01-setup.md
  build-basics-02-first-build.md
  build-basics-03-custom-base.md
  build-basics-04-artifacts.md
```

### Linking them together

Use a shared tag (e.g. `series:build-basics`) and cross-link at the bottom of
each article:

```markdown
---
title: "Build basics — 2. Your first build"
summary: "Submit a build request and inspect the resulting artifact."
tags:
  - build
  - series:build-basics
routes:
  - /forge/builds/new
---

<!-- body -->

---

**Series: Build basics**
← [1. Setup](build-basics-01-setup) · **2. Your first build** · [3. Custom base →](build-basics-03-custom-base)
```

The cross-links use slug references only. The frontend resolves them against
`/api/v1/help/<service>/tutorial/<slug>`.

### Index article (optional)

Create a `<series>-00-overview.md` as an entry point. Tag it identically, set
`routes` to the page a user lands on when they start the series, and list all
parts in the body.

---

## Verifying via the API

After pushing, wait one poll interval (or check the fusion-content logs), then
query the API. Port-forward locally if needed:

```sh
kubectl port-forward -n fusion service/fusion-content 18082:8080
```

### List all articles for a service

```sh
curl http://localhost:18082/api/v1/help?service=forge
```

### Filter by Diátaxis type

```sh
curl http://localhost:18082/api/v1/help?service=forge&type=tutorial
```

### Filter by tag

```sh
curl http://localhost:18082/api/v1/help?tag=series:build-basics
```

### Filter by frontend route

```sh
curl "http://localhost:18082/api/v1/help?route=/forge/builds/new"
```

### Full-text search (AND semantics)

```sh
curl "http://localhost:18082/api/v1/help?q=rebuild+artifact"
```

### Fetch a single article (with body)

```sh
curl http://localhost:18082/api/v1/help/forge/how-to/trigger-a-rebuild
```

### Pagination

```sh
curl "http://localhost:18082/api/v1/help?service=forge&page=2&pageSize=10"
```

Default page size is 20; maximum is 100.

### Response shapes

**List** (`GET /api/v1/help`):

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

**Single article** (`GET /api/v1/help/:service/:type/:slug`):

```json
{
  "service": "forge",
  "type": "how-to",
  "slug": "trigger-a-rebuild",
  "title": "Trigger a rebuild",
  "tags": ["build", "forge"],
  "routes": ["/forge/builds"],
  "summary": "Force fusion-forge to rebuild an existing artifact version.",
  "body": "## Prerequisites\n\n..."
}
```

---

## Common mistakes

| Mistake                                   | Effect                                      | Fix                                             |
|-------------------------------------------|---------------------------------------------|-------------------------------------------------|
| Wrong directory name (e.g. `howto`)       | Article rejected — unknown Diátaxis type    | Use exact names: `how-to`, `tutorial`, etc.     |
| Missing frontmatter fences                | Parse error — article skipped               | Open with `---\n` and close with `\n---\n`      |
| Frontmatter without `title`               | Title is empty string in API responses      | Always include `title:`                         |
| Adding `type:` to frontmatter             | Silently ignored — type comes from the path | Remove the field                                |
| Path depth > 3 segments                   | File rejected — path must be `<service>/<type>/<slug>.md` | Keep articles at exactly that depth   |
| Slug with spaces or uppercase             | Creates an ugly URL; avoid                  | Use lowercase hyphens only                      |
| Committing to a non-default branch        | fusion-content only polls the default branch | Merge to main/master before expecting it live  |

---

## Configuration reference

The help repository is configured in the repos Secret under the `help:` key:

```yaml
help:
  url: https://github.com/your-org/fusion-docs   # required
  token: ""                                       # omit for public repos
  dir: "help"                                     # subdirectory; default "help"
```

To disable the feature entirely, leave `url` empty or omit the `help:` key.
The changelog poller continues to run regardless.
