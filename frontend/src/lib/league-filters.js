// Finish loading the catalogue before deciding that a filter has no matches.
export async function loadLeagueCatalog(fetchPage, firstPage = null) {
  const rows = [];
  let offset = 0;
  let page = firstPage || await fetchPage(0);
  while (true) {
    if (!page.results.length && offset < page.meta.total) {
      throw new Error('The league list changed while searching. Please try again.');
    }
    rows.push(...page.results);
    offset += page.results.length;
    if (offset >= page.meta.total) break;
    page = await fetchPage(offset);
  }
  return [...new Map(rows.map(league => [league.league_id, league])).values()];
}

export function filterLeagues(leagues, { search = '', region = '', tier = '', liveOnly = false }) {
  const query = search.trim().toLowerCase();
  return leagues.filter(league =>
    (!liveOnly || league.is_live) &&
    (!region || String(league.region) === region) &&
    (!tier || String(league.tier) === tier) &&
    (!query || (league.name || '').toLowerCase().includes(query)));
}
