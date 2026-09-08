<script>
  import { countries } from './lib/countries.js';
  import FilterToolbar from "./FilterToolbar.svelte";
  import SearchField from "./SearchField.svelte";
  import Pagination from "./Pagination.svelte";
  import PageHeader from "./PageHeader.svelte";
  import { onMount, onDestroy } from 'svelte';
  import { apiResults } from './lib/api.js';
  import { pageOffset } from './lib/routes.js';
  import { regionText, countryText, teamLogo } from './lib/constants.js';
  import EntityImage from './EntityImage.svelte';
  import StatePanel from './StatePanel.svelte';
  import { takeDirectoryData } from './lib/directory.js';
  const initial = takeDirectoryData('/team');
  let teams = initial?.results ?? [];
  let total = initial?.meta.total ?? 0;
  let loading = !initial;
  let error = '';
  let search = new URLSearchParams(window.location.search).get('search') || '';
  let appliedSearch = search.trim();
  let offset = pageOffset(window.location.search);
  let requestId = 0;
  let searchTimer;
  const query = new URLSearchParams(window.location.search);
  let country = (query.get('country') || '').toUpperCase();
  let pro = query.get('pro') || '';
  let active = query.get('active') === 'true';
  let activeDays = query.get('active_days') || '90';
  let sorting = `${query.get('sort') || 'recommended'}:${query.get('order') || (query.get('sort') === 'name' ? 'asc' : 'desc')}`;
  function filters() {
    const [sort, order] = sorting.split(':');
    return {pro: pro || undefined, search: appliedSearch || undefined, country: country || undefined, active: active || undefined, active_days: active ? activeDays : undefined, sort, order};
  }
  function pageUrl(nextOffset) {
    const params = new URLSearchParams();
    for (const [key, value] of Object.entries(filters())) {
      if (value !== undefined) params.set(key, String(value));
    }
    if (nextOffset) params.set('offset', String(nextOffset));
    return `/team${params.size ? `?${params}` : ''}`;
  }
  async function load() {
    const current = ++requestId;
    loading = true; error = '';
    window.history.replaceState(null, '', pageUrl(offset));
    try {
      const data = await apiResults('/teams', {offset, limit:100, ...filters()});
      if (current !== requestId) return;
      teams = data.results; total = data.meta.total;
    } catch(err) { if (current === requestId) error = err.message; }
    finally { if (current === requestId) loading = false; }
  }
  function changeSearch(value) {
    search = value; ++requestId;
    clearTimeout(searchTimer);
    const applySearch = () => {
      appliedSearch = value.trim(); offset = 0;
      load();
    };
    if (!value.trim()) applySearch();
    else searchTimer = setTimeout(applySearch, 350);
  }
  function changeFilters() {
    clearTimeout(searchTimer);
    appliedSearch = search.trim(); offset = 0;
    load();
  }
  function resetFilters() {
    search = ''; country = ''; pro = ''; active = false; activeDays = '90'; sorting = 'recommended:desc';
    changeFilters();
  }
  onMount(() => { if (!initial) load(); });
  onDestroy(() => { clearTimeout(searchTimer); ++requestId; });
</script>
<svelte:head><title>Dota 2 Teams | Rosters &amp; results</title></svelte:head>
<PageHeader title="Teams" description="Find team rosters, player statistics, and tournament history." eyebrow="Team directory" />
<FilterToolbar label="Team filters" onreset={resetFilters} resetDisabled={!search && !country && !pro && !active && sorting === 'recommended:desc'}>
  {#snippet searchContent()}<SearchField label="Search teams" placeholder="Team name or tag…" value={search} oninput={event => changeSearch(event.currentTarget.value)} />{/snippet}
  <label>Country<select value={country} onchange={event => { country = event.currentTarget.value; changeFilters(); }}><option value="">All countries</option>{#each countries as item}<option value={item.code}>{item.name}</option>{/each}</select></label>
  <label>Professional<select value={pro} onchange={event => { pro = event.currentTarget.value; changeFilters(); }}><option value="">All teams</option><option value="true">Professional</option><option value="false">Non-professional</option></select></label>
  <label>Recent activity<select value={activeDays} disabled={!active} onchange={event => { activeDays = event.currentTarget.value; changeFilters(); }}>{#each ['7', '30', '90', '180', '365'] as days}<option value={days}>Last {days} days</option>{/each}</select></label>
  {#snippet optionsContent()}<label><input type="checkbox" checked={active} onchange={event => { active = event.currentTarget.checked; changeFilters(); }} />Active teams only</label>{/snippet}
  {#snippet sortContent()}<label>Sort by<select value={sorting} onchange={event => { sorting = event.currentTarget.value; changeFilters(); }}><option value="recommended:desc">Recommended</option><option value="activity:desc">Latest activity</option><option value="name:asc">Name A–Z</option><option value="name:desc">Name Z–A</option><option value="wins:desc">Most wins</option></select></label>{/snippet}
</FilterToolbar>
{#if active}<p class="activity-hint">A played game or recorded team/roster change in the last {activeDays} days, or playing live now.</p>{/if}
{#if loading}<StatePanel kind="loading" title="Loading teams" message="Finding team profiles and rosters." />
{:else if error}<StatePanel kind="error" title="Couldn’t load teams" message={error} actionLabel="Try again" onaction={load} />
{:else if !teams.length && (appliedSearch || country || pro || active)}<StatePanel title="No teams match your filters" message="Try changing the professional status, country, activity window, or team name." actionLabel="Clear filters" onaction={resetFilters} />
{:else if !teams.length}<StatePanel title="No teams on this page" message="There are no team profiles available here yet." actionLabel={offset ? 'Back to first page' : 'Refresh teams'} href={offset ? '/team' : ''} onaction={offset ? undefined : load} />

{:else}
<p class="muted result-count" role="status">Showing {offset + 1}–{offset + teams.length} of {total} {appliedSearch ? 'matching ' : ''}{total === 1 ? 'team' : 'teams'}</p>
<div class="teams">
{#each teams as team (team.team_id)}
<a class="team-card" href="/team/{team.team_id}">
  <div class="logo"><EntityImage src={teamLogo(team.team_id)} name={team.name} /></div>
  <div class="info"><h2>{team.name || `Team #${team.team_id}`}</h2><p>{countryText(team.country_code) || (regionText(team.region) !== 'Unknown' ? regionText(team.region) : 'Team profile')}{#if team.tag}{' · '}{team.tag}{/if}</p></div>
</a>
{/each}
</div>
{/if}
{#if !loading && !error && (offset > 0 || offset + 100 < total)}<Pagination {offset} {total} pageSize={100} previousHref={pageUrl(offset - 100)} nextHref={pageUrl(offset + 100)} label="Team pages" />{/if}
<style>
 .activity-hint { margin-top: 12px; }
 .result-count { font-size: 12px; margin: 20px 0; }
 .teams { display: grid; grid-template-columns: repeat(3,minmax(0,1fr)); gap: 16px; }
 .team-card { min-width: 0; display: flex; align-items: center; gap: 14px; background: var(--surface); border: 1px solid var(--line); padding: 16px; border-radius: 8px; text-decoration: none; }
 .team-card:hover { border-color: var(--line-strong); }
 .logo { width: 44px; height: 44px; flex-shrink: 0; border-radius: 8px; } .info { min-width: 0; }
 h2 { font-size: 15px; margin: 0 0 3px; overflow-wrap: anywhere; } p { font-size: 11px; color: var(--muted); margin: 0; overflow-wrap: anywhere; }
 @media(max-width: 1000px) { .teams { grid-template-columns: repeat(2,minmax(0,1fr)); } }
 @media(max-width: 600px) { .teams { grid-template-columns: 1fr; } }
</style>
