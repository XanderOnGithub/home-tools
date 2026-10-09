<!--
  One server: header (icon, name, status, ⋯ menu), who's online, who
  joined and left, and the live log. After an action the log reconnects, since a stop or restart
  ends the stream.
-->
<script lang="ts">
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { GAMES, details, type Action, type Server } from '@/features/servers'
  import { ActivityList } from '@/features/servers/activity-list'
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
  let error = $state('')
  let logKey = $state(0)

  function changed(s: Server) {
    error = ''
    live.update(s)
    logKey++ // a fresh stream for the new run
  }
</script>

<div class="page">
  <a class="back" href="/">
    <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M15 6l-6 6 6 6" /></svg>
    Servers
  </a>

  {#if live.status === 'error'}
    <LoadError what="the server" onretry={live.load} />
  {:else if live.status === 'ready' && !server}
    <h1 tabindex="-1">Not found</h1>
    <p>There's no server called “{id}”. <a href="/">All servers</a></p>
  {:else if server}
    <header class="head">
      <GameIcon game={server.game} />
      <div class="title">
        <h1 tabindex="-1">{server.name}</h1>
        {#if GAMES[server.game] !== server.name}<p class="muted">{GAMES[server.game]}</p>{/if}
        <ServerStatus {server} {pending} />
        {#if !pending && details(server)}<p class="muted">{details(server)}</p>{/if}
      </div>
      <ServerMenu {server} bind:pending onchange={changed} onerror={(m) => (error = m)} />
    </header>
    {#if error}<p class="error" role="alert">{error}</p>{/if}

    <div class="sections">
      <PlayerList {server} />
      {#if server.state}<ActivityList {server} />{/if}
      {#if server.state}
        {#key logKey}
          <LogView {id} />
        {/key}
      {/if}
    </div>
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
    margin-left: calc(-1 * var(--space-1));
    color: var(--color-text-muted);
    font-weight: var(--weight-semibold);
    text-decoration: none;
  }

  @media (hover: hover) {
    .back:hover {
      color: var(--color-text);
    }
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

  .head {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: start;
    gap: var(--space-4);
  }

  .title {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0;
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
    font-weight: var(--weight-medium);
  }

  .error {
    color: var(--color-danger-text);
  }

  .sections {
    display: flex;
    flex-direction: column;
    gap: var(--space-7);
  }
</style>
