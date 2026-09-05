# Docker deployment

The runtime is one Go application (HTTP server plus background collectors) and one ArangoDB instance. Node builds Svelte in a separate Docker stage; no Node/Vite server runs in production. Go serves the built frontend, SEO HTML, JSON-LD, API, robots.txt, and sitemaps from one origin.

## Repository layout

Frontend and backend source live in the same repository:

```text
frontend/          # Svelte source + package-lock.json, tracked with the backend
public/            # generated frontend output for non-Docker development
api/, delivery/, handler/, model/, repository/, worker/
deploy/            # Caddy configuration and deployment documentation
Dockerfile
compose.production.yml
```

A single checkout contains everything needed to build the app image. Docker builds `frontend/` and copies its output into the Go runtime image. For development without Docker, run `npm ci` and `npm run deploy` in `frontend/` to populate `public/`; downloaded league and team images are preserved. Generated files in `public/` are ignored by Git.

## Start locally

From the backend repository root:

```sh
cp .env.example .env
# Edit .env: set ARANGO_ROOT_PASSWORD to a strong password.
docker compose -f compose.production.yml up -d --build --wait
```

Visit http://localhost:1323. `SITE_URL` must match the URL visitors use because it controls canonical URLs and sitemaps. When changing APP_PORT, update SITE_URL too.

This production Compose file is separate from the existing database-only `docker-compose.yml`. Do not run both against the same host port; the production database has no published port. The application binds to loopback on the host. Existing local database contents are not automatically imported.

## Public HTTPS

Point the domain's DNS at the server. Set these values in `.env`:

```dotenv
SITE_DOMAIN=dota-leagues.example.com
SITE_URL=https://dota-leagues.example.com
GA_MEASUREMENT_ID=
```

Then start the HTTPS profile:

```sh
docker compose -f compose.production.yml --profile https up -d --build --wait
```

Caddy exposes ports 80/443 and obtains/manages certificates. The domain must resolve correctly and those ports must reach this server. If an existing reverse proxy already handles HTTPS, omit the profile and proxy to `127.0.0.1:1323` instead. Keep SITE_URL set to the public HTTPS origin.

## Persistent data and updates

- `database`: ArangoDB records.
- `database_apps`: ArangoDB application files.
- `logos`: downloaded league/team images under `/data`.
- `caddy_data` and `caddy_config`: certificate and proxy state.

Built frontend files stay inside the app image, **not** in a volume. Rebuilding therefore replaces HTML/CSS/JS together without stale bundles surviving in a data volume. `ASSET_DIR` changes where the app stores and serves downloaded logos; without it, non-Docker development still uses `public/`.

```sh
docker compose -f compose.production.yml up -d --build --wait
docker compose -f compose.production.yml logs --tail=100 -f app
docker compose -f compose.production.yml ps
```

Include `--profile https` when managing Caddy too. Normal `down` preserves data volumes. `down -v` deletes them and is only appropriate for disposable stacks. Back up the database before upgrading ArangoDB, and keep the original database password when reusing an initialized volume. Changing the environment variable does not rotate an existing database password.

The app image runs as UID 10001 with a read-only filesystem and a writable logo volume. Health checks wait for authenticated ArangoDB readiness before starting the app. The app's `/healthz` reports HTTP-process liveness; it is not an assertion that Valve or every upstream endpoint is available. Initial collection is asynchronous, so a new database may show empty pages while it fills.

## Database indexes

The application ensures missing persistent indexes before starting collectors and HTTP. Existing indexes are reused; this step never deletes records or rebuilds the updates feed. The normal application startup still rebuilds derived feed summaries as before.

After configuring `.env` as described above, create indexes separately before starting the app, or apply them to an existing database:

```sh
docker compose -f compose.production.yml up -d --wait arangodb
docker compose -f compose.production.yml build app
docker compose -f compose.production.yml run --rm --no-deps app -production -create-indexes
```

The maintenance container inherits the app's database credentials and network, needs no published database port, and exits when indexes are ready. It starts neither HTTP nor collectors. A failed operation exits nonzero; rerun after resolving the error. For a large existing database, run this command before bringing up the app so index creation does not delay its health check.

The definitions in [`db/indexes.go`](../db/indexes.go) cover eight indexes (field order matters):

| Collection | Indexed fields |
| --- | --- |
| `league_series` | `league_id, start_time` |
| `league_series` | `start_time` |
| `players` | `steam_next_refresh_at, account_id` |
| `updates` | `created_at, id` |
| `updates` | `entity, entity_id, group_kind, created_at, id` |
| `update_groups` | `created_at, id` |
| `update_groups` | `entity, created_at, id` |
| `source_snapshots` | `kind, entity_id, first_seen` |

All eight application indexes are non-unique and non-sparse. ArangoDB's built-in primary indexes are retained. Background index creation reduces write-lock duration but the command still waits for completion ([ArangoDB indexing documentation](https://docs.arango.ai/arangodb/stable/indexes-and-search/indexing/basics/)).

Successful logs contain `database index ready` with the collection, fields, and `created=true` for a new index or `created=false` for an existing one. A repeat run should reuse all eight indexes. The command creates missing indexed collections, but does not initialize unrelated collections or populate application data. Continue with the normal `up -d --build --wait` command after maintenance (include `--profile https` when using Caddy).

Outside Docker, run `go run . -create-indexes` (development configuration) or `go run . -production -create-indexes` from the repository root. Both require `config.json` and accept the same `DOTA_DEVELOPMENT_*` or `DOTA_PRODUCTION_*` environment overrides as the server. Compose loads `.env` for containers; a direct `go run` does not load that file automatically.

The Python script is an optional [query-plan inspection tool](../README.md#inspect-query-plans), not a requirement for deployment. It needs a source checkout, Python, and a reachable database configured in `config.json`; it is not included in the runtime image.

## Runtime configuration

Compose passes environment overrides to Viper as `DOTA_PRODUCTION_*`, for example `DOTA_PRODUCTION_DATABASE_URL` and `DOTA_PRODUCTION_VALVE_RPS`. `.env` and the existing `config.json` are excluded from Docker builds. The image contains only `config.json.example` defaults; credentials are supplied at runtime. `GA_MEASUREMENT_ID` is optional. Restart/recreate the app after configuration changes.

Before release, verify `/`, `/robots.txt`, `/sitemap.xml`, a real team/league/match page, and the canonical origin. Submit the sitemap and inspect representative URLs in Search Console after deployment.
