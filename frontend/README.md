# Dota 2 Leagues frontend

Svelte 5.57.0 and Vite 8. Requires Node.js 20.19+ or 22.12+.

From the repository root, run `cd frontend`, then:

- `npm ci` installs the locked dependencies.
- `npm run dev` starts Vite. Run the Go API on port 1323.
- `npm run check` checks Svelte components; `npm test` tests routing.
- `npm run build` creates `dist/`.
- `npm run deploy` builds and copies the frontend into the sibling `public/` directory used by Go. Downloaded league and team assets are preserved.

The app uses Svelte 5's `mount` API and event attributes. Existing legacy reactive components remain supported by Svelte 5; the clipboard component uses runes and the browser Clipboard API.

## Serving and indexing

Production must run the updated Go server from the backend repository root. It serves real HTML for `/`, `/league/:id`, `/team`, `/team/:id`, `/match/:leagueID/:matchID`, `/activity`, and `/dpc`. The initial HTML includes content from the application services, navigation links, a unique title, description, and canonical URL. Svelte replaces that content with the interactive application. API routes keep their existing URLs.

Set `SITE_URL` to the public origin (for example `https://dota-leagues.27n.gg`, the default) before starting Go. Build/copy the frontend and rebuild/restart Go together. A static file server or Vite preview alone does not provide server-rendered content or sitemaps.

Old `#/league/123?tab=live` links redirect in the browser to `/league/123?tab=live`. New links use ordinary page URLs. League and team lists paginate in batches of 100 (`?offset=100`); schedules use 20 series per page (`?offset=20#schedule`). Missing entities return HTTP 404 and service failures return HTTP 503.

Go serves `/robots.txt` and `/sitemap.xml`. The sitemap index lists paginated league, team, and match sitemaps (1,000 URLs per file). Match URLs come from unique stored league schedule records, without fetching individual matches from Valve. Sitemaps and initial page content require the configured database/services.

Search, filter, and sort variants of the directories carry `noindex,follow`. Unfiltered directories, pagination, and detail pages remain indexable. Data disclosures and tracking parameters canonicalize to the detail page.

Initial match HTML includes the outcome, UTC date, heroes, player-name fallbacks, statistics, and team links. The frontend retains a descriptive matchup title after rendering. `delivery/hero_names.json` is the backend name catalog extracted from `src/lib/heroes.js`; update both when adding heroes.

After deployment, verify the public origin in Google Search Console, submit `/sitemap.xml`, and inspect representative league, team, and match URLs. Google decides whether and when to index eligible pages.

## Visual testing

The UI includes a teams directory at `/team`, responsive league/team/match pages,
loading and empty states, retry actions, and a not-found page. League filters search the complete available catalogue and paginate matching results.
Team search applies to the current page.

The September 2026 Chrome MCP review checked desktop and mobile layouts,
live data, empty lists, service failures and retry recovery, missing records, sparse
match data, and legacy links. Regression checks are in `src/lib/*.test.js`.
The Go delivery tests cover directory pagination and HTTP 404/503 HTML responses.


The **Updates** navigation item opens `/activity`. It reads `GET /updates` with
20 entries per page, groups records by UTC date, and displays tournament, team,
roster, and player activity. See the backend README for the event design matrix.
Vite proxies `/updates` to the Go API. Restart the Go server after installing
these changes to register `/activity` and its crawlable initial content.

## Structured data

The Go HTML response includes JSON-LD in the document head. All pages describe the site and page; detail pages include `BreadcrumbList`. Team profiles use `SportsTeam`; leagues and matches use `SportsEvent`, with names, identifiers, and only known dates/durations/competitors. Stable entity IDs connect team references across pages. Radiant and Dire are competitors, not home/away teams.

Unknown venues, attendance modes, organizers, ticket prices, and event statuses are omitted. This is semantic markup, not a claim of eligibility for Google's event rich results. Validate deployed URLs in Google's Rich Results Test (breadcrumbs) and Schema.org's validator (full graph).

## Optional Google Analytics

Set `GA_MEASUREMENT_ID=G-XXXXXXXXXX` on the Go server to enable GA4. Leave it unset or blank to disable it completely: no Google Analytics script or configuration is emitted. Invalid measurement IDs fail startup with a configuration error. Restart Go after changing the value; no frontend rebuild is needed.

The server inserts Google's asynchronous `gtag.js` snippet once per HTML page. Normal navigation uses full page loads, so the standard configuration records page views without an additional frontend tracker. GA4 Enhanced Measurement can also record history changes used by filters; disable history-based page views in the GA4 web stream settings if those should not count. Vite development pages do not inject analytics.
