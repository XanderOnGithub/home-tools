<!--
  A titled panel. A <section> labelled by its own heading, so screen
  reader users can jump between cards by heading or by region.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  let { title, children }: { title: string; children: Snippet } = $props()

  const id = $props.id() // unique per instance, stable across renders
</script>

<section class="card" aria-labelledby="{id}-title">
  <h2 id="{id}-title">{title}</h2>
  {@render children()}
</section>

<style>
  .card {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-5);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
  }

  h2 {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }
</style>
