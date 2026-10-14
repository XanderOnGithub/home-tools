<!--
  A titled part of a page: <h2> + optional one-line description + actions
  on the right, then the content. Sections are separated by space, not
  boxes; put a <Card> inside when the content is a list or a group of
  controls. Pass `id` for in-page links (#section).
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  let {
    title,
    description,
    count,
    id,
    actions,
    children,
  }: {
    title: string
    description?: string
    count?: number // shown after the title, muted: "Names 5"
    id?: string
    actions?: Snippet
    children: Snippet
  } = $props()
</script>

<section class="section" {id} aria-labelledby={id ? `${id}-title` : undefined}>
  <div class="head">
    <div class="text">
      <h2 id={id ? `${id}-title` : undefined}>
        {title}{#if count !== undefined}<span class="count">{count}</span>{/if}
      </h2>
      {#if description}<p class="description">{description}</p>{/if}
    </div>
    {#if actions}<div class="actions">{@render actions()}</div>{/if}
  </div>
  {@render children()}
</section>

<style>
  .section {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    scroll-margin-top: calc(var(--header-height) + var(--space-4));
  }

  .section + :global(.section) {
    margin-top: var(--space-7);
  }

  .head {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    justify-content: space-between;
    gap: var(--space-2) var(--space-4);
  }

  .text {
    display: flex;
    flex: 1 1 14rem;
    flex-direction: column;
    gap: var(--space-1);
  }

  .count {
    margin-left: var(--space-2);
    color: var(--color-text-muted);
    font-weight: var(--weight-medium);
  }

  .description {
    margin: 0;
    max-width: var(--content-narrow);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .actions {
    display: flex;
    gap: var(--space-2);
  }
</style>
