<!--
  "Couldn't load …" with a Try again button: what every screen shows when
  its first fetch fails (usually the server is down or restarting).
  Extra actions (e.g. "Switch profile") go in as children. In an app, its
  mascot sits above, leaning over (decision #63).
-->
<script lang="ts">
  import type { Snippet } from 'svelte'
  import { ToolMascot } from '../../tool-mascot'
  import { appTool } from '../../tools'

  let {
    what,
    onretry,
    align = 'start',
    children,
  }: {
    what: string // completes "Couldn't load …", e.g. "your week"
    onretry: () => void
    align?: 'start' | 'center'
    children?: Snippet
  } = $props()

  const tool = appTool()
</script>

<div class="load-error" class:center={align === 'center'} role="alert">
  {#if tool}<span class="mascot"><ToolMascot {tool} /></span>{/if}
  <p>Couldn't load {what}. Check that the server is running.</p>
  <div class="actions">
    <button type="button" class="btn btn-primary" onclick={onretry}>Try again</button>
    {@render children?.()}
  </div>
</div>

<style>
  .load-error {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-3);
  }

  .center {
    align-items: center;
    text-align: center;
  }

  p {
    margin: 0;
  }

  /* The app's mascot, tipped over: something went wrong, nobody's fault. */
  .mascot {
    width: 3.5rem;
    rotate: -14deg;
    opacity: 0.85;
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
</style>
