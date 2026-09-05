<script>
  import Pagination from './Pagination.svelte';
  import { SCHEDULE_PAGE_SIZE } from './lib/schedule.js';
  let { id, offset, total } = $props();
</script>
<Pagination {offset} {total} pageSize={SCHEDULE_PAGE_SIZE} previousHref={`/league/${id}?offset=${offset - SCHEDULE_PAGE_SIZE}#schedule`} nextHref={`/league/${id}?offset=${offset + SCHEDULE_PAGE_SIZE}#schedule`} label="Schedule pages" previousLabel="Newer" nextLabel="Older" />
