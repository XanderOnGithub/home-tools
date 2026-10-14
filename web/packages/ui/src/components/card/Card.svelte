<!--
  A bordered surface for a group of things: a list, a set of controls, a
  chart. One look everywhere (border, 12px corners, no shadow), so cards
  never compete. `padding="none"` for lists whose rows bring their own
  padding (rows separated by dividers: use <ul class="card-list"> inside).
  `as` picks the element (an <a> card becomes one big link with a hover).
-->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import type { HTMLAttributes } from 'svelte/elements'

  let {
    padding = 'md',
    as = 'div',
    href,
    children,
    ...rest
  }: {
    padding?: 'none' | 'sm' | 'md' | 'lg'
    as?: 'div' | 'section' | 'article' | 'a' | 'li'
    href?: string
    children: Snippet
  } & HTMLAttributes<HTMLElement> = $props()
</script>

<svelte:element this={as} class="card pad-{padding}" class:link={as === 'a'} {href} {...rest}>
  {@render children()}
</svelte:element>

<style>
  .card {
    display: block;
    min-width: 0;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
    color: inherit;
  }

  .pad-none {
    overflow: hidden; /* rows' hover fills keep the corners */
  }
  .pad-sm {
    padding: var(--space-3);
  }
  .pad-md {
    padding: var(--space-4);
  }
  .pad-lg {
    padding: var(--space-5);
  }

  .link {
    text-decoration: none;
    transition:
      border-color var(--duration-fast) var(--ease-out),
      box-shadow var(--duration-fast) var(--ease-out);
  }

  @media (hover: hover) {
    .link:hover {
      border-color: var(--color-border-strong);
      box-shadow: var(--shadow-sm);
    }
  }

  .link:active {
    background: var(--color-surface-sunken);
  }

  /* Rows inside a padding="none" card. */
  .card :global(.card-list) {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .card :global(.card-list > li + li) {
    border-top: 1px solid var(--color-border);
  }
</style>
