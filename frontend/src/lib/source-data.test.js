import {test} from 'node:test';
import assert from 'node:assert/strict';
import {auditRows,groupRows,groupLinks} from './source-data.js';
test('only usable audit rows are displayed without decoding action codes',()=>{
 assert.deepEqual(auditRows(null),[]);
 assert.equal(auditRows({audit_entries:[null,{}, {audit_action:5,account_id:123,timestamp:100}]}).length,1);
});
test('nested groups and verified endpoints support explicit progression only',()=>{
 const groups=groupRows({node_groups:[{node_group_id:1,advancing_node_group_id:2,secondary_advancing_node_group_id:99,node_groups:[{node_group_id:2}]}]});
 assert.equal(groups.length,2);
 assert.deepEqual(groupLinks(groups),[{from:1,to:2,route:'Primary'}]);
 assert.deepEqual(groupRows({node_groups:null}),[]);
});
