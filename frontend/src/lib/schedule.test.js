import { test } from 'node:test';
import assert from 'node:assert/strict';
import { groupSchedule, SCHEDULE_PAGE_SIZE } from './schedule.js';
import { pageOffset } from './routes.js';

test('groups series by UTC day and preserves their ordering and game links', () => {
  const rows = [
    {series_id: 1, start_time: Date.parse('2026-09-05T01:30:00Z') / 1000, match_ids: ['100']},
    {series_id: 2, start_time: Date.parse('2026-09-05T00:05:00Z') / 1000},
    {series_id: 3, start_time: Date.parse('2026-09-04T23:59:00Z') / 1000}
  ];
  const groups = groupSchedule(rows);
  assert.deepEqual(groups.map(g => g.key), ['2026-09-05', '2026-09-04']);
  assert.deepEqual(groups[0].rows.map(r => r.series_id), [1, 2]);
  assert.equal(groups[0].rows[1].time, '00:05');
  assert.deepEqual(groups[0].rows[0].match_ids, ['100']);
  assert.equal(rows[0].time, undefined);
});

test('missing and invalid timestamps get a TBA group', () => {
  const groups = groupSchedule([{start_time: 0}, {}, {start_time: 'invalid'}]);
  assert.equal(groups.length, 1);
  assert.equal(groups[0].key, 'unscheduled');
  assert.ok(groups[0].rows.every(r => r.time === 'TBA'));
  assert.deepEqual(groupSchedule([]), []);
});

test('schedule pagination uses 20 rows and supports existing offset links', () => {
  assert.equal(SCHEDULE_PAGE_SIZE, 20);
  assert.equal(pageOffset('?offset=20', SCHEDULE_PAGE_SIZE), 20);
  assert.equal(pageOffset('?offset=100', SCHEDULE_PAGE_SIZE), 100);
  assert.equal(pageOffset('?offset=19', SCHEDULE_PAGE_SIZE), 0);
  assert.equal(pageOffset('?offset=20'), 0);
});
