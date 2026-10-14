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
          <div class="ghost-cover"><Skeleton height="100%" radius="lg" /></div>
          <Skeleton width="70%" height="1rem" />
          <Skeleton width="50%" height="0.8rem" />
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
  /* Covers in rows, like a browse page: 2 across on phones, as many as
   * fit (at least 11rem each) on wider screens. */
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(min(11rem, calc(50% - var(--space-2))), 1fr));
    gap: var(--space-6) var(--space-4);
  }

  @media (min-width: 40rem) {
    .grid {
      grid-template-columns: repeat(auto-fill, minmax(12.5rem, 1fr));
    }
  }

  .ghost {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .ghost-cover {
    aspect-ratio: 3 / 4;
  }
</style>
