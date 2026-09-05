<script>let { items } = $props();</script>
<nav class="breadcrumb" aria-label="Breadcrumb"><ol>{#each items as item}<li>{#if item.href}<a href={item.href}>{item.label}</a>{:else}<span aria-current="page">{item.label}</span>{/if}</li>{/each}</ol></nav>
<style>
  ol { display: flex; flex-wrap: wrap; gap: 6px 10px; list-style: none; padding: 0; margin: 0; }
  li { display: flex; gap: 10px; min-width: 0; overflow-wrap: anywhere; }
  li + li::before { content: '/'; color: var(--text-subtle); }
</style>
