import { test } from 'node:test';
import assert from 'node:assert/strict';
import { parsePath, legacyDestination } from './routes.js';

test('routes use distinct paths and preserve large match IDs', () => {
  assert.deepEqual(parsePath('/'), { name: 'home' });
  assert.deepEqual(parsePath('/dpc'), { name: 'dpc' });
  assert.deepEqual(parsePath('/league/123'), { name: 'league', id: '123' });
  assert.deepEqual(parsePath('/team/45'), { name: 'team', id: '45' });
  assert.deepEqual(parsePath('/match/123/9007199254740993'), { name: 'match', leagueId: '123', matchId: '9007199254740993' });
});
test('invalid routes do not silently render the home page', () => {
  for (const path of ['/unknown', '/league/0', '/league/01', '/league/nope', '/dpc/extra', '/team/1/', '//example.com']) {
    assert.equal(parsePath(path).name, 'notFound', path);
  }
});
test('old hash links preserve tabs without allowing external redirects', () => {
  assert.equal(legacyDestination('#/league/123?tab=live'), '/league/123?tab=live');
  assert.equal(legacyDestination('#/'), '/');
  assert.equal(legacyDestination('#//example.com'), null);
  assert.equal(legacyDestination('#section'), null);
});

test('teams directory has its own route', () => {
  assert.deepEqual(parsePath('/team'), { name: 'teams' });
  assert.equal(legacyDestination('#/team'), '/team');
});

test('pagination normalizes malformed and negative offsets', async () => {
  const { pageOffset } = await import('./routes.js');
  assert.equal(pageOffset('?offset=200'), 200);
  for (const query of ['', '?offset=-100', '?offset=abc', '?offset=99', '?offset=Infinity', '?offset=10000001']) {
    assert.equal(pageOffset(query), 0);
  }
});
