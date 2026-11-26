<!--
  One big card per tool (TOOLS in @home-tools/ui, decision #46): the
  tool's mascot on a stage in its color, then its name, one line and
  "Open". The whole card is the link; the mascot hops when it's hovered,
  tapped or focused. Cards have a thick bottom edge and press down into
  it, so they feel like something you can push.
-->
<script lang="ts">
  import { ToolMascot } from '@home-tools/ui/tool-mascot'
  import { TOOLS, toolUrl } from '@home-tools/ui/tools'
</script>

<section class="tools" aria-labelledby="tools-title">
  <h2 id="tools-title">Pick a tool</h2>
  <ul class="cards">
    {#each TOOLS as tool (tool.id)}
      <li>
        <a class="card" href={toolUrl(tool.id, location)} data-accent={tool.color}>
          <span class="stage">
            <span class="mascot"><ToolMascot {tool} /></span>
            <span class="shadow"></span>
          </span>
          <span class="body">
            <span class="name">{tool.name}</span>
            <span class="description">{tool.description}</span>
            <span class="open" aria-hidden="true">
              Open
              <svg viewBox="0 0 24 24"><path d="M5 12h14M13 6l6 6-6 6" /></svg>
            </span>
          </span>
        </a>
      </li>
    {/each}
  </ul>
</section>

<style>
  .tools {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    width: 100%;
    max-width: var(--page-width);
    margin: 0 auto;
    padding: 0 var(--page-gutter);
  }

  .cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 18rem), 1fr));
    gap: var(--space-5);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .card {
    --edge: var(--color-border);
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    border: 2px solid var(--edge);
    border-radius: var(--radius-xl);
    background: var(--color-surface);
    box-shadow: 0 4px 0 var(--edge);
    color: var(--color-text);
    text-decoration: none;
    transition:
      border-color var(--duration-fast) var(--ease-out),
      box-shadow var(--duration-fast) var(--ease-out),
      translate var(--duration-fast) var(--ease-out);
  }

  /* The stage: the tool's color, lightly, with the hero's grid on it. */
  .stage {
    --tint: color-mix(in oklab, var(--color-avatar) 22%, var(--color-surface));
    --line: color-mix(in oklab, var(--color-avatar) 20%, transparent);
    position: relative;
    display: grid;
    place-items: center;
    aspect-ratio: 16 / 9;
    background-color: var(--tint);
    background-image: linear-gradient(var(--line) 1px, transparent 1px),
      linear-gradient(90deg, var(--line) 1px, transparent 1px);
    background-position: center;
    background-size: var(--space-5) var(--space-5);
  }

  .mascot {
    position: relative;
    z-index: 1;
    width: 38%;
    max-width: 9rem;
    margin-top: 6%;
  }

  .shadow {
    position: absolute;
    bottom: 11%;
    width: 26%;
    max-width: 6rem;
    height: 6%;
    border-radius: 50%;
    background: color-mix(in oklab, var(--color-avatar-eye) 14%, transparent);
  }

  .body {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-4) var(--space-5) var(--space-5);
  }

  .name {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
    letter-spacing: var(--tracking-tight);
  }

  .description {
    color: var(--color-text-muted);
  }

  .open {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    margin-top: auto;
    padding-top: var(--space-3);
    color: var(--color-accent-text);
    font-weight: var(--weight-semibold);
  }

  .open svg {
    width: 1.125rem;
    height: 1.125rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.25;
    stroke-linecap: round;
    stroke-linejoin: round;
    transition: translate var(--duration-fast) var(--ease-out);
  }

  .card:focus-visible {
    outline-offset: 3px;
  }

  @media (hover: hover) {
    .card:hover {
      --edge: var(--color-avatar);
      translate: 0 -2px;
      box-shadow: 0 6px 0 var(--edge);
    }

    .card:hover .open svg {
      translate: 3px 0;
    }
  }

  /* Pressed: the card sinks into its edge. */
  .card:active {
    --edge: var(--color-avatar);
    translate: 0 3px;
    box-shadow: 0 1px 0 var(--edge);
  }
</style>
