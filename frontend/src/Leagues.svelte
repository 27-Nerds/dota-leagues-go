<script>
  import FilterToolbar from "./FilterToolbar.svelte";
  import SearchField from "./SearchField.svelte";
  import Pagination from "./Pagination.svelte";
  import PageHeader from "./PageHeader.svelte";
  import { onMount, onDestroy } from 'svelte';
  import League from './League.svelte';
  import LiveNow from './LiveNow.svelte';
  import StatePanel from './StatePanel.svelte';
  import { apiResults } from './lib/api.js';
  import { REGIONS, TIERS } from './lib/constants.js';
  import { pageOffset } from './lib/routes.js';
  const pageSize = 100;
  const query = new URLSearchParams(window.location.search);
  let offset = pageOffset(window.location.search);
  let status = ['active', 'completed', 'all'].includes(query.get('status')) ? query.get('status') : 'active';
  let search = query.get('search') || '';
  let regionFilter = query.get('region') || '';
  let tierFilter = query.get('tier') || '';
  let liveOnly = query.get('live') === 'true';
  let sorting = `${query.get('sort') || 'recommended'}:${query.get('order') || (query.get('sort') === 'name' ? 'asc' : 'desc')}`;
  let leagues = [];
  let total = 0;
  let loading = true;
  let loadError = null;
  let timer;
  let requestId = 0;
  function params(nextOffset) {
    const [sort, order] = sorting.split(':');
    return {status, sort, order, search: search.trim() || undefined, region: regionFilter || undefined, tier: tierFilter || undefined, live: liveOnly || undefined, offset: nextOffset, limit: pageSize};
  }
  function pageUrl(nextOffset) {
    const values = params(nextOffset);
    const qs = new URLSearchParams();
    for (const [key, value] of Object.entries(values)) {
      if (key !== 'limit' && value !== undefined && !(key === 'offset' && value === 0)) qs.set(key, String(value));
    }
    return `/?${qs}`;
  }
  async function loadInitial() {
    const current = ++requestId;
    loading = true; loadError = null;
    window.history.replaceState(null, '', pageUrl(offset));
    try {
      const data = await apiResults('/leagues', params(offset));
      if (current !== requestId) return;
      leagues = data.results; total = data.meta.total;
    } catch(err) { if (current === requestId) loadError = err.message; }
    finally { if (current === requestId) loading = false; }
  }
  function changed(debounce = false) {
    ++requestId; clearTimeout(timer);
    const apply = () => { offset = 0; loadInitial(); };
    if (debounce && search.trim()) timer = setTimeout(apply, 350); else apply();
  }
  function resetFilters() { search = ''; regionFilter = ''; tierFilter = ''; liveOnly = false; sorting = 'recommended:desc'; changed(); }
  onMount(loadInitial);
  onDestroy(() => { clearTimeout(timer); ++requestId; });
  $: filtered = !!(search.trim() || regionFilter || tierFilter || liveOnly);
</script>
<svelte:head><title>Dota 2 Leagues | Tournaments &amp; live games</title></svelte:head>
<PageHeader title="Leagues" description="Find tournaments, live games, and match results." eyebrow="Tournaments" />
<LiveNow />
<FilterToolbar label="Tournament filters" onreset={resetFilters} resetDisabled={!filtered && sorting === 'recommended:desc'}>
  {#snippet searchContent()}<SearchField label="Search leagues" placeholder="League name…" value={search} oninput={event => { search = event.currentTarget.value; changed(true); }} />{/snippet}
  <label>Status<select value={status} onchange={event => { status = event.currentTarget.value; changed(); }}><option value="active">Active</option><option value="completed">Completed</option><option value="all">All tournaments</option></select></label>
  <label>Tier<select value={tierFilter} onchange={event => { tierFilter = event.currentTarget.value; changed(); }}><option value="">All tiers</option>{#each Object.entries(TIERS) as [value, label]}<option {value}>{label}</option>{/each}</select></label>
  <label>Region<select value={regionFilter} onchange={event => { regionFilter = event.currentTarget.value; changed(); }}><option value="">All regions</option>{#each Object.entries(REGIONS) as [value, label]}<option {value}>{label}</option>{/each}</select></label>
  {#snippet optionsContent()}<label><input type="checkbox" checked={liveOnly} onchange={event => { liveOnly = event.currentTarget.checked; changed(); }} />Live only</label>{/snippet}
  {#snippet sortContent()}<label>Sort by<select value={sorting} onchange={event => { sorting = event.currentTarget.value; changed(); }}><option value="recommended:desc">Recommended</option><option value="name:asc">Name A–Z</option><option value="name:desc">Name Z–A</option><option value="start_date:desc">Latest start date</option><option value="start_date:asc">Earliest start date</option><option value="end_date:desc">Latest end date</option><option value="prize_pool:desc">Largest prize pool</option><option value="tier:desc">Highest tier</option></select></label>{/snippet}
</FilterToolbar>
{#if loading}
  <StatePanel kind="loading" title="Loading leagues" message="Finding tournaments and their latest details." />
{:else if loadError}
  <StatePanel kind="error" title="Couldn’t load leagues" message={loadError} actionLabel="Try again" onaction={loadInitial} />

{:else}
  {#if leagues.length}
    <p class="results-count" role="status">Showing {offset + 1}–{offset + leagues.length} of {total} {filtered ? 'matching' : ''} {total === 1 ? 'league' : 'leagues'}</p>
    <div class="leagues">{#each leagues as league (league.league_id)}<League {league} />{/each}</div>
  {:else if filtered}
    <StatePanel title="No leagues match your filters" message="No available leagues match this combination. Try another tier, region, or league name." />
  {:else}
    <StatePanel title={offset ? 'No leagues on this page' : 'No leagues available yet'} message="There are no tournaments to show here right now. Check again for the latest schedule." actionLabel={offset ? 'Back to first page' : 'Refresh leagues'} href={offset ? '/' : ''} onaction={offset ? undefined : loadInitial} />
  {/if}
  {#if leagues.length && (offset > 0 || total > pageSize)}
    <Pagination {offset} {total} {pageSize} previousHref={pageUrl(offset - pageSize)} nextHref={pageUrl(offset + pageSize)} label="League pages" />
  {/if}
{/if}
<style>
  .results-count { color: var(--muted); font-size: 12px; margin: 16px 0; }
  .leagues { display: grid; grid-template-columns: repeat(2,minmax(0,1fr)); gap: 12px; }
  @media(max-width:650px) { .leagues { grid-template-columns: 1fr; } }
</style>
