import test from 'node:test';
import assert from 'node:assert/strict';
import { takeDirectoryData } from './directory.js';

test('directory bootstrap preserves filtered results and is consumed once', () => {
  let node = { textContent: JSON.stringify({path:'/team',search:'country=UA',results:[{team_id:2}],meta:{total:101}}), remove() { node = null; } };
  const doc = {getElementById: () => node};
  const location = {search:'?country=UA'};
  assert.equal(takeDirectoryData('/team', doc, location).meta.total, 101);
  assert.equal(takeDirectoryData('/team', doc, location), null);
});

test('invalid or mismatched bootstrap falls back to fetching', () => {
  for (const textContent of ['bad json', JSON.stringify({path:'/player',search:'',results:[],meta:{total:0}}), JSON.stringify({path:'/team',search:'country=UA',results:[],meta:{total:0}})]) {
    assert.equal(takeDirectoryData('/team', {getElementById: () => ({textContent, remove() {}})}, {search:''}), null);
  }
});
