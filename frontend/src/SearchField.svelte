<script>let { label, placeholder, value, oninput } = $props();</script>
<label class="search-field"><span>{label}</span><span class="search-input"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true"><circle cx="10.5" cy="10.5" r="6.5"/><path d="m16 16 5 5"/></svg><input type="search" {placeholder} {value} {oninput} /></span></label>
<style>
  .search-field { display: grid; gap: 6px; min-width: 0; font-size: 12px; line-height: 20px; font-weight: 500; color: var(--muted); }
  .search-input { position: relative; display: block; }
  svg { position: absolute; width: 17px; height: 17px; top: 13px; left: 12px; color: var(--text-subtle); pointer-events: none; }
  input { padding-left: 37px; width: 100%; }
</style>
