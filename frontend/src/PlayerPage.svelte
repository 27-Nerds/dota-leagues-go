<script>
  import PlayerAvatar from "./PlayerAvatar.svelte";
  import Breadcrumbs from "./Breadcrumbs.svelte";
  import StatePanel from "./StatePanel.svelte";
  import { onMount } from "svelte";
  import { api } from "./lib/api.js";
  import ChangeLog from "./ChangeLog.svelte";
  import { countryText, fantasyRoleText, formatDate, formatDateTime, formatMoney, steamProfileUrl } from "./lib/constants.js";
  import { steamState, playerDisplayName, isListedPro, proRegistration, proRegistrations, teamHistory } from "./lib/players.js";
  export let id;
  let player = null;
  let loadError = null;
  let loading = true;
  $: name = playerDisplayName(player, id);
  $: steam = steamState(player);
  $: results = (Array.isArray(player?.results) ? player.results : []).filter(r => r && r.league_id > 0);
  $: listed = isListedPro(player);
  $: registration = proRegistration(player);
  $: registrations = proRegistrations(player);
  $: history = teamHistory(player);
  $: staleFeed = listed && player?.dpc_seen_at > 0 && Date.now() - Number(player.dpc_seen_at) > 7 * 86400000;
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
<Breadcrumbs items={[{label:'Players',href:'/player'}, {label:name}]} />
{#if loading || loadError}<h1>{loadError?.status === 404 ? "Player not found" : "Player profile"}</h1>{/if}
{#if loading}
  <StatePanel kind="loading" title="Loading player" message="Fetching the professional and Steam profile." />
{:else if loadError}
  <StatePanel kind={loadError.status === 404 ? "notFound" : "error"} title={loadError.status === 404 ? "Player not found" : "Unable to load player"} message={loadError.message} actionLabel={loadError.status === 404 ? "Browse players" : "Try again"} href={loadError.status === 404 ? "/player" : undefined} onaction={loadError.status === 404 ? undefined : load} />
{:else if player}
  <header class="player-header profile-header">
    <div class="portrait" aria-hidden="true"><span class="initial">{name.trim().charAt(0).toUpperCase() || "?"}</span><PlayerAvatar src={player.avatar_url} /></div>
    <div class="info">
      <p class="eyebrow">{listed ? "Professional player" : "Steam profile only"}{#if steam.hasData || steam.kind !== "unchecked"}<span class="status" class:public={steam.kind === "public"} title="Steam profile visibility at the last check"><span class="dot" aria-hidden="true"></span>Steam {steam.label.toLowerCase()}</span>{/if}</p>
      <h1>{name}</h1>
      {#if player.real_name?.trim()}<p class="real-name">{player.real_name}</p>{/if}
      <dl class="meta" aria-label="Player details">
        <div><dt>Team</dt><dd>{#if team}{#if team.href}<a href={team.href}>{team.name}</a>{:else}{team.name}{/if}{#if team.tag && team.tag !== team.name}<span class="tag">{team.tag}</span>{/if}{:else}{listed ? "No team" : "—"}{/if}</dd></div>
        <div><dt>Role</dt><dd>{fantasyRoleText(player.fantasy_role) || "—"}</dd></div>
        <div><dt>Country</dt><dd>{countryText(player.country_code) || "—"}</dd></div>
        <div><dt>Account</dt><dd>{player.account_id}</dd></div>
      </dl>
    </div>
  </header>

  <div class="profile-grid">
    <section class="column" aria-labelledby="pro-title">
      <h2 id="pro-title">Professional profile</h2>
      {#if listed}
        {#if staleFeed}<p class="callout">Not present in Valve’s pro player feed since {formatDateTime(millis(player.dpc_seen_at))} UTC. The details below are from that last listing.</p>{/if}
        {#if registration || rosterDiffers || player.sponsor?.trim() || player.is_locked || player.total_earnings > 0}
          <dl class="facts">
            {#if registration}<div><dt>Pro registration</dt><dd>Period {registration.registration_period}<span class="tag">registered {formatDate(registration.timestamp)}</span></dd></div>{/if}
            {#if rosterDiffers}<div><dt>Current roster <small title="From the team's Valve roster, which can differ from the DPC player feed">(Valve roster)</small></dt><dd><a href={`/team/${roster.id}`}>{roster.name || `Team #${roster.id}`}</a>{#if roster.joined_at}<span class="tag">since {formatDate(roster.joined_at)}</span>{/if}</dd></div>{/if}
            {#if player.total_earnings > 0}<div><dt>Recorded earnings</dt><dd>{formatMoney(player.total_earnings)}</dd></div>{/if}
            {#if player.sponsor?.trim()}<div><dt>Sponsor</dt><dd>{player.sponsor}</dd></div>{/if}
            {#if player.is_locked}<div><dt>Roster</dt><dd>Locked</dd></div>{/if}
          </dl>
        {/if}
        {#if history.length}
          <h3>Team history <small title="Valve’s per-player membership log from the pro player feed">Valve</small></h3>
          <!-- svelte-ignore a11y_no_noninteractive_tabindex (Keyboard users need to scroll the table horizontally.) -->
          <div class="table-scroll" role="region" aria-label="Team history" tabindex="0"><table>
            <thead><tr><th scope="col">Team</th><th scope="col" class="date">Since (UTC)</th></tr></thead>
            <tbody>{#each history as h}<tr><td>{#if h.team_available}<a href={`/team/${h.team_id}`}>{h.team_name || `Team #${h.team_id}`}</a>{:else}{h.team_name || `Team #${h.team_id}`}{/if}{#if h.team_tag && h.team_tag !== h.team_name}<span class="tag">{h.team_tag}</span>{/if}</td><td class="date">{formatDate(h.start_timestamp)}</td></tr>{/each}</tbody>
          </table></div>
        {/if}
        {#if registrations.length}
          <h3>Registration history <small title="Valve’s pro circuit registration windows, newest first">Valve</small></h3>
          <!-- svelte-ignore a11y_no_noninteractive_tabindex (Keyboard users need to scroll the table horizontally.) -->
          <div class="table-scroll" role="region" aria-label="Registration history" tabindex="0"><table>
            <thead><tr><th scope="col">Registration period</th><th scope="col" class="date">Registered (UTC)</th></tr></thead>
            <tbody>{#each registrations as r}<tr><td>Period {r.registration_period}</td><td class="date">{formatDate(r.timestamp)}</td></tr>{/each}</tbody>
          </table></div>
        {/if}
        {#if results.length}
          <h3>Tournament placements</h3>
          <!-- svelte-ignore a11y_no_noninteractive_tabindex (Keyboard users need to scroll the table horizontally.) -->
          <div class="table-scroll" role="region" aria-label="Tournament placements" tabindex="0"><table>
            <thead><tr><th scope="col">Tournament</th><th scope="col" class="num">Placement</th><th scope="col" class="num">Earnings</th></tr></thead>
            <tbody>{#each results as r}<tr><td>{#if r.league_available}<a href={`/league/${r.league_id}`}>{r.league_name || `Tournament #${r.league_id}`}</a>{:else}{r.league_name || `Tournament #${r.league_id}`}<small class="unavailable">Details unavailable</small>{/if}</td><td class="num">{r.placement || "—"}</td><td class="num">{formatMoney(r.earnings)}</td></tr>{/each}</tbody>
          </table></div>
        {/if}
        {#if !history.length && !registrations.length && !results.length && !registration}
          <p class="callout">Valve’s pro feed lists this player without registration or team history yet.</p>
        {/if}
      {:else}
        <StatePanel title="No professional profile" message="This player is not listed in the Dota Pro Circuit player feed. The profile comes from Steam." />
      {/if}
    </section>

    <aside class="column side" aria-labelledby="steam-title">
      <h2 id="steam-title">Steam profile</h2>
      <div class="panel">
        {#if steam.kind === "private"}
          <p class="callout">This Steam profile is {steam.label.toLowerCase()}. Steam still shows the name and avatar; location and other details are hidden.
            {#if player.steam_public_at}Last seen public on {formatDateTime(millis(player.steam_public_at))} UTC; the location below dates from then.{:else}Not seen public since collection started.{/if}</p>
        {:else if steam.kind === "unavailable"}
          <p class="callout">Steam did not return this profile at the last check.{#if steam.hasData} Showing the data last collected{#if player.steam_updated_at} on {formatDateTime(millis(player.steam_updated_at))} UTC{/if}.{/if}</p>
        {/if}
        {#if steam.hasData}
          <dl class="facts stack">
            <div><dt>Steam name</dt><dd>{player.steam_name}</dd></div>
            <div><dt>Location <small title="Self-reported on Steam, not nationality">self-reported</small></dt><dd>{player.steam_location?.trim() || (steam.kind === "private" ? "Hidden" : "—")}</dd></div>
            {#if player.steam_updated_at}<div><dt>Last updated</dt><dd>{formatDateTime(millis(player.steam_updated_at))} UTC</dd></div>{/if}
            {#if player.steam_checked_at}<div><dt>Last checked</dt><dd>{formatDateTime(millis(player.steam_checked_at))} UTC</dd></div>{/if}
          </dl>
          {#if steamProfileUrl(player.account_id)}<a class="button secondary external" href={steamProfileUrl(player.account_id)} rel="noopener noreferrer">Open on Steam<span aria-hidden="true">↗</span></a>{/if}
        {:else}
          <p class="callout">{steam.kind === "unchecked" ? "The Steam lookup for this account has not run yet." : "Steam returned no public data for this account."}</p>
          {#if steamProfileUrl(player.account_id)}<a class="button secondary external" href={steamProfileUrl(player.account_id)} rel="noopener noreferrer">Open on Steam<span aria-hidden="true">↗</span></a>{/if}
        {/if}
      </div>
    </aside>
  </div>

  <ChangeLog entities={["player"]} entityID={id} />
{/if}

<style>
  /* Header: portrait, eyebrow, name, labeled facts. Shares the profile-header divider with the league page. */
  .player-header { display: flex; align-items: flex-start; gap: 24px; margin: 4px 0 var(--space-section); }
  .portrait { position: relative; width: 96px; height: 96px; flex: 0 0 96px; border-radius: var(--radius); overflow: hidden; background: var(--surface-subtle); border: 1px solid var(--line); }
  .initial { position: absolute; inset: 0; display: grid; place-items: center; font-size: 34px; font-weight: 600; color: var(--text-subtle); }
  .portrait :global(.player-avatar) { position: relative; width: 100%; height: 100%; flex-basis: 100%; border-radius: 0; }
  .info { min-width: 0; flex: 1; padding-top: 4px; }
  .eyebrow { display: flex; flex-wrap: wrap; align-items: center; gap: 6px 14px; margin-bottom: 6px; }
  .status { display: inline-flex; align-items: center; gap: 6px; color: var(--muted); letter-spacing: 0; text-transform: none; }
  .dot { width: 7px; height: 7px; border-radius: 50%; background: var(--line-strong); }
  .status.public .dot { background: var(--positive); }
  h1 { margin: 0; }
  .real-name { margin: -4px 0 0; color: var(--muted); font-size: 14px; line-height: 1.5; }
  .meta { display: flex; flex-wrap: wrap; gap: 10px 32px; margin: 16px 0 0; }
  .meta div, .facts div { display: grid; gap: 4px; min-width: 0; }
  dt { color: var(--muted); font-size: 11px; font-weight: 500; letter-spacing: .2px; }
  dt small { font-weight: 400; color: var(--text-subtle); }
  dd { margin: 0; font-size: 14px; font-weight: 500; line-height: 1.4; overflow-wrap: anywhere; font-variant-numeric: tabular-nums; }
  .tag { margin-left: 8px; color: var(--muted); font-weight: 400; font-size: 12px; }

  /* Two columns on wide screens: the professional record leads, Steam sits beside it. */
  .profile-grid { display: grid; grid-template-columns: minmax(0, 1fr) 300px; gap: 0 40px; align-items: start; }
  .column h2 { margin-top: 0; }
  .side .panel { background: var(--surface); border: 1px solid var(--line); border-radius: var(--radius); padding: 18px; }
  .facts { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 14px 24px; margin: 0 0 24px; }
  .facts.stack { grid-template-columns: 1fr; gap: 14px; margin-bottom: 18px; }
  .callout { margin: 0 0 16px; padding: 10px 12px; border-left: 3px solid var(--gold); background: var(--accent-soft); border-radius: 0 var(--radius-sm) var(--radius-sm) 0; color: var(--text); font-size: 13px; line-height: 1.55; }
  .side .callout { margin-bottom: 14px; }
  h3 { display: flex; align-items: baseline; gap: 8px; font-size: 14px; font-weight: 600; letter-spacing: -.1px; margin: 24px 0 10px; }
  h3 small { font-weight: 500; font-size: 10px; letter-spacing: .5px; text-transform: uppercase; color: var(--text-subtle); }
  .table-scroll + h3 { margin-top: 28px; }
  .date, .num { white-space: nowrap; } .num { text-align: right; } td.date { color: var(--muted); }
  .unavailable { display: block; margin-top: 3px; color: var(--muted); font-size: 11px; }
  .external { width: 100%; } .external span { transition: transform 160ms ease; } .external:hover span { transform: translate(2px, -2px); }

  @media (max-width: 900px) { .profile-grid { grid-template-columns: 1fr; } .side { margin-top: 8px; } .side h2 { margin-top: 24px; } }
  @media (max-width: 600px) {
    .player-header { gap: 16px; } .portrait { width: 64px; height: 64px; flex-basis: 64px; } .initial { font-size: 24px; }
    .meta { gap: 10px 22px; } .profile-grid { gap: 0; } .side .panel { padding: 14px; }
  }
  @media (prefers-reduced-motion: reduce) { .external span { transition: none; } }
</style>
