<script>
  let { text, label = 'Copy', copiedLabel = 'Copied' } = $props();
  let state = $state('');
  $effect(() => { text; state = ''; });
  async function copy() {
    try { await navigator.clipboard.writeText(String(text)); state = copiedLabel; }
    catch { state = 'Select and copy manually'; }
  }
</script>
<button class="secondary" onclick={copy} aria-live="polite">{state || label}</button>
