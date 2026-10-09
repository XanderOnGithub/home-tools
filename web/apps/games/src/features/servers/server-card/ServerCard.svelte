<!--
  One game server in the list: icon, name, status, players. The whole card
  opens the server's page (a "stretched" link on the name, so it's one real
  <a> for screen readers and right-click); the ⋯ menu sits on top of it,
  top-right, for Start / Stop / Restart.
-->
<script lang="ts">
  import { GAMES, details, type Action, type Server } from '@/features/servers'
  import { GameIcon } from '@/features/servers/game-icon'
  import { ServerMenu } from '@/features/servers/server-menu'
  import { ServerStatus } from '@/features/servers/server-status'

  let { server, onchange }: { server: Server; onchange: (updated: Server) => void } = $props()

  let pending = $state<Action | null>(null)
  let error = $state('')
  let info = $derived(details(server))
</script>

<article class="card" aria-labelledby="name-{server.id}">
  <GameIcon game={server.game} />
  <div class="body">
    <h2 id="name-{server.id}"><a class="name" href="/servers/{server.id}">{server.name}</a></h2>
    <!-- Only when it adds something ("Old world" → Minecraft). -->
    {#if GAMES[server.game] !== server.name}
      <p class="game">{GAMES[server.game] ?? server.game}</p>
    {/if}
    <ServerStatus {server} {pending} />
    {#if info && !pending}<p class="details">{info}</p>{/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
  </div>
  <div class="menu">
    <ServerMenu {server} bind:pending onchange={(s) => ((error = ''), onchange(s))} onerror={(m) => (error = m)} />
  </div>
</article>

<style>
  .card {
    position: relative;
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: start;
    gap: var(--space-4);
    padding: var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
    transition: border-color var(--duration-fast) var(--ease-out);
  }

  @media (hover: hover) {
    .card:hover {
      border-color: var(--color-border-strong);
    }
  }

  .body {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0;
  }

  h2 {
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
  }

  /* The name's link covers the whole card. */
  .name {
    color: inherit;
    text-decoration: none;
  }

  .name::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: inherit;
  }

  .name:focus-visible {
    outline: none;
  }

  .card:has(.name:focus-visible) {
    outline: var(--focus-ring);
    outline-offset: var(--focus-offset);
  }

  /* Above the stretched link, so the ⋯ gets its own clicks. */
  .menu {
    position: relative;
    z-index: 1;
    margin: calc(-1 * var(--space-2)) calc(-1 * var(--space-2)) 0 0;
  }

  p {
    margin: 0;
  }

  .game,
  .details {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
  }

  .error {
    position: relative;
    z-index: 1;
    color: var(--color-danger-text);
    font-size: var(--text-sm);
  }
</style>
