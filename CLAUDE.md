# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

`yarr` is a self-hosted RSS/Atom/RDF feed reader written in Go, with a server-rendered
index page plus a vanilla-JS/Vue frontend served from embedded assets. This is
`tianyl1984/yarr`, a personal fork of `nkanaev/yarr` — it has diverged from upstream in
notable ways (see "Fork-specific differences" below), so don't assume upstream docs,
Makefile targets, or the sqlite storage backend still apply.

## Repo layout

```
backend/   Go module root (go.mod/go.sum/vendor, cmd/, src/) — all server + frontend-asset code
frontend/  placeholder for a standalone frontend once it's split out (currently empty — see Frontend below)
scripts/   helper shell scripts (e.g. scripts/deploy-docker.sh)
deploy/    deployment config: Dockerfile, docker-compose.yml, .env.example
```

`backend/` is the Go module root — run `go build`/`go run`/`go vet` etc. from inside
`backend/`, not the repo root.

## Commands

There is no Makefile in this fork (the CI workflows under `.github/workflows/` still
reference `make test` / `make darwin_arm64_gui` etc. from the upstream template but
these targets no longer exist here — treat those workflows as stale/inoperative rather
than authoritative).

- Build the server binary (from `backend/`): `go build -o out/yarr ./cmd/yarr`
- Run directly (from `backend/`): `go run ./cmd/yarr -db "$YARR_DB"` (requires a reachable MySQL instance; see Configuration)
- Serve assets from disk instead of the embedded FS (for frontend iteration without rebuilding, from `backend/`): `go build -tags debug -o out/yarr ./cmd/yarr` — this excludes `src/assets/assetsfs.go` (guarded by `//go:build !debug`) so `assets.FS.Open` falls back to `os.DirFS("src/assets")`, meaning JS/CSS/HTML edits are picked up on refresh, no recompile needed. Note this `DirFS` path is relative to the process's working directory, so run the binary from `backend/` too.
- There are no `*_test.go` files in this repo currently — `go test ./...` (from `backend/`) will report "no test files" everywhere.
- Small standalone debugging utilities live under `backend/cmd/` and can be run directly, e.g. `go run ./cmd/feed2json <url|filepath>`, `go run ./cmd/readability <url>`, `go run ./cmd/htmlfeed` (hardcoded scraping test).
- Docker: `./scripts/deploy-docker.sh` (runnable from anywhere; it `cd`s to the repo root) runs `docker compose -f deploy/docker-compose.yml up --build -d` — builds `deploy/Dockerfile` with the repo root as build context (so it can `COPY backend/ .`), reads env from `deploy/.env` (copy `deploy/.env.example` there first).

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

Currently still lives under `backend/src/assets` (not yet split into `frontend/` — see
Repo layout above; that's a planned future refactor, not done). Server-rendered
`src/assets/index.html` (Go `html/template`, delimiters `{% %}`) bootstrapped with
initial settings, plus a Vue 2 (`vue.min.js`) SPA in `src/assets/javascripts/app.js`
talking to the JSON API via `api.js`. No JS build step / bundler / package.json — these
are plain scripts served as-is (embedded into the binary at build time via
`src/assets/assetsfs.go`, unless built with `-tags debug`).

## Fork-specific differences from upstream nkanaev/yarr

Worth knowing if consulting upstream docs/issues, since they'll describe behavior that
doesn't apply here:
- Storage backend changed from sqlite to MySQL (`YARR_DB` is a DSN, not a file path).
- Added: `src/htmlfeed` custom-scraper pipeline and the `feed_config` table (`custom_config.md`).
- Added: `/opml/compare` endpoint to diff an uploaded OPML against existing subscriptions.
- Added: optional cf-worker-auth SSO login layer (`src/server/auth`), off by default.
- Removed: the upstream Makefile/multi-platform build tooling isn't present in this repo (see Commands above) even though the GitHub Actions workflows referencing it are still checked in.
- Restructured into `backend/`, `frontend/` (placeholder), `scripts/`, `deploy/` top-level directories (this repo previously had `cmd/`, `src/`, `go.mod` etc. directly at the repo root — the Go module root is now `backend/`).
