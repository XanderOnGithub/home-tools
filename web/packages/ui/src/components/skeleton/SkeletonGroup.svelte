<!--
  Wraps a loading skeleton: tells screen readers "Loading <label>" once
  (the bones themselves are aria-hidden), and waits a moment before
  showing anything, so fast loads never flash gray (delay, default 150 ms).
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  let { label, delay = 150, children }: { label: string; delay?: number; children: Snippet } = $props()

  let shown = $state(false)

  $effect(() => {
    const t = setTimeout(() => (shown = true), delay)
    return () => clearTimeout(t)
  })
</script>

<div class="skeleton-group" role="status" aria-label="Loading {label}" aria-busy="true">
  {#if shown}{@render children()}{/if}
</div>

<style>
  .skeleton-group {
    display: contents;
  }
</style>
