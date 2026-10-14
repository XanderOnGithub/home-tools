<!--
  One server, laid out like a dashboard: the header (icon, name, status,
  ⋯ menu), then the live log and console as the main column, with who's
  online and who joined and left beside it (stacked on phones, players
  first). After an action the log reconnects, since a stop or restart
  ends the stream.
-->
<script lang="ts">
  import { Card } from '@home-tools/ui/components/card'
  import { EmptyState } from '@home-tools/ui/components/empty-state'
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { PageHeader } from '@home-tools/ui/components/page-header'
  import { Skeleton, SkeletonGroup } from '@home-tools/ui/components/skeleton'
  import { GAMES, details, type Action, type Server } from '@/features/servers'
  import { ActivityList } from '@/features/servers/activity-list'
  import { ConsoleView } from '@/features/servers/console-view'
  import { GameIcon } from '@/features/servers/game-icon'
  import { liveServers } from '@/features/servers/live-servers'
  import { LogView } from '@/features/servers/log-view'
  import { PlayerList } from '@/features/servers/player-list'
  import { ServerMenu } from '@/features/servers/server-menu'
  import { ServerStatus } from '@/features/servers/server-status'

  let { id }: { id: string } = $props()

  const live = liveServers()
  let server = $derived(live.servers.find((s) => s.id === id))
  let pending = $state<Action | null>(null)
  let logKey = $state(0)

  let subtitle = $derived.by(() => {
    if (!server) return undefined
    const parts: string[] = []
    if (GAMES[server.game] !== server.name) parts.push(GAMES[server.game])
    if (!pending && details(server)) parts.push(details(server))
    return parts.join(' · ') || undefined
  })

  function changed(s: Server) {
    live.update(s)
    logKey++ // a fresh stream for the new run
  }

  const BACK = { href: '/', label: 'Servers' }
</script>

{#if live.status === 'error'}
  <PageHeader title="Server" back={BACK} />
  <LoadError what="the server" onretry={live.load} />
{:else if live.status === 'loading'}
  <SkeletonGroup label="the server">
    <div class="ghost-head">
      <Skeleton width="4rem" height="0.9rem" />
      <Skeleton width="14rem" height="1.75rem" />
      <Skeleton width="10rem" height="0.9rem" />
    </div>
    <div class="layout">
      <div class="side"><Skeleton height="10rem" radius="lg" /></div>
      <div class="main"><Skeleton height="24rem" radius="lg" /></div>
    </div>
  </SkeletonGroup>
{:else if !server}
  <PageHeader title="Server not found" back={BACK} />
  <EmptyState title="No server called “{id}”" description="It may have been renamed or removed from the data folder.">
    {#snippet action()}<a class="btn btn-secondary" href="/">All servers</a>{/snippet}
  </EmptyState>
{:else}
  <PageHeader title={server.name} description={subtitle} back={BACK}>
    {#snippet leading()}<GameIcon game={server!.game} />{/snippet}
    {#snippet meta()}<ServerStatus server={server!} {pending} />{/snippet}
    {#snippet actions()}<ServerMenu server={server!} bind:pending onchange={changed} />{/snippet}
  </PageHeader>

  <div class="layout">
    <div class="side">
      <Card><PlayerList {server} /></Card>
      {#if server.state}<Card><ActivityList {server} /></Card>{/if}
    </div>
    <div class="main">
      {#if server.state}
        <Card padding="none">
          {#key logKey}
            <LogView {id} />
          {/key}
          {#if server.console && server.state.running}<ConsoleView {id} />{/if}
        </Card>
      {/if}
    </div>
  </div>
{/if}

<style>
  .ghost-head {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-bottom: var(--space-6);
  }

  .layout {
    display: grid;
    gap: var(--space-4);
    align-items: start;
  }

  .side,
  .main {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    min-width: 0;
  }

  @media (min-width: 56rem) {
    .layout {
      grid-template-columns: 20rem 1fr;
    }
  }
</style>
