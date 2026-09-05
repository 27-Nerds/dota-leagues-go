export const SCHEDULE_PAGE_SIZE = 20;

export function groupSchedule(series) {
  const groups = new Map();
  for (const row of series) {
    const date = new Date(Number(row.start_time) * 1000);
    const valid = !!row.start_time && Number.isFinite(date.getTime());
    const key = valid ? date.toISOString().slice(0, 10) : 'unscheduled';
    if (!groups.has(key)) {
      groups.set(key, {
        key,
        label: valid ? new Intl.DateTimeFormat('en-GB', {
          weekday: 'long', day: 'numeric', month: 'short', year: 'numeric', timeZone: 'UTC'
        }).format(date) : 'Date to be announced',
        rows: []
      });
    }
    groups.get(key).rows.push({
      ...row,
      time: valid ? new Intl.DateTimeFormat('en-GB', {
        hour: '2-digit', minute: '2-digit', hourCycle: 'h23', timeZone: 'UTC'
      }).format(date) : 'TBA'
    });
  }
  return [...groups.values()];
}
