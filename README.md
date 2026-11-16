# Dashboard - Releases

Dashboard listing the GitHub Release History of Repositories, can be used as a public changelog for private repositories.

Go backend serving a Vue (Vite + Tailwind) frontend, deployed as a single Docker image.

## Environment

1. `GITHUB_TOKEN` — read-only PAT with access to the target repos.
2. Repo Configuration (either/or will combine and dedupe):
    - `GITHUB_REPOS` — comma or space separated repos, e.g. `org/repo,org/other` (or just `org/repo`).
    - `GITHUB_REPOS_FILE` — path to a YAML config file listing repos under `repos:`.

Repos can be given as `owner/name` or as a GitHub URL, e.g. `https://github.com/ben-burwood/certui`. The first repo is the default.

The server exits on startup if the token or repos are missing, or a repo is invalid.

```yaml
# config.yml
repos:
  - ben-burwood/released
  - https://github.com/ben-burwood/certui
```

With Docker, mount the file and point `GITHUB_REPOS_FILE` at it.

The selected repo is kept in the `?repo=org/repo` query param, so links open on that repo.

Releases are cached for 60 seconds. Release notes are rendered from Markdown server-side, with GitHub's `Full Changelog` trailer removed.

## API

- `GET /api/repos` — JSON list of the configured repos, the default first
- `GET /api/releases?repo=org/repo` — JSON of every release of the repo (default if omitted), drafts first then newest first; `404` for repos that aren't configured
- `GET /api/assets/{id}` — redirect to the signed URL of a private release-note attachment

## Development

```sh
# Backend (http://localhost:8080, serves ./frontend/dist)
GITHUB_TOKEN=... GITHUB_REPOS=org/repo go run .

# Frontend
cd frontend && npm install && npm run build   # or `npm run dev` (proxies /api to :8080)
```

## Docker

```sh
docker compose up --build
```
