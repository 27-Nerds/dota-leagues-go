import { test } from 'node:test';
import assert from 'node:assert/strict';
import { formatDate, formatDateTime, formatMoney, steamProfileUrl, normalizeUrl } from './constants.js';

test('Steam IDs preserve every digit and reject invalid accounts', () => {
  assert.equal(steamProfileUrl(262476000), 'https://steamcommunity.com/profiles/76561198222741728/');
  assert.equal(steamProfileUrl('4294967295'), 'https://steamcommunity.com/profiles/76561202255233023/');
  for (const id of [0, null, undefined, -1, 'oops', '4294967296']) assert.equal(steamProfileUrl(id), null);
});

test('external links only allow valid HTTP(S) destinations', () => {
  assert.equal(normalizeUrl('example.com'), 'https://example.com/');
  assert.equal(normalizeUrl('https://example.com/x'), 'https://example.com/x');
  for (const value of ['javascript:alert(1)', 'data:text/html,hello', 'mailto:x@example.com', '-', null]) assert.equal(normalizeUrl(value), null);
});

test('missing and invalid values never render Invalid Date or NaN', () => {
  for (const value of [undefined, null, 0, 'bad', Infinity]) {
    assert.equal(formatDate(value), '—');
    assert.equal(formatDateTime(value), '—');
  }
  assert.equal(formatMoney('bad'), '—');
  assert.equal(formatMoney(12000), '$12,000');
  assert.equal(formatDate(86400), '02/01/1970');
});
