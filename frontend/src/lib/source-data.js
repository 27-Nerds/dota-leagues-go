export function auditRows(payload) {
  return (Array.isArray(payload?.audit_entries) ? payload.audit_entries : []).filter(row => row && Number.isInteger(row.audit_action) && row.account_id > 0 && row.timestamp > 0).sort((a,b) => b.timestamp - a.timestamp);
}
export function groupRows(payload) {
  const rows = new Map();
  function visit(groups) {
    if (!Array.isArray(groups)) return;
    for (const group of groups) {
      if (!group || typeof group !== 'object') continue;
      if (group.node_group_id > 0) rows.set(group.node_group_id, group);
      visit(group.node_groups);
    }
  }
  visit(payload?.node_groups);
  return [...rows.values()];
}
// Show only explicit source links whose two endpoints are actually available.
export function groupLinks(groups) {
  const known = new Set(groups.map(g => g.node_group_id));
  return groups.flatMap(g => ['advancing_node_group_id','secondary_advancing_node_group_id','tertiary_advancing_node_group_id'].flatMap((field, index) => {
    const target = g[field];
    return target > 0 && target !== g.node_group_id && known.has(target) ? [{from:g.node_group_id,to:target,route:['Primary','Secondary','Tertiary'][index]}] : [];
  }));
}
