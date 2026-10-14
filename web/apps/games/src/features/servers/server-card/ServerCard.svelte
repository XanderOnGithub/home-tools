<!--
  One game server in the list, like a game on a streaming site's browse
  page (decision #58): a tall cover (the server's own art, or a drawn
  scene), its state and who's on laid over it, then the name and a line
  of details. The whole card opens the server's page (a "stretched" link
  on the name, so it's one real <a> for screen readers and right-click);
  the ⋯ menu sits on top of it for Start / Stop / Restart.

  Stopped servers fade their cover, so what's running stands out at a
  glance (the badge says it in words too).
-->
<script lang="ts">
  import { GAMES, artUrl, uptime, type Action, type Server } from '@/features/servers'
  import { GameCover } from '@/features/servers/game-cover'
  import { ServerMenu } from '@/features/servers/server-menu'

  let { server, onchange }: { server: Server; onchange: (updated: Server) => void } = $props()

  let pending = $state<Action | null>(null)

  const DOING: Record<Action, string> = { start: 'Starting…', stop: 'Stopping…', restart: 'Restarting…' }
  let running = $derived(server.state?.running ?? false)
  let doing = $derived(pending ?? server.busy ?? null)
  let phase = $derived<'doing' | 'running' | 'stopped' | 'unknown'>(
    doing ? 'doing' : !server.state ? 'unknown' : running ? 'running' : 'stopped',
  )
  let label = $derived(
    doing ? DOING[doing] : phase === 'unknown' ? 'Unavailable' : phase === 'running' ? 'Running' : 'Stopped',
  )

  let players = $derived(running && server.players ? server.players : null)
  let line = $derived.by(() => {
    const parts: string[] = []
    // The game, unless the name already says it ("Valheim").
    const game = GAMES[server.game] ?? server.game
    if (game !== server.name) parts.push(game)
    if (phase === 'running' && server.state?.started_at) parts.push(`Up ${uptime(server.state.started_at)}`)
    if (phase === 'stopped' && parts.length === 0) parts.push('Not running')
    if (phase === 'unknown') parts.push(server.error ?? 'Unknown error')
    return parts.join(' · ')
  })
</script>

<article class="card {phase}" aria-labelledby="name-{server.id}">
  <div class="art">
    <GameCover game={server.game} src={artUrl(server, 'cover')} />
    <span class="state">
      <span class="dot" aria-hidden="true"></span>
      {label}
    </span>
    {#if players}
      <span class="players">
        <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></svg>
        {players.online}{players.max ? ` / ${players.max}` : ''}
        <span class="visually-hidden">online</span>
      </span>
    {/if}
  </div>

  <div class="meta">
    <div class="text">
      <h2 id="name-{server.id}"><a class="name" href="/servers/{server.id}">{server.name}</a></h2>
      <p class="line" class:error={phase === 'unknown'}>{line}</p>
    </div>
    <div class="actions">
      <ServerMenu {server} bind:pending {onchange} />
    </div>
  </div>
</article>

<style>
  .card {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    min-width: 0;
  }

  /* The cover: the accent sits behind it and shows as it lifts on hover
   * (a quiet nod to how streaming sites do it). */
  .art {
    position: relative;
    border-radius: var(--radius-lg);
    background: var(--color-accent);
  }

  .art :global(.cover) {
    border-radius: var(--radius-lg);
    box-shadow: inset 0 0 0 1px rgb(0 0 0 / 0.06);
    transition:
      translate var(--duration-normal) var(--ease-out),
      filter var(--duration-normal) var(--ease-out),
      opacity var(--duration-normal) var(--ease-out);
  }

  @media (hover: hover) {
    .card:hover .art :global(.cover) {
      translate: -0.375rem -0.375rem;
    }
  }

  .card:active:not(:has(.actions:active)) .art :global(.cover) {
    translate: -0.125rem -0.125rem;
  }

  .stopped .art :global(.cover),
  .unknown .art :global(.cover) {
    filter: grayscale(0.85);
    opacity: 0.75;
  }

  .stopped:hover .art :global(.cover) {
    filter: grayscale(0.4);
    opacity: 0.9;
  }

  /* Over the picture: fixed dark chips, readable on any art. */
  .state,
  .players {
    position: absolute;
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    min-height: 1.5rem;
    padding: 0 var(--space-2);
    border-radius: var(--radius-sm);
    background: rgb(0 0 0 / 0.72); /* on any art: fixed, not themed */
    color: var(--color-on-overlay);
    backdrop-filter: blur(6px);
    font-size: var(--text-xs);
    font-weight: var(--weight-semibold);
    pointer-events: none;
    transition: translate var(--duration-normal) var(--ease-out);
  }

  @media (hover: hover) {
    .card:hover .state,
    .card:hover .players {
      translate: -0.375rem -0.375rem;
    }
  }

  .state {
    top: var(--space-2);
    left: var(--space-2);
  }

  .players {
    bottom: var(--space-2);
    left: var(--space-2);
    font-variant-numeric: tabular-nums;
  }

  .players svg {
    width: 0.8rem;
    height: 0.8rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.5;
  }

  .dot {
    width: 0.4375rem;
    height: 0.4375rem;
    border-radius: var(--radius-full);
    background: var(--color-on-overlay-muted);
  }

  /* Fixed colors: they sit on the dark chip, not on a themed surface. */
  .running .dot {
    background: #4ade80;
    animation: pulse 2s ease-in-out infinite;
  }

  .doing .dot {
    background: #facc15;
    animation: pulse 1s ease-in-out infinite;
  }

  .unknown .dot {
    background: #f87171;
  }

  @keyframes pulse {
    50% {
      opacity: 0.35;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .dot {
      animation: none !important;
    }
  }

  .meta {
    display: flex;
    align-items: flex-start;
    gap: var(--space-1);
  }

  .text {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
  }

  h2 {
    font-size: var(--text-md);
    font-weight: var(--weight-semibold);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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
  }

  @media (hover: hover) {
    .card:hover .name {
      color: var(--color-accent-text);
    }
  }

  .name:focus-visible {
    outline: none;
  }

  .card:has(.name:focus-visible) .art {
    outline: var(--focus-ring);
    outline-offset: var(--focus-offset);
  }

  /* Up to two lines on narrow cards (phones), then cut. */
  .line {
    display: -webkit-box;
    margin: 0;
    overflow: hidden;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    line-clamp: 2;
  }

  .line.error {
    color: var(--color-danger-text);
  }

  /* Above the stretched link, so the ⋯ gets its own clicks. */
  .actions {
    position: relative;
    z-index: 1;
    margin: calc(-1 * var(--space-2)) calc(-1 * var(--space-2)) 0 0;
  }
</style>
