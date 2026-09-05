const id = /^[1-9][0-9]*$/;

export function parsePath(pathname) {
  if (pathname === '/') return { name: 'home' };
  if (pathname === '/activity') return { name: 'updates' };
  if (pathname === '/team') return { name: 'teams' };
  if (pathname === '/dpc') return { name: 'dpc' };
  const parts = pathname.split('/').slice(1);
  if (parts.length === 2 && id.test(parts[1])) {
    if (parts[0] === 'league') return { name: 'league', id: parts[1] };
    if (parts[0] === 'team') return { name: 'team', id: parts[1] };
  }
  if (parts.length === 3 && parts[0] === 'match' && id.test(parts[1]) && id.test(parts[2])) {
    return { name: 'match', leagueId: parts[1], matchId: parts[2] };
  }
  return { name: 'notFound' };
}

export function legacyDestination(hash) {
  if (!hash.startsWith('#/')) return null;
  const path = hash.slice(1);
  return parsePath(path.split('?')[0]).name === 'notFound' ? null : path;
}

export function pageOffset(search, pageSize = 100) {
  const value = new URLSearchParams(search).get('offset');
  const offset = Number(value);
  return Number.isSafeInteger(offset) && offset >= 0 && offset <= 10000000 && offset % pageSize === 0 ? offset : 0;
}
