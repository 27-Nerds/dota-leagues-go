import test from 'node:test';
import assert from 'node:assert/strict';
import { dotaTVUrl } from './dotatv.js';

test('DotaTV link preserves the exact server Steam ID', () => {
  assert.equal(dotaTVUrl('90283335803849746'), 'steam://run/570//%2Bwatch_server%2090283335803849746/');
});

test('missing, imprecise, or invalid IDs cannot become launch commands', () => {
  for (const id of [null, undefined, '', '0', 90283335803849746, '12;quit', '12/quit', ' 123', '18446744073709551616']) {
    assert.equal(dotaTVUrl(id), null);
  }
});
