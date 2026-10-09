<!--
  A server's state in words: a dot plus text (color is never the only
  signal). Running = filled accent dot, stopped = hollow dot,
  unavailable = danger text with the reason; an action in flight
  ("Stopping…", from this page or anyone else's, e.g. the Discord bot)
  replaces it all. Uptime and players go on the line below
  (see `details`), so this one never wraps mid-phrase.
-->
<script lang="ts">
  import type { Action, Server } from '@/features/servers'

  let { server, pending = null }: { server: Server; pending?: Action | null } = $props()

  const DOING: Record<Action, string> = { start: 'Starting…', stop: 'Stopping…', restart: 'Restarting…' }
  let running = $derived(server.state?.running ?? false)
  let doing = $derived(pending ?? server.busy ?? null)
</script>

<p class="status" class:on={running && !doing} class:unknown={!server.state && !doing}>
  <span class="dot" aria-hidden="true"></span>
  {#if doing}
    {DOING[doing]}
  {:else if !server.state}
    Unavailable: {server.error ?? 'unknown error'}
  {:else if running}
    Running
  {:else}
    Stopped
  {/if}
</p>

<style>
  .status {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    margin: 0;
    font-weight: var(--weight-semibold);
  }

  .dot {
    flex: none;
    width: 0.7rem;
    height: 0.7rem;
    border: 2px solid var(--color-border-strong);
    border-radius: var(--radius-full);
  }

  .on .dot {
    border-color: var(--color-accent);
    background: var(--color-accent);
  }

  .unknown {
    color: var(--color-danger-text);
    font-weight: var(--weight-medium);
  }

  .unknown .dot {
    border-color: var(--color-danger-text);
  }
</style>
