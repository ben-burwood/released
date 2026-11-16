# Dashboard - Releases

Dashboard listing the GitHub Release History of a Repository, can be used as a public changelog for private repositories.

Go backend serving a Vue (Vite + Tailwind) frontend, deployed as a single Docker image.

## Environment

1. `GITHUB_TOKEN` — read-only PAT with access to the target repo.
2. `GITHUB_REPO` — repo to list releases from, e.g. `org/repo`.

Both are required; the server exits on startup if either is missing.

Releases are cached for 60 seconds. Release notes are rendered from Markdown server-side, with GitHub's `Full Changelog` trailer removed and
private `github.com/user-attachments` image links resolved to signed URLs so they load without a GitHub session.

## API

- `GET /api/releases` — JSON of every release of the repo, drafts first then newest first

## Development

```sh
# Backend (http://localhost:8080, serves ./frontend/dist)
GITHUB_TOKEN=... GITHUB_REPO=streetwave/web-commercial go run .

# Frontend
cd frontend && npm install && npm run build   # or `npm run dev` (proxies /api to :8080)
```

## Docker

```sh
docker compose up --build
```
