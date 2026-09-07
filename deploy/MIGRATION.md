# Server migration

## Checkpoint 1: destination prepared

Inventory taken on 5 September 2026. No source services were stopped and no source data was changed.

| Item | Source | Destination |
| --- | --- | --- |
| Address | `46.101.217.107` | `134.122.118.225` |
| OS | Ubuntu 18.04.4 | Ubuntu 26.04 |
| Application | `dota-league.service`, user `deploy` | Docker Compose, not started yet |
| Repository | Deployed binary; no Git metadata | `/opt/dota-leagues/repo` |
| Release | `ad63e3114b57c74dedc6a8999f87a8b528be25a7` (operator supplied) | `f9a0958aa144da6179b0b2b9922519350d0d7e73` |
| Database | System ArangoDB 3.6.5, `dota_leagues`, localhost:8529 | Planned ArangoDB 3.12 container |
| Downloaded files | `/home/deploy/work/dota_league/public`, 860 MB | Planned Compose `logos` volume mounted at `/data` |
| Resources | Database directory about 402 MB | 2 GB RAM, no swap, 46 GB free before Docker installation |

Docker Engine 29.8.0, Compose 5.5.1, Git, and rsync are available on the destination. The public repository is cloned over HTTPS without copying private SSH keys. The release is checked out detached so later changes to `master` do not silently change this migration.

The operator confirmed that `dota-leagues.27n.gg` will remain the public hostname and DNS can be updated at cutover. The destination now has a mode-0600 `.env` with that HTTPS origin and a newly generated database password. Credentials are not stored in this document. Certificates, application containers, and restored data are not yet prepared.

Initial collection counts (the live collectors can change these):

| Collection | Documents |
| --- | ---: |
| `team_rosters` | 2,988 |
| `teams` | 2,988 |
| `players` | 5,742 |
| `live_game_details` | 1 |
| `league_details` | 4,648 |
| `leagues` | 9,914 |
| `games` | 37 |

## Latest-code and backup preparation

GitHub `master`, the local application checkout, and the destination checkout were rechecked and all point to `f9a0958aa144da6179b0b2b9922519350d0d7e73`. The destination build uses this latest published release; no old-server binary or frontend bundle is used. Recheck GitHub before final cutover and rebuild/revalidate if the release changes.

The destination image build completed successfully, including Svelte checks, 31 frontend tests, the frontend build, and Go compilation. Image tags are `dota-leagues:local` and `dota-leagues:f9a0958`; the image manifest digest is `sha256:8c3cf5ef61e061eeb7dbfbb2814bbd9fcfe7f7987f0f57f83bbea1af0de77bda`. No application containers were started.

The source rehearsal dump is `/var/backups/dota-leagues/rehearsal-20260905T192840Z`. It contains 2,118,441 bytes of dump files plus a per-file SHA-256 manifest. The temporary credential file was removed after the dump completed. The dump was copied to the same path on the destination; all 16 dump files passed SHA-256 verification. The source service remained active, so this is a rehearsal backup, not the final cutover snapshot.

The inspected source `public/` contains 4,153 numeric league directories and a `teams/` directory. League directories contain `logo.png`; this layout matches the new asset root. No old frontend bundles were present at its top level.

## Restore checkpoint

The verified dump has been restored into `dota_leagues` on the destination's ArangoDB 3.12.4-3 container. Before restoring, the destination was checked to confirm that this database did not exist; no existing database was overwritten. This is an isolated destination copy of the rehearsal snapshot, using the production database name for the later application check.

Every restored document was compared with the decompressed dump after excluding ArangoDB `_id`/`_rev` metadata. All content hashes and counts matched:

| Collection | Restored documents |
| --- | ---: |
| `games` | 33 |
| `league_details` | 4,648 |
| `leagues` | 9,914 |
| `live_game_details` | 1 |
| `players` | 5,742 |
| `team_rosters` | 2,988 |
| `teams` | 2,988 |

The `games` count differs from the earlier inventory because the source collectors continued running before the dump. Verification used the actual backup rather than the earlier live inventory. The destination report is `/var/backups/dota-leagues/rehearsal-20260905T192840Z/restore-verification.json`.

The latest app's `-production -create-indexes` maintenance command completed successfully and created all eight indexes. It also created the missing indexed collections. The application server and collectors have not been started; only ArangoDB is running, with no published host database port.

## Downloaded files restored

All **6,902 files (858,280,069 bytes)** were transferred from the source `public/` directory over SSH. Each file's SHA-256 hash and size matched the source manifest both in destination staging and after copying into the `dota-leagues_logos` volume. Existing relative paths were preserved, including numeric league directories and `teams/`.

The volume was checked to be empty before copying. Files and directories are owned by `10001:10001`; a temporary container running as the app user confirmed read/write access to a league image and write access to the teams directory. No HTTP server or collectors were started for these checks.

Destination files retained for verification and final synchronization:

- `/var/backups/dota-leagues/rehearsal-20260905T192840Z/public/`: transferred source images.
- `public-manifest.json` in the same backup directory: source SHA-256 and size for every file.
- `public-verification.json` in that directory: successful volume verification summary.

The database and images are now restored for rehearsal. The source remains live, so a final synchronized dump and asset copy are still required before cutover.

## Application started for rehearsal

The operator requested application startup after restoration. `docker compose -f compose.production.yml up -d --no-build --wait app` completed successfully on the destination. Both app and ArangoDB are healthy. The app listens at `127.0.0.1:1323` on the destination; Caddy and DNS cutover have not been activated.

HTTP checks passed for health, leagues, teams, activity, DPC, sample league/team profiles, directory/update APIs, robots.txt, the sitemap index and every advertised child sitemap, and the JS/CSS bundles. Page canonical URLs use `https://dota-leagues.27n.gg`. Responses for `/15280/logo.png` and `/teams/10000260/logo.png` match the restored files by SHA-256. The initial check guessed an unsupported `/sitemaps/pages.xml` URL (404); the actual advertised route `/sitemaps/pages/1` and all other advertised sitemap routes passed.

For browser testing from your computer, keep this tunnel running:

```sh
ssh -N -L 18080:127.0.0.1:1323 root@134.122.118.225
```

Then open `http://localhost:18080`. The old public domain still reaches the old server. New-server collectors are now writing rehearsal data, so stop the destination app before the final restore. Preserve the original backup and repeat the agreed final-transfer process before DNS cutover.

## Supplemental local logos

The operator found that `/20206/logo.png` existed in the development checkout but not in the source-server backup. Comparison of local `public/<league-id>/logo.png` and `public/teams/<team-id>/logo.png` against the destination identified 3,609 missing files (605,070,424 bytes), out of 5,031 local downloaded logos.

These supplemental images are copied with `rsync --ignore-existing`, restricted to an explicit logo file list. Existing destination images are preserved. The destination remains `dota-leagues_logos`, with UID/GID `10001:10001`, directory mode 0755 and file mode 0644.

The completed sync copied **3,605 files (604,833,132 bytes)**. Four files became present on the server after the initial scan and were skipped by `--ignore-existing`. Every copied file passed SHA-256, size, ownership, and mode verification. A final dry run found zero local logos missing on the destination. `/20206/logo.png` returned HTTP 200 with the same SHA-256 as the local 73,338-byte image.

The destination backup directory contains `local-logo-manifest.json` and `local-logo-verification.json` for these additions. Existing server images were not overwritten, and no files were deleted.

During the final asset sync from the old server, preserve these local-only images: **do not use `rsync --delete`, empty the logo volume, or replace it wholesale**. Merge the final source-server images into the existing volume. Database restoration does not require recreating the logo volume.

## Supplemental local database merge

The operator requested local history to be merged after the source-server restoration. The new app was stopped briefly; the old public service remained running. A full JSON backup of all 15 destination collections was saved before changing records. Local exports excluded live `games`, `live_game_details`, and derived `update_groups`; only the canonical `latest` DPC standings document was considered.

Merge rules:

- Existing server entity, standings, result, and match records were preserved by `_key`; only absent records were inserted.
- Schedule rows were imported only for leagues with no destination schedule rows, avoiding mixtures with already refreshed server schedules.
- Updates were deduplicated by globally generated event `id`. Imported storage keys use `local_<id>` to avoid collisions between ArangoDB-generated numeric keys from separate databases. Event IDs and group IDs were retained.
- Source snapshots and missing source-status records were inserted. Derived source-version counts were recalculated from the combined archive; existing observed status fields were preserved.
- `update_groups` was rebuilt from events by normal application startup.

Verified additions, before collectors resumed:

| Collection | Inserted | Total after merge |
| --- | ---: | ---: |
| `leagues` | 0 | 9,914 |
| `league_details` | 31 | 4,681 |
| `teams` | 3,250 | 6,446 |
| `players` | 6,633 | 12,520 |
| `team_rosters` | 3,250 | 6,446 |
| `league_series` | 110,541 | 112,697 |
| `league_results` | 11 | 11 |
| `match_minimal` | 10 | 10 |
| `dpc_standings` | 0 | 1 |
| `updates` | 6,506 | 6,868 |
| `source_snapshots` | 3,565 | 3,802 |
| `source_status` | 3,335 | 3,572 |

The merge checked each original record and each inserted record against normalized SHA-256 content hashes and checked collection counts. All checks passed; the only permitted change to an existing record was its derived source-version count. The app was restarted and became healthy.

Destination audit directory: `/var/backups/dota-leagues/local-db-20260905/`. It contains the checksummed local `export/`, complete destination `before-merge/` JSON backup, `merge.py`, preview/final plans, merge logs, and verification reports. The apply script deliberately refuses to run with the app active or reuse an existing pre-merge backup directory; inspect and prepare a new backup path before any future apply run.

**Final cutover must preserve these additions.** Restore the final old-server dump into a separate temporary database first. Reconcile its seven collections by key, retaining destination-only records and reviewing overlapping record freshness. Do not drop/recreate the production database or blindly restore over its collections: that would discard local-only teams, players, and roster data. Preserve the new historical collections and merge backups.

## Activity history trimmed

At the operator's request, the destination activity history was reduced to the **latest 1,000 raw update records**, ordered by `created_at DESC, id DESC, _key DESC`. With the new app stopped, 5,990 of 6,990 records were removed and derived activity groups were rebuilt in the same exclusive transaction. Verification confirmed that all 1,000 retained events were unchanged and all 1,000 resulting groups had valid references. Other collections were not modified.

The complete pre-cleanup activity and group backups, SHA-256 manifest, and verification report are retained at `/var/backups/dota-leagues/activity-trim-20260905T200015Z/`. The new app was restarted afterward. This was a one-time cleanup; new observations will increase the count again.

Do not reimport the trimmed local activity during final migration. Preserve the current destination activity history; the old server's original seven collections do not include `updates` or `update_groups`.

## Public domain activated with Cloudflare

The operator changed DNS and confirmed Cloudflare **Flexible** mode. The HTTPS Compose profile is now running Caddy on the destination's ports 80/443. Caddy obtained a public certificate for `dota-leagues.27n.gg`.

Cloudflare Flexible initially looped because Caddy redirected the HTTP origin connection back to HTTPS. The local checkout and destination `compose.production.yml` and `deploy/Caddyfile` now support an optional `CADDY_SITE_ADDRESSES` setting. The destination `.env` sets it to `http://dota-leagues.27n.gg, https://dota-leagues.27n.gg`, allowing HTTP origin requests while retaining valid direct HTTPS. Original proxy configuration is backed up at `/var/backups/dota-leagues/caddy-before-flexible/`.

The public home page returned HTTP 200 through Cloudflare without a redirect loop, with the correct canonical origin. Direct origin HTTPS using SNI and normal certificate verification also returned HTTP 200. This proxy adjustment is a working-tree change on the destination and in the local checkout; it must be committed/published before a clean future checkout can reproduce it. Do not reset the destination checkout without preserving it.

The public domain is now active. The previously planned final source-data reconciliation and retirement of the old service have not been performed; the old server has not been stopped by this migration session.

## Stale Cloudflare image cache

Chrome reproduced 404s for league logos after public activation. `/20206/logo.png` returned `CF-Cache-Status: HIT`, `Cache-Control: max-age=14400`, and an HTML 404 body identifying the old `nginx/1.16.1` server. This is an old-origin cached response, not a missing destination asset. Cache-busting requests in the same Chrome session for league logos 20206, 19944, and 20145 returned HTTP 200 / cache MISS and decoded correctly at 1024×400.

Cloudflare cache must be purged for hostname `dota-leagues.27n.gg`, then the browser hard-reloaded to remove locally cached failures. No Cloudflare dashboard/API connection is available to this session, so the account operator must perform that purge. Do not overwrite or re-copy valid destination images to address these cached errors.

## Remaining sequence
1. **Verify application behavior.** Test the latest app with restored data behind loopback access, without changing DNS. Verify sample league/team image requests and all key routes. Any rehearsal collector writes are disposable and must not be treated as migrated source data. The source has no stored schedule collection; missing newer collections will be initialized and populated by the new app.
2. **Final transfer.** Once rehearsal passes and the cutover window is agreed, stop the source `dota-league.service` to stop its writers. Keep source ArangoDB running. Take a final dump and collection counts and synchronize assets again without deleting supplemental logos. Restore into a separate temporary database, then reconcile the seven source collections into production while its app is stopped, preserving the supplemental local data described above. Verify the reconciliation before startup changes derived collections. Retain the backup and source installation.
3. **Start and switch.** Recheck the GitHub release; rebuild and validate if it changed. Create indexes, start the destination application, and verify health, pages, API, images, canonical URLs, robots, and sitemaps. Arrange HTTPS before switching traffic; then update DNS and check the public origin. Keep the source collectors stopped to avoid two independent histories.
4. **Observe and retain rollback.** Monitor application/database logs and resource usage. Keep the old server and final backups until the new deployment is accepted. Do not remove source data or run Compose `down -v` as part of migration.

## Rollback boundary

Before the new app writes production data, recovery is to stop the destination app, restart the old service, and restore routing to the old server if it changed. After the new collectors have written data, the databases diverge: pause writers and decide how to retain the new observations before reverting. Restarting the old service alone does not transfer those observations back.

## References

- [Deployment and index maintenance](README.md)
- [Docker installation on Ubuntu](https://docs.docker.com/engine/install/ubuntu/)
- [ArangoDB logical upgrades](https://docs.arango.ai/arangodb/stable/operations/upgrading/)
- [ArangoDB restore behavior and older dumps](https://docs.arango.ai/arangodb/stable/components/tools/arangorestore/examples/)
