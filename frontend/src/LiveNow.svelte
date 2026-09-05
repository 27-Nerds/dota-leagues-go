<script>
  import { onMount, onDestroy } from 'svelte';
  import { apiResults } from './lib/api.js';
  let liveLeagues = [];
  let total = 0;
  let catalogError = false;
  let entries = [];
  let loading = false;
  let failures = 0;
  let expanded = false;
  let request = 0;

  async function refresh() {
    const current = ++request;
    loading = true; catalogError = false;
    try {
      const data = await apiResults('/leagues', {status:'all', live:true, limit:100, offset:0});
      if (current !== request) return;
      liveLeagues = data.results; total = data.meta.total;
    } catch { if (current === request) { catalogError = true; loading = false; } return; }
    const settled = await Promise.allSettled(liveLeagues.slice(0,3).map(async league => {
      try { return {league,games:(await apiResults(`/leagues/${league.league_id}/live-games`)).results}; }
      catch(err) { if(err.status === 404) return {league,games:[]}; throw err; }
    }));
    if(current !== request) return;
    failures = settled.filter(s => s.status === 'rejected').length;
    entries = liveLeagues.map((league, index) => settled[index]?.status === 'fulfilled' ? settled[index].value : { league, games: null, failed: settled[index]?.status === 'rejected' });
    loading = false;
  }
  onMount(refresh);
  onDestroy(() => ++request);
</script>
{#if liveLeagues.length || catalogError}
<section class="live-strip" aria-label="Live games">
  <div class="strip-heading"><h2><span aria-hidden="true">●</span> Live now</h2><span>{#if catalogError}Across all leagues{:else}Across all leagues · {total} {total === 1 ? 'league' : 'leagues'} reporting live activity{/if}</span></div>
  {#if catalogError}<p class="notice" role="alert">Couldn’t check live activity. <button class="secondary" onclick={refresh}>Try again</button></p>
  {:else if loading && !entries.length}<p role="status">Checking live games…</p>
  {:else if failures && !entries.length}<p role="alert">Live updates are unavailable. <button class="secondary" onclick={refresh}>Try again</button></p>
  {:else}
  <div class="live-grid">
    {#each (expanded ? entries : entries.slice(0,3)) as entry (entry.league.league_id)}
      <a class="live-card" href="/league/{entry.league.league_id}?tab=live">
        <span class="league-name">{entry.league.name}</span>
        {#if entry.games?.length}
        {@const game = entry.games[0]}
        <strong>{game.team1_name || 'Radiant'} <span>vs</span> {game.team2_name || 'Dire'}</strong>
        <small>{game.spectators > 0 ? `${game.spectators.toLocaleString()} watching` : 'View live games'}{#if entry.games.length > 1} · +{entry.games.length - 1} more{/if}</small>
        {:else}<small>{entry.failed ? 'Live updates unavailable — view league' : entry.games ? 'No live games reported yet' : 'View live games →'}</small>{/if}
      </a>
    {/each}
  </div>
  {#if failures}<p class="notice">Some live updates are unavailable. <button class="secondary" onclick={refresh}>Retry updates</button></p>{/if}
  {#if entries.length > 3}<button class="expand" aria-expanded={expanded} onclick={() => expanded = !expanded}>{expanded ? 'Show fewer' : `Show ${entries.length} live leagues`} <span aria-hidden="true">{expanded ? '−' : '+'}</span></button>{/if}
  {/if}
</section>
{/if}
<style>
  .live-strip { border: 1px solid #ddd6bb; background: var(--accent-soft); padding: 18px; border-radius: 9px; }
  .strip-heading { display: flex; flex-wrap: wrap; gap: 10px; align-items: center; margin-bottom: 14px; }
  h2 { font-size: 12px; text-transform: uppercase; letter-spacing: 1.3px; margin: 0; } h2 span { color: var(--accent); margin-right: 5px; }
  .strip-heading > span { font-size: 11px; color: var(--muted); }
  .live-grid { display: grid; grid-template-columns: repeat(3,minmax(0,1fr)); gap: 12px; }
  .live-card { min-width: 0; display: grid; gap: 5px; text-decoration: none; padding: 12px; background: var(--surface); border: 1px solid var(--line); border-radius: 6px; }
  .live-card:hover { border-color: var(--line-strong); }
  .league-name { font-size: 11px; color: var(--muted); overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
  strong { font-size: 13px; overflow-wrap: anywhere; } strong span { font-size: 10px; color: var(--muted); padding: 0 3px; font-weight: 400; }
  small { color: var(--muted); font-size: 10px; }
  .expand { display: flex; width: 100%; justify-content: space-between; font-size: 11px; background: none; border: 0; color: var(--muted); min-height: 36px; padding: 12px 0 0; }
  .notice { font-size: 12px; }
  @media(max-width: 650px) { .live-grid { grid-template-columns: 1fr; } .live-strip { padding: 14px; } }
</style>
