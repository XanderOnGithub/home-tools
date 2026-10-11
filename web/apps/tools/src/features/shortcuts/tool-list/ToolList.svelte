<!--
  One link per tool (TOOLS in @home-tools/ui, decision #46): its blob in its
  color, name and a short line about it. The whole tile is the link, so its
  accessible name is the tool's name + description; the blob is
  decorative (ProfileAvatar is aria-hidden).

  Phones: a list of rows (blob left, text right), so descriptions fit.
  Wider screens: tiles side by side, like the profile picker.
-->
<script lang="ts">
  import { ProfileAvatar } from '@home-tools/ui/profiles/profile-avatar'
  import { TOOLS, toolUrl } from '@home-tools/ui/tools'
</script>

<ul class="tool-list">
  {#each TOOLS as tool (tool.id)}
    <li>
      <!-- data-accent gives the blob (and focus ring) this tool's color. -->
      <a class="tool" href={toolUrl(tool.id, location)} data-accent={tool.color}>
        <span class="tool-blob"><ProfileAvatar name={tool.seed} /></span>
        <span class="tool-text">
          <span class="tool-name">{tool.name}</span>
          <span class="tool-description">{tool.description}</span>
        </span>
      </a>
    </li>
  {/each}
</ul>

<style>
  .tool-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    width: 100%;
    max-width: 28rem;
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .tool {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    min-height: var(--touch-target);
    padding: var(--space-3);
    border-radius: var(--radius-lg); /* rounds the focus ring and hover fill */
    color: var(--color-text);
    text-decoration: none;
    transition: background-color var(--duration-fast) var(--ease-out);
  }

  .tool-blob {
    flex: none;
    width: 4.5rem;
    height: 4.5rem;
    transition: scale var(--duration-fast) var(--ease-out);
  }

  .tool-text {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0; /* lets long text wrap instead of widening the row */
  }

  .tool-name {
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
  }

  .tool-description {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  /* Hover only where a pointer exists; on touch, :hover sticks after a tap. */
  @media (hover: hover) {
    .tool:hover {
      background: var(--color-surface);
    }
    .tool:hover .tool-blob {
      scale: 1.05;
    }
  }

  /* Pressed beats hover: instant feedback on tap. */
  .tool:active .tool-blob {
    scale: 0.97;
  }

  /* Room for tiles: blob on top, text centered below. */
  @media (min-width: 40rem) {
    .tool-list {
      flex-direction: row;
      flex-wrap: wrap;
      justify-content: center;
      gap: var(--space-4);
      max-width: 60rem;
    }

    .tool {
      flex-direction: column;
      width: 13rem;
      padding: var(--space-4);
      text-align: center;
    }

    .tool-blob {
      width: 8rem;
      height: 8rem;
    }
  }
</style>
