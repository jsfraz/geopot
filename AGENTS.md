# AGENTS.md

Honeypot: an SSH server that always denies auth, records the attempted credentials, geolocates the source IP, and streams results to a live dashboard. Two codebases in one repo:

- `geopot/` — Go backend (module `jsfraz/geopot`): SSH listener on `:2222`, HTTP API + WebSocket on `:8080`, and it serves the built frontend from `./static`.
- `frontend/` — Vue 3 + Vite + Tailwind v4 dashboard, client-rendered only.

## Commands

Backend, from `geopot/`:

```bash
go test ./...     # only tests in the repo; no DB or services needed
go build ./...    # compile check
```

`go.mod`'s `go` directive is a hard floor. It is currently `1.27.3`; with `GOTOOLCHAIN=auto` an older local Go silently downloads that toolchain, but the `Dockerfile` builder image must be at least that version — official `golang` images set `GOTOOLCHAIN=local`, so a stale pin fails `docker build`. Keep the pin exact and in sync with `go.mod`.

Frontend, from `frontend/` (Node `^20.19.0 || >=22.12.0`):

```bash
npm install
npm run dev          # Vite on :5173
npm run type-check   # vue-tsc --build
npm run build        # type-check + vite build (npm-run-all) -> dist/
```

There is no linter, formatter, or frontend test suite. Don't assume ESLint/Prettier/Vitest exist or add them unasked.

Verification gate before declaring work done: `go test ./...` + `go build ./...` for backend changes, `npm run type-check` for frontend changes.

## Local development

1. `sudo docker compose -f docker-compose.dev.yml --env-file .env.dev up -d` starts only TimescaleDB and Valkey; it publishes no ports.
2. Start the backend yourself (the VS Code `go: debug` launch config loads `.env.dev` with `GIN_MODE=debug`).
3. `npm run dev` in `frontend/`.

Gotchas:

- `.env.dev` hardcodes the containers' bridge IPs (`172.20.0.2` / `.3`). They change whenever the compose network is recreated; re-resolve with `docker inspect` when the backend cannot connect.
- Postgres **must** be TimescaleDB. Startup runs GORM `AutoMigrate`, then the embedded `database/sql/migration.sql` (hypertable, compression, continuous aggregates) and refreshes every aggregate, which can take minutes on large data. Plain Postgres fails.
- The Vite dev server does not proxy. `App.vue` hardcodes API + WebSocket at `http://localhost:8080` when `import.meta.env.DEV`, and the backend accepts `localhost:*` origins only in debug mode — so the Go process must run with `GIN_MODE=debug` for `:5173` to work.
- `GIN_MODE` defaults to `release`. Debug also enables `/swagger`, `/api/openapi.json`, gin's request logger and SQL logging, all of which expose captured credentials. Use it locally only.
- The UI served by Go on `:8080` comes from the gitignored `geopot/static/`, which is a stale snapshot locally (the Docker image copies `frontend/dist` there at build time). For UI work use Vite, not `:8080`.

## Data flow

In `main.go`, loopback/private source IPs are written straight to Postgres (in debug they are replaced by a random public IP and coordinates so `./test.sh` produces map data). Public IPs are pushed to a Valkey list; a background worker pops one roughly every 1.05s, geolocates it via freeipapi.com, stores it, and broadcasts it to WebSocket clients. Public-IP records therefore appear with delay and depend on that external API.

Stats endpoints are under `/api/stats/*`; live updates are JSON `ModelsConnection` frames on `/ws`. Per-IP rate limit is 120/min for `/api` and `/ws`.

## Generated code

`frontend/src/api/` is OpenAPI-generated (typescript-rxjs, generator 7.15.0) and committed. Never hand-edit it. Regenerate with `npm run openapi` from `frontend/`; it needs the backend running on `localhost:8080` in debug mode (the spec exists only there) plus `java` and `wget`, and it overwrites the whole directory.

Components take a `Configuration` prop and subscribe to `StatsApi` rxjs observables; `App.vue` owns the base-path and WebSocket URL logic. Tailwind v4 is configured through `@theme` tokens in `src/assets/main.css` — there is no `tailwind.config.js`.

## Conventions

- Work on `development`; `master` is release. `.github/workflows/docker-image.yml` builds and pushes `ghcr.io/jsfraz/geopot:latest` on every push to `master`. Don't commit or push unless asked.
- Go style: doc comment above every function with `@param`/`@return`, stdlib `log`, config via `envconfig` in `utils/config.go`, shared DB/Valkey/WebSocket handles through `utils.GetSingleton()`.
- Security-sensitive: never add real captured credentials, database dumps, or `.env*` values to the repo, tests, or fixtures.
