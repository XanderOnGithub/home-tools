<!--
  Home of the games tool: every configured game server with its status and
  actions. Servers are configured in files (decision #37), so "none yet"
  says where.
-->
<script lang="ts">
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { ServerCard } from '@/features/servers/server-card'
  import { liveServers } from '@/features/servers/live-servers'

  const live = liveServers()
  let shown = $derived(live.servers.filter((s) => !s.archived))
</script>

<div class="page">
  <h1 tabindex="-1">Game servers</h1>

  {#if live.status === 'error'}
    <LoadError what="the servers" onretry={live.load} />
  {:else if live.status === 'ready' && shown.length === 0}
    <p class="empty">
      No servers yet. Add one file per server in <code>games/servers/</code> in the data folder
      (see <code>deploy/README.md</code>), then restart Home Tools.
    </p>
  {:else}
    <div class="grid">
      {#each shown as server (server.id)}
        <ServerCard {server} onchange={live.update} />
      {/each}
    </div>
  {/if}
</div>

<style>
  .page {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  h1 {
    font-size: var(--text-3xl);
    font-weight: var(--weight-extrabold);
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(18rem, 1fr));
    gap: var(--space-4);
  }

  .empty {
    margin: 0;
    color: var(--color-text-muted);
  }

  code {
    font-family: var(--font-mono);
    font-size: 0.9em;
  }
</style>
