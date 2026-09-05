import { test } from 'node:test';
import assert from 'node:assert/strict';
import { rosterChanges, updateValue, updatePresentation, groupUpdates, updateDate } from './updates.js';
import { parsePath, pageOffset } from './routes.js';

test('roster maps distinguish membership from active status without order changes', () => {
  assert.deepEqual(rosterChanges({'1':true,'2':false}, {'2':false,'1':true}), []);
  assert.deepEqual(rosterChanges({'1':true,'2':false,'3':true}, {'2':true,'3':false,'4':false}).map(r=>[r.id,r.kind]), [['1','left'],['2','activated'],['3','deactivated'],['4','joined']]);
  assert.equal(rosterChanges(null, {'1':false})[0].status, 'Inactive');
  assert.equal(rosterChanges({'1':true}, null)[0].kind, 'left');
});
test('field values preserve zero and use domain labels and UTC dates', () => {
  assert.equal(updateValue('total_prize_pool', 0), '$0');
  assert.equal(updateValue('total_prize_pool', 123456), '$123,456');
  assert.equal(updateValue('wins', 0), '0');
  assert.equal(updateValue('status', 3), 'Accepted');
  assert.equal(updateValue('status', 5), 'Concluded');
  assert.equal(updateValue('region', 3), 'Europe');
  assert.equal(updateValue('country_code', 'ua'), 'Ukraine');
  assert.match(updateValue('tier', 5), /The International/);
  assert.match(updateValue('start_timestamp', 1788609600), /5 Sept 2026/);
  assert.equal(updateValue('team_captain', 123), 'Account #123');
  assert.equal(updateValue('tag', null), 'Not set');
  assert.equal(updateValue('new_field', false), 'No');
});
test('every entity has a creation presentation and only existing internal routes are linked', () => {
  for (const entity of ['tournament','team','roster','player']) {
    const view = updatePresentation({entity,entity_id:7,action:'created',changes:[],url:'javascript:alert(1)'});
    assert.match(view.title, /added$/);
    assert.equal(view.changes.length, 0);
    assert.equal(view.href, entity === 'tournament' ? '/league/7' : entity === 'player' ? null : '/team/7');
    assert.ok(view.note);
  }
  assert.equal(updatePresentation({entity:'team',entity_id:'//evil.test'}).href, null);
  assert.equal(updatePresentation({entity:'future',entity_id:7}).title, 'Record updated');
  assert.equal(updatePresentation({entity:'team',changes:[{field:'wins'},{field:'losses'}]}).title, 'Team record updated');
  assert.equal(updatePresentation({entity:'tournament',changes:[{field:'total_prize_pool'}]}).title, 'Prize pool updated');
});
test('millisecond timestamps group chronologically without reordering equal-time entries', () => {
  const rows = [{id:'a',created_at:1788609600000},{id:'b',created_at:1788609600000},{id:'c',created_at:1788523200000}];
  assert.deepEqual(groupUpdates(rows).map(g=>g.rows.map(r=>r.id)), [['a','b'],['c']]);
  assert.equal(updateDate(rows[0].created_at).key, '2026-09-05');
  assert.equal(updateDate('invalid').key, 'unknown');
  assert.deepEqual(groupUpdates([]), []);
  assert.equal(parsePath('/activity').name, 'updates');
  assert.equal(parsePath('/updates').name, 'notFound');
  assert.equal(pageOffset('?offset=20',20),20);
});

test('compact summaries preserve the meaning of changes without long descriptions', async () => {
  const { updateSummary } = await import('./updates.js');
  assert.equal(updateSummary({changes:[{field:'members',before:{1:true,2:true},after:{2:false,3:true}}]}), '1 removed · 1 deactivated · 1 added');
  assert.equal(updateSummary({changes:[{field:'total_prize_pool',before:1000,after:2000}]}), 'Prize pool: $1,000 → $2,000');
  assert.equal(updateSummary({changes:[{field:'description',before:'x'.repeat(1000),after:'y'.repeat(1000)},{field:'url'},{field:'name'}]}), 'Description, Website + 1 more');
});

test('player entries link to the recorded team with safe missing-data fallbacks', async () => {
  const { playerTeam } = await import('./updates.js');
  assert.deepEqual(playerTeam({team:{id:7,name:'Team A',available:true}}), {label:'Team',name:'Team A',href:'/team/7'});
  assert.deepEqual(playerTeam({team:{id:7,name:'Team A',available:false}}), {label:'Team',name:'Team A',href:null});
  assert.equal(playerTeam({team:{id:7,name:'Team A'}}).href, null);
  assert.equal(playerTeam({team:{id:7,name:''}}).name, 'Team #7');
  assert.equal(playerTeam({team:{id:0,name:''}}).href, null);
  assert.equal(playerTeam({team:{id:'//example.com'}}).href, null);
});

test('roster summaries infer replacements, rebuilds and possible disbands conservatively', () => {
  const roster = (before, after, extra = {}) => updatePresentation({entity:'roster',action:'updated',changes:[{field:'members',before,after}],...extra}).title;
  assert.equal(roster({1:true,2:true},{2:true,3:true}), 'Player replaced');
  assert.equal(roster({1:true,2:true,3:true},{3:true,4:true,5:true}), 'Roster rebuilt');
  assert.equal(roster({1:true,2:true,3:true,4:true,5:true},{5:true}), 'Possible disband');
  assert.equal(roster({1:true,2:true,3:true,4:true,5:true,6:true},{5:true,6:true}), 'Players left');
  assert.equal(roster({1:true,2:true,3:true,4:true,5:true,6:true},{5:true,6:true},{roster_admins:{6:true}}), 'Possible disband');
  assert.equal(roster({1:true},{1:true,2:true,3:true,4:true,5:true}), 'Roster reformed');
  assert.equal(roster({1:false,2:false,3:false,4:false},{}), 'Roster updated');
  assert.equal(roster({1:true},{1:true,2:false}), 'Roster updated');
  assert.equal(roster({1:true,2:true,3:true,4:true,5:true},{6:true}), 'Roster updated');
  assert.equal(roster({}, {1:true,2:true,3:true,4:true,5:true},{action:'created'}), 'Roster added');
  assert.equal(roster({1:true},{1:false,2:true}), 'Player replaced');
});
test('related field changes receive meaningful titles', () => {
  assert.equal(updatePresentation({entity:'team',changes:[{field:'name'},{field:'tag'}]}).title,'Team rebranded');
  assert.equal(updatePresentation({entity:'team',changes:[{field:'team_captain'}]}).title,'Captain changed');
  assert.equal(updatePresentation({entity:'tournament',changes:[{field:'start_timestamp'},{field:'end_timestamp'}]}).title,'Schedule changed');
});
