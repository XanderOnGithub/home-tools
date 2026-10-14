<!--
  The top of every page (decision #57): an optional back link, the page's
  <h1> (focused on navigation by AppShell, so tabindex -1), a one-line
  description, and actions on the right (wrapping under on phones). One
  header, so every page in every tool starts the same way.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  let {
    title,
    description,
    back,
    eyebrow,
    meta,
    leading,
    actions,
    heading = $bindable(),
  }: {
    title: string
    description?: string
    back?: { href: string; label: string } // "Plans" → "‹ Plans"
    eyebrow?: Snippet // small line above the title
    meta?: Snippet // under the title/description: badges, a status
    leading?: Snippet // beside the title block, left: an icon
    actions?: Snippet
    heading?: HTMLHeadingElement // for pages that focus it themselves
  } = $props()
</script>

<header class="page-header">
  {#if back}
    <a class="back" href={back.href}>
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M15 18l-6-6 6-6" /></svg>
      {back.label}
    </a>
  {/if}
  <div class="row">
    {#if leading}<div class="leading">{@render leading()}</div>{/if}
    <div class="text">
      {#if eyebrow}<div class="eyebrow">{@render eyebrow()}</div>{/if}
      <h1 bind:this={heading} tabindex="-1">{title}</h1>
      {#if description}<p class="description">{description}</p>{/if}
      {#if meta}<div class="meta">{@render meta()}</div>{/if}
    </div>
    {#if actions}<div class="actions">{@render actions()}</div>{/if}
  </div>
</header>

<style>
  .page-header {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-bottom: var(--space-6);
  }

  .back {
    display: inline-flex;
    align-items: center;
    align-self: flex-start;
    gap: var(--space-1);
    min-height: 2rem;
    margin-left: calc(-1 * var(--space-1));
    padding-right: var(--space-2);
    border-radius: var(--radius-md);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
    text-decoration: none;
  }

  @media (hover: hover) {
    .back:hover {
      color: var(--color-text);
    }
  }

  .back svg {
    width: 1.125rem;
    height: 1.125rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    justify-content: space-between;
    gap: var(--space-3) var(--space-5);
  }

  .leading {
    flex: none;
    align-self: flex-start;
    margin-top: 0.125rem;
  }

  .text {
    display: flex;
    flex: 1 1 12rem;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0;
  }

  .eyebrow {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  h1 {
    font-size: var(--text-title);
    font-weight: var(--weight-bold);
    overflow-wrap: anywhere;
  }

  .description {
    margin: 0;
    max-width: var(--content-narrow);
    color: var(--color-text-muted);
  }

  .meta {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    margin-top: var(--space-2);
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
</style>
