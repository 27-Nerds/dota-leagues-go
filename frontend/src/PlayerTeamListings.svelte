<script>
  import { formatDate, formatDateTime } from './lib/constants.js';
  import { isListedPro } from './lib/players.js';
  let { player } = $props();
  const dpc = $derived(isListedPro(player));
  const roster = $derived(player?.roster_team?.id > 0 ? player.roster_team : null);
  const differs = $derived(dpc && player?.team_id > 0 && roster && player.team_id !== roster.id);
</script>

{#if dpc || roster}
<section class="team-listings" aria-labelledby="team-listings-title">
  <div class="section-heading">
    <h2 id="team-listings-title">Team listings</h2>
    {#if differs}<span class="source-status">Listings differ</span>{/if}
  </div>
  <p class="explanation">{differs ? 'The DPC profile and team roster name different teams. Neither listing confirms the player’s current active team.' : 'Stored listings from Valve. A listing does not confirm that the player still actively plays for the team.'}</p>
  <div class="sources">
    {#if dpc}
    <article class="source">
      <h3>DPC-listed team</h3>
      <div class="team-name">
        {#if player.team_id > 0}
          {#if player.team_available}<a href={`/team/${player.team_id}`}>{player.team_name || `Team #${player.team_id}`}</a>{:else}<span>{player.team_name || `Team #${player.team_id}`}</span>{/if}
          {#if player.team_tag && player.team_tag !== player.team_name}<small class="team-tag">{player.team_tag}</small>{/if}
        {:else}<span class="empty">No team listed</span>{/if}
      </div>
      <p class="source-date">{#if player.dpc_seen_at > 0}Last seen in pro feed <time>{formatDateTime(player.dpc_seen_at / 1000)} UTC</time>{:else}Last seen in pro feed: unknown{/if}</p>
    </article>
    {/if}
    <article class="source">
      <h3>Listed on team roster</h3>
      <div class="team-name">
        {#if roster}<a href={`/team/${roster.id}`}>{roster.name || `Team #${roster.id}`}</a>{#if roster.tag && roster.tag !== roster.name}<small class="team-tag">{roster.tag}</small>{/if}
        {:else}<span class="empty">No stored roster found</span>{/if}
      </div>
      {#if roster}
        <p class="join-date">Roster join date <time>{roster.joined_at > 0 ? formatDate(roster.joined_at) : 'Unknown'}</time></p>
        <p class="source-date">Roster retrieved <time>{roster.retrieved_at > 0 ? `${formatDateTime(roster.retrieved_at)} UTC` : 'Unknown'}</time></p>
      {:else}<p class="source-date">This does not establish that the player is teamless.</p>{/if}
    </article>
  </div>
  {#if roster}<p class="roster-note">If several stored rosters list this account, we show the one with the latest join date.</p>{/if}
</section>
{/if}

<style>
  .team-listings { margin-bottom: 32px; }
  .section-heading { display: flex; align-items: baseline; flex-wrap: wrap; gap: 8px 14px; }
  h2 { margin: 0; }
  .source-status { color: var(--accent); background: var(--accent-soft); padding: 4px 8px; border-radius: var(--radius-sm); font-size: 11px; font-weight: 500; }
  .explanation { color: var(--muted); font-size: 13px; max-width: 65ch; margin: 10px 0 18px; }
  .sources { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); background: var(--surface); border: 1px solid var(--line); border-radius: var(--radius); overflow: hidden; }
  .source { padding: 18px; min-width: 0; }
  .source:only-child { grid-column: 1 / -1; }
  .source + .source { border-left: 1px solid var(--line); }
  h3 { margin: 0 0 12px; color: var(--muted); font-size: 12px; font-weight: 500; }
  .team-name { display: flex; align-items: baseline; flex-wrap: wrap; gap: 6px 10px; font-size: 17px; font-weight: 600; line-height: 1.4; overflow-wrap: anywhere; }
  .team-name > a, .team-name > span { min-width: 0; }
  .team-tag { font-size: 11px; font-weight: 400; color: var(--muted); padding: 2px 5px; background: var(--surface-subtle); border-radius: 2px; }
  .empty { color: var(--muted); font-weight: 400; font-size: 14px; }
  .source-date, .join-date { margin: 14px 0 0; color: var(--muted); font-size: 12px; line-height: 1.6; }
  time { display: block; font-variant-numeric: tabular-nums; }
  .join-date { color: var(--text); }
  .roster-note { font-size: 11px; line-height: 1.6; color: var(--muted); margin: 10px 0 0; }
  @media (max-width: 600px) {
    .sources { grid-template-columns: 1fr; }
    .source + .source { border-left: 0; border-top: 1px solid var(--line); }
    .source { padding: 16px; }
  }
</style>
