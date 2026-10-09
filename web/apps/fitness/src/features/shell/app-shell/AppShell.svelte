<!--
  The frame around every signed-in screen: navigation + profile menu.

  One set of links, two layouts: in the header on wide screens, a bottom
  tab bar on phones (within thumb reach). The current page's link gets
  aria-current="page", which screen readers announce as "current page".

  On navigation, focus moves to the new page's <h1> (single-page apps don't
  reload, so without this screen readers wouldn't notice the page change).
-->
<script lang="ts">
  import { tick, type Snippet } from 'svelte'
  import { ProfileMenu } from '@/features/profiles/profile-menu'
  import type { Profile } from '@/features/profiles/types'
  import { router } from '@/router'

  let { profile, onswitch, children }: { profile: Profile; onswitch: () => void; children: Snippet } =
    $props()

  const LINKS = [
    { href: '/', label: 'Home', icon: 'M3 11l9-8 9 8 M5 9.5V20h5v-6h4v6h5V9.5' },
    { href: '/plans', label: 'Plans', icon: 'M8 6h12 M8 12h12 M8 18h12 M4 6h.01 M4 12h.01 M4 18h.01' },
    { href: '/history', label: 'History', icon: 'M12 7v5l3 2 M3.05 11a9 9 0 1 1 .5 4 M3 4v5h5' },
  ]

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
    {#each LINKS as link (link.href)}
      <li>
        <a href={link.href} aria-current={router.path === link.href ? 'page' : undefined}>
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
      <nav class="header-nav" aria-label="Main">
        {@render navLinks('inline')}
      </nav>
      <ProfileMenu {profile} {onswitch} />
    </div>
  </header>

  <main class="main" bind:this={main}>
    {@render children()}
  </main>

  <nav class="tab-bar" aria-label="Main">
    {@render navLinks('tabs')}
  </nav>
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
    z-index: 1;
    border-bottom: 1px solid var(--color-border);
    background: var(--color-bg);
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
    justify-content: space-between;
    gap: var(--space-4);
    min-height: 3.5rem;
    padding-block: var(--space-1);
  }

  .main {
    flex: 1;
    padding-block: var(--space-5) var(--space-7);
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
    display: flex;
    align-items: center;
    color: var(--color-text-muted);
    font-weight: var(--weight-semibold);
    text-decoration: none;
    transition:
      color var(--duration-fast) var(--ease-out),
      background var(--duration-fast) var(--ease-out);
  }

  .links svg {
    flex: none;
    width: 1.375rem;
    height: 1.375rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .links a[aria-current='page'] {
    color: var(--color-accent-text);
  }

  @media (hover: hover) {
    .links a:not([aria-current='page']):hover {
      color: var(--color-text);
    }
  }

  /* Inline links in the header (wide screens): icon + label pills. The
   * first pill's padding is pulled back so its icon lines up with the
   * page content's left edge. */
  .links.inline {
    gap: var(--space-1);
    margin-left: calc(-1 * var(--space-4));
  }

  .links.inline a {
    gap: var(--space-2);
    min-height: var(--touch-target);
    padding: 0 var(--space-4);
    border-radius: var(--radius-full);
  }

  .links.inline a[aria-current='page'] {
    background: var(--color-accent-subtle);
  }

  /* Bottom tabs (phones): icon above label, evenly spread. */
  .tab-bar {
    position: sticky;
    bottom: 0;
    z-index: 1;
    border-top: 1px solid var(--color-border);
    background: var(--color-bg);
    padding-bottom: env(safe-area-inset-bottom); /* iPhone home bar */
  }

  .links.tabs li {
    flex: 1;
  }

  .links.tabs a {
    flex-direction: column;
    justify-content: center;
    gap: 2px;
    min-height: 3.5rem;
    font-size: var(--text-xs);
  }

  /* Phones: tabs only. Wide screens: header links only. */
  .header-nav {
    display: none;
  }

  @media (min-width: 40rem) {
    .header-nav {
      display: block;
    }

    .tab-bar {
      display: none;
    }

    .header-inner {
      padding-block: var(--space-2);
    }
  }

  /* Phones: nothing on the header's left, so keep the menu on the right. */
  @media (max-width: 39.99rem) {
    .header-inner {
      justify-content: flex-end;
    }
  }
</style>
