<script>
  import SourceData from "./SourceData.svelte";
  import Breadcrumbs from "./Breadcrumbs.svelte";
  import { onMount, onDestroy } from 'svelte';
  import Schedule from './Schedule.svelte';
  import { SCHEDULE_PAGE_SIZE } from './lib/schedule.js';
  import HoverCommand from './HoverCommand.svelte';
  import EntityImage from './EntityImage.svelte';
  import StatePanel from './StatePanel.svelte';
  import { api, apiResults } from './lib/api.js';
  import { formatDate, regionText, tierText, formatMoney, normalizeUrl, leagueLogo } from './lib/constants.js';
  import { pageOffset } from './lib/routes.js';
  import { dotaTVUrl } from './lib/dotatv.js';
  export let id;
  let league = null;
  let loading = true;
  let loadError = null;
  let notFound = false;
  const query = new URLSearchParams(window.location.search);
  let tab = ['schedule','live','results'].includes(query.get('tab')) ? query.get('tab') : 'schedule';
  const initialOffset = pageOffset(window.location.search, SCHEDULE_PAGE_SIZE);
  let series = [];
  let seriesTotal = 0;
  let seriesLoaded = false;
  let loadingSeries = false;
  let seriesError = '';
  let liveGames = [];
  let loadingLive = false;
  let liveLoaded = false;
  let liveError = '';
  let liveTimer;
  let standings = [];
  let loadingStandings = false;
  let standingsLoaded = false;
  let standingsError = '';
  let disposed = false;
  const pageSize = SCHEDULE_PAGE_SIZE;
  async function loadLeague() {
    loading = true; loadError = null; notFound = false;
    try {
      const result = await api(`/leagues/${id}`);
      if(!result?.league_id) { notFound=true; league=null; return; }
      league = result;
      activateTab();
    } catch(err) { notFound=err.status===404; loadError=err.message; league=null; }
    finally { loading=false; }
  }
  onMount(() => { loadLeague(); return () => { disposed=true; }; });
  onDestroy(() => clearInterval(liveTimer));

  async function loadSeries() {
    if(loadingSeries) return;
    loadingSeries=true; seriesError='';
    try {
      const data=await apiResults(`/leagues/${id}/series`,{limit:pageSize,offset:initialOffset});
      series=data.results;
      seriesTotal=data.meta.total; seriesLoaded=true;
    } catch(err) {
      if(err.status===404) { series=[]; seriesLoaded=true; seriesTotal=0; }
      else seriesError=err.message;
    } finally { loadingSeries=false; }
  }
  async function loadLive() {
    if(loadingLive) return;
    loadingLive=true; liveError='';
    try { liveGames=(await apiResults(`/leagues/${id}/live-games`)).results; liveLoaded=true; }
    catch(err) { if(err.status===404) { liveGames=[]; liveLoaded=true; } else liveError=err.message; }
    finally { loadingLive=false; }
  }
  async function loadStandings() {
    if(loadingStandings) return;
    loadingStandings=true; standingsError='';
    try {
      const data=await api(`/leagues/${id}/results`);
      standings=Array.isArray(data?.results?.results) ? data.results.results : [];
      standingsLoaded=true;
    } catch(err) { standingsError=err.message; }
    finally { loadingStandings=false; }
  }
  function activateTab() {
    clearInterval(liveTimer);
    if(disposed) return;
    if(tab==='schedule' && !seriesLoaded) loadSeries();
    if(tab==='live') { loadLive(); liveTimer=setInterval(loadLive,60000); }
    if(tab==='results' && !standingsLoaded) loadStandings();
  }
  function selectTab(next) {
    tab=next;
    const url=new URL(window.location.href);
    if(next==='schedule') url.searchParams.delete('tab'); else url.searchParams.set('tab',next);
    window.history.replaceState(null,'',url);
    activateTab();
  }
  function teamName(name, id) { return name || (id ? `Team #${id}` : 'To be announced'); }
</script>
<svelte:head><title>{notFound ? 'League not found' : league?.name || 'League'} | Dota 2 Leagues</title>{#if notFound}<meta name="robots" content="noindex" />{/if}</svelte:head>
<Breadcrumbs items={[{label:'Leagues',href:'/'},{label:league?.name || `League #${id}`}]} />
{#if loading}<StatePanel kind="loading" title="Loading league" message="Finding the tournament and its schedule." />
{:else if notFound}<h1>League not found</h1><StatePanel kind="notFound" title="This league isn’t available" message="Check the league link or browse the current tournaments." actionLabel="Browse leagues" href="/" />
{:else if loadError}<h1>League unavailable</h1><StatePanel kind="error" title="Couldn’t load this league" message={loadError} actionLabel="Try again" onaction={loadLeague} />
{:else if league}
  <div class="banner profile-header">
    <div class="league-art"><EntityImage src={leagueLogo(id)} name={league.name} wide contain /></div>
    <div class="info"><p class="eyebrow">{regionText(league.region)} · {tierText(league.tier)}{#if league.is_live}<span class="badge">● Live now</span>{/if}</p><h1>{league.name || `League #${id}`}</h1>
      <div class="facts"><div><span>Dates (UTC)</span><strong>{formatDate(league.start_timestamp)} – {formatDate(league.end_timestamp)}</strong></div><div><span>Prize pool</span><strong>{formatMoney(league.total_prize_pool)}</strong></div></div>
      {#if normalizeUrl(league.url)}<a class="ext" href={normalizeUrl(league.url)} rel="noopener noreferrer">Tournament website ↗</a>{/if}
    </div>
  </div>
  {#if league.description}<p class="description">{league.description}</p>{/if}
  <HoverCommand leagueId={id} />
  <nav class="tabs" aria-label="League sections">
    {#each [['schedule','Schedule'],['live','Live games'],['results','DPC results']] as [value,label]}
      <button class:active={tab===value} aria-pressed={tab===value} onclick={() => selectTab(value)}>{label}</button>
    {/each}
  </nav>
  {#if tab==='schedule'}
    <Schedule {id} {series} total={seriesTotal} offset={initialOffset} loading={loadingSeries} error={seriesError} onretry={loadSeries} />
  {:else if tab==='live'}
    <section aria-label="Live games">
      <div class="refresh-row"><p class="muted">Updates every minute while this section is open.</p><button class="secondary" onclick={loadLive} disabled={loadingLive}>{loadingLive?'Refreshing…':'Refresh'}</button></div>
      {#if loadingLive && !liveLoaded}<StatePanel kind="loading" title="Checking live games" />{/if}
      {#if liveError}<StatePanel kind="error" title="Live games are unavailable" message={liveError} actionLabel="Try again" onaction={loadLive} />{/if}
      {#if liveGames.length}<div class="live-games">{#each liveGames as game}<article class="live-game"><p class="eyebrow">● Live game</p><div class="live-match">{#if game.team1_id}<a href="/team/{game.team1_id}">{teamName(game.team1_name,game.team1_id)}</a>{:else}<strong>{game.team1_name || 'Radiant'}</strong>{/if}<span class="vs">vs</span>{#if game.team2_id}<a href="/team/{game.team2_id}">{teamName(game.team2_name,game.team2_id)}</a>{:else}<strong>{game.team2_name || 'Dire'}</strong>{/if}</div><div class="live-actions">{#if game.spectators>0}<small>{game.spectators.toLocaleString()} watching</small>{/if}
        </div>{#if dotaTVUrl(game.game_id)}{#key game.game_id}<HoverCommand serverSteamId={game.game_id} />{/key}{/if}</article>{/each}</div>
      {:else if liveLoaded && !liveError}<StatePanel title="No live games right now" message="Check the schedule for matchups, or come back when a game starts." actionLabel="View schedule" onaction={() => selectTab('schedule')} />{/if}
    </section>
  {:else}
    <section aria-label="DPC results">
      {#if loadingStandings}<StatePanel kind="loading" title="Loading DPC results" message="Checking tournament standings and earnings." />
      {:else if standingsError}<StatePanel kind="error" title="Couldn’t load DPC results" message={standingsError} actionLabel="Try again" onaction={loadStandings} />
      {:else if standings.length}
      <!-- svelte-ignore a11y_no_noninteractive_tabindex (The wide results table must be keyboard-scrollable.) -->
      <div class="table-scroll" role="region" aria-label="DPC results table" tabindex="0"><table><thead><tr><th>Place</th><th>Team</th><th>Points</th><th>Earnings</th></tr></thead><tbody>{#each standings as row}<tr><td>{row.standing || '—'}</td><td>{#if row.team_id}<a href="/team/{row.team_id}">{teamName(row.team_name,row.team_id)}</a>{:else}{row.team_name || 'Unknown team'}{/if}</td><td>{row.points ?? '—'}</td><td>{formatMoney(row.earnings)}</td></tr>{/each}</tbody></table></div>
      {:else if standingsLoaded}<StatePanel title="No DPC results available" message="This league has no results in the DPC feed. Its game results may still be available in the schedule." actionLabel="View schedule" onaction={() => selectTab('schedule')} />{/if}
    </section>
  {/if}
{/if}
{#if league}<SourceData kind="league" entityID={id} />{/if}
<style>
 .banner { display: flex; gap: 28px; align-items: center; }
 .league-art { width: 210px; aspect-ratio: 1.8; flex-shrink: 0; border: 1px solid var(--line); border-radius: 9px; overflow: hidden; }
 .info { min-width: 0; flex: 1; } .eyebrow { line-height: 1.7; } .badge { display: inline-block; margin-left: 10px; background: var(--gold); color: var(--text); border-radius: 4px; padding: 2px 7px; font-size: 10px; letter-spacing: 0; }
 .facts { display: flex; flex-wrap: wrap; gap: 16px 36px; margin: 18px 0 12px; } .facts div { display: grid; gap: 5px; } .facts span { font-size: 11px; color: var(--muted); } .facts strong { font-size: 13px; font-weight: 500; }
 .ext { font-size: 12px; } .description { max-width: 900px; color: var(--muted); font-size: 13px; margin: 24px 0 16px; white-space: pre-line; overflow-wrap: anywhere; }
 .tabs { display: flex; gap: 4px; border-bottom: 1px solid var(--line-strong); margin: 28px 0 22px; }
 .tabs button { border: 0; border-bottom: 3px solid transparent; border-radius: 0; padding: 12px 20px; background: none; color: var(--muted); font-size: 13px; }
 .tabs button.active { border-color: var(--accent); color: var(--text); } .tabs button:hover { background: var(--surface-subtle); }
 .live-actions { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 12px; margin-top: 16px; }
 .vs { color: var(--muted); font-size: 11px; }
 .refresh-row { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 16px; } .refresh-row p { font-size: 12px; }
 .live-games { display: grid; grid-template-columns: repeat(2,minmax(0,1fr)); gap: 16px; } .live-game { background: var(--surface); padding: 18px; border: 1px solid var(--line); border-radius: 8px; min-width: 0; } .live-match { display: flex; flex-wrap: wrap; gap: 12px; align-items: center; } .live-match a,.live-match strong { font-size: 15px; overflow-wrap: anywhere; } small { color: var(--muted); font-size: 11px; }
 @media(max-width: 750px) { .banner { align-items: flex-start; flex-direction: column; gap: 20px; } .league-art { width: 100%; aspect-ratio: 2.5; } .tabs button { flex: 1; padding: 12px 8px; } .live-games { grid-template-columns: 1fr; } }
</style>
