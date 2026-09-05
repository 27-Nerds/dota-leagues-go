<script>
  let { src, alt = '', name = '', wide = false, contain = false } = $props();
  let failed = $state(false);
  $effect(() => { src; failed = false; });
</script>

<div class="entity-image" class:wide class:contain>
  {#if src && !failed}
    <img {src} {alt} loading="lazy" onerror={() => failed = true} />
  {:else}
    <span class="fallback" role={alt ? 'img' : undefined} aria-label={alt || undefined}>
      <span aria-hidden="true">—</span>
      {#if wide}<small>{name || 'Image unavailable'}</small>{/if}
    </span>
  {/if}
</div>

<style>
  .entity-image { width: 100%; height: 100%; min-width: 0; overflow: hidden; background: var(--surface-subtle); border-radius: inherit; display: grid; place-items: center; }
  img { width: 100%; height: 100%; object-fit: contain; filter: drop-shadow(0 0 .5px #68745f); }
  .wide img { object-fit: cover; }
  .contain img { object-fit: contain; }
  .fallback { color: var(--muted); text-align: center; display: grid; gap: 8px; padding: 12px; width: 100%; }
  .fallback > span { font-size: 24px; }
  small { font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
