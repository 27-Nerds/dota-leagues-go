<script>let { value } = $props();</script>
{#snippet renderValue(item)}
  {#if item === null || item === undefined || item === ''}—
  {:else if Array.isArray(item)}
    {#if item.length}<ul>{#each item as child}<li>{@render renderValue(child)}</li>{/each}</ul>{:else}—{/if}
  {:else if typeof item === 'object'}
    <dl>{#each Object.entries(item).filter(([key]) => !key.startsWith('_')) as [key,child]}<div><dt>{key.replaceAll('_',' ')}</dt><dd>{@render renderValue(child)}</dd></div>{/each}</dl>
  {:else if typeof item === 'boolean'}{item ? 'Yes' : 'No'}
  {:else}{item}{/if}
{/snippet}
{@render renderValue(value)}
<style>
 dl,ul { margin: 0; padding: 0; } ul { list-style: none; } li + li { margin-top: 6px; }
 dl div + div { margin-top: 6px; } dt { color: var(--muted); font-size: 11px; text-transform: capitalize; } dd { margin: 0; }
</style>
