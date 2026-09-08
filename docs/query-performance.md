# Query performance

Measured on local ArangoDB 3.12.4-3 using a disposable copy of application data:
111,117 series, 4,614 teams, 4,553 events, and 4,552 update groups. Each value is
the median server query execution time across five runs, with the query result
cache disabled and exact totals enabled. Before/after rows and totals matched.
These are query timings, not end-to-end HTTP latency or a load test.

| Query | Before | After |
| --- | ---: | ---: |
| Teams, default ranking | 288.3 ms | 18.9 ms |
| Teams, active filter | 290.0 ms | 27.0 ms |
| Updates, default page | 8.85 ms | 2.51 ms |
| Updates, roster + last day | 8.27 ms | 1.53 ms |
| Updates, text search for “team” | 14.12 ms | 13.01 ms |

## Teams

The historical aggregation is cached for 30 seconds per repository instance.
Refreshing it took 294.8 ms in this measurement. The current implementation warms
the cache at startup and serves the previous snapshot while an expired cache
refreshes in the background. Concurrent requests share the refresh; failures retain
the last successful snapshot and retry after 30 seconds. A cold request still waits
if startup warming failed. Cache hits avoid reading 111,117 series, while live
games, team profiles, and roster changes remain fresh. Competition history can
remain stale during refreshes or upstream database failures. Name/wins sorting
without an activity filter skips the cache.

The warm page plan scans games and teams and sorts teams; it contains no series
scan or aggregation. The activity map is passed as a bind variable: measured
peak query memory rose from about 3.8 MB to 9.6 MB for the default page. If that
becomes significant at higher concurrency, persistent per-team activity summaries
are the next option. Those require handling series replacement, league-tier
corrections, future timestamps and the rolling 90-day tier window.

## Updates

`update_groups` stores the latest observation, original document keys, historical
search names, and entity IDs for each group. Event and summary writes share a
transaction. Startup atomically rebuilds summaries under exclusive write locks,
including legacy singleton events. This shifts grouping work from each read to
writes and startup. Original observations remain intact.

The default page uses `[created_at, id]`; entity-filtered pages use
`[entity, created_at, id]`. `EXPLAIN` shows pagination before source-history
loading, with no collection scan or `COLLECT` in the feed query. Sorting within
selected groups preserves observation order and timestamp tie-breaking.

Default-feed peak query memory fell from about 5.5 MB to 98 KB. Exact totals
still scan matching group index entries; substring search still examines group
search metadata. Large groups also cost more to update and return. The small
search improvement is expected, and search semantics remain unchanged.

A preliminary indexed per-event latest-group lookup took about 24 ms versus
11 ms for the former feed query, so that approach was rejected.

## Index setup

The current application ensures all configured indexes at startup. To prepare them without running collectors or rebuilding feed summaries, use `go run . -create-indexes` from the repository root. This operation is repeatable and preserves existing records. The complete list and production container commands are in [database index maintenance](../deploy/README.md#database-indexes).

The Python script below inspects plans; it is not required to initialize indexes. Its `--explain-only` mode makes no index changes. Without that flag, it ensures only its targeted subset of indexes. After application startup, before/after plans may match because indexes already exist. It reads `config.json` directly and does not inherit `DOTA_*` environment overrides.

## Verification

Integration tests compare the new feed against the former query for legacy
entries, date boundaries, historical names, timestamp ties, merged reversals,
pagination and totals. They check repeatable backfills and rollback of event
inserts when a summary write fails. Cache tests cover expiration, failed refreshes,
concurrent misses and cancellation. Existing team tests verify rankings, filters,
profile edits and immediate live-game changes. The full `make quality` suite
passed with `ARANGO_TEST_URL=http://127.0.0.1:8529`.

After starting the current application, inspect the current plans with:

```bash
python3 scripts/arango_indexes.py --explain-only
```

The script explains the historical cache refresh separately from warm page
queries. Warm-page explanations use an empty activity-map placeholder; runtime
measurements above used the complete activity map from the same snapshot.
