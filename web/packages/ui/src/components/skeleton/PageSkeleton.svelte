<!--
  A generic loading layout for a page whose real shape is a list of rows
  ("list"), cards in a grid ("grid") or a form ("form"). Pages with a more
  specific shape draw their own from <Skeleton>s.
-->
<script lang="ts">
  import Skeleton from './Skeleton.svelte'
  import SkeletonGroup from './SkeletonGroup.svelte'

  let { label, variant = 'list', rows = 4 }: { label: string; variant?: 'list' | 'grid' | 'form'; rows?: number } =
    $props()
</script>

<SkeletonGroup {label}>
  {#if variant === 'grid'}
    <div class="grid">
      {#each { length: rows } as _, i (i)}<Skeleton height="7rem" radius="lg" />{/each}
    </div>
  {:else if variant === 'form'}
    <div class="form">
      {#each { length: rows } as _, i (i)}
        <div class="field"><Skeleton width="6rem" height="0.875rem" /><Skeleton height="2.75rem" /></div>
      {/each}
    </div>
  {:else}
    <div class="list">
      {#each { length: rows } as _, i (i)}
        <div class="row">
          <div class="text">
            <Skeleton width="{40 + ((i * 17) % 30)}%" height="1rem" />
            <Skeleton width="{25 + ((i * 11) % 20)}%" height="0.8rem" />
          </div>
        </div>
      {/each}
    </div>
  {/if}
</SkeletonGroup>

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(16rem, 1fr));
    gap: var(--space-3);
  }

  .form {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    max-width: var(--content-narrow);
  }

  .field,
  .text {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .list {
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
  }

  .row {
    padding: var(--space-4);
  }

  .row + .row {
    border-top: 1px solid var(--color-border);
  }
</style>
