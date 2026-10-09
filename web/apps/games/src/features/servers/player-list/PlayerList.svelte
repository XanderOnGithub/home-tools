<!--
  Who's on a server right now. Minecraft: from RCON, each with their head
  (from Mojang via our server; a plain icon when there's none). Valheim:
  character names from its log (decision #40). Every state says something
  useful: stopped, not set up (Minecraft without RCON), couldn't ask,
  nobody on. Unnamed players (a count without names) are summed up.
-->
<script lang="ts">
  import { SvelteSet } from 'svelte/reactivity'
  import { headUrl, type Server } from '@/features/servers'

  let { server }: { server: Server } = $props()

  // Players whose head didn't load (offline-mode server, Mojang down…).
  const noHead = new SvelteSet<string>()

  let players = $derived(server.players)
  let unnamed = $derived(players ? Math.max(0, players.online - players.names.length) : 0)
</script>

<section class="players" aria-labelledby="players-title">
  <div class="head">
    <h2 id="players-title">Players</h2>
    {#if players}<span class="count">{players.online}{players.max ? ` / ${players.max}` : ''}</span>{/if}
  </div>

  {#if !server.state?.running}
    <p class="muted">Nobody can be online while it's stopped.</p>
  {:else if players}
    {#if players.online === 0}
      <p class="muted">Nobody's online.</p>
    {:else}
      <ul>
        {#each players.names as name (name)}
          <li>
            {#if server.game === 'minecraft' && !noHead.has(name)}
              <img class="head" src={headUrl(server.id, name)} alt="" onerror={() => noHead.add(name)} />
            {:else}
              <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></svg>
            {/if}
            {name}
          </li>
        {/each}
      </ul>
      {#if unnamed > 0}
        <p class="muted">
          {players.names.length ? `+${unnamed} more` : `${unnamed} online`}; the game doesn't share names.
        </p>
      {/if}
    {/if}
  {:else if server.players_error}
    <p class="muted">Couldn't find out who's online.</p>
  {:else}
    <p class="muted">
      Not set up: add RCON (<code>query</code> and <code>rcon_password</code>) to this server's file to see who's on.
    </p>
  {/if}
</section>

<style>
  .players {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .head {
    display: flex;
    align-items: baseline;
    gap: var(--space-3);
  }

  h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  .count {
    color: var(--color-text-muted);
    font-weight: var(--weight-semibold);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
  }

  ul {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-4) var(--space-2) var(--space-3);
    border-radius: var(--radius-full);
    background: var(--color-accent-subtle);
    font-weight: var(--weight-semibold);
  }

  /* Minecraft faces are 8×8 pixels: scaled up, kept blocky. */
  .head {
    width: 1.5rem;
    height: 1.5rem;
    margin: calc(-1 * var(--space-1)) 0;
    border-radius: var(--radius-sm);
    image-rendering: pixelated;
  }

  li svg {
    width: 1.1rem;
    height: 1.1rem;
    fill: none;
    stroke: var(--color-accent-text);
    stroke-width: 2;
    stroke-linecap: round;
  }

  code {
    font-family: var(--font-mono);
    font-size: 0.9em;
  }
</style>
