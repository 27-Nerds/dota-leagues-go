<script>
 import DataValue from "./DataValue.svelte";
  import PageHeader from "./PageHeader.svelte";
 import { onMount } from 'svelte';
 import { api } from './lib/api.js';
 import StatePanel from './StatePanel.svelte';
 let data = null;
 let error = '';
 let loading = true;
 const sections = [['results','League results'],['standings','Standings'],['major_wildcard_standings','Major wildcard standings'],['major_group_standings','Major group standings'],['major_playoff_standings','Major playoff standings']];
 async function load() { loading=true; error=''; try { const body=await api('/dpc/standings'); data=body?.results || {}; } catch(err) {error=err.message;} finally {loading=false;} }
 onMount(load);
 function rows(key) { return Array.isArray(data?.[key]) ? data[key].filter(row => row && typeof row === 'object' && !Array.isArray(row)) : []; }
 function columns(rows) { return [...new Set(rows.flatMap(Object.keys).filter(key => !key.startsWith('_')))]; }
 $: hasData = data && sections.some(([key]) => rows(key).length > 0);
</script>
<svelte:head><title>DPC Standings | Dota 2 Leagues</title></svelte:head>
<PageHeader title="DPC standings" description="Available league results and standings from the Dota Pro Circuit." eyebrow="Dota Pro Circuit" />
{#if loading}<StatePanel kind="loading" title="Loading standings" message="Checking the latest available records." />
{:else if error}<StatePanel kind="error" title="Couldn’t load standings" message={error} actionLabel="Try again" onaction={load} />
{:else if !hasData}<StatePanel title="No DPC standings available" message="The standings feed has no records to display. You can still explore current tournaments and their match results." actionLabel="Browse leagues" href="/" />
{:else}
 {#each sections as [key,label]}
 {@const entries = rows(key)}
 {#if entries.length}
 <section><h2>{label}</h2>
<!-- svelte-ignore a11y_no_noninteractive_tabindex (The table region must be keyboard-scrollable.) -->
<div class="table-scroll" role="region" aria-label={label} tabindex="0"><table><thead><tr>{#each columns(entries) as col}<th scope="col">{col.replaceAll('_',' ')}</th>{/each}</tr></thead><tbody>{#each entries as row}<tr>{#each columns(entries) as col}<td class:entity={/name|team/.test(col)}><DataValue value={row[col]} /></td>{/each}</tr>{/each}</tbody></table></div></section>
 {/if}
 {/each}
{/if}
<style>td { max-width: 280px; overflow-wrap: anywhere; min-width: 120px; } .entity { min-width: 200px; } th { text-transform: capitalize; }</style>
