<script>
  let { src } = $props();
  let failed = $state(false);
  $effect(() => { src; failed = false; });
</script>

{#if src && !failed}
  <img class="player-avatar" {src} alt="" width="32" height="32" loading="lazy" decoding="async" onerror={() => failed = true} />
{/if}

<style>
  .player-avatar { width: 2rem; height: 2rem; flex: 0 0 2rem; object-fit: cover; border-radius: .25rem; background: var(--surface-subtle); }
</style>
