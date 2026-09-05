<script>
  import EntityImage from './EntityImage.svelte';
  import { regionText, tierText, formatDate, formatMoney, leagueLogo } from './lib/constants.js';
  let { league } = $props();
</script>
<article class="league">
  <a class="cover" href="/league/{league.league_id}" aria-label={`View ${league.name}`} tabindex="-1">
    <EntityImage src={leagueLogo(league.league_id)} name={league.name} wide contain />
  </a>
  <div class="content">
    <div class="metadata"><p class="eyebrow" title={`${regionText(league.region)} · ${tierText(league.tier)}`}>{regionText(league.region)} · {tierText(league.tier)}</p>{#if league.is_live}<span class="badge"><span aria-hidden="true">●</span> Live now</span>{/if}</div>
    <h2><a href="/league/{league.league_id}">{league.name || 'Unnamed league'}</a></h2>
    <div class="facts"><span>{formatDate(league.start_timestamp)} – {formatDate(league.end_timestamp)}</span><strong>{formatMoney(league.total_prize_pool)}</strong></div>


  </div>
</article>
<style>
  .league { min-width: 0; border: 1px solid var(--line); border-radius: 10px; background: var(--surface); overflow: hidden; display: flex; flex-direction: row; align-items: center; }
  .cover { display: block; width: 112px; height: 86px; flex: 0 0 112px; margin: 16px 0 16px 16px; border-radius: 5px; overflow: hidden; position: relative; background: var(--surface-subtle); }
  .metadata { display: grid; grid-template-columns: minmax(0,1fr) 72px; align-items: center; gap: 10px; min-height: 20px; margin-bottom: 8px; }
  .badge { justify-self: end; display: inline-flex; align-items: center; white-space: nowrap; background: var(--gold); color: var(--text); padding: 3px 8px; border-radius: 4px; font-size: 10px; font-weight: 500; line-height: 1.4; }
  .badge span { color: var(--accent); margin-right: 4px; }
  .content { min-width: 0; padding: 16px; flex: 1; display: flex; flex-direction: column; }
  .eyebrow { font-size: 10px; letter-spacing: 0; margin: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  h2 { font-size: 16px; margin: 0 0 8px; letter-spacing: -.4px; overflow-wrap: anywhere; }
  h2 a { color: var(--text); text-decoration: none; } h2 a:hover { text-decoration: underline; }
  .facts { display: flex; justify-content: space-between; flex-wrap: wrap; gap: 6px; font-size: 12px; color: var(--muted); }
  .facts strong { color: var(--muted); }
  .league:hover { border-color: var(--line-strong); }
  @media(max-width:600px) { .cover { width: 72px; flex-basis: 72px; height: 64px; margin: 12px 0 12px 12px; } .content { padding: 12px; } h2 { font-size: 15px; } .facts { font-size: 11px; } .metadata { grid-template-columns: minmax(0,1fr) 64px; gap: 8px; } .badge { font-size: 9px; padding: 2px 4px; } }

</style>
