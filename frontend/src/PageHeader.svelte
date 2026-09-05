<script>
  let { title, description = '', eyebrow = '' } = $props();
</script>
<div class="page-heading">
  {#if eyebrow}<p class="eyebrow">{eyebrow}</p>{/if}
  <h1>{title}</h1>
  {#if description}<p class="page-description">{description}</p>{/if}
</div>
