# Direct Valve data collection

## Implemented scope

All valid team and league responses fetched by normal workers are archived before domain decoding. There is no pilot cohort, 40-record limit, or separate collection schedule. Existing freshness guards and the shared Valve rate limiter apply.

`source_snapshots` retains changed JSON responses. Consecutive identical responses update their last observation time; A → B → A remains three versions. Object-key ordering is normalized for comparison, arrays keep source ordering, and numeric precision is preserved. Failed or mismatched responses do not replace the last valid snapshot. A domain-model schema mismatch does not prevent saving valid source JSON.

`source_status` tracks each record's last collection attempt, last success, failure category, version count, and audit/group coverage.

## Inspection

Team and tournament pages show a compact Data disclosure when a saved snapshot is available:

- raw JSON download;
- source collection timestamps;
- version selector when two or more versions exist (latest 100 listed);
- audit rows with resolved player names and original account IDs;
- nested tournament groups and explicit advancement links whose endpoints are present.

Empty reports and unavailable progression links stay hidden. Audit action codes remain untranslated until their meanings are verified. Resolved names are read-time enrichment and do not alter archived JSON.

The Updates page includes Collection coverage for all attempted records, with server-side name/ID search and 20-row pagination. Failures remain visible alongside successful collections. Legacy cohort markers are no longer used.

## API

- `GET /source-data/{team|league}/{id}`: status, recent version metadata and selected snapshot.
- `?version=<snapshot-key>`: select a historical snapshot.
- `&download=1`: download its raw JSON.
- `GET /source-data/coverage?offset=0&limit=20&search=`: paginated coverage with `results` and `meta.total`.

Read endpoints never trigger upstream requests.

## Verified access

On 2026-09-05, six unauthenticated direct Valve requests returned HTTP 200 with matching IDs. NAVI and Liquid each supplied five audit entries, Gaimin supplied none. TI 2019 and TI 2025 supplied bracket groups; the sampled smaller qualifier did not. Exact requests and aggregate results are in `valve-access-check-2026-09-05.json`.

An initial bounded collection saved 40 valid records. This was a bootstrap, not an ongoing restriction. Normal refreshes expand the archive automatically.

## Next investigations

Use saved versions to investigate audit-action semantics, bracket relationships and coverage gaps. Retain raw codes where evidence is insufficient. Do not infer confirmed transfers, kicks or universal bracket rules from sparse examples.

### Steam identity fallback

When Valve's DPC player endpoint explicitly returns no profile, the player worker
reads `https://steamcommunity.com/profiles/<SteamID64>/?xml=1` without an API key.
It validates the returned Steam ID and saves the display name, avatar URL,
optional self-reported location (`steam_location`),
`profile_source: "steam"`, and `profile_checked_at` (Unix milliseconds). Steam
location and other profile fields do not establish nationality, team membership,
or professional status.

Team refreshes can enqueue a fallback profile again after 24 hours, including
after a backend restart. Each attempt checks DPC first. The player catalogue can
also upgrade a Steam record immediately when a professional profile appears.
Steam writes cannot overwrite a DPC profile. If both sources lack a usable name,
the existing account-ID placeholder remains, with a bounded in-memory 24-hour
retry cooldown. Transport and malformed-response errors remain visible. Steam
requests share the existing Valve request limiter; there are no browser requests
for profile XML. Avatar URLs are cached as data; no new avatar UI is introduced.

### Steam enrichment for all players

An independent background sweep starts 45 seconds after backend startup. It
selects up to 50 due player records at a time, including DPC players and inactive
players, with five seconds between requests and a minute between batches. It
shares the Valve limiter. Successful and unavailable profiles are eligible again
after 24 hours. A transport/malformed-response failure is deferred for an hour
and stops that batch. Persistent retry timestamps let collection resume after
restart; cancellation does not record a failed lookup.

Steam enrichment stores `steam_name`, `steam_location`, `avatar_url`,
`steam_updated_at`, `steam_next_refresh_at`, and `steam_status`. DPC names,
country codes, teams, and statistics remain untouched. A DPC upgrade preserves
previously collected Steam metadata. Successful responses with an empty location
clear the previous location; failed requests preserve it. This is background
collection, with no new UI or extra requests from page loads.
