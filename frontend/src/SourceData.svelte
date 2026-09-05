<script>
  import { onMount, onDestroy } from 'svelte';
  import { api } from './lib/api.js';
  import { auditRows, groupRows, groupLinks } from './lib/source-data.js';
  import { formatDateTime, steamProfileUrl } from './lib/constants.js';
  export let kind;
  export let entityID;
  let data = null;
  let selected = '';
  let busy = false;
  let error = '';
  let request = 0;
  const opened = new URLSearchParams(window.location.search).get('data') === '1';
  $: snapshot = data?.snapshot;
  $: audits = auditRows(snapshot?.payload);
  $: groups = groupRows(snapshot?.payload);
  $: links = groupLinks(groups);
  $: download = `${process.env.baseUrl}/source-data/${kind}/${entityID}?download=1&version=${encodeURIComponent(snapshot?._key || '')}`;
  async function load(version = '') {
    const current = ++request; busy = true; error = '';
    try {
      const result = await api(`/source-data/${kind}/${entityID}`,version ? {version} : undefined);
      if (current !== request) return;
      if (version && !result.snapshot) throw new Error('This version is not available.');
      data = result; selected = result.snapshot?._key || '';
    } catch (err) { if (current === request && data) {error=err.message;selected=data.snapshot?._key || '';} }
    finally {if(current===request)busy=false;}
  }
  onMount(() => load());
  onDestroy(() => ++request);
</script>
{#if snapshot?.payload}
  <details class="source-data source-log" open={opened}>
    <summary>Data</summary>
    <div class="data-heading">
      <p>Collected {formatDateTime(snapshot.first_seen / 1000)} <span>· Checked {formatDateTime(data.status?.last_attempt / 1000)} UTC</span></p>
      <a href={download}>Download JSON</a>
    </div>
    {#if data.status?.failure}<p class="note">Latest collection failed ({data.status.failure.replaceAll('_',' ')}). Showing the last saved data. Last success: {formatDateTime(data.status.last_success / 1000)}.</p>{/if}
    {#if data.versions?.length > 1}
      <label class="versions">Version <select value={selected} disabled={busy} onchange={event => load(event.currentTarget.value)}>{#each data.versions as version}<option value={version._key}>{formatDateTime(version.first_seen / 1000)}</option>{/each}</select></label>
      {#if data.status?.versions > data.versions.length}<p class="note">Showing the latest {data.versions.length} versions.</p>{/if}
    {/if}
    {#if error}<p role="alert" class="note">{error}</p>{/if}
    {#if audits.length}
      <h3>Source audit</h3>
      <p class="note">Valve action codes are unverified. History may be incomplete.</p>
      <div class="table-scroll"><table class="audit-table"><thead><tr><th>Source time (UTC)</th><th>Player</th><th>Action code</th></tr></thead><tbody>{#each audits as row}<tr><td>{formatDateTime(row.timestamp)}</td><td>{#if data.player_names?.[row.account_id]}<a href={steamProfileUrl(row.account_id)}>{data.player_names[row.account_id]}</a><small class="account-id">#{row.account_id}</small>{:else}Account #{row.account_id}{/if}</td><td>{row.audit_action}</td></tr>{/each}</tbody></table></div>
    {/if}
    {#if groups.length}
      <h3>Tournament groups</h3>
      <div class="table-scroll"><table><thead><tr><th>Group</th><th>Name</th><th>Phase / round</th><th>Completed</th><th>Standings</th></tr></thead><tbody>{#each groups as group}<tr id={`source-group-${group.node_group_id}`}><td>{group.node_group_id}</td><td>{group.name || '—'}</td><td>{group.phase ?? '—'} / {group.round ?? '—'}</td><td>{group.is_completed === true ? 'Yes' : group.is_completed === false ? 'No' : '—'}</td><td>{Array.isArray(group.team_standings) ? group.team_standings.length : '—'}</td></tr>{/each}</tbody></table></div>
      {#if links.length}
        <h3>Advancement links</h3>
        <div class="table-scroll"><table><thead><tr><th>From group</th><th>To group</th><th>Route</th></tr></thead><tbody>{#each links as link}<tr><td><a href={`#source-group-${link.from}`}>{link.from}</a></td><td><a href={`#source-group-${link.to}`}>{link.to}</a></td><td>{link.route}</td></tr>{/each}</tbody></table></div>
      {/if}
    {/if}
    {#if !audits.length && !groups.length}<p class="note">This response supplies no usable audit or group rows. The original fields are available in the JSON download.</p>{/if}
  </details>
{/if}
<style>
  .source-data { margin-top: 24px; border-top: 1px solid var(--line); padding-top: 6px; }
  summary { cursor: pointer; font-size: 12px; font-weight: 500; padding: 8px 0; color: var(--muted); }
  .data-heading { display: flex; align-items: baseline;  flex-wrap: wrap; gap: 8px 16px; margin: 4px 0 8px; font-size: 11px; }
  .data-heading p { margin: 0; color: var(--muted); } .data-heading span { display: inline-block; }
  .note { font-size: 12px; color: var(--muted); margin: 4px 0 8px; }
  h3 { font-size: 12px; font-weight: 500; color: var(--muted); margin: 14px 0 4px; }
  .versions { display: flex; align-items: center; flex-wrap: wrap; gap: 10px; color: var(--muted); font-size: 12px; }
  .account-id { display: inline; color: var(--muted); font-size: 11px; margin-left: 8px; font-variant-numeric: tabular-nums; }
  .audit-table th:first-child { width: 180px; }
  .audit-table th:last-child { width: 100px; }
  .audit-table td { white-space: nowrap; }
  td { font-variant-numeric: tabular-nums; } tr:target { background: var(--accent-soft); }
</style>
