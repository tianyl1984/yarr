# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

`yarr` is a self-hosted RSS/Atom/RDF feed reader. The backend is a Go JSON API server
(`/api/*`); the frontend is a separate Vue 3 + Vite single-page app that talks to that
API. This is `tianyl1984/yarr`, a personal fork of `nkanaev/yarr` — it has diverged from
upstream in notable ways (see "Fork-specific differences" below), so don't assume
upstream docs, Makefile targets, the sqlite storage backend, or the old embedded
server-rendered frontend still apply.

## Repo layout

```
backend/   Go module root (go.mod/go.sum/vendor, cmd/, src/) — the JSON API server
frontend/  standalone Vue 3 + Vite SPA (package.json, vite.config.js, src/) — see Frontend below
scripts/   helper shell scripts (deploy-docker.sh, start-local.sh)
deploy/    deployment config: Dockerfile.backend, Dockerfile.frontend, nginx.conf, docker-compose.yml, .env.example
```

`backend/` is the Go module root — run `go build`/`go run`/`go vet` etc. from inside
`backend/`, not the repo root.

## Commands

There is no Makefile in this fork (the CI workflows under `.github/workflows/` still
reference `make test` / `make darwin_arm64_gui` etc. from the upstream template but
these targets no longer exist here — treat those workflows as stale/inoperative rather
than authoritative).

Backend (from `backend/`):

- Build the server binary: `go build -o out/yarr ./cmd/yarr`
- Run directly: `go run ./cmd/yarr -db "$YARR_DB"` (requires a reachable MySQL instance; see Configuration). Serves the JSON API on `YARR_ADDR` (default `127.0.0.1:7070`); it no longer serves any frontend assets.
- There are no `*_test.go` files in this repo currently — `go test ./...` will report "no test files" everywhere.
- Small standalone debugging utilities live under `backend/cmd/` and can be run directly, e.g. `go run ./cmd/feed2json <url|filepath>`, `go run ./cmd/readability <url>`, `go run ./cmd/htmlfeed` (hardcoded scraping test).

Frontend (from `frontend/`):

- Install deps: `npm ci` (or `npm install`).
- Dev server with hot reload: `npm run dev` — serves on `:5173` and proxies `/api` to `http://127.0.0.1:7070` (see `vite.config.js`), so run the backend alongside it.
- Production build: `npm run build` (outputs to `frontend/dist/`); `npm run preview` serves the built output.

Docker / other:

- `./scripts/deploy-docker.sh` (runnable from anywhere; it `cd`s to the repo root) runs `docker compose -f deploy/docker-compose.yml up --build -d` — builds two services from the repo root as build context: `yarr-backend` (`deploy/Dockerfile.backend`, `COPY backend/ .`) and `yarr-frontend` (`deploy/Dockerfile.frontend`, builds the Vite app and serves `dist/` via nginx, proxying `/api/` to the backend per `deploy/nginx.conf`). Only the frontend publishes a port (`7070:80`). Reads env from `deploy/.env` (copy `deploy/.env.example` there first).
- `./scripts/start-local.sh` — local (non-Docker) startup helper.

## Configuration

Runtime config is via CLI flags or environment variables (see `flag` definitions in
`backend/cmd/yarr/main.go`); env vars are the defaults for the flags. Key ones (see
`deploy/.env.example`):

- `YARR_DB` — MySQL DSN, e.g. `user:pass@tcp(127.0.0.1:3306)/yarr?charset=utf8mb4&parseTime=true&multiStatements=true`. Required; the process fatals on startup if unset.
- `YARR_ADDR` — listen address (default `127.0.0.1:7070`); `unix:<path>` is supported for a Unix socket.
- `YARR_BASE` — base path if served behind a subpath.
- `YARR_PROXY` — outbound proxy for feed fetching.
- `YARR_BROWSERLESS` — browserless endpoint (used for JS-rendered page scraping).
- `YARR_AUTH_URL` / `YARR_AUTH_SECRET` — enable cf-worker-auth SSO login (see Auth below); if `YARR_AUTH_URL` is unset, the whole app is unauthenticated.
- `YARR_CERTFILE` / `YARR_KEYFILE` — enable HTTPS.

## Architecture

All paths below are relative to `backend/`.

### Storage: MySQL, not sqlite

Unlike upstream yarr (which uses sqlite), this fork's `src/storage` package targets
MySQL only (`github.com/go-sql-driver/mysql`). `storage.New(dsn)` opens the DB and runs
`migrate()`. Migrations are plain SQL files embedded via `//go:embed` in
`src/storage/migration.go` and applied sequentially, tracked by a `db_version` row in
the `settings` table (see `src/storage/sql/*.sql`). To add a migration: add a new
`mNN_*.sql` file, embed it, add a corresponding `mNN_*` function, and append it to the
`migrations` slice in that exact order — order is significant since it's positional
(`migrations[v-1]`).

### Request flow

`cmd/yarr/main.go` parses flags/env → builds a `storage.Storage` → builds a
`server.Server` (`src/server/server.go`) → `srv.Start()`. `Server.handler()` in
`src/server/routes.go` wires up a small custom router (`src/server/router`, regex-based,
supports `:param` and `*wildcard` path segments and middleware chains via `c.Next()`).
All routes and handlers live in one file, `src/server/routes.go`; request bodies are
decoded via the form structs in `src/server/forms.go`.

Background feed refreshing is owned by `src/worker/worker.go`: a fixed pool of 4
goroutines pulls feeds off a channel, fetches/parses them (`src/worker/crawler.go`), and
writes new items back through `storage`. `Server.Start()` kicks off an initial favicon
sweep and, if `refresh_rate` (a DB setting) is nonzero, starts a ticker-driven
auto-refresh loop (`worker.SetRefreshRate`).

### Feed parsing pipeline

- `src/parser` — RSS/Atom/RDF/JSON feed parsing (`rss.go`, `atom.go`, `rdf.go`, `json.go`) unified into common `models.go` types.
- `src/htmlfeed` — a **custom-config-driven scraper** for sites without a real feed: it reads a per-URL scraping config from the `feed_config` DB table (`storage.GetFeedConfig`) describing CSS selectors for item root/title/link/date/description, scrapes the page with `goquery`, and synthesizes an RSS XML document from the result. See `custom_config.md` for the config JSON schema and selector syntax (tag / `.class` / `:eq(n)`), and the `/htmlFeed?url=` endpoint for testing a config against a live page.
- `src/content/readability` — Mozilla-Readability-style article extraction, used by the `/page` reader-mode endpoint.
- `src/content/sanitizer` — HTML sanitization (whitelist-based) applied to all item content and readability output before it's returned to the client, to prevent stored/reflected XSS from feed content.
- `src/content/silo` — special-cased embeds (e.g. resolving video iframe embeds, redirect URL rewriting) for known hosting platforms.
- `src/httpclient` — shared HTTP client construction, including a separate proxy-aware client (`NewProxyClient`) used when a feed has `UseProxy` set or `YARR_PROXY` is configured.

### Auth (cf-worker-auth SSO)

Optional, enabled only when `YARR_AUTH_URL` is set (`src/server/auth`). This is a
lightweight session layer that defers actual login to an external `cf-worker-auth`
service:
- `auth.Middleware` wraps all routes except `Public` paths (`/static`, `/fever`,
  `/manifest.json`, `/auth/callback`); browser GETs without a valid session are
  redirected to `<AuthURL>/login?callback=...`, API calls get a bare 401.
- `/auth/callback` (`handleAuthCallback` in `routes.go`) exchanges the SSO one-time
  `token` for user info (`auth.FetchUserInfo`) and sets a signed session cookie
  (`auth.SetSession`) — HMAC-SHA256 over the username, keyed by `YARR_AUTH_SECRET`
  (`username:hmac`), verified with constant-time comparison. There's no local user
  table; identity is entirely delegated to the SSO service's `login` field.
- If `YARR_AUTH_SECRET` is unset when auth is enabled, a random secret is generated at
  startup (logged as a warning) — sessions won't survive a restart in that case.

### Frontend

Lives in the top-level `frontend/` directory as a standalone Vue 3 + Vite SPA
(`package.json` name `yarr-frontend`) — fully split out from the backend. Entry is
`frontend/index.html` → `src/main.js` → `src/App.vue`; source is organized under
`src/` (`api/`, `components/`, `state/`, `directives/`, `icons/`, `styles/`,
`keybindings.js`). It talks to the backend purely over the `/api/*` JSON API. The `@`
import alias resolves to `frontend/src`.

The old embedded server-rendered frontend (`backend/src/assets`: a Go `html/template`
`index.html` plus a Vue 2 SPA served from an embedded FS, with a `-tags debug` DirFS
fallback) has been **removed** — the backend no longer serves any HTML/JS/CSS.

In production the built `dist/` is served by nginx (`deploy/Dockerfile.frontend` +
`deploy/nginx.conf`), which also reverse-proxies `/api/` to the backend container.

## Fork-specific differences from upstream nkanaev/yarr

Worth knowing if consulting upstream docs/issues, since they'll describe behavior that
doesn't apply here:
- Storage backend changed from sqlite to MySQL (`YARR_DB` is a DSN, not a file path).
- Added: `src/htmlfeed` custom-scraper pipeline and the `feed_config` table (`custom_config.md`).
- Added: `/opml/compare` endpoint to diff an uploaded OPML against existing subscriptions.
- Added: optional cf-worker-auth SSO login layer (`src/server/auth`), off by default.
- Removed: the upstream Makefile/multi-platform build tooling isn't present in this repo (see Commands above) even though the GitHub Actions workflows referencing it are still checked in.
- Restructured into `backend/`, `frontend/`, `scripts/`, `deploy/` top-level directories (this repo previously had `cmd/`, `src/`, `go.mod` etc. directly at the repo root — the Go module root is now `backend/`).
- Frontend split out of the backend: it's now a standalone Vue 3 + Vite app under `frontend/`, and the old embedded server-rendered assets (`backend/src/assets`, Vue 2) have been removed — the backend is a pure `/api/*` JSON server. Deployed as two containers (`yarr-backend` + nginx-served `yarr-frontend`).
