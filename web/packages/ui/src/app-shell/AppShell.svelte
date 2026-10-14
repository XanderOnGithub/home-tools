<!--
  The frame around every signed-in screen: tool switcher, navigation,
  profile menu, and the toasts.

  Shared by every tool, so they all read as one product (decision #57):
  the same header with the same switcher in the same place. One set of
  links (the tool passes its own), two layouts: in the header on wide
  screens, a bottom tab bar on phones (within thumb reach). With a single
  link there's nothing to switch between, so it isn't shown at all. A
  link is current on its own page and the pages under it (/plans/push
  → Plans); it gets aria-current="page", announced as "current page".

  On navigation, focus moves to the new page's <h1> (single-page apps don't
  reload, so without this screen readers wouldn't notice the page change).
-->
<script lang="ts">
  import { tick, type Snippet } from 'svelte'
  import { ProfileMenu } from '../profiles/profile-menu'
  import type { Profile } from '../profiles/types'
  import { router } from '../router'
  import { Toaster } from '../toast'
  import { toolById } from '../tools'
  import ToolSwitcher from './ToolSwitcher.svelte'
  import type { NavLink } from './types'

  let {
    tool: toolId,
    links,
    profile,
    onswitch,
    children,
  }: {
    tool: string // TOOLS id: whose name and icon the switcher shows
    links: NavLink[]
    profile: Profile
    onswitch: () => void
    children: Snippet
  } = $props()

  let tool = $derived(toolById(toolId)!)

  const isCurrent = (href: string) =>
    router.path === href || (href !== '/' && router.path.startsWith(`${href}/`))

  let main = $state<HTMLElement>()
  let first = true

  $effect(() => {
    void router.path // re-run on every navigation
    if (first) {
      first = false // initial load: leave focus where the browser put it
      return
    }
    tick().then(() => main?.querySelector<HTMLElement>('h1')?.focus())
  })
</script>

{#snippet navLinks(variant: 'inline' | 'tabs')}
  <ul class="links {variant}">
    {#each links as link (link.href)}
      <li>
        <a href={link.href} aria-current={isCurrent(link.href) ? 'page' : undefined}>
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d={link.icon} /></svg>
          <span>{link.label}</span>
        </a>
      </li>
    {/each}
  </ul>
{/snippet}

<div class="shell">
  <header class="header">
    <div class="header-inner">
      <ToolSwitcher {tool} />
      {#if links.length > 1}
        <nav class="header-nav" aria-label="Main">
          {@render navLinks('inline')}
        </nav>
      {/if}
      <div class="header-end">
        <ProfileMenu {profile} {onswitch} />
      </div>
    </div>
  </header>

  <main class="main" bind:this={main}>
    {@render children()}
  </main>

  {#if links.length > 1}
    <nav class="tab-bar" aria-label="Main">
      {@render navLinks('tabs')}
    </nav>
  {/if}

  <Toaster />
</div>

<style>
  .shell {
    display: flex;
    flex-direction: column;
    min-height: 100dvh;
  }

  /* The header's background and border span the window; its content
   * lines up with main's (same --page-width and gutter). */
  .header {
    position: sticky;
    top: 0;
    z-index: 2;
    border-bottom: 1px solid var(--color-border);
    background: var(--color-surface);
  }

  .header-inner,
  .main {
    width: 100%;
    max-width: calc(var(--page-width) + 2 * var(--page-gutter));
    margin: 0 auto;
    padding-inline: var(--page-gutter);
  }

  .header-inner {
    display: flex;
    align-items: center;
    gap: var(--space-5);
    height: var(--header-height);
  }

  .header-end {
    margin-left: auto;
  }

  .main {
    flex: 1;
    padding-block: var(--space-5) var(--space-8);
  }

  /* Page headings receive focus on navigation; it's not an interactive
   * element, so no focus ring. */
  .main :global(h1:focus) {
    outline: none;
  }

  .links {
    display: flex;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .links a {
    position: relative;
    display: flex;
    align-items: center;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
    text-decoration: none;
    transition: color var(--duration-fast) var(--ease-out);
  }

  .links svg {
    flex: none;
    width: 1.25rem;
    height: 1.25rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.75;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .links a[aria-current='page'] {
    color: var(--color-text);
  }

  @media (hover: hover) {
    .links a:not([aria-current='page']):hover {
      color: var(--color-text);
    }
  }

  /* Header links (wide screens): text only, the current one underlined
   * along the header's bottom edge, like tabs. */
  .links.inline {
    gap: var(--space-5);
    height: var(--header-height);
  }

  .links.inline a {
    height: 100%;
  }

  .links.inline svg {
    display: none;
  }

  .links.inline a[aria-current='page']::after {
    content: '';
    position: absolute;
    inset: auto 0 -1px;
    height: 2px;
    border-radius: 2px 2px 0 0;
    background: var(--color-accent);
  }

  /* Bottom tabs (phones): icon above label, evenly spread. */
  .tab-bar {
    position: sticky;
    bottom: 0;
    z-index: 2;
    border-top: 1px solid var(--color-border);
    background: var(--color-surface);
    padding-bottom: env(safe-area-inset-bottom); /* iPhone home bar */
  }

  .links.tabs li {
    flex: 1;
  }

  .links.tabs a {
    flex-direction: column;
    justify-content: center;
    gap: var(--space-1);
    min-height: 3.25rem;
    font-size: var(--text-xs);
  }

  .links.tabs a[aria-current='page'] {
    color: var(--color-accent-text);
  }

  .links.tabs a[aria-current='page'] svg {
    stroke-width: 2.25;
  }

  /* Phones: tabs only. Wide screens: header links only. */
  .header-nav {
    display: none;
    align-self: stretch;
  }

  @media (min-width: 40rem) {
    .header-nav {
      display: block;
    }

    .tab-bar {
      display: none;
    }
  }
</style>
