<script>
  import { onMount, onDestroy } from 'svelte';
  import { apiResults } from './lib/api.js';
  import { formatDateTime } from './lib/constants.js';
  import SearchField from './SearchField.svelte';
  let rows = [];
  let total = 0;
  let offset = 0;
  let search = '';
  let available = false;
  let loading = false;
  let error = '';
  let request = 0;
  let timer;
  const pageSize = 20;
  async function load(nextOffset = offset) {
    const current = ++request;
    loading = true; error = '';
    try {
      const result = await apiResults('/source-data/coverage',{offset:nextOffset,limit:pageSize,search:search.trim()});
      if(current!==request)return;
      rows=result.results;total=result.meta.total;offset=nextOffset;
      available = available || total > 0;
    } catch(err) { if(current===request)error=err.message; }
    finally {if(current===request)loading=false;}
  }
  function changed(value) {
    search=value;clearTimeout(timer);++request;loading=true;
    if(search.trim())timer=setTimeout(()=>load(0),350);else load(0);
  }
  onMount(()=>load(0));
  onDestroy(()=>{clearTimeout(timer);++request;});
</script>
{#if available || error}
  <details class="coverage source-log">
    <summary>Collection coverage</summary>
    <div class="coverage-toolbar"><SearchField label="Search collected records" placeholder="Name or ID…" value={search} oninput={event=>changed(event.currentTarget.value)} /><button class="secondary" disabled={loading} onclick={()=>load()}>Refresh</button></div>
    {#if error}<p role="alert" class="note">{error}</p>
    {:else}
      <p class="note" role="status">{loading ? 'Loading…' : total ? `${offset+1}–${offset+rows.length} of ${total.toLocaleString()} records` : 'No matching records'}</p>
      {#if rows.length}
        <div class="table-scroll" aria-busy={loading}><table><thead><tr><th>Record</th><th>Collection</th><th>Last success (UTC)</th><th>Audit / groups</th><th>Versions</th></tr></thead><tbody>
          {#each rows as row}<tr><td><a href={`/${row.kind === 'league' ? 'league' : 'team'}/${row.entity_id}?data=1`}>{row.name || `${row.kind} #${row.entity_id}`}</a></td><td>{row.failure ? row.failure.replaceAll('_',' ') : row.last_success ? 'Collected' : 'Pending'}</td><td>{row.last_success ? formatDateTime(row.last_success/1000) : '—'}</td><td>{row.last_success ? (row.kind === 'team' ? row.audit_count || 0 : row.group_count || 0) : '—'}</td><td>{row.versions || '—'}</td></tr>{/each}
        </tbody></table></div>
      {/if}
      {#if offset>0 || offset+pageSize<total}
        <nav class="coverage-pages" aria-label="Coverage pages">
          <button class="secondary" disabled={loading || offset===0} onclick={()=>load(Math.max(0,offset-pageSize))}>← Previous</button>
          <span>Page {Math.floor(offset/pageSize)+1} of {Math.max(1,Math.ceil(total/pageSize))}</span>
          <button class="secondary" disabled={loading || offset+pageSize>=total} onclick={()=>load(offset+pageSize)}>Next →</button>
        </nav>
      {/if}
    {/if}
  </details>
{/if}
<style>
  .coverage { border-top: 1px solid var(--line); padding-top: 6px; margin-top: 24px; }
  summary { cursor: pointer; font-size: 12px; padding: 8px 0; color: var(--muted); }
  .coverage-toolbar { display: grid; grid-template-columns: minmax(0,1fr) auto; gap: 12px; align-items: end; margin: 12px 0; }
  .note { color: var(--muted); font-size: 12px; margin: 6px 0 8px; } td { white-space: nowrap; }
  .coverage-pages { display: flex; justify-content: center; align-items: center; flex-wrap: wrap; gap: 12px; margin-top: 16px; font-size: 12px; color: var(--muted); }
</style>
