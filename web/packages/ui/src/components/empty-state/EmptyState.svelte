<!--
  What a list or page shows before there's anything in it: an icon, a
  short title, one line on what goes here, and the action that fills it.
  Centered in a dashed outline so it reads as "space for something", not
  as content.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  let {
    title,
    description,
    icon,
    size = 'md',
    action,
  }: {
    title: string
    description?: string
    icon?: string // SVG path, 24×24 box, stroked
    size?: 'sm' | 'md'
    action?: Snippet
  } = $props()
</script>

<div class="empty {size}">
  {#if icon}
    <span class="icon" aria-hidden="true"><svg viewBox="0 0 24 24"><path d={icon} /></svg></span>
  {/if}
  <p class="title">{title}</p>
  {#if description}<p class="description">{description}</p>{/if}
  {#if action}<div class="action">{@render action()}</div>{/if}
</div>

<style>
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-7) var(--space-5);
    border: 1px dashed var(--color-border-strong);
    border-radius: var(--radius-lg);
    text-align: center;
  }

  .sm {
    padding: var(--space-5) var(--space-4);
  }

  .icon {
    display: grid;
    place-items: center;
    width: 2.5rem;
    height: 2.5rem;
    margin-bottom: var(--space-2);
    border-radius: var(--radius-md);
    background: var(--color-surface-sunken);
    color: var(--color-text-muted);
  }

  .icon svg {
    width: 1.25rem;
    height: 1.25rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  p {
    margin: 0;
  }

  .title {
    font-weight: var(--weight-semibold);
  }

  .description {
    max-width: 28rem;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .action {
    margin-top: var(--space-3);
  }
</style>
