<script>
  import heroes from "./lib/heroes.js";
  import items from "./lib/items.js";
  import { playerHref } from "./lib/constants.js";
  import EntityImage from "./EntityImage.svelte";
  export let p;
  $: hero = p.hero_id ? heroes[p.hero_id] : null;
  $: profile = playerHref(p.account_id);
  $: inventory = Array.from({ length:6 }, (_, index) => Array.isArray(p.items) ? p.items[index] || 0 : 0);
</script>
<div class="player-row">
  <div class="player-hero"><EntityImage src={hero?.img || hero?.icon} alt={hero?.name || "Unknown hero"} name={hero?.name || "Unknown hero"} contain /></div>
  <div class="player-info">
    {#if profile}<a href={profile}>{p.pro_name || `Player #${p.account_id}`}</a>{:else}<span>{p.pro_name || "Anonymous player"}</span>{/if}
    <div class="hero-name">{hero?.name || "Unknown hero"}</div>
  </div>
  <div class="kda" aria-label={`Kills ${p.kills ?? "unknown"}, deaths ${p.deaths ?? "unknown"}, assists ${p.assists ?? "unknown"}, level ${p.level ?? "unknown"}`}><b>{p.kills ?? "-"}</b> / <b>{p.deaths ?? "-"}</b> / <b>{p.assists ?? "-"}</b><span>Level {p.level ?? "-"}</span></div>
  <div class="items-row" aria-label="Items">{#each inventory as itemId}
    <span class="item-slot" title={itemId ? items[itemId]?.name || `Unknown item #${itemId}` : "Empty item slot"}>
      {#if itemId && items[itemId]}<span class="item-empty" aria-hidden="true">?</span><img src={items[itemId].icon} alt={items[itemId].name} onerror={e => { e.currentTarget.hidden = true; }} />{:else}<span class="item-empty" aria-label={itemId ? `Unknown item #${itemId}` : "Empty item slot"}>{itemId ? "?" : "·"}</span>{/if}
    </span>
  {/each}</div>
</div>
<style>
  .player-row { display:grid; grid-template-columns:64px minmax(0,1fr) auto; align-items:center; column-gap:12px; row-gap:6px; padding:14px 16px; border-bottom:1px solid var(--surface-subtle); }
  .player-hero { width:64px; height:36px; border-radius:2px; overflow:hidden; align-self:start; margin-top:3px; grid-row:1 / 3; } .player-info { min-width:0; } .player-info a,.player-info > span { display:block; color:var(--text); text-decoration:none; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; } .player-info a:hover {text-decoration:underline;} .hero-name {color:var(--muted); font-size:12px;}
  .items-row { grid-column:2 / 4; display:flex; gap:4px; } .item-slot {position:relative; display:flex; justify-content:center; align-items:center; width:30px; height:24px; background:var(--surface-subtle); border:1px solid var(--line); border-radius:2px; overflow:hidden;} .item-slot img {position:absolute; inset:0; width:100%; height:100%; object-fit:contain;} .item-empty {color:var(--muted); font-size:12px;}
  .kda {font-family:'Fira Code',monospace; font-variant-numeric:tabular-nums; font-size:12px; color:var(--muted); white-space:nowrap; text-align:right;} .kda span {display:block; color:var(--muted); font-family:'Fira Sans',sans-serif; font-size:11px;}
  @media(max-width:400px) { .player-row {padding:12px; column-gap:8px; grid-template-columns:48px minmax(0,1fr) auto;} .player-hero {width:48px; height:27px;} .kda {font-size:11px;} .item-slot {width:26px;} }
</style>
