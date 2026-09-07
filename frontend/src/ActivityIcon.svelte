<script>
  // Existing Lucide icons and matching activity-specific line symbols. Lucide: https://lucide.dev — license in public/licenses/lucide.txt.
  let { name = 'file-clock' } = $props();
</script>
<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false" data-icon={name}>
  {#if name === 'swap'}
    <path d="M4 7h16m-4-4 4 4-4 4M20 17H4m4-4-4 4 4 4" />
  {:else if name === 'rebuild'}
    <path d="M3 8a9 9 0 0 1 15-3l3 3M21 3v5h-5M21 16a9 9 0 0 1-15 3l-3-3M3 21v-5h5" />
  {:else if name === 'roster-alert'}
    <path d="M12 3 2 21h20L12 3Z M12 9v5m0 3v.1" />
  {:else if name === 'reform'}
    <circle cx="9" cy="7" r="4" /><path d="M2 21v-2a7 7 0 0 1 14 0v2m3-13v6m-3-3h6" />
  {:else if name === 'user-plus' || name === 'user-minus'}
    <circle cx="9" cy="7" r="4" /><path d="M2 21v-2a7 7 0 0 1 14 0v2m1-11h6" />
    {#if name === 'user-plus'}<path d="M20 7v6" />{/if}
  {:else if name === 'identity'}
    <path d="M3 3h8l10 10-8 8L3 11V3Z" /><circle cx="7.5" cy="7.5" r=".5" />
  {:else if name === 'captain'}
    <path d="m3 6 5 4 4-6 4 6 5-4-2 12H5L3 6Zm3 15h12" />
  {:else if name === 'calendar-clock'}
    <path d="M20 10V5a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h5M7 1v4m9-4v4M3 8h17" />
    <circle cx="17" cy="17" r="5" /><path d="M17 14v3l2 1" />
  {:else if name === 'flag'}
    <path d="M4 22V4a1 1 0 0 1 .4-.8A6 6 0 0 1 8 2c3 0 5 2 7.333 2q2 0 3.067-.8A1 1 0 0 1 20 4v10a1 1 0 0 1-.4.8A6 6 0 0 1 16 16c-3 0-5-2-7.333-2A6 6 0 0 0 4 15" />
  {:else if name === 'circle-check'}
    <circle cx="12" cy="12" r="10" /><path d="m9 12 2 2 4-4" />
  {:else if name === 'circle-x'}
    <circle cx="12" cy="12" r="10" /><path d="m15 9-6 6m0-6 6 6" />
  {:else if name === 'lock' || name === 'unlock'}
    <rect x="3" y="11" width="18" height="11" rx="2" />
    {#if name === 'lock'}<path d="M7 11V7a5 5 0 0 1 10 0v4" />{:else}<path d="M7 11V7a5 5 0 0 1 9.9-1" />{/if}
  {:else if name === 'trophy'}
    <path d="M10 14.66V17a1 1 0 0 1-1 1 2 2 0 0 0-2 2v2" />
    <path d="M14 14.66V17a1 1 0 0 0 1 1 2 2 0 0 1 2 2v2" />
    <path d="M17.916 10H19.5A2.5 2.5 0 0 0 22 7.5V5a1 1 0 0 0-1-1h-3" />
    <path d="M4 22h16" />
    <path d="M6 9a6 6 0 0 0 12 0V3a1 1 0 0 0-1-1H7a1 1 0 0 0-1 1z" />
    <path d="M6.084 10H4.5A2.5 2.5 0 0 1 2 7.5V5a1 1 0 0 1 1-1h3" />
  {:else if name === 'shield'}
    <path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z" />
  {:else if name === 'users-round'}
    <path d="M18 21a8 8 0 0 0-16 0" />
    <circle cx="10" cy="8" r="5" />
    <path d="M22 20c0-3.37-2-6.5-4-8a5 5 0 0 0-.45-8.3" />
  {:else if name === 'user-round'}
    <circle cx="12" cy="8" r="5" />
    <path d="M20 21a8 8 0 0 0-16 0" />
  {:else if name === 'coins'}
    <path d="M13.744 17.736a6 6 0 1 1-7.48-7.48" />
    <path d="M15 6h1v4" />
    <path d="m6.134 14.768.866-.5 2 3.464" />
    <circle cx="16" cy="8" r="6" />
  {:else if name === 'chart-no-axes-column'}
    <path d="M5 21v-6" />
    <path d="M12 21V3" />
    <path d="M19 21V9" />
  {:else}
    <path d="M16 22h2a2 2 0 0 0 2-2V8a2.4 2.4 0 0 0-.706-1.706l-3.588-3.588A2.4 2.4 0 0 0 14 2H6a2 2 0 0 0-2 2v2.85" />
    <path d="M14 2v5a1 1 0 0 0 1 1h5" />
    <path d="M8 14v2.2l1.6 1" />
    <circle cx="8" cy="16" r="6" />
  {/if}
</svg>
<style>
  svg { display: block; width: 18px; height: 18px; }
  @media(max-width:600px) { svg { width: 16px; height: 16px; } }
</style>
