<!--
  A tool's icon: its glyph in white on a rounded tile of the tool's color,
  like an app icon (TOOLS in ../tools). The same drawing is the favicon
  (apps/fitness/scripts/favicon). Decorative: the tool's name is always
  next to it, so it's hidden from screen readers.
-->
<script lang="ts">
  import type { Tool } from '../tools'

  let { tool, size = 'md' }: { tool: Tool; size?: 'sm' | 'md' | 'lg' } = $props()
</script>

<span class="tile {size}" data-accent={tool.color} aria-hidden="true">
  <svg viewBox="0 0 24 24"><path d={tool.icon} /></svg>
</span>

<style>
  .tile {
    display: inline-grid;
    flex: none;
    place-items: center;
    border-radius: 28%; /* scales with the tile, like an app icon */
    background: var(--color-tile);
    color: #fff; /* on --color-tile, in both themes */
  }

  .sm {
    width: 1.5rem;
    height: 1.5rem;
  }

  .md {
    width: 2.25rem;
    height: 2.25rem;
  }

  .lg {
    width: 3.5rem;
    height: 3.5rem;
  }

  svg {
    width: 62%;
    height: 62%;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .lg svg {
    stroke-width: 1.75;
  }
</style>
