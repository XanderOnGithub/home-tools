<script lang="ts">
  import { EmptyState } from '@home-tools/ui/components/empty-state'
  import { PageHeader } from '@home-tools/ui/components/page-header'
</script>

<PageHeader title="Page not found" />
<EmptyState title="There's nothing at this address" description="The link may be old, or the page was moved.">
  {#snippet action()}<a class="btn btn-secondary" href="/">Go to Summary</a>{/snippet}
</EmptyState>
