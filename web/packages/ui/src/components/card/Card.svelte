<!--
  A bordered surface for a group of things: a list, a set of controls, a
  chart. One look everywhere (border, 12px corners, a thin solid edge
  underneath: decision #63), so cards never compete. A link card stands on
  a thicker edge and sinks into it when pressed, like the tool cards on
  the Home Tools page. `padding="none"` for lists whose rows bring their own
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

<svelte:element this={as} class="card pad-{padding}" class:link={as === 'a'} class:pressable={as === 'a'} {href} {...rest}>
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

  .card:not(.link) {
    box-shadow: var(--edge-card); /* link cards: .pressable in app.css */
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
