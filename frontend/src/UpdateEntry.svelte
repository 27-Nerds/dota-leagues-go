<script>
  import UpdateChange from './UpdateChange.svelte';
  import ActivityIcon from './ActivityIcon.svelte';
  import { updatePresentation, updateDate, updateSummary, playerTeam } from './lib/updates.js';
  let { update } = $props();
  let view = $derived(updatePresentation(update));
  let team = $derived(playerTeam(update));
  let date = $derived(updateDate(update.created_at));
</script>

{#snippet identity()}
  <span class="identity">
    <span class="name">{#if view.href}<a href={view.href}>{view.name}</a>{:else}{view.name}{/if}</span>
    <span class="event-label">{view.title}</span>
  </span>
{/snippet}

<article class="entry" aria-label={`${view.title}: ${view.name}`}>
  <time datetime={date.iso} title={date.label + ' · ' + date.time + ' UTC'}>{date.time}</time>
  <span class="event-icon" class:caution={view.inferred}><ActivityIcon name={view.icon} /></span>
  <div class="entry-body">
    {#if !view.created && (view.changes.length || update.sources?.length)}
      <details>
        <summary class="event-summary">
          {@render identity()}
          {#if view.transition && (view.transition.left > 1 || view.transition.joined > 1)}
            <span class="movements">
              {#if view.transition.left}<span class="out" aria-label={`${view.transition.left} players left`}>−{view.transition.left}</span>{/if}
              {#if view.transition.joined}<span class="in" aria-label={`${view.transition.joined} players joined`}>+{view.transition.joined}</span>{/if}
            </span>
          {:else if update.entity !== 'roster' && view.changes.every(c => ['wins','losses','total_prize_pool','status','total_earnings','steam_profile'].includes(c.field))}
            <span class="value-summary">{updateSummary(update)}</span>
          {/if}
          <span class="disclosure" aria-hidden="true"><svg viewBox="0 0 16 16"><path d="m5 6 3 3 3-3" /></svg></span>
        </summary>
        <div class="changes">
          {#if view.inferred}<p class="note">Roster cleared; disband not confirmed.</p>{/if}
          {#each view.changes as change}<UpdateChange {change} />{/each}
          {#if update.sources?.length > 1}
            <details class="history">
              <summary>History <span class="history-count">{update.sources.length}</span></summary>
              {#each update.sources as source}
                <div class="observation">
                  <time datetime={updateDate(source.created_at).iso}>{updateDate(source.created_at).key} · {updateDate(source.created_at).time} UTC</time>
                  {#each source.changes as change}<UpdateChange {change} />{/each}
                </div>
              {/each}
            </details>
          {/if}
        </div>
      </details>
    {:else}
      <div class="plain-row">
        {@render identity()}
        {#if update.entity === 'player'}<span class="player-team">{#if team.href}<a href={team.href}>{team.name}</a>{:else}{team.name}{/if}</span><span class="account">#{update.entity_id}</span>{/if}
      </div>
    {/if}
  </div>
</article>

<style>
  .entry { display: grid; grid-template-columns: 40px 18px minmax(0,1fr); gap: 12px; padding: 12px 18px; align-items: start; background: var(--surface); border-bottom: 1px solid var(--line); }
  .entry:hover { background: var(--surface-hover); }
  time { font-size: 11px; color: var(--muted); font-variant-numeric: tabular-nums; line-height: 22px; }
  .event-icon { color: var(--muted); padding-top: 2px; } .event-icon.caution { color: var(--accent); }
  .entry-body { min-width: 0; }
  .event-summary,.plain-row { display: flex; align-items: baseline; gap: 16px; min-height: 22px; }
  .event-summary { cursor: pointer; list-style: none; }
  .event-summary::-webkit-details-marker { display: none; }
  .identity { display: flex; align-items: baseline; flex-wrap: wrap; gap: 2px 12px; min-width: 0; }
  .name { font-size: 13px; font-weight: 500; line-height: 22px; overflow-wrap: anywhere; }
  .name a { text-decoration: none; } .name a:hover { text-decoration: underline; }
  .event-label { font-size: 12px; color: var(--muted); line-height: 22px; }
  .movements { display: flex; gap: 10px; font-size: 12px; font-variant-numeric: tabular-nums; white-space: nowrap; }
  .out { color: var(--negative); } .in { color: var(--positive); }
  .value-summary { color: var(--muted); font-size: 12px; line-height: 22px; }
  .disclosure { margin-left: auto; flex-shrink: 0; align-self: center; color: var(--muted); }
  .disclosure svg { display: block; width: 16px; height: 16px; fill: none; stroke: currentColor; stroke-width: 1.5; }
  details[open] > .event-summary .disclosure { transform: rotate(180deg); }
  .changes { margin: 12px 0 4px; max-width: 850px; }
  .note { margin: 0 0 10px; font-size: 12px; color: var(--muted); }
  .history { margin-top: 12px; }
  .history > summary { cursor: pointer; font-size: 12px; color: var(--muted); padding: 6px 0; }
  .history-count { margin-left: 6px; font-variant-numeric: tabular-nums; }
  .observation { margin-top: 12px; padding-left: 14px; border-left: 1px solid var(--line); }
  .player-team,.account { color: var(--muted); font-size: 12px; line-height: 22px; overflow-wrap: anywhere; }
  .player-team a { color: inherit; } .account { margin-left: auto; white-space: nowrap; font-size: 11px; }
  @media(max-width:600px) {
    .entry { padding: 11px 12px; grid-template-columns: 32px 16px minmax(0,1fr); gap: 8px; }
    .event-summary,.plain-row { flex-wrap: wrap; gap: 2px 10px; position: relative; padding-right: 18px; }
    .identity { flex-basis: 100%; gap: 0 8px; }
    .identity .name { flex-basis: 100%; }
    .disclosure { position: absolute; right: 0; top: 3px; }
    .account { margin-left: 0; }
  }
</style>
