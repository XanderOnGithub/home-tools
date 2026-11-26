<!--
  Top-left: the current tool's mascot and name, opening a menu of every tool
  (TOOLS + the Home Tools page, decision #57). This is what makes four
  subdomains read as one product: the same switcher, in the same place,
  on every page. Links swap the host's first label (toolUrl), so the real
  domain stays out of git.
-->
<script lang="ts">
  import { placeUnder } from '../menu'
  import { ToolMascot } from '../tool-mascot'
  import { HOME, TOOLS, toolUrl, type Tool } from '../tools'

  let { tool }: { tool: Tool } = $props()

  let trigger = $state<HTMLButtonElement>()
  let menu = $state<HTMLDivElement>()
  let open = $state(false)

  const href = (t: Tool) => toolUrl(t.id, location)
</script>

<button
  bind:this={trigger}
  type="button"
  class="trigger"
  popovertarget="tool-menu"
  aria-label="{tool.name}: switch tool"
>
  <span class="mascot lg"><ToolMascot {tool} /></span>
  <span class="name">{tool.name}</span>
  <svg class="caret" class:open viewBox="0 0 24 24" aria-hidden="true"><path d="M7 10l5 5 5-5" /></svg>
</button>

<div
  bind:this={menu}
  id="tool-menu"
  class="menu"
  popover="auto"
  onbeforetoggle={(e) => (open = placeUnder(e, trigger, menu, 'start'))}
>
  <p class="menu-label">Tools</p>
  {#each TOOLS as t (t.id)}
    <a class="menu-item" href={href(t)} aria-current={t.id === tool.id ? 'page' : undefined}>
      <span class="mascot"><ToolMascot tool={t} /></span>
      {t.name}
      {#if t.id === tool.id}
        <svg class="check" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12.5l4.5 4.5L19 7.5" /></svg>
      {/if}
    </a>
  {/each}
  <hr class="menu-separator" />
  <a class="menu-item" href={href(HOME)}>
    <span class="mascot"><ToolMascot tool={HOME} /></span>
    All tools
  </a>
</div>

<style>
  .trigger {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-height: var(--touch-target);
    margin-left: calc(-1 * var(--space-2));
    padding: 0 var(--space-2);
    border: none;
    border-radius: var(--radius-md);
    background: none;
    color: var(--color-text);
    cursor: pointer;
    transition: background var(--duration-fast) var(--ease-out);
  }

  @media (hover: hover) {
    .trigger:hover {
      background: var(--color-surface-sunken);
    }
  }

  /* Mascots (decision #63) in place of icons: a little bigger in the
   * trigger, where it stands for the whole app. Their props may stick out
   * of the box (overflow is visible), so they get a little room. */
  .mascot {
    flex: none;
    width: 1.5rem;
    height: 1.5rem;
    margin-block: 0.125rem;
  }

  .mascot.lg {
    width: 2rem;
    height: 2rem;
  }

  .name {
    font-weight: var(--weight-bold);
    letter-spacing: var(--tracking-tight);
  }

  .caret {
    width: 1rem;
    height: 1rem;
    fill: none;
    stroke: var(--color-text-muted);
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
    transition: rotate var(--duration-fast) var(--ease-out);
  }

  .caret.open {
    rotate: 180deg;
  }

  .check {
    width: 1rem;
    height: 1rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.25;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
</style>
