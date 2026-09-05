<script>
  import FilterToolbar from "./FilterToolbar.svelte";
  import SourceCoverage from "./SourceCoverage.svelte";
  import SearchField from "./SearchField.svelte";
  import Pagination from "./Pagination.svelte";
  import PageHeader from "./PageHeader.svelte";
  import { onMount, onDestroy } from 'svelte';
  import { apiResults } from './lib/api.js';
  import { pageOffset } from './lib/routes.js';
  import { groupUpdates, UPDATE_PAGE_SIZE } from './lib/updates.js';
  import StatePanel from './StatePanel.svelte';
  import UpdateEntry from './UpdateEntry.svelte';
  let offset = pageOffset(window.location.search, UPDATE_PAGE_SIZE);
  const params = new URLSearchParams(window.location.search);
  let search = params.get('search') || '';
  let entity = ['roster','team','tournament','player'].includes(params.get('entity')) ? params.get('entity') : '';
  let days = ['1','7','30'].includes(params.get('days')) ? params.get('days') : '';
  $: filtered = Boolean(search.trim() || entity || days);
  let filterTimer;
  function changed(debounce = false) {
    clearTimeout(filterTimer);
    ++request;
    const apply = () => { offset = 0; history.replaceState(null, "", pageHref(0)); load(); };
    if (debounce && search.trim()) filterTimer = setTimeout(apply, 350); else apply();
  }
  function resetFilters() { search = ""; entity = ""; days = ""; changed(); }
  function pageHref(nextOffset) {
    const query = new URLSearchParams();
    if (search.trim()) query.set('search',search.trim());
    if (entity) query.set('entity',entity);
    if (days) query.set('days',days);
    if (nextOffset > 0) query.set('offset',nextOffset);
    return '/activity' + (query.size ? '?' + query : '');
  }
  let updates = [];
  let total = 0;
  let loading = true;
  let error = '';
  let request = 0;
  $: groups = groupUpdates(updates);
  async function load() {
    const current = ++request;
    loading = true; error = '';
    try {
      const data = await apiResults('/updates', { offset, limit: UPDATE_PAGE_SIZE, search, entity, days });
      if (current !== request) return;
      updates = data.results; total = data.meta.total;
    } catch (err) { if (current === request) error = err.message; }
    finally { if (current === request) loading = false; }
  }
  onMount(load);
  onDestroy(() => { clearTimeout(filterTimer); ++request; });
</script>
<svelte:head><title>Updates | Dota 2 Leagues</title></svelte:head>
<PageHeader title="Updates" description="Tournament changes, team records, and roster moves as they’re recorded." eyebrow="Activity log" />
<FilterToolbar label="Activity filters" onreset={resetFilters} resetDisabled={!filtered}>
  {#snippet searchContent()}<SearchField label="Search activity" placeholder="Name or ID…" value={search} oninput={event => { search = event.currentTarget.value; changed(true); }} />{/snippet}
  <label>Activity type<select value={entity} onchange={event => { entity = event.currentTarget.value; changed(); }}>
    <option value="">All activity</option><option value="roster">Rosters</option><option value="team">Teams</option><option value="tournament">Tournaments</option><option value="player">Players</option>
  </select></label>
  <label>Time range<select value={days} onchange={event => { days = event.currentTarget.value; changed(); }}>
    <option value="">All time</option><option value="1">Last 24 hours</option><option value="7">Last 7 days</option><option value="30">Last 30 days</option>
  </select></label>
  {#snippet optionsContent()}<span class="activity-order">Newest first · All times UTC</span>{/snippet}
</FilterToolbar>
{#if updates.length && !error}
  <div class="feed-heading">
    <p class="result-count" role="status">{#if loading}Refreshing updates…{:else}Showing {offset + 1}–{offset + updates.length} of {total.toLocaleString()} updates{/if}</p>
    <button class="refresh" onclick={load} disabled={loading}>Refresh updates</button>
  </div>
{/if}
{#if loading}<StatePanel kind="loading" title="Loading updates" message="Checking the latest tournament, team, roster, and player activity." />
{:else if error}<StatePanel kind="error" title="Couldn’t load updates" message={error} actionLabel="Try again" onaction={load} />
{:else if !updates.length && filtered}<StatePanel title="No matching updates" message="Try another name, activity type, or time range." actionLabel="Clear filters" onaction={resetFilters} />
{:else if !updates.length}<StatePanel title={offset ? 'No updates on this page' : 'No updates recorded yet'} message={offset ? 'New activity may have changed the page count. Return to the latest updates.' : 'Tournament, team, roster, and player changes will appear here as they’re recorded. Earlier changes aren’t included.'} actionLabel={offset ? 'Latest updates' : 'Refresh updates'} href={offset ? '/activity' : ''} onaction={offset ? undefined : load} />
{:else}
  <div class="feed">{#each groups as group (group.key)}<section aria-label={group.label} class="date-group"><h2>{group.label}</h2>{#each group.rows as update}<UpdateEntry {update} />{/each}</section>{/each}</div>
  {#if offset > 0 || offset + UPDATE_PAGE_SIZE < total}<Pagination {offset} {total} pageSize={UPDATE_PAGE_SIZE} previousHref={pageHref(offset - UPDATE_PAGE_SIZE)} nextHref={pageHref(offset + UPDATE_PAGE_SIZE)} label="Updates pages" previousLabel="Newer" nextLabel="Older" />{/if}
  <p class="history-note">This log starts when activity recording was enabled. Times show when changes were recorded.</p>
{/if}
<SourceCoverage />
<style>
  .feed-heading { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: .25rem 1rem; margin: .5rem 0 .75rem; }
  .refresh { min-height: 2.75rem; padding: .5rem 0; background: transparent; border: 0; box-shadow: none; color: var(--muted); font-size: .8125rem; font-weight: 500; }
  .refresh:hover:not(:disabled) { background: transparent; color: var(--text); text-decoration: underline; text-underline-offset: .2em; }
  .refresh:disabled { opacity: .5; cursor: default; }
  .activity-order { font-size: .8125rem; color: var(--muted); }
  .result-count,.history-note { color: var(--muted); font-size: 12px; } .result-count { margin: 0; }
  .feed { border: 1px solid var(--line); border-radius: 8px; overflow: hidden; }
  .date-group h2 { margin: 0; padding: 8px 18px; background: var(--surface-subtle); font-size: 13px; font-weight: 500; color: var(--text); border-bottom: 1px solid var(--line); }
  .history-note { margin-top: 24px; }
  @media(max-width:600px) { .date-group h2 { padding: 8px 12px; } }
</style>
