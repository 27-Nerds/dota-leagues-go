<script>
  import PageHeader from "./PageHeader.svelte";
  import Breadcrumbs from "./Breadcrumbs.svelte";
  import { onMount } from "svelte";
  import PlayerRow from "./PlayerRow.svelte";
  import StatePanel from "./StatePanel.svelte";
  import EntityImage from "./EntityImage.svelte";
  import { api } from "./lib/api.js";
  import { matchOutcome } from "./lib/match-outcome.js";
  import { formatDateTime, teamLogo } from "./lib/constants.js";
  export let leagueId;
  export let matchId;
  const initialTitle = document.title;
  let match = null;
  let loadError = null;
  let loading = true;
  let sides = [];
  $: outcome = matchOutcome(match?.match_outcome);
  async function load() {
    loading = true; loadError = null;
    try {
      const body = await api(`/leagues/${leagueId}/matches/${matchId}/minimal`);
      const result = body?.results;
      if (!result?.match_id) throw Object.assign(new Error("This match is not available."), { status:404 });
      const tourney = result.tourney || {};
      const players = (Array.isArray(result.players) ? result.players : []).filter(Boolean);
      sides = [{key:"radiant", label:"Radiant", id:tourney.radiant_team_id, logo:tourney.radiant_team_logo_url, name:tourney.radiant_team_name || "Radiant team", score:result.radiant_score, players:players.filter(p => p.team_number === 0).sort((a,b) => a.player_slot-b.player_slot)}, {key:"dire", label:"Dire", id:tourney.dire_team_id, logo:tourney.dire_team_logo_url, name:tourney.dire_team_name || "Dire team", score:result.dire_score, players:players.filter(p => p.team_number === 1).sort((a,b) => a.player_slot-b.player_slot)}];
      match = result;
    } catch(err) { loadError = err; }
    finally { loading = false; }
  }
  onMount(load);
</script>
<svelte:head><title>{match ? `${sides[0].name} vs ${sides[1].name} - Match ${matchId} | Dota 2 Leagues` : initialTitle}</title>{#if loadError?.status === 404}<meta name="robots" content="noindex" />{/if}</svelte:head>
<Breadcrumbs items={[{label:'Leagues',href:'/'},{label:`League #${leagueId}`,href:`/league/${leagueId}`},{label:`Match #${matchId}`}]} />
<PageHeader title={match ? `${sides[0].name} vs ${sides[1].name}` : `Match #${matchId}`} />
{#if loading}<StatePanel kind="loading" title="Loading match" message="Fetching the scoreboard and player statistics." />
{:else if loadError}<StatePanel kind={loadError.status === 404 ? "notFound" : "error"} title={loadError.status === 404 ? "Match not found" : "Unable to load match"} message={loadError.message} actionLabel={loadError.status === 404 ? "Back to league" : "Try again"} href={loadError.status === 404 ? `/league/${leagueId}` : undefined} onaction={loadError.status === 404 ? undefined : load} />
{:else if match}
  <div class="match-summary">
  <div class="scoreboard">
    {#each sides as side, i}
      <div class="team" class:dire={side.key === "dire"}>
        <div class="logo"><EntityImage src={side.logo || teamLogo(side.id)} name={side.name} /></div>
        {#if side.id > 0}<a href={`/team/${side.id}`}>{side.name}</a>{:else}<span>{side.name}</span>{/if}
        <span class="side-label">{side.label}{#if outcome.winner === side.key}<strong class="winner">Winner</strong>{:else if outcome.winner}<span>Defeat</span>{/if}</span>
      </div>
      {#if i === 0}<div class="score"><span class="score-label">Kills</span><div class="nums"><b>{sides[0].score ?? "-"}</b><span>:</span><b>{sides[1].score ?? "-"}</b></div><div class="muted">{match.duration > 0 ? `${Math.floor(match.duration / 60)}m ${match.duration % 60}s` : "Duration unavailable"}</div></div>{/if}
    {/each}
  </div>
  {#if !outcome.winner}<p class="result-status">{outcome.label}</p>{/if}
  <div class="match-context">
    <p class="match-date">{match.start_time ? `${formatDateTime(match.start_time)} UTC` : "Match date unavailable"}</p>
    <span class="match-reference">Match ID <code class="match-id">{matchId}</code></span>
  </div>
  </div>
  <div class="columns">{#each sides as side}
    <section class="team-col" class:radiant={side.key === "radiant"} class:dire={side.key === "dire"} aria-label={`${side.label} players`}>
      <h2>{side.label}<span>K / D / A</span></h2>
      {#if side.players.length}{#each side.players as player}<PlayerRow p={player} />{/each}
      {:else}<StatePanel title="No player statistics" message="Player details are not available for this side." />{/if}
    </section>
  {/each}</div>
{/if}
<style>
  .match-summary { margin: 0 0 24px; border: 1px solid var(--line); border-radius: var(--radius); background: var(--surface); }
  .scoreboard { display: grid; grid-template-columns: minmax(0,1fr) 180px minmax(0,1fr); align-items: center; gap: 24px; padding: 24px 32px; }
  .team { display: grid; grid-template-columns: 56px minmax(0,1fr); column-gap: 16px; row-gap: 4px; align-items: center; min-width: 0; font-size: 20px; font-weight: 500; overflow-wrap: anywhere; }
  .team a { text-decoration: none; } .team a:hover { text-decoration: underline; }
  .logo { width: 56px; height: 56px; grid-row: span 2; border-radius: var(--radius-sm); }
  .side-label { display: flex; align-items: baseline; flex-wrap: wrap; gap: 4px 10px; color: var(--muted); font-size: 12px; font-weight: 400; }
  .winner { color: var(--positive); font-weight: 600; }
  .team.dire .side-label { justify-content: flex-end; }
  .result-status { margin: -8px 16px 16px; text-align: center; color: var(--muted); font-size: 12px; }
  .team.dire { grid-template-columns: minmax(0,1fr) 56px; text-align: right; }
  .team.dire .logo { grid-column: 2; grid-row: 1 / 3; } .team.dire a, .team.dire > span { grid-column: 1; } .team.dire .side-label { grid-row: 2; }
  .score { text-align: center; } .score-label { color: var(--muted); font-size: 11px; }
  .nums { display: flex; justify-content: center; gap: 14px; align-items: center; font-family: 'Fira Code',monospace; font-size: 36px; line-height: 1.4; font-variant-numeric: tabular-nums; }
  .nums b { font-weight: 500; } .nums b:first-child { color: var(--positive); } .nums b:last-child { color: var(--negative); } .nums span { color: var(--text-subtle); font-size: 24px; }
  .score .muted { font-size: 12px; }
  .match-context { display: flex; justify-content: space-between; align-items: baseline; flex-wrap: wrap; gap: 12px 24px; border-top: 1px solid var(--line); padding: 12px 20px; }
  .match-date { color: var(--muted); margin: 0; font-size: 12px; }
  .match-id { color: var(--muted); font-size: 12px; font-variant-numeric: tabular-nums; user-select: all; }
  .match-reference { display: inline-flex; align-items: baseline; gap: 8px; color: var(--muted); font-size: 12px; }
  .columns { display: grid; grid-template-columns: repeat(2,minmax(0,1fr)); gap: 24px; }
  .team-col { background: var(--surface); min-width: 0; border: 1px solid var(--line); border-radius: var(--radius); overflow: hidden; }
  .team-col h2 { display: flex; align-items: center; justify-content: space-between; margin: 0; padding: 12px 16px; font-size: 13px; font-weight: 500; background: var(--surface-subtle); color: var(--text); }
  .team-col h2 span { color: var(--muted); font-size: 11px; font-family: 'Fira Code',monospace; }
  .radiant h2 { border-left: 3px solid var(--positive); } .dire h2 { border-left: 3px solid var(--negative); }
  @media(max-width:1000px) { .columns { grid-template-columns: 1fr; } }
  @media(max-width:767px) {
    .scoreboard { gap: 12px; padding: 20px 12px; grid-template-columns: minmax(0,1fr) 96px minmax(0,1fr); }
    .team, .team.dire { display: flex; flex-direction: column; align-self: start; text-align: center; gap: 6px; font-size: 14px; }
    .team .side-label, .team.dire .side-label { justify-content: center; }
    .logo { width: 44px; height: 44px; } .nums { font-size: 28px; gap: 6px; }
    .score .muted { font-size: 11px; } .match-context { padding: 12px; }
  }
</style>
