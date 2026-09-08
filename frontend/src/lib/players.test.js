import test from 'node:test';
import assert from 'node:assert/strict';
import { teamMoves } from './players.js';

test('a listed join date and a later observed change retain their different meanings', () => {
  const joined = Date.UTC(2026, 5, 9);
  const observed = Date.UTC(2026, 8, 7);
  const rows = teamMoves({ audit_entries: [{ team_id: 9, team_name: 'Rune Eaters', start_timestamp: joined / 1000 }] }, [
    { created_at: observed, changes: [{ field: 'team_id', before: 1, after: 9 }, { field: 'team', before: 'nouns', after: 'Rune Eaters' }] }
  ]);
  assert.deepEqual(rows.map(r => [r.at, r.kind, r.fromId, r.toId]), [
    [observed, 'observed', 1, 9],
    [joined, 'listed_join', 0, 9]
  ]);
});

test('missing join dates do not create an invented historical event', () => {
  assert.deepEqual(teamMoves({ audit_entries: [{ team_id: 9 }] }, []), []);
  const rows = teamMoves({}, [{ created_at: 123, changes: [{ field: 'roster_team_id', before: 9, after: 0 }] }]);
  assert.equal(rows[0].kind, 'observed');
  assert.equal(rows[0].source, 'Team roster');
});
