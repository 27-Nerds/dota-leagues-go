import { test } from 'node:test';
import assert from 'node:assert/strict';
import { api, apiResults } from './api.js';

test('errors retain HTTP status for missing, rate-limited and failed states', async t => {
  for (const status of [404, 429, 503]) {
    t.mock.method(globalThis, 'fetch', async () => new Response(JSON.stringify({message:'Unavailable'}),{status}));
    await assert.rejects(api('/test'), error => error.status === status);
    t.mock.restoreAll();
  }
});

test('successful empty lists are data, not loading or failures', async t => {
  t.mock.method(globalThis, 'fetch', async () => new Response(JSON.stringify({results:[]})));
  assert.deepEqual(await apiResults('/test'), {results:[],meta:{total:0}});
});

test('malformed responses fail with a recoverable error', async t => {
  t.mock.method(globalThis, 'fetch', async () => new Response('<html>not JSON</html>'));
  await assert.rejects(apiResults('/test'), /unexpected response shape/);
});
