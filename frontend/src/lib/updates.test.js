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
    assert.equal(view.href, entity === 'tournament' ? '/league/7' : entity === 'player' ? '/player/7' : '/team/7');
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
test('tournament status transitions receive lifecycle titles', () => {
  const status = (before, after, extra = []) => updatePresentation({entity:'tournament',action:'updated',changes:[{field:'status',before,after},...extra]});
  assert.equal(status(3, 5).title, 'Tournament concluded');
  assert.equal(status(3, 5).icon, 'flag');
  assert.equal(status(2, 3).title, 'Tournament approved');
  assert.equal(status(2, 4).title, 'Tournament rejected');
  assert.equal(status(3, 6).title, 'Tournament deleted');
  assert.equal(status(0, 2).title, 'Tournament status changed');
  assert.equal(status(3, 5, [{field:'end_timestamp'}]).title, 'Tournament concluded');
  assert.equal(updatePresentation({entity:'tournament',changes:[{field:'status',before:3,after:3},{field:'url'}]}).title, 'Tournament updated');
});
test('player changes describe team moves, privacy flips, and identity updates', () => {
  const title = changes => updatePresentation({entity:'player',action:'updated',changes}).title;
  assert.equal(title([{field:'team_id',before:0,after:7},{field:'team',before:'',after:'Team A'}]), 'Player joined team');
  assert.equal(title([{field:'team_id',before:7,after:0},{field:'team',before:'Team A',after:''}]), 'Player left team');
  assert.equal(title([{field:'team_id',before:7,after:8},{field:'team',before:'Team A',after:'Team B'}]), 'Player transferred');
  assert.equal(title([{field:'roster_team',before:'',after:'LGD'},{field:'roster_team_id',before:0,after:5}]), 'Joined roster');
  assert.equal(title([{field:'roster_team',before:'LGD',after:''},{field:'roster_team_id',before:5,after:0}]), 'Left roster');
  assert.equal(title([{field:'steam_profile',before:'public',after:'private'}]), 'Steam profile private');
  assert.equal(title([{field:'steam_profile',before:'private',after:'public'}]), 'Steam profile public');
  assert.equal(title([{field:'steam_profile',before:'public',after:'unavailable'}]), 'Steam profile unavailable');
  assert.equal(title([{field:'is_pro',before:false,after:true},{field:'name',before:'x',after:'y'}]), 'Pro profile added');
  assert.equal(title([{field:'name',before:'x',after:'y'}]), 'Player renamed');
  assert.equal(title([{field:'total_earnings',before:1,after:2}]), 'Earnings updated');
  assert.equal(title([{field:'steam_location',before:'a',after:'b'}]), 'Steam profile updated');
  assert.equal(title([{field:'sponsor',before:'a',after:'b'}]), 'Player updated');
  assert.equal(updateValue('steam_profile', 'friendsonly'), 'Friends only');
  assert.equal(updateValue('fantasy_role', 2), 'Support');
  assert.equal(updateValue('total_earnings', 1500), '$1,500');
  assert.equal(parsePath('/player/42').name, 'player');
  assert.equal(parsePath('/player').name, 'players');
  assert.equal(parsePath('/player/0').name, 'notFound');
});
test('steam state keeps last public data visible after a profile goes private', async () => {
  const { steamState, playerDisplayName } = await import('./players.js');
  assert.deepEqual(steamState({}), {kind:'unchecked',label:'Not checked yet',hasData:false});
  assert.deepEqual(steamState({steam_name:'n',steam_status:'ok',steam_privacy:'public'}), {kind:'public',label:'Public',hasData:true});
  assert.deepEqual(steamState({steam_name:'n',steam_status:'ok',steam_privacy:'private'}), {kind:'private',label:'Private',hasData:true});
  assert.deepEqual(steamState({steam_name:'n',steam_status:'unavailable'}), {kind:'unavailable',label:'Unavailable',hasData:true});
  assert.deepEqual(steamState({steam_status:'request_failed'}), {kind:'unknown',label:'Check failed',hasData:false});
  assert.equal(playerDisplayName({name:' ',steam_name:'Steam'}, 3), 'Steam');
  const { isListedPro, proRegistration, teamHistory } = await import('./players.js');
  assert.equal(isListedPro({profile_source:''}), true);
  assert.equal(isListedPro({profile_source:'steam'}), false);
  assert.equal(isListedPro(null), false);
  assert.deepEqual(proRegistration({pro_registration:[{registration_period:10,timestamp:1},{registration_period:11,timestamp:5}]}), {registration_period:11,timestamp:5});
  assert.equal(proRegistration({}), null);
  const { proRegistrations } = await import('./players.js');
  assert.deepEqual(proRegistrations({pro_registration:[{registration_period:10,timestamp:1},{registration_period:9,timestamp:0},{registration_period:11,timestamp:5}]}).map(r=>r.registration_period), [11,10]);
  assert.deepEqual(teamHistory({audit_entries:[{team_id:1,start_timestamp:1},{team_id:0,start_timestamp:9},{team_id:2,start_timestamp:3}]}).map(h=>h.team_id), [2,1]);
  assert.equal(playerDisplayName(null, 3), 'Player #3');
});
test('related field changes receive meaningful titles', () => {
  assert.equal(updatePresentation({entity:'team',changes:[{field:'name'},{field:'tag'}]}).title,'Team rebranded');
  assert.equal(updatePresentation({entity:'team',changes:[{field:'team_captain'}]}).title,'Captain changed');
  assert.equal(updatePresentation({entity:'tournament',changes:[{field:'start_timestamp'},{field:'end_timestamp'}]}).title,'Schedule changed');
});
