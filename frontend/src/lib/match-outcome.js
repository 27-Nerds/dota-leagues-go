// Dota EMatchOutcome values (not team_number or kill totals).
// https://s2v.app/SchemaExplorer/dota2/client/EMatchOutcome
export function matchOutcome(value) {
  if (value === 2) return { winner: 'radiant', label: 'Radiant victory' };
  if (value === 3) return { winner: 'dire', label: 'Dire victory' };
  if (value === 5) return { winner: null, label: 'No winner' };
  if (Number.isInteger(value) && value >= 64 && value <= 69) return { winner: null, label: 'Not scored' };
  return { winner: null, label: 'Result unavailable' };
}
