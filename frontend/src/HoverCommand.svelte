<script>
  import CopyButton from "./CopyButton.svelte";
  import { dotaTVUrl } from './lib/dotatv.js';
  let { leagueId = null, serverSteamId = null } = $props();
  let commandText = $derived(serverSteamId ? `watch_server ${serverSteamId}` : `dota_spectator_auto_spectate_games ${leagueId}`);
  let steamUrl = $derived(serverSteamId ? dotaTVUrl(serverSteamId) : `steam://run/570//${encodeURIComponent(`+${commandText}`)}/`);

</script>

<section class="spectate" aria-label="Spectate in Dota 2">
  <h3>Spectate in Dota 2</h3>
  <div class="command-row">
    <code aria-label="Dota 2 console command">{commandText}</code>
    <div class="actions"><CopyButton text={commandText} label="Copy command" /><a class="button" href={steamUrl}>Open in Steam</a></div>
  </div>
</section>
<style>
  .spectate { min-width: 0; margin-top: 16px; padding: 16px; border: 1px solid var(--line); border-radius: var(--radius); background: var(--surface-subtle); }
  h3 { margin: 0 0 10px; color: var(--text); font-size: 13px; font-weight: 500; }
  .command-row { display: flex; flex-wrap: wrap; align-items: center; gap: 12px; }
  code { display: block; flex: 1 1 300px; min-width: 0; font-family: 'Fira Code', monospace; font-size: 12px; line-height: 1.6; overflow-wrap: anywhere; color: var(--text); background: var(--surface); border: 1px solid var(--line); padding: 10px 12px; border-radius: var(--radius-sm); user-select: all; }
  .actions { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
  a { font-size: 12px; white-space: nowrap; }
  @media(max-width:600px) { .spectate { padding: 12px; } code { flex-basis: 100%; font-size: 11px; } }
</style>
