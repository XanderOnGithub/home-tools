<!--
  One link per tool (TOOLS in @home-tools/ui, decision #46), as rows in
  one card: the tool's icon (the same tile as its favicon and the header's
  tool switcher, decision #57), its name and a short line, and a chevron.
  The whole row is the link, so its accessible name is the name +
  description.
-->
<script lang="ts">
  import { ToolIcon } from '@home-tools/ui/tool-icon'
  import { TOOLS, toolUrl } from '@home-tools/ui/tools'
</script>

<ul class="tool-list">
  {#each TOOLS as tool (tool.id)}
    <li>
      <a class="tool" href={toolUrl(tool.id, location)}>
        <ToolIcon {tool} size="md" />
        <span class="tool-text">
          <span class="tool-name">{tool.name}</span>
          <span class="tool-description">{tool.description}</span>
        </span>
        <svg class="chevron" viewBox="0 0 24 24" aria-hidden="true"><path d="M9 6l6 6-6 6" /></svg>
      </a>
    </li>
  {/each}
</ul>

<style>
  .tool-list {
    width: 100%;
    margin: 0;
    padding: 0;
    overflow: hidden;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
    list-style: none;
  }

  li + li {
    border-top: 1px solid var(--color-border);
  }

  .tool {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    min-height: 4rem;
    padding: var(--space-3) var(--space-4);
    color: var(--color-text);
    text-decoration: none;
    transition: background-color var(--duration-fast) var(--ease-out);
  }

  .tool:focus-visible {
    outline-offset: -2px;
  }

  .tool-text {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.125rem;
    min-width: 0; /* lets long text wrap instead of widening the row */
  }

  .tool-name {
    font-weight: var(--weight-semibold);
  }

  .tool-description {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .chevron {
    flex: none;
    width: 1.125rem;
    height: 1.125rem;
    fill: none;
    stroke: var(--color-text-muted);
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
    transition: translate var(--duration-fast) var(--ease-out);
  }

  /* Hover only where a pointer exists; on touch, :hover sticks after a tap. */
  @media (hover: hover) {
    .tool:hover {
      background: var(--color-surface-sunken);
    }

    .tool:hover .chevron {
      translate: 2px 0;
    }
  }

  .tool:active {
    background: var(--color-accent-subtle);
  }
</style>
