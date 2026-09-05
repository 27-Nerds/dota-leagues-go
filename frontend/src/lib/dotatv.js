// game_id is the server's uint64 Steam ID, not the match ID. Keep it as a
// string: converting it to a JavaScript number loses precision.
export function dotaTVUrl(serverSteamId) {
  if (typeof serverSteamId !== 'string' || !/^[1-9][0-9]{0,19}$/.test(serverSteamId)) return null;
  if (BigInt(serverSteamId) > 18446744073709551615n) return null;
  return `steam://run/570//${encodeURIComponent(`+watch_server ${serverSteamId}`)}/`;
}
