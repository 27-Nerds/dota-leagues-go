import { countryText, fantasyRoleText, formatDateTime, formatMoney, LEAGUE_STATUS, REGIONS, TIERS } from './constants.js';

export const UPDATE_PAGE_SIZE = 20;
const entities = { tournament: 'Tournament', team: 'Team', roster: 'Roster', player: 'Player' };
const labels = {
  name: 'Name', tag: 'Team tag', region: 'Region', country_code: 'Country', url: 'Website',
  url_logo: 'Logo URL', wins: 'Wins', losses: 'Losses', team_captain: 'Captain', tier: 'Tier',
  start_timestamp: 'Starts (UTC)', end_timestamp: 'Ends (UTC)', status: 'Status',
  total_prize_pool: 'Prize pool', description: 'Description', members: 'Roster members', admins: 'Administrators',
  real_name: 'Real name', team: 'Team', team_id: 'Team ID', fantasy_role: 'Role', is_pro: 'Professional', sponsor: 'Sponsor',
  total_earnings: 'Earnings', steam_name: 'Steam name', steam_location: 'Steam location', steam_profile: 'Steam profile'
};
const steamProfileLabels = { public: 'Public', private: 'Private', friendsonly: 'Friends only', unavailable: 'Unavailable' };
export function fieldLabel(field) { return labels[field] || String(field || 'Details').replaceAll('_', ' '); }
export function updateValue(field, value) {
  if (value === null || value === undefined || value === '') return 'Not set';
  if (field === 'total_prize_pool' || field === 'total_earnings') return formatMoney(value);
  if (field === 'fantasy_role') return fantasyRoleText(value) || `Role ${value}`;
  if (field === 'steam_profile') return steamProfileLabels[value] || String(value);
  if (field === 'region') return REGIONS[value] || `Region ${value}`;
  if (field === 'tier') return TIERS[value] || `Tier ${value}`;
  if (field === 'status') return LEAGUE_STATUS[value] || `Status ${value}`;
  if (field === 'country_code') return countryText(value) || String(value);
  if (field.endsWith('_timestamp')) return formatDateTime(value);
  if (field === 'admins') return Object.entries(value || {}).filter(([, admin]) => admin === true).map(([id]) => `Account #${id}`).join(', ') || 'None';
  if (field === 'team_captain') return Number(value) > 0 ? `Account #${value}` : 'Not set';
  if (typeof value === 'boolean') return value ? 'Yes' : 'No';
  if (typeof value === 'number') return value.toLocaleString('en-US');
  if (typeof value === 'object') return JSON.stringify(value);
  return String(value);
}

// The roster payload is a map of account IDs to active flags, not player names.
export function rosterChanges(before, after) {
  const old = before && typeof before === 'object' ? before : {};
  const next = after && typeof after === 'object' ? after : {};
  return [...new Set([...Object.keys(old), ...Object.keys(next)])].sort((a,b) => Number(a)-Number(b)).flatMap(id => {
    if (!Object.hasOwn(old, id)) return [{ id, kind: 'joined', label: 'Added', status: next[id] ? 'Active' : 'Inactive' }];
    if (!Object.hasOwn(next, id)) return [{ id, kind: 'left', label: 'Removed', status: old[id] ? 'Previously active' : 'Previously inactive' }];
    if (old[id] !== next[id]) return [{ id, kind: next[id] ? 'activated' : 'deactivated', label: next[id] ? 'Activated' : 'Deactivated', status: next[id] ? 'Inactive → Active' : 'Active → Inactive' }];
    return [];
  });
}
// These describe observed roster transitions, not confirmed transfer announcements.
export function rosterTransition(update) {
  const change = update.changes?.find(c => c.field === 'members');
  if (!change) return null;
  const active = value => new Set(Object.entries(value || {}).filter(([id, flag]) => Number(id) > 0 && flag === true).map(([id]) => id));
  const before = active(change.before), after = active(change.after);
  const left = [...before].filter(id => !after.has(id));
  const joined = [...after].filter(id => !before.has(id));
  const admins = update.roster_admins || update.changes?.find(c => c.field === 'admins')?.after || {};
  const playersRemaining = [...after].filter(id => admins[id] !== true).length;
  let title = 'Roster updated';
  if (left.length >= 4 && joined.length === 0 && playersRemaining <= 1) title = 'Possible disband';
  else if (before.size <= 1 && joined.length >= 4) title = 'Roster reformed';
  else if (left.length === 1 && joined.length === 1) title = 'Player replaced';
  else if (left.length >= 2 && joined.length >= 2) title = 'Roster rebuilt';
  else if (joined.length && !left.length) title = joined.length === 1 ? 'Player joined' : 'Players joined';
  else if (left.length && !joined.length) title = left.length === 1 ? 'Player left' : 'Players left';
  return { title, left: left.length, joined: joined.length };
}
// ELeagueStatus lifecycle: 0 Not set, 1 Unsubmitted, 2 Submitted, 3 Accepted, 4 Rejected, 5 Concluded, 6 Deleted.
const statusTitles = { 3: 'Tournament approved', 4: 'Tournament rejected', 5: 'Tournament concluded', 6: 'Tournament deleted' };
export function tournamentStatusTitle(changes) {
  const change = changes.find(c => c.field === 'status');
  if (!change || change.before === change.after) return null;
  return statusTitles[change.after] || 'Tournament status changed';
}
const fieldsOnly = (changes, allowed) => changes.length > 0 && changes.every(c => allowed.includes(c.field));
export function playerChangeTitle(changes) {
  const team = changes.find(c => c.field === 'team_id');
  if (team) {
    const before = Number(team.before) > 0, after = Number(team.after) > 0;
    if (before && after) return 'Player transferred';
    if (after) return 'Player joined team';
    if (before) return 'Player left team';
  }
  const steam = changes.find(c => c.field === 'steam_profile');
  if (steam) {
    if (steam.after === 'public') return 'Steam profile public';
    if (steam.after === 'unavailable') return 'Steam profile unavailable';
    if (steam.after) return 'Steam profile private';
  }
  const pro = changes.find(c => c.field === 'is_pro');
  if (pro && pro.after === true && pro.before !== true) return 'Pro profile added';
  if (changes.some(c => c.field === 'name')) return 'Player renamed';
  if (fieldsOnly(changes, ['total_earnings'])) return 'Earnings updated';
  if (fieldsOnly(changes, ['steam_name', 'steam_location'])) return 'Steam profile updated';
  return null;
}
export function updatePresentation(update) {
  const entity = entities[update.entity] || 'Record';
  const changes = Array.isArray(update.changes) ? update.changes : [];
  const created = update.action === 'created';
  let title = `${entity} ${created ? 'added' : 'updated'}`;
  let icon = { tournament: 'trophy', team: 'shield', roster: 'users-round', player: 'user-round' }[update.entity] || 'file-clock';
  if (!created && update.entity === 'tournament' && changes.length === 1 && changes[0].field === 'total_prize_pool') {
    title = 'Prize pool updated'; icon = 'coins';
  }
  if (!created && update.entity === 'team' && changes.length && changes.every(c => ['wins','losses'].includes(c.field))) {
    title = 'Team record updated'; icon = 'chart-no-axes-column';
  }
  if (!created) {
    if (update.entity === 'roster') title = rosterTransition(update)?.title || title;
    if (update.entity === 'team' && changes.some(c => ['name', 'tag'].includes(c.field))) title = 'Team rebranded';
    if (update.entity === 'team' && changes.length === 1 && changes[0].field === 'team_captain') title = 'Captain changed';
    if (update.entity === 'tournament' && changes.some(c => ['start_timestamp','end_timestamp'].includes(c.field))) title = 'Schedule changed';
    if (update.entity === 'tournament') title = tournamentStatusTitle(changes) || title;
    if (update.entity === 'player') title = playerChangeTitle(changes) || title;
  }
  const treatments = {
    'Player replaced': ['swap', 'change'], 'Roster rebuilt': ['rebuild', 'change'],
    'Possible disband': ['roster-alert', 'caution'], 'Roster reformed': ['reform', 'positive'],
    'Player joined': ['user-plus', 'positive'], 'Players joined': ['user-plus', 'positive'],
    'Player left': ['user-minus', 'departure'], 'Players left': ['user-minus', 'departure'],
    'Team rebranded': ['identity', 'change'], 'Captain changed': ['captain', 'change'],
    'Schedule changed': ['calendar-clock', 'change'], 'Prize pool updated': ['coins', 'financial'],
    'Team record updated': ['chart-no-axes-column', 'neutral'],
    'Tournament concluded': ['flag', 'neutral'], 'Tournament approved': ['circle-check', 'positive'],
    'Tournament rejected': ['circle-x', 'departure'], 'Tournament deleted': ['circle-x', 'departure'],
    'Tournament status changed': ['trophy', 'change'],
    'Player joined team': ['user-plus', 'positive'], 'Player left team': ['user-minus', 'departure'], 'Player transferred': ['swap', 'change'],
    'Player renamed': ['identity', 'change'], 'Earnings updated': ['coins', 'financial'], 'Steam profile private': ['lock', 'caution'],
    'Steam profile public': ['unlock', 'positive'], 'Steam profile unavailable': ['lock', 'caution'], 'Steam profile updated': ['user-round', 'neutral'],
    'Pro profile added': ['circle-check', 'positive']
  };
  const [eventIcon, tone] = treatments[title] || [icon, 'neutral'];
  icon = eventIcon;
  const inferred = title === 'Possible disband';
  const transition = !created && update.entity === 'roster' ? rosterTransition(update) : null;
  const path = update.entity === 'tournament' ? 'league' : ['team','roster'].includes(update.entity) ? 'team' : update.entity === 'player' ? 'player' : null;
  const href = path && /^[1-9][0-9]*$/.test(String(update.entity_id)) ? `/${path}/${update.entity_id}` : null;
  return { entity, title, icon, tone, inferred, transition, created, changes, href, name: update.name || `${entity} #${update.entity_id}`, note: {
    tournament: 'Tournament added to the directory.', team: 'Team added to the directory.',
    roster: 'First roster recorded for this team.', player: 'Player added to the database.'
  }[update.entity] || 'Record added to the directory.' };
}
export function updateDate(timestamp) {
  const date = new Date(Number(timestamp));
  if (!timestamp || !Number.isFinite(date.getTime())) return { key: 'unknown', label: 'Date unavailable', time: '—', iso: undefined };
  return { key: date.toISOString().slice(0,10), iso: date.toISOString(),
    label: new Intl.DateTimeFormat('en-GB', { dateStyle: 'full', timeZone: 'UTC' }).format(date),
    time: new Intl.DateTimeFormat('en-GB', { timeStyle: 'short', timeZone: 'UTC', hourCycle:'h23' }).format(date) };
}
export function groupUpdates(updates) {
  const groups = new Map();
  for (const update of updates) {
    const date = updateDate(update.created_at);
    if (!groups.has(date.key)) groups.set(date.key, { ...date, rows: [] });
    groups.get(date.key).rows.push(update);
  }
  return [...groups.values()];
}

export function updateSummary(update) {
  const changes = Array.isArray(update.changes) ? update.changes : [];
  const membership = changes.find(c => c.field === 'members');
  if (!changes.length && update.sources?.length) return 'No net change';
  if (membership) {
    const counts = new Map();
    for (const member of rosterChanges(membership.before, membership.after)) {
      const label = member.label.toLowerCase();
      counts.set(label, (counts.get(label) || 0) + 1);
    }
    return [...counts].map(([label, count]) => `${count} ${label}`).join(' · ') || 'Roster updated';
  }
  if (changes.length <= 2 && changes.every(c => ['total_prize_pool', 'wins', 'losses', 'status', 'total_earnings', 'steam_profile', 'team'].includes(c.field))) {
    return changes.map(c => `${fieldLabel(c.field)}: ${updateValue(c.field, c.before)} → ${updateValue(c.field, c.after)}`).join(' · ');
  }
  const fields = changes.slice(0, 2).map(c => fieldLabel(c.field)).join(', ');
  return changes.length > 2 ? `${fields} + ${changes.length - 2} more` : fields;
}

export function playerTeam(update) {
  const team = update.team;
  const label = 'Team';
  if (!team || !/^[1-9][0-9]*$/.test(String(team.id))) return { label, name: 'Not recorded', href: null };
  return { label, name: team.name || `Team #${team.id}`, href: team.available === true ? `/team/${team.id}` : null };
}
