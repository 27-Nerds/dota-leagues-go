<script>
  import { fieldLabel, updateValue, rosterChanges } from './lib/updates.js';
  import { playerHref } from './lib/constants.js';
  let { change } = $props();
  let members = $derived(change.field === 'members' ? rosterChanges(change.before, change.after) : []);
</script>
{#if change.field === 'members'}
  <ul class="members" aria-label="Roster changes">
    {#each members as member}
      <li><span class="change-kind" class:removed={member.kind === 'left' || member.kind === 'deactivated'}>{member.label}</span><strong>{#if playerHref(member.id)}<a href={playerHref(member.id)}>Account #{member.id}</a>{:else}Account #{member.id}{/if}</strong><span class="member-status">{member.status}</span></li>
    {:else}<li>No membership or active-status changes.</li>{/each}
  </ul>
{:else}
  <div class="field-change" class:prize={change.field === 'total_prize_pool'}>
    <span class="field-name">{fieldLabel(change.field)}</span>
    <div class="values"><span class="before"><span class="sr-only">Before: </span>{updateValue(change.field, change.before)}</span><span aria-hidden="true" class="arrow">→</span><strong><span class="sr-only">After: </span>{updateValue(change.field, change.after)}</strong></div>
  </div>
{/if}
<style>
  .field-change { padding: 10px 0; border-top: 1px solid var(--line); display: grid; grid-template-columns: 130px minmax(0,1fr); gap: 16px; }
  .field-name { color: var(--muted); font-size: 12px; }
  .values { display: grid; grid-template-columns: minmax(0,1fr) 20px minmax(0,1fr); gap: 12px; align-items: start; font-size: 13px; white-space: pre-wrap; overflow-wrap: anywhere; }
  .before { color: var(--muted); } .values strong { font-weight: 500; color: var(--text); } .arrow { color: var(--text-subtle); text-align: center; }
  .prize .values { font-size: 18px; font-variant-numeric: tabular-nums; }
  .members { list-style: none; padding: 0; margin: 0; }
  .members li { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 14px; padding: 10px 0; border-top: 1px solid var(--line); font-size: 12px; }
  .members strong { font-weight: 500; } .member-status { color: var(--muted); margin-left: auto; }
  .change-kind { min-width: 82px; color: var(--positive); }
  .removed { color: var(--negative); }
  .sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); }
  @media(max-width:600px) { .field-change { grid-template-columns: 1fr; gap: 5px; } .values { gap: 7px; } .prize .values { font-size: 16px; } .member-status { width: 100%; margin-left: 96px; } }
</style>
