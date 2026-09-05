<script>
  let { offset, total, pageSize, previousHref, nextHref, label = 'Pages', previousLabel = 'Previous', nextLabel = 'Next' } = $props();
</script>
{#if offset > 0 || offset + pageSize < total}
<nav class="pagination" aria-label={label}>
  {#if offset > 0}<a href={previousHref}>← {previousLabel}</a>{:else}<span class="page-disabled" aria-disabled="true">← {previousLabel}</span>{/if}
  <span class="page-count">Page {Math.floor(offset / pageSize) + 1} of {Math.max(1,Math.ceil(total / pageSize))}</span>
  {#if offset + pageSize < total}<a href={nextHref}>{nextLabel} →</a>{:else}<span class="page-disabled" aria-disabled="true">{nextLabel} →</span>{/if}
</nav>
{/if}
