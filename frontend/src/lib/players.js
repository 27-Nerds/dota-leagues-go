// Steam data is kept from the last successful public lookup even after the profile goes private,
// so the visibility state is derived separately from whether any data exists.
export function steamState(player) {
  if (!player) return null;
  const hasData = Boolean(player.steam_name);
  const status = player.steam_status || '';
  if (!status && !hasData) return { kind: 'unchecked', label: 'Not checked yet', hasData };
  if (status === 'unavailable') return { kind: 'unavailable', label: 'Unavailable', hasData };
  if (status === 'request_failed' && !hasData) return { kind: 'unknown', label: 'Check failed', hasData };
  const privacy = player.steam_privacy || 'public';
  if (privacy === 'private') return { kind: 'private', label: 'Private', hasData };
  if (privacy === 'friendsonly') return { kind: 'private', label: 'Friends only', hasData };
  return { kind: 'public', label: 'Public', hasData };
}

// Valve stopped sending is_pro; a profile that came from the pro player feed is the professional signal.
export function isListedPro(player) {
  return Boolean(player) && player.profile_source !== 'steam';
}

// Newest registration period first; entries without a date are dropped.
export function proRegistrations(player) {
  const entries = Array.isArray(player?.pro_registration) ? player.pro_registration.filter(r => r && r.timestamp > 0) : [];
  return entries.sort((a, b) => b.timestamp - a.timestamp);
}

export function proRegistration(player) {
  return proRegistrations(player)[0] || null;
}

export function teamHistory(player) {
  return (Array.isArray(player?.audit_entries) ? player.audit_entries : []).filter(h => h && h.team_id > 0).sort((a, b) => (b.start_timestamp || 0) - (a.start_timestamp || 0));
}

export function playerDisplayName(player, id) {
  return player?.name?.trim() || player?.steam_name?.trim() || `Player #${id}`;
}
