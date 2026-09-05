<script>
  let { label, searchContent, children, optionsContent, sortContent, onreset, resetDisabled = false } = $props();
</script>

<section class="filter-bar" aria-label={label}>
  <div class="primary">
    <div class="search">{@render searchContent()}</div>
    <div class="fields">
      {@render children()}
      {#if sortContent}<div class="sorting">{@render sortContent()}</div>{/if}
    </div>
    <button class="reset" onclick={onreset} disabled={resetDisabled}>Clear filters</button>
  </div>
  {#if optionsContent}<div class="options">{@render optionsContent()}</div>{/if}
</section>

<style>
  .filter-bar { margin: 1.5rem 0 .75rem; padding-bottom: .875rem; border-bottom: 1px solid var(--line); }
  .primary { display: flex; align-items: end; gap: .875rem; }
  .search { flex: 1; min-width: 14rem; }
  .fields { display: grid; grid-auto-flow: column; grid-auto-columns: minmax(0,9rem); gap: .875rem; }
  .fields :global(label), .options :global(label), .sorting :global(label) { display: flex; flex-direction: column; gap: .375rem; min-width: 0; font-size: .75rem; line-height: 1.25rem; font-weight: 500; color: var(--muted); }
  .filter-bar :global(select) { width: 100%; min-height: 2.75rem; border: 1px solid var(--line-strong); background: var(--surface); color: var(--text); padding: .5rem .625rem; border-radius: .3125rem; max-width: 100%; font-size: .8125rem; transition: border-color 180ms ease, background 180ms ease; }
  .filter-bar :global(select:hover:not(:disabled)) { border-color: var(--accent); }
  .filter-bar :global(select:disabled) { opacity: .5; cursor: not-allowed; }
  .options { margin-top: .75rem; }
  .options :global(label) { flex-direction: row; align-items: center; min-height: 1.5rem; font-size: .8125rem; gap: .5rem; cursor: pointer; }
  .options :global(label:has(input:checked)) { color: var(--accent); }
  .options :global(input[type=checkbox]) { accent-color: var(--accent); width: 1rem; height: 1rem; margin: 0; cursor: pointer; }
  .reset { flex-shrink: 0; min-height: 2.75rem; padding: .5rem 0; background: transparent; border: 0; border-radius: .125rem; color: var(--muted); font-size: .8125rem; font-weight: 500; box-shadow: none; transition: color 180ms ease; }
  .reset:hover:not(:disabled) { background: transparent; color: var(--text); text-decoration: underline; text-underline-offset: .2em; }
  .reset:active:not(:disabled) { transform: translateY(1px); }
  .reset:disabled { cursor: default; opacity: .45; }
  @media(max-width:1100px) { .primary { display: grid; grid-template-columns: minmax(0,1fr) auto; } .search { grid-column: 1 / -1; min-width: 0; } .fields { grid-auto-columns: minmax(0,1fr); } }
  @media(max-width:600px) { .primary, .fields { gap: .625rem; } .fields { grid-column: 1 / -1; grid-auto-flow: row; grid-template-columns: repeat(2,minmax(0,1fr)); } .reset { grid-column: 2; justify-self: end; min-height: 2rem; } }
  @media(prefers-reduced-motion:reduce) { .filter-bar :global(select), .reset { transition: none; } }
</style>
