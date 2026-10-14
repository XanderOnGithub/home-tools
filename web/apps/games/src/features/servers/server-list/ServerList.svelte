<!--
  Home of the games tool: every configured game server as a card (status,
  players). Servers are configured in files (decision #37), so the empty
  state says where.
-->
<script lang="ts">
  import { EmptyState } from '@home-tools/ui/components/empty-state'
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { PageHeader } from '@home-tools/ui/components/page-header'
  import { Skeleton, SkeletonGroup } from '@home-tools/ui/components/skeleton'
  import { ServerCard } from '@/features/servers/server-card'
  import { liveServers } from '@/features/servers/live-servers'

  const live = liveServers()
  let shown = $derived(live.servers.filter((s) => !s.archived))
  let running = $derived(shown.filter((s) => s.state?.running).length)
  let online = $derived(shown.reduce((n, s) => n + (s.players?.online ?? 0), 0))

  let summary = $derived.by(() => {
    if (live.status !== 'ready' || shown.length === 0) return undefined
    const parts = [`${running} of ${shown.length} running`]
    if (online) parts.push(`${online} ${online === 1 ? 'player' : 'players'} online`)
    return parts.join(' · ')
  })
</script>

<PageHeader title="Servers" description={summary} />

{#if live.status === 'error'}
  <LoadError what="the servers" onretry={live.load} />
{:else if live.status === 'loading'}
  <SkeletonGroup label="servers">
    <div class="grid">
      {#each { length: 3 } as _, i (i)}
        <div class="ghost">
          <Skeleton width="2.75rem" height="2.75rem" radius="md" />
          <div class="ghost-text">
            <Skeleton width="60%" height="1.1rem" />
            <Skeleton width="5rem" height="1.5rem" radius="sm" />
            <Skeleton width="80%" height="0.8rem" />
          </div>
        </div>
      {/each}
    </div>
  </SkeletonGroup>
{:else if shown.length === 0}
  <EmptyState
    icon="M4 4h16v7H4z M4 13h16v7H4z M8 7.5h.01 M8 16.5h.01"
    title="No servers yet"
    description="Add one file per server in games/servers/ in the data folder (see deploy/README.md), then restart Home Tools."
  />
{:else}
  <div class="grid">
    {#each shown as server (server.id)}
      <ServerCard {server} onchange={live.update} />
    {/each}
  </div>
{/if}

<style>
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(18rem, 1fr));
    gap: var(--space-3);
  }

  .ghost {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: var(--space-4);
    padding: var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
  }

  .ghost-text {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
</style>
