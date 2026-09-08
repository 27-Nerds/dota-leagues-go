# Valve Dota 2 webapi — undocumented data (reverse-engineering notes)

Method: fetched every endpoint this repo calls at ~1 rps with the stock UA
(`Valve/Steam HTTP Client 1.0 (570)`), diffed raw JSON against our parsed
structs, and mined Valve's own React bundle (`dota_react/main.js`, contenthash
`qA3udhsb6x_S`, 2 MB) for endpoints we don't use. Raw captures live in
/tmp/opencode/valve-probe/raw (session of 2026-09-08); they are volatile, so
sizes/statuses below double as a reproducible manifest.

## Summary of findings

| # | Finding | Endpoint | Impact |
|---|---------|----------|--------|
| 1 | **Bulk league fetch** `GetLeaguesData` (plural) exists — N leagues in one request, with `delay_seconds` throttle | new | replaces per-league polling for live leagues |
| 2 | **Search endpoint** `GetDPCSearchResults` (`search`, `count`, `resultflags`) | new | fuzzy team/league search |
| 3 | Bracket `nodes[]` tree (incl. **per-game VOD YouTube URLs**) is fully unparsed in our model | GetLeagueData | match placement + vod links |
| 4 | Node-group fields beyond standings are dropped: `name`, timing, `win_loss_limit`, advancing-group pointers, etc. | GetLeagueData | bracket semantics |
| 5 | Standings drop all **tiebreak** columns incl. a Valve typo (`tiebereak_average_game_length`) and logo url / abbreviation | GetLeagueData | full standings |
| 6 | Top-level `prize_pool` drops `prize_split_pct_x100[]`, `prize_pool_items[]` | GetLeagueData | prize split detail |
| 7 | Player endpoints: `has_played_in_international`, `team_abbreviation` unparsed; model fields `results/is_locked/is_pro/birthdate` never observed in 3834 records | GetProPlayerInfo / GetPlayerInfo | dead code + missing flags |
| 8 | `GetLeagueMatchMinimal`: `tourney.series_type` / `series_game` dropped | minimal match | series placement of a single game |
| 9 | `GetSingleTeamInfo` flags `get_dpc_info`/`get_player_stats` are **no-ops** in 2026; v001 and v0001 byte-identical | teams | simplification |
| 10 | `GetRealtimeStats` returns HTTP 200 body `null` for many live games (incl. production) — game/state gated, not endpoint-gated | match stats | expectations for nulls |
| 11 | Soft error envelope `{"success":false,"message":"..."}` with HTTP 200 (GetLeaguesData: "Missing league_ids") | bulk league data | robustness gap in our transport (only non-2xx treated as errors) |
| 12 | `GetLiveGames` has a v002 alias; `time` uses int32 cast sentinel for negative values (~4294967xxx) | live games | quirk doc |

## Endpoint-by-endpoint detail

### IDOTA2League/GetLeaguesData/v001 — NEW (from main.js)

Site call: `params:{league_ids:a, delay_seconds:t}`. Response: `{leagues:[<same payload as GetLeagueData>]}`
(`admins`, `info`, `node_groups`, `prize_pool`, `registered_players`, `series_infos`, `streams`).

- `league_ids` is a comma-separated list; non-existent ids come back as zeroed stubs (verified: `36,9999999` returned both).
- `delay_seconds` — client-facing throttle hint; with 0 the response is immediate.
- Soft errors return HTTP 200 + `{"success":false,"message":"Missing league_ids"}` (evidence: new_bulk_delay5.json, 48 B) — our transport decodes this into an empty struct instead of failing.

Sample sizes: single TI2025 payload 159,805 B; list of two tiny leagues 1,060 B.

### IDOTA2DPC/GetDPCSearchResults/v001 — NEW (from main.js)

Site call: `params:{search:"",count:40,resultflags:2}` (used as "popular teams" when search is empty).
Response shape: `{players:[], teams:[{id,name,url}], leagues:[{id,name}]}`.

Verified flag bits: **2 = teams**, **4 = leagues**. Empty `search` returns popular teams.
The `players` bit was not verifiable in this session (flag 1/8/9 with pro names returned no player entries;
search matching appears substring-based on team names, case-insensitive). Team `url` is a logo CDN url, may be empty.

### IDOTA2League/GetLeagueData/v001 (and bulk) — unparsed keys

Captured leagues: TI 2019 (10749), TI 2025 (18324), DreamLeague S28/S29, ESL One Birmingham 2026,
two current community tournaments. Key sets are stable across all of them.

Top level `admins[]` and `registered_players[]`: always empty in ~15 sampled leagues spanning 2012–2026; shape unknown (likely legacy).

`info`: we drop **`image_bits`** (observed values: 0 for ancient community leagues, 1537 = bits 0+9+10 for essentially all modern ones) and **`notes`**.
Model-only fields never present in any payload: `base_prize_pool`, `total_prize_pool`, `is_live`, `updated_timestamp`.

Node group (24 keys raw vs 13 parsed). Dropped:
- `name`, `region`, `start_time`, `end_time` — group timing/label
- `default_node_type`, `win_loss_limit`, `elimination_dpc_points`
- `is_tiebreaker`, `advancing_team_count`
- `secondary_advancing_node_group_id`, `secondary_advancing_team_count`
- `tertiary_advancing_node_group_id`, `tertiary_advancing_team_count`
- **nested `node_groups[]`** (sub-brackets) and **`nodes[]`** — see below

Bracket node entry (completely unparsed; 20 keys):

```
node_id, name, node_type, series_id, scheduled_time, actual_time,
team_id_1, team_id_2, team_1_wins, team_2_wins, is_completed, has_started,
winning_node_id, losing_node_id, incoming_node_id_1, incoming_node_id_2,
matches: [{match_id, winning_team_id, duration}],
stream_ids: [int]            # broadcast channels (up to ~6 ids on finals)
vods:     [{series_game, stream_id, url}]   # per-game YouTube VOD urls
```

`vods[]` is populated on TI 2025 (58/60 nodes carry at least one; TI 2019 does not — pre-VOD era).
TI 2025 sample: node "Semifinals, Game 3" → vod url `https://www.youtube.com/watch?v=...` per series_game.

Team standing (17 keys raw vs 9 parsed). Dropped:
- `team_abbreviation`, `team_logo_url`
- **all tiebreak columns**: `tiebreak_game_win_pct`, `tiebreak_opponent_match_wins`,
  `tiebreak_opponent_game_win_pct`, `tiebreak_coinflip`, and the Valve-typo'd
  `tiebereak_average_game_length`.

Top-level `prize_pool` (we parse only base/total). Dropped: `prize_split_pct_x100[]`, `prize_pool_items[]`.

### IDOTA2DPC/GetPlayerInfo/v001 & GetProPlayerInfo/v001

Same record schema in both. Full key union across all 3834 fantasy players:
`account_id, name, country_code, fantasy_role, team_id, team_name*, team_tag*, sponsor,
real_name, total_earnings, team_url_logo*, audit_entries (3788/3834), team_abbreviation*
(pro_registration {registration_period, timestamp}, has_played_in_international)` (* = present only for rostered players).

- Unparsed: **`has_played_in_international`** (bool; false for all 3834 in this snapshot — may be set only around TI or stale), **`team_abbreviation`**.
- Never observed on any record: `results[]`, `is_locked`, `is_pro`, `birthdate` → model fields are dead for these endpoints.
- Non-pro account (e.g. 99999999) returns HTTP 200 body `{"name":""}` — our "empty object" handling (`ID==0`) is correct; the comment in api/client.go should say so.
- Note: player-side `audit_entries` entries are `{start_timestamp, team_id, team_name, team_tag, team_url_logo}` — a different shape from team-side audit entries `{account_id, audit_action, timestamp}`. The model handles both correctly (PlayerTeamEntry vs TeamAuditEntry).

### IDOTA2DPC/GetLeagueResults/v001 (per league)

Fully mapped: `results[] {earnings, points, standing, team_abbreviation, team_id, team_logo(_url), team_name, timestamp}` + top-level `dollars[]`, `points[]`. TI 2025 returned 16 standings and 20 per-team dollar buckets.

### IDOTA2DPC/GetLeagueMatchMinimal/v001

Top level fully mapped (match_id is **string** in this endpoint — model already string-typed).
`tourney` drops: **`series_type`**, **`series_game`** (which game of the series) on top of team block.
Verified against a completed TI 2025 match (8451547891, Spirit vs Falcons): `series_type:1, series_game:0`.
Live match probe returned HTTP 200 body `null` → endpoint is populated after the match ends.

### IDOTA2Teams/GetSingleTeamInfo

- All of bare / `get_dpc_info=true|false` / `get_player_stats=true` / v0001 vs v0002 / site's **v001** variant are byte-identical (md5-verified for team 36) → the flags currently change nothing; member_stats + audit_entries + dpc_results always present.
- `registered_member_account_ids` / `updated_timestamp` not observed on any sampled team so far (model-only).

### IDOTA2MatchStats/GetRealtimeStats/v001

~13 live games probed across tiers (community bo1s to Pro League): all HTTP 200 body `null`
(4 bytes), including by `match_id`. Production app.log shows the same null responses, so this is normal gating (likely only specific match phases / officially featured broadcasts emit stats), not a client-side issue. When populated it yields the LiveGameDetails schema we already parse.

### IDOTA2League/GetLiveGames/v001 (+ v002 alias)

Fully mapped; v002 exists and returns identical game-entry keys (42 vs 44 games at slightly different times). Quirk: `time` is int32-cast on the server — negative values arrive as ~4294967xxx (observed -79, -10, -90); clamp or decode before display.

### IDOTA2League/GetPrizePool/v001

Fully mapped. Works for ended leagues too: TI 2025 returns `{"prize_pool":2881791,"increment_per_second":0}`; null (HTTP 200 "null") is also a normal response for some leagues.

### IDOTA2DPC/GetDPCStandings/v001

All five arrays empty today (seasonal endpoint post-DPC); `?region=1` returns the same envelope (115 B). No version-specific surprises.

## Recommendations (not yet code changes)

1. Switch live-league polling to one `GetLeaguesData?v001?league_ids=<csv>` call per cycle — big rate-limit win; keep GetLeagueData as fallback for non-live singles.
2. Parse bracket `nodes[]` (+vods, stream_ids) — this is the only place VOD URLs exist server-side.
3. Add standings tiebreak columns + abbreviation/logo_url; add group `name/start_time/end_time/win_loss_limit`.
4. Minimal match: capture `tourney.series_type/series_game`.
5. Player model: either drop or populate `results/is_locked/is_pro/birthdate`; add `has_played_in_international`, `team_abbreviation`.
6. Transport: treat HTTP 200 + `{"success":false}` as an error where endpoints use it.
7. Optional new feature: site-style popular-teams list via GetDPCSearchResults (`search=""`,`resultflags=2`).
