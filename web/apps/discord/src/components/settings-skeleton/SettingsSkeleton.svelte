<!--
  What a Discord page shows while the settings load: a card, then a few
  sections, in the shape of the real thing (decision #57).
-->
<script lang="ts">
  import { Skeleton, SkeletonGroup } from '@home-tools/ui/components/skeleton'

  let { sections = 2 }: { sections?: number } = $props()
</script>

<SkeletonGroup label="the bot's settings">
  <div class="ghost">
    <Skeleton height="8.5rem" radius="lg" />
    {#each { length: sections } as _, i (i)}
      <div class="section">
        <Skeleton width="9rem" height="1.25rem" />
        <Skeleton width="70%" height="0.875rem" />
        <Skeleton height="3.5rem" radius="lg" />
      </div>
    {/each}
  </div>
</SkeletonGroup>

<style>
  .ghost {
    display: flex;
    flex-direction: column;
    gap: var(--space-7);
  }

  .section {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }
</style>
