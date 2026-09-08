<script>
  import { countries } from './lib/countries.js';
  import FilterToolbar from "./FilterToolbar.svelte";
  import SearchField from "./SearchField.svelte";
  import Pagination from "./Pagination.svelte";
  import PageHeader from "./PageHeader.svelte";
  import PlayerAvatar from "./PlayerAvatar.svelte";
  import StatePanel from "./StatePanel.svelte";
  import { onMount, onDestroy } from 'svelte';
  import { apiResults } from './lib/api.js';
  import { pageOffset } from './lib/routes.js';
  import { countryText, fantasyRoleText, formatMoney } from './lib/constants.js';
  import { playerDisplayName, isListedPro } from './lib/players.js';
  import { takeDirectoryData } from './lib/directory.js';
  const initial = takeDirectoryData('/player');
  let players = initial?.results ?? [];
  let total = initial?.meta.total ?? 0;
  let loading = !initial;
  let error = '';
  const query = new URLSearchParams(window.location.search);
  let search = query.get('search') || '';
  let appliedSearch = search.trim();
  let offset = pageOffset(window.location.search);
  let country = (query.get('country') || '').toUpperCase();
  let pro = query.get('pro') || '';
  let team = query.get('team') || '';
  let sorting = `${query.get('sort') || 'recommended'}:${query.get('order') || (query.get('sort') === 'name' ? 'asc' : 'desc')}`;
  let requestId = 0;
  let searchTimer;
  $: filtered = Boolean(appliedSearch || country || pro || team || sorting !== 'recommended:desc');
  function filters() {
    const [sort, order] = sorting.split(':');
    return {search: appliedSearch || undefined, country: country || undefined, pro: pro || undefined, team: team || undefined, sort, order};
  }
  function pageUrl(nextOffset) {
    const params = new URLSearchParams();
    for (const [key, value] of Object.entries(filters())) if (value !== undefined) params.set(key, String(value));
    if (nextOffset) params.set('offset', String(nextOffset));
    return `/player${params.size ? `?${params}` : ''}`;
  }
  async function load() {
    const current = ++requestId;
    loading = true; error = '';
    window.history.replaceState(null, '', pageUrl(offset));
    try {
      const data = await apiResults('/players', {offset, limit: 100, ...filters()});
      if (current !== requestId) return;
      players = data.results; total = data.meta.total;
    } catch (err) { if (current === requestId) error = err.message; }
    finally { if (current === requestId) loading = false; }
  }
  function changeSearch(value) {
    search = value; ++requestId;
    clearTimeout(searchTimer);
    const apply = () => { appliedSearch = value.trim(); offset = 0; load(); };
    if (!value.trim()) apply(); else searchTimer = setTimeout(apply, 350);
  }
  function changeFilters() { clearTimeout(searchTimer); appliedSearch = search.trim(); offset = 0; load(); }
  function resetFilters() { search = ''; country = ''; pro = ''; team = ''; sorting = 'recommended:desc'; changeFilters(); }
  const detail = p => [countryText(p.country_code), fantasyRoleText(p.fantasy_role), p.team_id > 0 ? (p.team_name || `Team #${p.team_id}`) : (isListedPro(p) ? 'No team' : '')].filter(Boolean).join(' · ') || (isListedPro(p) ? 'Professional player' : 'Steam profile only');
  onMount(() => { if (!initial) load(); });
  onDestroy(() => { clearTimeout(searchTimer); ++requestId; });
</script>
<svelte:head><title>Dota 2 Players | Profiles &amp; teams</title></svelte:head>
<PageHeader title="Players" description="Find professional profiles, current teams, earnings, and Steam profiles." eyebrow="Player directory" />
<FilterToolbar label="Player filters" onreset={resetFilters} resetDisabled={!filtered}>
  {#snippet searchContent()}<SearchField label="Search players" placeholder="Name, real name, or account ID…" value={search} oninput={event => changeSearch(event.currentTarget.value)} />{/snippet}
  <label>Country<select value={country} onchange={event => { country = event.currentTarget.value; changeFilters(); }}><option value="">All countries</option>{#each countries as item}<option value={item.code}>{item.name}</option>{/each}</select></label>
  <label>Professional<select value={pro} onchange={event => { pro = event.currentTarget.value; changeFilters(); }}><option value="">All players</option><option value="true">Listed pro players</option><option value="false">Steam profiles only</option></select></label>
  <label>Team<select value={team} onchange={event => { team = event.currentTarget.value; changeFilters(); }}><option value="">Any</option><option value="true">With a team</option><option value="false">Without a team</option></select></label>
  {#snippet sortContent()}<label>Sort by<select value={sorting} onchange={event => { sorting = event.currentTarget.value; changeFilters(); }}><option value="recommended:desc">Recommended</option><option value="name:asc">Name A–Z</option><option value="name:desc">Name Z–A</option></select></label>{/snippet}
</FilterToolbar>
{#if loading}<StatePanel kind="loading" title="Loading players" message="Finding player profiles." />
{:else if error}<StatePanel kind="error" title="Couldn’t load players" message={error} actionLabel="Try again" onaction={load} />
{:else if !players.length && filtered}<StatePanel title="No players match your filters" message="Try another name, country, professional status, or team filter." actionLabel="Clear filters" onaction={resetFilters} />
{:else if !players.length}<StatePanel title="No players on this page" message="There are no player profiles available here yet." actionLabel={offset ? 'Back to first page' : 'Refresh players'} href={offset ? '/player' : ''} onaction={offset ? undefined : load} />
{:else}
<p class="muted result-count" role="status">Showing {offset + 1}–{offset + players.length} of {total} {appliedSearch ? 'matching ' : ''}{total === 1 ? 'player' : 'players'}</p>
<div class="players">
{#each players as p (p.account_id)}
<a class="player-card" href="/player/{p.account_id}">
  <div class="portrait"><PlayerAvatar src={p.avatar_url} /></div>
  <div class="info"><h2>{playerDisplayName(p, p.account_id)}</h2><p>{detail(p)}</p>{#if p.total_earnings > 0}<p class="earnings">{formatMoney(p.total_earnings)} recorded</p>{/if}</div>
</a>
{/each}
</div>
{/if}
{#if !loading && !error && (offset > 0 || offset + 100 < total)}<Pagination {offset} {total} pageSize={100} previousHref={pageUrl(offset - 100)} nextHref={pageUrl(offset + 100)} label="Player pages" />{/if}
<style>
 .result-count { font-size: 12px; margin: 20px 0; }
 .players { display: grid; grid-template-columns: repeat(3,minmax(0,1fr)); gap: 16px; }
 .player-card { min-width: 0; display: flex; align-items: center; gap: 14px; background: var(--surface); border: 1px solid var(--line); padding: 16px; border-radius: 8px; text-decoration: none; color: inherit; }
 .player-card:hover { border-color: var(--line-strong); }
 .portrait { width: 44px; height: 44px; flex-shrink: 0; border-radius: 8px; overflow: hidden; background: var(--surface-subtle); }
 .portrait :global(.player-avatar) { width: 100%; height: 100%; flex-basis: 100%; border-radius: 8px; }
 .info { min-width: 0; }
 h2 { font-size: 15px; margin: 0 0 3px; overflow-wrap: anywhere; } p { font-size: 11px; color: var(--muted); margin: 0; overflow-wrap: anywhere; }
 .earnings { margin-top: 2px; font-variant-numeric: tabular-nums; }
 @media(max-width: 1000px) { .players { grid-template-columns: repeat(2,minmax(0,1fr)); } }
 @media(max-width: 600px) { .players { grid-template-columns: 1fr; } }
</style>
