<!--
  A small picture for each game, drawn here (not the games' logos):
  Minecraft = a block, Valheim = a horned helmet, in the person's accent
  color. Decorative: the game's name is always next to it.
-->
<script lang="ts">
  import type { Game } from '@/features/servers'

  let { game }: { game: Game } = $props()
</script>

<span class="tile {game}" aria-hidden="true">
  <svg viewBox="0 0 48 48">
    {#if game === 'minecraft'}
      <!-- Isometric block: grass top, two dirt sides, a grass lip. -->
      <path class="top" d="M24 6 42 15 24 24 6 15z" />
      <path class="left" d="M6 15 24 24v18L6 33z" />
      <path class="right" d="M42 15 24 24v18l18-9z" />
      <path class="lip" d="M6 15 24 24v5L6 20z M42 15 24 24v5l18-9z" />
    {:else}
      <!-- Helmet: dome, brim, nose guard, two horns. -->
      <path class="horn" d="M12 22c-6-2-8-9-6-14 2 5 6 7 10 8z M36 22c6-2 8-9 6-14-2 5-6 7-10 8z" />
      <path class="dome" d="M11 27c0-9 6-15 13-15s13 6 13 15z" />
      <rect class="brim" x="9" y="26" width="30" height="5" rx="2" />
      <path class="brim" d="M22 31h4v9l-2 2-2-2z" />
    {/if}
  </svg>
</span>

<style>
  .tile {
    display: grid;
    flex: none;
    place-items: center;
    width: 3.5rem;
    height: 3.5rem;
    border-radius: var(--radius-md);
    background: var(--color-accent-subtle);
  }

  svg {
    width: 2.25rem;
    height: 2.25rem;
  }

  /* One color (the person's accent), shaded with opacity like the rest
   * of the UI: the shape says which game, not the colors. */
  svg {
    color: var(--color-accent);
    fill: currentColor;
  }
  .top,
  .dome {
    opacity: 1;
  }
  .lip,
  .brim {
    opacity: 0.8;
  }
  .left {
    opacity: 0.55;
  }
  .right,
  .horn {
    opacity: 0.4;
  }
</style>
