import assert from 'node:assert/strict';
import { matchOutcome } from './match-outcome.js';

assert.equal(matchOutcome(2).winner, 'radiant');
assert.equal(matchOutcome(3).winner, 'dire');
assert.equal(matchOutcome(5).label, 'No winner');
for (const code of [64, 65, 66, 67, 68, 69]) {
  assert.deepEqual(matchOutcome(code), { winner: null, label: 'Not scored' });
}
for (const code of [undefined, null, 0, 1, 4, 6, 13, 70, '2']) {
  assert.deepEqual(matchOutcome(code), { winner: null, label: 'Result unavailable' });
}
