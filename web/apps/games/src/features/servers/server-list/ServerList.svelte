<!--
  Home of the games tool: every configured game server as a card (status,
  players), with who's playing and the latest joins and leaves beside
  them on wide screens (below on phones). Servers are configured in files
  (decision #37), so the empty state says where. When every server is
  unavailable for the same reason (Docker isn't connected), the reason is
  said once in a banner at the top, not on every card.
-->
<script lang="ts">
  import { Banner } from '@home-tools/ui/components/banner'
  import { EmptyState } from '@home-tools/ui/components/empty-state'
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { PageHeader } from '@home-tools/ui/components/page-header'
  import { Skeleton, SkeletonGroup } from '@home-tools/ui/components/skeleton'
  import { ServerCard } from '@/features/servers/server-card'
  import { liveServers } from '@/features/servers/live-servers'
  import { NowPlaying } from '@/features/servers/now-playing'

  const live = liveServers()
  let shown = $derived(live.servers.filter((s) => !s.archived))
  let running = $derived(shown.filter((s) => s.state?.running).length)
  let online = $derived(shown.reduce((n, s) => n + (s.players?.online ?? 0), 0))

  // The one reason every server is unavailable, if they share one.
  let sharedError = $derived.by(() => {
    const errors = new Set(shown.map((s) => (!s.state && !s.busy ? (s.error ?? '') : '')))
    const [only] = errors
    return errors.size === 1 && only ? only : null
  })

  let summary = $derived.by(() => {
    if (live.status !== 'ready' || shown.length === 0) return undefined
    if (sharedError) return `${shown.length} ${shown.length === 1 ? 'server' : 'servers'}` // "0 running" would be a guess
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
    <div class="layout">
      <div class="grid">
        {#each { length: 3 } as _, i (i)}
          <div class="ghost">
            <div class="ghost-cover"><Skeleton height="100%" radius="lg" /></div>
            <Skeleton width="70%" height="1rem" />
            <Skeleton width="50%" height="0.8rem" />
          </div>
        {/each}
      </div>
      <Skeleton height="14rem" radius="lg" />
    </div>
  </SkeletonGroup>
{:else if shown.length === 0}
  <EmptyState
    icon="M4 4h16v7H4z M4 13h16v7H4z M8 7.5h.01 M8 16.5h.01"
    title="No servers yet"
    description="Add one file per server in games/servers/ in the data folder (see deploy/README.md), then restart Home Tools."
  />
{:else}
  {#if sharedError}
    <div class="banner">
      <Banner tone="danger" title={sharedError}>
        Servers can't be checked, started or stopped until it's back. Setup steps are in deploy/README.md.
      </Banner>
    </div>
  {/if}
  <div class="layout">
    <div class="grid">
      {#each shown as server (server.id)}
        <ServerCard {server} onchange={live.update} showError={!sharedError} />
      {/each}
    </div>
    <NowPlaying servers={shown} />
  </div>
{/if}

<style>
  /* Phones: covers, then who's playing. Wide screens: side by side. */
  .layout {
    display: grid;
    gap: var(--space-7);
  }

  @media (min-width: 64rem) {
    .layout {
      grid-template-columns: minmax(0, 1fr) 20rem;
      align-items: start;
      gap: var(--space-6);
    }
  }

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

  .banner {
    margin-bottom: var(--space-6);
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
