import { test } from 'node:test';
import assert from 'node:assert/strict';
import { loadLeagueCatalog, filterLeagues } from './league-filters.js';

const rows = Array.from({length: 202}, (_, i) => ({
  league_id: i + 1, name: `League ${i + 1}`, tier: i < 100 ? 2 : 4, region: 3, is_live: i === 150
}));
const page = offset => ({results: rows.slice(offset, offset + 100), meta: {total: rows.length}});

test('filters find matches beyond the first loaded page', async () => {
  const requests = [];
  const catalog = await loadLeagueCatalog(async offset => { requests.push(offset); return page(offset); }, page(0));
  assert.deepEqual(requests, [100, 200]);
  assert.equal(filterLeagues(catalog, {tier: '4'}).length, 102);
  assert.equal(filterLeagues(catalog, {search: '  LEAGUE 151 ', tier: '4', region: '3', liveOnly: true})[0].league_id, 151);
  assert.equal(filterLeagues(catalog, {tier: '5'}).length, 0);
});

test('starting on another page still searches the catalogue from the beginning', async () => {
  const requests = [];
  const catalog = await loadLeagueCatalog(async offset => { requests.push(offset); return page(offset); });
  assert.deepEqual(requests, [0, 100, 200]);
  assert.equal(catalog[0].league_id, 1);
});

test('failed or incomplete pages cannot be mistaken for zero matches', async () => {
  await assert.rejects(loadLeagueCatalog(async () => { throw new Error('Offline'); }, page(0)), /Offline/);
  await assert.rejects(loadLeagueCatalog(async () => ({results: [], meta: {total: 202}}), page(0)), /list changed/);
});

test('a complete empty catalogue needs no additional requests', async () => {
  const catalog = await loadLeagueCatalog(() => { throw new Error('Unexpected fetch'); }, {results: [], meta: {total: 0}});
  assert.deepEqual(catalog, []);
});
