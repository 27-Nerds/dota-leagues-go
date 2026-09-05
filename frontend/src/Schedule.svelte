<script>
  import { tick } from 'svelte';
  import StatePanel from './StatePanel.svelte';
  import EntityImage from './EntityImage.svelte';
  import SchedulePager from './SchedulePager.svelte';
  import { teamLogo } from './lib/constants.js';
  import { groupSchedule, SCHEDULE_PAGE_SIZE } from './lib/schedule.js';
  let { id, series, total, offset, loading, error, onretry } = $props();
  let groups = $derived(groupSchedule(series));
  let scheduleElement;
  let positioned = false;
  $effect(() => {
    if (!loading && !positioned && window.location.hash === '#schedule') {
      positioned = true;
      tick().then(() => scheduleElement?.scrollIntoView({ block: 'start' }));
    }
  });
  function name(team, id) { return team || (id ? `Team #${id}` : 'To be announced'); }
</script>

<section id="schedule" aria-labelledby="schedule-title" bind:this={scheduleElement}>
  <div class="schedule-heading">
    <div><h2 id="schedule-title">Schedule</h2><p>{#if total > 0}{total.toLocaleString()} series · {/if}Newest first · All times UTC</p></div>
  </div>
  {#if loading}
    <StatePanel kind="loading" title="Loading schedule" message="Checking matchups and game links." />
  {:else if error}
    <StatePanel kind="error" title="Couldn’t load the schedule" message={error} actionLabel="Try again" onaction={onretry} />
  {:else if series.length}
    <p class="range" role="status">Showing series {offset + 1}–{offset + series.length} of {total}</p>
    <div class="schedule">
      {#each groups as group (group.key)}
        <section class="date-group" aria-label={group.label}>
          <div class="date-heading"><h3>{group.label}</h3><span>{group.rows.length} series on this page</span></div>
          {#each group.rows as row}
            <article class="series">
              <div class="start-time"><time datetime={group.key === 'unscheduled' ? undefined : `${group.key}T${row.time}:00Z`}>{row.time}</time></div>
              <div class="matchup">
                <div class="side first">
                  {#if row.team_id_1}<a href="/team/{row.team_id_1}">{name(row.team_name_1,row.team_id_1)}</a>{:else}<span class="tba">To be announced</span>{/if}
                  <span class="team-logo"><EntityImage src={teamLogo(row.team_id_1)} name={row.team_name_1} /></span>
                </div>
                <span class="vs">vs</span>
                <div class="side second">
                  <span class="team-logo"><EntityImage src={teamLogo(row.team_id_2)} name={row.team_name_2} /></span>
                  {#if row.team_id_2}<a href="/team/{row.team_id_2}">{name(row.team_name_2,row.team_id_2)}</a>{:else}<span class="tba">To be announced</span>{/if}
                </div>
              </div>
              <div class="game-links">
                {#each row.match_ids || [] as mid,i}<a href="/match/{id}/{mid}" aria-label={`Game ${i + 1}: ${name(row.team_name_1,row.team_id_1)} vs ${name(row.team_name_2,row.team_id_2)}`}>Game {i+1}<span aria-hidden="true">↗</span></a>
                {:else}<span class="pending">Game links pending</span>{/each}
              </div>
            </article>
          {/each}
        </section>
      {/each}
    </div>
    {#if offset > 0 || total > SCHEDULE_PAGE_SIZE}<div class="bottom-pager"><SchedulePager {id} {offset} {total} /></div>{/if}
  {:else}
    <StatePanel title={offset ? 'No series on this page' : 'No schedule published yet'} message={offset ? 'The schedule may have changed. Return to the first page to see the latest matchups.' : 'Matchups and game links will appear here when they’re available.'} actionLabel={offset ? 'Back to schedule' : 'Refresh schedule'} href={offset ? `/league/${id}#schedule` : ''} onaction={offset ? undefined : onretry} />
  {/if}
</section>

<style>
  #schedule { scroll-margin-top: 20px; }
  .schedule-heading { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; margin: 24px 0 14px; }
  h2 { margin: 0 0 4px; font-size: 21px; }
  .schedule-heading p { margin: 0; color: var(--muted); font-size: 12px; }
  .range { margin: 0 0 14px; color: var(--muted); font-size: 11px; }
  .schedule { border: 1px solid var(--line); border-radius: 9px; overflow: hidden; background: var(--surface); }
  .date-heading { display: flex; flex-wrap: wrap; justify-content: space-between; gap: 8px; align-items: center; padding: 11px 18px; background: var(--surface-subtle); border-block: 1px solid var(--line); }
  .date-group:first-child .date-heading { border-top: 0; }
  h3 { font-size: 13px; font-weight: 500; margin: 0; color: var(--text); }
  .date-heading > span { font-size: 11px; color: var(--muted); }
  .series { display: grid; grid-template-columns: 58px minmax(0,1fr) 240px; align-items: center; gap: 18px; padding: 12px 18px; min-height: 68px; border-bottom: 1px solid var(--line); }
  .series:last-child { border-bottom: 0; }
  .series:hover { background: var(--surface-hover); }
  .start-time { font-variant-numeric: tabular-nums; color: var(--muted); font-size: 12px; }
  .matchup { display: grid; grid-template-columns: minmax(0,1fr) 24px minmax(0,1fr); align-items: center; gap: 10px; }
  .side { display: flex; align-items: center; gap: 10px; min-width: 0; }
  .first { justify-content: flex-end; text-align: right; }
  .side a,.tba { font-size: 13px; font-weight: 500; line-height: 1.4; overflow-wrap: anywhere; }
  .side a { text-decoration: none; } .side a:hover { text-decoration: underline; }
  .tba { color: var(--muted); font-weight: 400; }
  .team-logo { width: 28px; height: 28px; flex-shrink: 0; border-radius: 5px; overflow: hidden; }
  .vs { text-align: center; font-size: 10px; color: var(--muted); }
  .game-links { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 6px; max-width: 260px; }
  .game-links a { display: inline-flex; align-items: center; gap: 6px; border: 1px solid var(--line); border-radius: 4px; padding: 6px 9px; min-height: 34px; font-size: 11px; text-decoration: none; background: var(--surface-hover); white-space: nowrap; }
  .game-links a:hover { background: var(--surface-hover); border-color: var(--line-strong); }
  .game-links a span { color: var(--muted); }
  .pending { font-size: 11px; color: var(--muted); }
  .bottom-pager { display: flex; justify-content: center; margin-top: 0; }
  @media(max-width: 950px) {
    .series { grid-template-columns: 50px minmax(0,1fr); gap: 8px 14px; }
    .game-links { grid-column: 2; max-width: none; justify-content: flex-start; }
  }
  @media(max-width: 600px) {
    .series { padding: 14px; grid-template-columns: 42px minmax(0,1fr); gap: 10px; align-items: start; }
    .start-time { padding-top: 5px; }
    .matchup { grid-template-columns: minmax(0,1fr); gap: 8px; }
    .side { justify-content: flex-start; text-align: left; gap: 8px; }
    .first .team-logo { order: -1; }
    .vs { display: none; }
    .side a,.tba { font-size: 12px; }
    .team-logo { width: 24px; height: 24px; }
    .date-heading { padding: 10px 14px; }
    .date-heading > span { display: none; }
    .game-links a { min-height: 38px; }
  }
</style>
