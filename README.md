## Development

Use Go 1.27.1. With Go 1.21 or newer and `GOTOOLCHAIN=auto`, the Go command downloads the version required by `go.mod`.

Run code quality checks with `make quality`. To include the disposable ArangoDB integration test:

```bash
ARANGO_TEST_URL=http://127.0.0.1:8529 make quality
```

start ArangoDB with docker (no manual install needed):

```bash
docker compose up -d
```

create config file `config.json` (you can edit it to set arangodb credentials), or just copy an example

```bash
cp config.json.example config.json
```

auth is disabled on the test instance, so any credentials work — the example values (`login`/`pass`) are fine as-is. the database itself is created automatically on first run.

run project

```bash
go run .
```

navigate to http://localhost:1323

stop the db when you are done: `docker compose down`

## Database indexes

The Go application creates missing indexes automatically before starting HTTP and collectors. To prepare indexes separately, run from the repository root with `config.json` configured:

```sh
go run . -create-indexes
# Use production.database settings instead:
go run . -production -create-indexes
```

This command creates missing indexed collections, preserves records, reuses existing indexes, and exits without starting HTTP, collectors, or the updates-summary rebuild. Logs report each index as `created=true` or `created=false`; failures exit nonzero and can be retried after fixing the cause.

See [Docker index maintenance](deploy/README.md#database-indexes) for Compose commands, the complete index list, and guidance for existing databases. Index definitions live in [`db/indexes.go`](db/indexes.go).

### Inspect query plans

After starting the current application to initialize all collections and feed summaries, run:

```sh
python3 scripts/arango_indexes.py --explain-only
```

This optional tool requires Python 3 and a source checkout. It reads `development.database` directly from `config.json`; it does not use the Go application's `DOTA_*` overrides or Compose `.env`. Use `--environment production` to select production configuration, `--league-id ID` to inspect a particular league, or `--output PATH` to change the default `/tmp/arango-index-explain.json` report.

Without `--explain-only`, the script also ensures its targeted schedule, Steam-refresh, and update-group indexes, then saves before/after optimizer plans. It is not the complete index initializer and is not included in the production runtime image. Since startup now creates these indexes, before/after plans may be identical. The script never executes the explained queries, including the series-removal query; these reports are not runtime benchmarks.

The teams page caches competition history for 30 seconds, shared across filters
and pages. Live games, team profiles, and roster changes are read on every
request. Concurrent cache misses share one refresh; errors are not cached.
Name/wins sorting without an activity filter skips the history query entirely.

The updates feed reads one derived document per group from `update_groups`.
Event inserts and group-summary changes commit in the same transaction. Startup
rebuilds these summaries from the original events under exclusive write locks;
original events remain unchanged. Rebuild failures roll back and stop startup.
Run one application version at a time during this upgrade: older writers do not
maintain summaries. Date/entity filters and pagination use group indexes, while
source history is loaded only for the requested page. Historical names remain
searchable, and exact totals still require counting all matching groups.

See [query performance measurements](docs/query-performance.md) for the measured
query changes and remaining costs.

## Frontend updates log

Related new observations for the same entity are grouped before pagination:
roster, branding, and schedule changes use a 48-hour window; score-record and
prize-pool changes use one hour. Windows start at the first observation and do
not slide. Initial imports stay separate. Original events remain stored and
are expandable in the feed, including changes later reversed within a group.
Older events without grouping metadata remain separate.

Roster titles describe net active-membership changes: one out/one in is a
replacement, at least two out/two in is a rebuild, and a near-empty roster
adding at least four active members is a reformation. “Possible disband” means
at least four active members left, nobody joined, and at most one non-admin
remains. This is an inference from snapshots, not a confirmed announcement or
proof anyone was kicked. Times are observation times. Admin status comes from
Valve when available. Active teams are eligible for refresh every two hours, with daily checks for
inactive teams. A sweep runs every ten minutes, prioritizing teams with matches
in the last 90 days or next 30 days, live games, or roster changes in the last
30 days. Player discovery and current tournament refresh jobs run every two
hours. Actual freshness depends on queue length and upstream availability.
The shared Valve rate limit remains unchanged. The 48-hour grouping window
combines related observations without delaying their publication;
a later observation can revise the summary. Group assignment is serialized
within the application's shared repository instance.


Open `/activity` or choose **Updates** in the navigation. The page requests
`/updates` in batches of 20, groups entries by their recorded UTC date, and has
one set of newer/older page links below the feed. `/updates` remains a JSON API.
The HTML page includes initial feed content, canonical pagination URLs, and a
sitemap entry for indexing.

| Update | Presentation |
| --- | --- |
| Tournament added | Tournament name, directory link, and a compact “added” label. |
| Tournament updated | Before/after metadata, readable tier/region/status, UTC start/end dates, description and website changes. Prize-pool changes show USD amounts prominently. |
| Team added | Team name, profile link, and a compact “added” label. |
| Team updated | Before/after name, tag, country, region, website, logo URL, captain account ID, and win/loss counts. Record-only changes have a dedicated heading. |
| Roster added | Team link and a compact roster-added label; creation payloads contain no member snapshot. |
| Roster updated | Account IDs labeled Added, Removed, Activated, or Deactivated, preserving active/inactive status. Names are not included in the payload. |
| Player added | Player name, account ID, and recorded team name linking to the team page. |

Unknown future fields use a readable field label and before/after values; unknown
entities use a generic record entry. Entries use compact rows; changed fields and roster counts are summarized, with full before/after details collapsed until expanded.
Loading, retryable errors, empty history, and empty later pages have explicit states.

## Sorting and list filters

Teams: `GET /teams?country=UA&active=true&active_days=90&sort=name&order=asc`.
Country codes are case-insensitive. `active_days` defaults to 90 and accepts
1–3650 days. Active teams are playing live, have a played series in that window,
or have a recorded team/roster update. Importing a team or refreshing its metadata
timestamp does not count; recorded changes are available only from the start of
the updates feed. Future and unplayed series are excluded.

Team sorts: `recommended` (default), `name`, `wins`, `activity`.
Tournament sorts on `/leagues`: `recommended` (default), `name`, `start_date`,
`end_date`, `prize_pool`, `tier`. Explicit sorts support `order=asc` or `desc`;
names default to ascending, other fields to descending. Recommended sorting
keeps the existing fixed ranking. IDs break ties for consistent pagination.
Filters and sorting apply before pagination, and `meta.total` is the filtered
count. Existing search, tournament status, tier, region, and live filters still
combine with sorting.

The team and tournament pages expose these controls and preserve them in
pagination URLs. Changing a filter or sort returns to the first page.

## Updates API

`GET /updates?offset=0&limit=20` returns persisted activity, newest first. It uses
the same `meta` and `results` envelope as the other list endpoints. The default
limit is 10 and the maximum is 100.

```json
{
  "meta": { "limit": 20, "offset": 0, "total": 1 },
  "results": [
    {
      "id": "example-event-id",
      "entity": "team",
      "entity_id": 7,
      "name": "Team B",
      "action": "updated",
      "created_at": 1788609600000,
      "url": "/team/7",
      "changes": [{ "field": "name", "before": "Team A", "after": "Team B" }]
    }
  ]
}
```

`created_at` is a Unix timestamp in milliseconds. `entity` is `tournament`,
`team`, `roster`, or `player`; `action` is `created` or `updated`. Creation
entries have an empty `changes` array. `url` is omitted for players because
there is no player detail page. Roster changes contain account IDs and active
status in the `members` field.

The feed records tournament details and prize-pool changes, team metadata and
win/loss changes, roster membership changes, and newly stored players. It
ignores refresh timestamps, DB keys, and roster ordering. It contains public
data changes, not operational logs or refresh errors.

The `updates` collection and its chronological index are created on startup.
History starts when this feature is enabled; existing data is not backfilled.
Entries are written after successful source-data writes. A feed-write failure
is logged without failing the completed refresh, so this is a best-effort
activity feed rather than a transactional audit trail.

------

## Deployment

copy files to the server:

```bash
rsync  --cvs-exclude -av ./dota_league deploy@46.101.217.107:work
```

build executable

```bash
go build dota_league
```

run

```
./dota_league
```

or start as a service

```bash 
sudo cp dota-league.service /lib/systemd/system/dota-league.service

sudo systemctl start dota-league
sudo systemctl enable dota-league
```

----

### help tasks

get service logs:

```bash
journalctl -u dota-league.service
```

arango database dump:

```bash
arangodump --server.database dota_leagues

```

Player creation events include a `team` object (`id`, `name`) captured when the
player is stored. The frontend links to `/team/{id}`; if the team name is absent,
it shows `Team #{id}`. When no team ID is supplied, it shows “Team: Not recorded”.
The feed does not look up or backfill team context for old events.

## Docker application stack

See [Docker deployment](deploy/README.md) for the multi-stage Go/Svelte image, persistent ArangoDB and image storage, optional Caddy HTTPS profile, repository layout, and release commands. The existing `docker-compose.yml` remains the local database-only setup; the complete stack is in `compose.production.yml`.

## Frontend development

The Svelte source lives in [`frontend/`](frontend/README.md), tracked in this repository. With the Go API running on port 1323:

```sh
cd frontend
npm ci
npm run dev
```

Use the local URL printed by Vite. Run `npm run deploy` from `frontend/` to build and copy assets into `public/` for Go to serve. `public/` is generated output and downloaded images; it is not the frontend source directory.
