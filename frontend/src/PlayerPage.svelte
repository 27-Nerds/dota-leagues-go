<script>
  import PlayerAvatar from "./PlayerAvatar.svelte";
  import Breadcrumbs from "./Breadcrumbs.svelte";
  import StatePanel from "./StatePanel.svelte";
  import { onMount } from "svelte";
  import { api } from "./lib/api.js";
  import ChangeLog from "./ChangeLog.svelte";
  import { countryText, fantasyRoleText, formatDate, formatDateTime, formatMoney, steamProfileUrl } from "./lib/constants.js";
  import { steamState, playerDisplayName } from "./lib/players.js";
  export let id;
  let player = null;
  let loadError = null;
  let loading = true;
  $: name = playerDisplayName(player, id);
  $: steam = steamState(player);
  $: results = (Array.isArray(player?.results) ? player.results : []).filter(r => r && r.league_id > 0);
  // Prefer the DPC team; fall back to the team whose Valve roster lists the player.
  $: roster = player?.roster_team?.id > 0 ? player.roster_team : null;
  $: team = player?.team_id > 0 ? { id: player.team_id, name: player.team_name || `Team #${player.team_id}`, tag: player.team_tag, href: player.team_available ? `/team/${player.team_id}` : null }
    : roster ? { id: roster.id, name: roster.name || `Team #${roster.id}`, tag: roster.tag, href: `/team/${roster.id}` } : null;
  $: rosterDiffers = roster && roster.id !== player?.team_id;
  const millis = value => Number(value) > 0 ? Number(value) / 1000 : 0;
  async function load() {
    loading = true; loadError = null;
    try {
      const result = await api(`/players/${id}`);
      if (!result?.account_id) throw Object.assign(new Error("This player is not available."), { status: 404 });
      player = result;
    } catch (err) { loadError = err; }
    finally { loading = false; }
  }
  onMount(load);
</script>

<svelte:head><title>{loadError?.status === 404 ? "Player not found" : name} · Dota 2 Leagues</title>{#if loadError?.status === 404}<meta name="robots" content="noindex" />{/if}</svelte:head>
<Breadcrumbs items={team ? [{label:'Teams',href:'/team'}, {label: team.name, href: team.href || undefined}, {label:name}] : [{label:'Players'}, {label:name}]} />
{#if loading || loadError}<h1>{loadError?.status === 404 ? "Player not found" : "Player profile"}</h1>{/if}
{#if loading}
  <StatePanel kind="loading" title="Loading player" message="Fetching the professional and Steam profile." />
{:else if loadError}
  <StatePanel kind={loadError.status === 404 ? "notFound" : "error"} title={loadError.status === 404 ? "Player not found" : "Unable to load player"} message={loadError.message} actionLabel={loadError.status === 404 ? "Browse teams" : "Try again"} href={loadError.status === 404 ? "/team" : undefined} onaction={loadError.status === 404 ? undefined : load} />
{:else if player}
  <header class="player-header">
    <div class="identity">
      <div class="portrait"><PlayerAvatar src={player.avatar_url} /></div>
      <div class="info">
        <h1>{name}</h1>
        <p class="player-status">{player.is_pro ? "Professional player" : player.profile_source === "steam" ? "Steam profile only" : "Player profile"}{#if player.real_name?.trim()}<span class="separator" aria-hidden="true">·</span><span class="real-name">{player.real_name}</span>{/if}</p>
        <ul class="metadata" aria-label="Player details">
          <li>Account #{player.account_id}</li>
          {#if countryText(player.country_code)}<li>{countryText(player.country_code)}</li>{/if}
          {#if fantasyRoleText(player.fantasy_role)}<li>{fantasyRoleText(player.fantasy_role)}</li>{/if}
          {#if team}<li>{#if team.href}<a href={team.href}>{team.name}</a>{:else}{team.name}{/if}</li>{:else if player.is_pro}<li>No team</li>{/if}
        </ul>
      </div>
    </div>
  </header>

  <section aria-labelledby="pro-title">
    <h2 id="pro-title">Professional profile</h2>
    {#if player.is_pro || player.profile_source !== "steam"}
      <dl class="facts">
        <div><dt>Team</dt><dd>{#if player.team_id > 0}{#if player.team_available}<a href={`/team/${player.team_id}`}>{player.team_name || `Team #${player.team_id}`}</a>{:else}{player.team_name || `Team #${player.team_id}`}{/if}{#if player.team_tag}<span class="tag">{player.team_tag}</span>{/if}{:else}—{/if}</dd></div>
        <div><dt>Role</dt><dd>{fantasyRoleText(player.fantasy_role) || "—"}</dd></div>
        <div><dt>Country</dt><dd>{countryText(player.country_code) || "—"}</dd></div>
        <div><dt>Recorded earnings</dt><dd>{player.total_earnings > 0 ? formatMoney(player.total_earnings) : "—"}</dd></div>
        {#if rosterDiffers}<div><dt>Current roster <small title="From the team's Valve roster, which can differ from the DPC player feed">(Valve roster)</small></dt><dd><a href={`/team/${roster.id}`}>{roster.name || `Team #${roster.id}`}</a>{#if roster.joined_at}<span class="tag">since {formatDate(roster.joined_at)}</span>{/if}</dd></div>{/if}
        {#if player.sponsor?.trim()}<div><dt>Sponsor</dt><dd>{player.sponsor}</dd></div>{/if}
        {#if player.is_locked}<div><dt>Roster</dt><dd>Locked</dd></div>{/if}
      </dl>
      {#if results.length}
        <!-- svelte-ignore a11y_no_noninteractive_tabindex (Keyboard users need to scroll the table horizontally.) -->
        <div class="table-scroll" role="region" aria-label="Tournament placements" tabindex="0"><table>
          <thead><tr><th scope="col">Tournament</th><th scope="col" class="num">Placement</th><th scope="col" class="num">Earnings</th></tr></thead>
          <tbody>{#each results as r}<tr><td>{#if r.league_available}<a href={`/league/${r.league_id}`}>{r.league_name || `Tournament #${r.league_id}`}</a>{:else}{r.league_name || `Tournament #${r.league_id}`}<small class="unavailable">Details unavailable</small>{/if}</td><td class="num">{r.placement || "—"}</td><td class="num">{formatMoney(r.earnings)}</td></tr>{/each}</tbody>
        </table></div>
      {/if}
    {:else}
      <StatePanel title="No professional profile" message="This player is not listed in the Dota Pro Circuit player feed. The profile below comes from Steam." />
    {/if}
  </section>

  <section aria-labelledby="steam-title">
    <h2 id="steam-title">Steam profile <span class="badge" class:muted={steam.kind !== 'public'} title="Visibility reported by Steam at the last check">{steam.label}</span></h2>
    {#if steam.kind === "private"}
      <p class="note">This Steam profile is {steam.label.toLowerCase()}. Steam still shows the name and avatar; location and other details are hidden.
        {#if player.steam_public_at}It was last seen public on {formatDateTime(millis(player.steam_public_at))} UTC, and the location below dates from then.{:else}It has not been seen public since collection started.{/if}
        {#if player.steam_checked_at}<span class="checked">Checked {formatDateTime(millis(player.steam_checked_at))} UTC.</span>{/if}</p>
    {:else if steam.kind === "unavailable"}
      <p class="note">Steam did not return this profile at the last check.
        {#if steam.hasData}Showing the data last collected{#if player.steam_updated_at} on {formatDateTime(millis(player.steam_updated_at))} UTC{/if}.{/if}
        {#if player.steam_checked_at}<span class="checked">Checked {formatDateTime(millis(player.steam_checked_at))} UTC.</span>{/if}</p>
    {/if}
    {#if steam.hasData}
      <dl class="facts">
        <div><dt>Steam name</dt><dd>{player.steam_name}</dd></div>
        <div><dt>Location <small title="Self-reported on Steam, not nationality">(self-reported)</small></dt><dd>{player.steam_location?.trim() || (steam.kind === "private" ? "Hidden" : "—")}</dd></div>
        {#if player.steam_updated_at}<div><dt>Last updated</dt><dd>{formatDateTime(millis(player.steam_updated_at))}</dd></div>{/if}
        {#if steamProfileUrl(player.account_id)}<div><dt>Profile</dt><dd><a class="external" href={steamProfileUrl(player.account_id)} rel="noopener noreferrer">Open on Steam<span class="external-arrow" aria-hidden="true">↗</span></a></dd></div>{/if}
      </dl>
    {:else}
      <StatePanel title={steam.kind === "unchecked" ? "Steam profile not collected yet" : "Steam profile unavailable"} message={steam.kind === "unchecked" ? "The Steam lookup for this account has not run yet." : "Steam returned no public data for this account."} actionLabel={steamProfileUrl(player.account_id) ? "Open on Steam" : ""} href={steamProfileUrl(player.account_id) || ""} />
    {/if}
  </section>

  <ChangeLog entities={["player"]} entityID={id} />
{/if}

<style>
  .player-header { margin: 1.5rem 0 2rem; padding-bottom: 1.5rem; border-bottom: 1px solid var(--line); }
  .identity { display: flex; align-items: center; gap: 1.75rem; }
  .portrait { width: 6rem; height: 6rem; flex: 0 0 6rem; border-radius: .5rem; overflow: hidden; background: var(--surface-subtle); }
  .portrait :global(.player-avatar) { width: 100%; height: 100%; flex-basis: 100%; border-radius: .5rem; }
  .info { min-width: 0; flex: 1; }
  h1 { margin: 0; font-size: clamp(1.75rem, 3.5vw, 2.75rem); line-height: 1.12; letter-spacing: -.035em; overflow-wrap: anywhere; text-wrap: balance; }
  .player-status { margin: .5rem 0 0; color: var(--muted); font-size: .875rem; }
  .separator { margin: 0 .5rem; } .real-name { color: var(--text); }
  .metadata { display: flex; flex-wrap: wrap; gap: .375rem 0; list-style: none; padding: 0; margin: .875rem 0 0; color: var(--muted); font-size: .8125rem; }
  .metadata li + li::before { content: "·"; padding: 0 .625rem; color: var(--muted); }
  section { margin: 32px 0; }
  h2 { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
  .badge { display: inline-block; background: var(--gold); color: var(--text); border-radius: 4px; padding: 2px 7px; font-size: 11px; font-weight: 500; letter-spacing: 0; }
  .badge.muted { background: var(--surface-subtle); color: var(--muted); }
  .facts { display: grid; grid-template-columns: repeat(auto-fit, minmax(11rem, 1fr)); gap: 16px 24px; margin: 0 0 20px; }
  .facts div { min-width: 0; display: grid; gap: 5px; }
  dt { color: var(--muted); font-size: 11px; } dt small { font-weight: 400; }
  dd { margin: 0; font-size: 14px; font-weight: 500; overflow-wrap: anywhere; }
  .tag { margin-left: 8px; color: var(--muted); font-weight: 400; }
  .note { color: var(--muted); font-size: 13px; margin: 0 0 16px; } .checked { display: block; margin-top: 4px; }
  .external { text-underline-offset: .2em; } .external-arrow { display: inline-block; margin-left: .25rem; }
  table { width:100%; border-collapse:collapse; } th,td { padding:12px; border-bottom:1px solid var(--line); text-align:left; } th { font-size:12px; color:var(--muted); white-space:nowrap; } .num { text-align:right; white-space:nowrap; }
  .unavailable { display: block; margin-top: 3px; color: var(--muted); font-size: 11px; }
  @media(max-width:600px) { .identity { gap: 1rem; align-items: flex-start; } .portrait { width: 4rem; height: 4rem; flex-basis: 4rem; } }
</style>
