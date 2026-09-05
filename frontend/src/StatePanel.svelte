<script>
  let { kind = 'empty', title, message, actionLabel = '', onaction = undefined, href = '' } = $props();
</script>

<div class="state-panel" class:failure={kind === 'error'} role={kind === 'error' ? 'alert' : 'status'} aria-live={kind === 'error' ? 'assertive' : 'polite'} aria-busy={kind === 'loading'}>
  <span class="state-mark" aria-hidden="true">{kind === 'loading' ? '…' : kind === 'error' ? '!' : kind === 'notFound' ? '404' : '—'}</span>
  <div class="state-content"><h2>{title}</h2>
    {#if message}<p>{message}</p>{/if}
    {#if kind === 'loading'}<div class="skeleton" aria-hidden="true"><span></span><span></span></div>{/if}
    {#if actionLabel && onaction}<button class="secondary" onclick={onaction}>{actionLabel}</button>
    {:else if actionLabel && href}<a class="button secondary" {href}>{actionLabel}</a>{/if}
  </div>
</div>
<style>
  .state-panel { width: 100%; display: flex; align-items: flex-start; gap: 16px; padding: 24px; border: 1px solid var(--line); background: var(--surface); border-radius: var(--radius); margin: 16px 0; }
  .state-mark { display: inline-grid; place-items: center; flex: 0 0 34px; height: 34px; border-radius: var(--radius-sm); background: var(--surface-subtle); color: var(--muted); font-size: 12px; font-weight: 600; }
  .state-content { min-width: 0; flex: 1; }
  h2 { font-size: 16px; margin: 0 0 6px; }
  p { margin: 0; max-width: 65ch; color: var(--muted); font-size: 13px; overflow-wrap: anywhere; }
  button,a { margin-top: 14px; }
  .failure { border-left: 3px solid var(--negative); }
  .failure .state-mark { color: var(--negative); }
  .skeleton { display: grid; gap: 8px; margin-top: 16px; max-width: 400px; }
  .skeleton span { height: 7px; background: var(--surface-subtle); border-radius: 2px; animation: breathe 1.6s ease-in-out infinite alternate; }
  .skeleton span:last-child { width: 62%; }
  @keyframes breathe { to { opacity: .4; } }
  @media(max-width:600px) { .state-panel { padding: 18px; gap: 12px; } }
</style>
