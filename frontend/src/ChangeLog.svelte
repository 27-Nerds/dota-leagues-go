<script>
  // Compact "Changes" block shared by tournament, team, and player pages, styled like the Data block.
  import { onMount, onDestroy } from 'svelte';
  import { apiResults } from './lib/api.js';
  import { formatDateTime } from './lib/constants.js';
  import { updatePresentation, fieldLabel, updateValue } from './lib/updates.js';
  export let entities;
  export let entityID;
  let updates = [];
  let total = 0;
  let loaded = false;
  let error = '';
  let request = 0;
  const opened = new URLSearchParams(window.location.search).get('changes') === '1';
  const millis = value => Number(value) > 0 ? Number(value) / 1000 : 0;
  const changeDetails = changes => changes.map(c => `${fieldLabel(c.field)}: ${updateValue(c.field, c.before)} → ${updateValue(c.field, c.after)}`).join(' · ') || 'No field changes';
  $: activityHref = `/activity?search=${encodeURIComponent(entityID)}${entities.length === 1 ? `&entity=${entities[0]}` : ''}`;
  async function load() {
    const current = ++request;
    loaded = false; error = '';
    try {
      const pages = await Promise.all(entities.map(entity => apiResults('/updates', { entity, search: String(entityID), limit: 20 })));
      if (current !== request) return;
      updates = pages.flatMap(page => page.results).filter(u => String(u.entity_id) === String(entityID) && entities.includes(u.entity))
        .sort((a, b) => Number(b.created_at) - Number(a.created_at)).slice(0, 20);
      total = pages.reduce((sum, page) => sum + page.meta.total, 0);
    } catch (err) { if (current === request) error = err.message; }
    finally { if (current === request) loaded = true; }
  }
  onMount(load);
  onDestroy(() => ++request);
</script>
<details class="source-data source-log" open={opened}>
  <summary>Changes</summary>
  <div class="data-heading"><p>{#if !loaded}Loading recorded changes…{:else if error}Couldn’t load changes: {error}{:else if updates.length}{total} recorded · newest first <span>· All times UTC</span>{:else}No recorded changes yet. Changes appear here once observed.{/if}</p>
    {#if error}<button type="button" class="link" onclick={load}>Try again</button>{:else if updates.length}<a href={activityHref}>Open in activity log</a>{/if}</div>
  {#if updates.length}
    <div class="table-scroll"><table class="changes-table"><thead><tr><th>Recorded (UTC)</th><th>Change</th><th>Details</th></tr></thead><tbody>{#each updates as update (update.id)}{@const view = updatePresentation(update)}<tr><td>{formatDateTime(millis(update.created_at))}</td><td>{view.title}</td><td>{#if view.created}{view.note}{:else}{changeDetails(view.changes)}{/if}</td></tr>{/each}</tbody></table></div>
  {/if}
</details>
<style>
  .source-data { margin-top: 24px; border-top: 1px solid var(--line); padding-top: 6px; }
  summary { cursor: pointer; font-size: 12px; font-weight: 500; padding: 8px 0; color: var(--muted); }
  .data-heading { display: flex; align-items: baseline; flex-wrap: wrap; gap: 8px 16px; margin: 4px 0 8px; font-size: 11px; }
  .data-heading p { margin: 0; color: var(--muted); } .data-heading span { display: inline-block; }
  .link { background: none; border: 0; padding: 0; font: inherit; color: var(--accent); cursor: pointer; text-decoration: underline; }
  .changes-table th:first-child { width: 180px; } .changes-table th:nth-child(2) { width: 180px; }
  .changes-table td:first-child, .changes-table td:nth-child(2) { white-space: nowrap; }
  .changes-table td:last-child { overflow-wrap: anywhere; }
  td { font-variant-numeric: tabular-nums; }
</style>
