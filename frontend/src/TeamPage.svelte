<script>
  import PlayerAvatar from "./PlayerAvatar.svelte";
  import SourceData from "./SourceData.svelte";
  import Breadcrumbs from "./Breadcrumbs.svelte";
  import { onMount } from "svelte";
  import heroes from "./lib/heroes.js";
  import { api } from "./lib/api.js";
  import { formatDate, regionText, countryText, steamProfileUrl, teamLogo, normalizeUrl } from "./lib/constants.js";
  import StatePanel from "./StatePanel.svelte";
  import EntityImage from "./EntityImage.svelte";
  export let id;
  let team = null;
  let loadError = null;
  let loading = true;
  let members = [];
  let history = [];
  $: hasSteamLocation = members.some(m => typeof m.steam_location === "string" && m.steam_location.trim());
  const number = value => Number.isFinite(Number(value)) ? Number(value) : 0;
  const average = value => value == null ? "—" : number(value).toFixed(1);
  function creationDate(timestamp) {
    const date = new Date(Number(timestamp) * 1000);
    if (!timestamp || !Number.isFinite(date.getTime())) return "—";
    return new Intl.DateTimeFormat('en-GB', { day: 'numeric', month: 'short', year: 'numeric', timeZone: 'UTC' }).format(date);
  }
  async function load() {
    loading = true;
    loadError = null;
    try {
      const result = await api(`/teams/${id}`);
      if (!result?.team_id) throw Object.assign(new Error("This team is not available."), { status: 404 });
      const stats = new Map((Array.isArray(result.member_stats) ? result.member_stats : []).filter(Boolean).map(s => [s.account_id, s]));
      members = (Array.isArray(result.members) ? result.members : []).filter(Boolean).map(m => ({ ...m, stat: stats.get(m.account_id) }));
      history = (Array.isArray(result.dpc_results) ? result.dpc_results : []).filter(Boolean);
      team = result;
    } catch (err) { loadError = err; }
    finally { loading = false; }
  }
  onMount(load);
</script>

<svelte:head><title>{loadError?.status === 404 ? "Team not found" : team?.name || "Team"} · Dota 2 Leagues</title>{#if loadError?.status === 404}<meta name="robots" content="noindex" />{/if}</svelte:head>
<Breadcrumbs items={[{label:'Teams',href:'/team'},{label:team?.name || `Team #${id}`}]} />
{#if loading || loadError}<h1>{loadError?.status === 404 ? "Team not found" : "Team profile"}</h1>{/if}
{#if loading}
  <StatePanel kind="loading" title="Loading team" message="Fetching the roster and tournament history." />
{:else if loadError}
  <StatePanel kind={loadError.status === 404 ? "notFound" : "error"} title={loadError.status === 404 ? "Team not found" : "Unable to load team"} message={loadError.message} actionLabel={loadError.status === 404 ? "Browse teams" : "Try again"} href={loadError.status === 404 ? "/team" : undefined} onaction={loadError.status === 404 ? undefined : load} />
{:else if team}
  <header class="team-header">
    <div class="identity">
      <div class="team-logo"><EntityImage src={teamLogo(team.team_id)} name={team.name || "Team"} /></div>
      <div class="info">
        <h1>{team.name || `Team #${id}`}</h1>
        <p class="team-status">{#if team.tag && team.tag !== team.name}<span class="tag">{team.tag}</span><span class="status-separator" aria-hidden="true">·</span>{/if}{team.pro == null ? "Team profile" : team.pro ? "Professional team" : "Non-professional team"}</p>
        <ul class="metadata" aria-label="Team details">
          <li>Team #{team.team_id}</li>
          <li>Created {creationDate(team.time_created)}</li>
          {#if countryText(team.country_code)}<li>{countryText(team.country_code)}</li>{/if}
          {#if regionText(team.region) !== 'Unknown'}<li>{regionText(team.region)}</li>{/if}
          {#if normalizeUrl(team.url)}<li><a class="external" href={normalizeUrl(team.url)}>{team.url}<span class="external-arrow" aria-hidden="true">↗</span></a></li>{/if}
        </ul>
      </div>
    </div>
    <dl class="team-stats" aria-label="Team statistics">
      <div><dt>Total games</dt><dd>{team.games_played_total == null ? "—" : number(team.games_played_total).toLocaleString()}</dd></div>
      <div><dt>Official games</dt><dd>{team.games_played_total == null ? "—" : Math.max(0, number(team.games_played_total) - number(team.games_played_matchmaking)).toLocaleString()}</dd></div>
      <div><dt>Reported record</dt><dd>{number(team.wins).toLocaleString()}<span class="record-unit">W</span> <span class="record-separator">–</span> {number(team.losses).toLocaleString()}<span class="record-unit">L</span></dd></div>
    </dl>
  </header>
  <section aria-labelledby="roster-title">
    <h2 id="roster-title">Roster</h2>
    {#if members.length}
      <!-- svelte-ignore a11y_no_noninteractive_tabindex (Keyboard users need to scroll the table horizontally.) -->
      <div class="table-scroll" role="region" aria-label="Team roster" tabindex="0"><table>
        <thead><tr><th scope="col">Player</th>{#if hasSteamLocation}<th scope="col" title="Self-reported location on Steam">Steam location</th>{/if}<th scope="col">Joined</th><th scope="col">Top hero</th><th scope="col" class="num">Average K / D / A</th></tr></thead>
        <tbody>{#each members as m}
          <tr><td><div class="roster-player"><PlayerAvatar src={m.avatar_url} /><div class="roster-player-name">{#if steamProfileUrl(m.account_id)}<a href={steamProfileUrl(m.account_id)}>{m.pro_name || m.player_name || `Player #${m.account_id}`}</a>{:else}{m.pro_name || m.player_name || "Unknown player"}{/if}{#if m.admin}<span class="admin">Team admin</span>{/if}{#if m.real_name?.trim()}<span class="real-name">{m.real_name}</span>{/if}</div></div></td>
            {#if hasSteamLocation}<td class="steam-location">{m.steam_location?.trim() || "—"}</td>{/if}
            <td>{formatDate(m.time_joined)}</td>
            <td>{#if m.stat?.top_heroes?.[0]}{heroes[m.stat.top_heroes[0].hero_id]?.name || `Hero #${m.stat.top_heroes[0].hero_id}`} ({number(m.stat.top_heroes[0].picks)}){:else}—{/if}</td>
            <td class="num">{m.stat ? `${average(m.stat.avg_kills)} / ${average(m.stat.avg_deaths)} / ${average(m.stat.avg_assists)}` : "—"}</td></tr>
        {/each}</tbody>
      </table></div>
    {:else}<StatePanel title="No roster available" message="Player information has not been published for this team." />{/if}
  </section>
  <section aria-labelledby="history-title">
    <h2 id="history-title">DPC history</h2>
    {#if history.length}
      <!-- svelte-ignore a11y_no_noninteractive_tabindex (Keyboard users need to scroll the table horizontally.) -->
      <div class="table-scroll" role="region" aria-label="DPC history" tabindex="0"><table>
        <thead><tr><th scope="col">Tournament</th><th scope="col">Date (UTC)</th><th scope="col" class="num">Standing</th><th scope="col" class="num">Points</th><th scope="col" class="num">Earnings</th></tr></thead>
        <tbody>{#each history as r}<tr><td>{#if r.league_available && r.league_id > 0}<a href={`/league/${r.league_id}`}>{r.league_name || `Tournament #${r.league_id}`}</a>{:else}{r.league_name || (r.league_id > 0 ? `Tournament #${r.league_id}` : 'Unknown tournament')}<small class="unavailable">Details unavailable</small>{/if}</td><td class="history-date">{formatDate(r.timestamp)}</td><td class="num">{r.standing || "—"}</td><td class="num">{number(r.points)}</td><td class="num">${number(r.earnings).toLocaleString()}</td></tr>{/each}</tbody>
      </table></div>
    {:else}<StatePanel title="No DPC history available" message="There are no recorded DPC tournament results for this team." />{/if}
  </section>
{/if}
{#if team}<SourceData kind="team" entityID={id} />{/if}
<style>
  .team-header { margin: 1.5rem 0 2rem; }
  .identity { display: flex; align-items: center; gap: 1.75rem; }
  .team-logo { width: 6rem; height: 6rem; flex: 0 0 6rem; border-radius: .5rem; overflow: hidden; }
  .info { min-width: 0; flex: 1; }
  h1 { margin: 0; font-size: clamp(1.75rem, 3.5vw, 2.75rem); line-height: 1.12; letter-spacing: -.035em; overflow-wrap: anywhere; text-wrap: balance; }
  .team-status { margin: .5rem 0 0; color: var(--muted); font-size: .875rem; }
  .status-separator { margin: 0 .5rem; }
  .tag { color: var(--text); font-weight: 500; }
  .metadata { display: flex; flex-wrap: wrap; gap: .375rem 0; list-style: none; padding: 0; margin: .875rem 0 0; color: var(--muted); font-size: .8125rem; }
  .metadata li + li::before { content: "·"; padding: 0 .625rem; color: var(--muted); }
  .external { overflow-wrap: anywhere; text-underline-offset: .2em; transition: color 180ms ease; }
  .external:hover { color: var(--accent); }
  .external-arrow { display: inline-block; margin-left: .25rem; }
  .team-stats { display: grid; grid-template-columns: repeat(3,minmax(0,1fr)); margin: 1.75rem 0 0; padding-top: 1.25rem; border-top: 1px solid var(--line); }
  .team-stats div { min-width: 0; padding-right: 1rem; }
  .team-stats div + div { padding-left: 1.5rem; border-left: 1px solid var(--line); }
  dt { color: var(--muted); font-size: .75rem; }
  dd { margin: .375rem 0 0; font-size: clamp(1.25rem, 2.5vw, 1.875rem); line-height: 1.2; font-weight: 600; font-variant-numeric: tabular-nums; }
  .record-unit { font-size: .6em; font-weight: 500; margin-left: .125rem; }
  .record-separator { color: var(--muted); font-weight: 400; }
  section { margin:32px 0; } table { width:100%; border-collapse:collapse; } th,td { padding:12px; border-bottom:1px solid var(--line); text-align:left; } th { font-size:12px; color:var(--muted); white-space:nowrap; } .num { text-align:right; white-space:nowrap; }
  .unavailable { display: block; margin-top: 3px; color: var(--muted); font-size: 11px; }
  .steam-location { min-width: 9rem; max-width: 18rem; overflow-wrap: anywhere; color: var(--muted); }
  .roster-player { display: flex; align-items: center; gap: .625rem; }
  .roster-player-name { min-width: 0; }
  .real-name { display: block; margin-top: .25rem; color: var(--muted); font-size: .8125rem; }
  .history-date { white-space: nowrap; font-variant-numeric: tabular-nums; }
  .admin { display:inline-block; margin-left:8px; font-size:11px; background:var(--surface-subtle); padding:2px 6px; border-radius:3px; color:var(--muted); }
  @media(max-width:600px) {
    .identity { gap: 1rem; align-items: flex-start; }
    .team-logo { width: 4rem; height: 4rem; flex-basis: 4rem; }
    .team-stats { margin-top: 1.25rem; }
    .team-stats div { padding-right: .5rem; }
    .team-stats div + div { padding-left: .75rem; }
  }
  @media(prefers-reduced-motion:reduce) { .external { transition: none; } }
</style>
