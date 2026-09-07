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

export function playerDisplayName(player, id) {
  return player?.name?.trim() || player?.steam_name?.trim() || `Player #${id}`;
}
