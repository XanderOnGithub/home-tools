<!--
  One server: its card (status + actions) and its live log. After an
  action the log reconnects, since a stop or restart ends the stream.
-->
<script lang="ts">
  import { LoadError } from '@home-tools/ui/components/load-error'
  import type { Server } from '@/features/servers'
  import { LogView } from '@/features/servers/log-view'
  import { ServerCard } from '@/features/servers/server-card'
  import { liveServers } from '@/features/servers/live-servers'

  let { id }: { id: string } = $props()

  const live = liveServers()
  let server = $derived(live.servers.find((s) => s.id === id))
  let logKey = $state(0)

  function changed(s: Server) {
    live.update(s)
    logKey++ // a fresh stream for the new run
  }
</script>

<div class="page">
  <a class="back" href="/">
    <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M15 6l-6 6 6 6" /></svg>
    All servers
  </a>

  {#if live.status === 'error'}
    <LoadError what="the server" onretry={live.load} />
  {:else if live.status === 'ready' && !server}
    <p>There's no server called “{id}”. <a href="/">All servers</a></p>
  {:else if server}
    <ServerCard {server} onchange={changed} logsLink={false} />
    {#if server.state}
      {#key logKey}
        <LogView {id} />
      {/key}
    {/if}
  {/if}
</div>

<style>
  .page {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  .back {
    display: inline-flex;
    align-items: center;
    align-self: flex-start;
    gap: var(--space-1);
    min-height: var(--touch-target);
    color: var(--color-text-muted);
    font-weight: var(--weight-semibold);
    text-decoration: none;
  }

  .back svg {
    width: 1.25rem;
    height: 1.25rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
</style>
