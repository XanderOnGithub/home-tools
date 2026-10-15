<!--
  Everything the bot does that has settings (decision #60), one at a time:
  a list of features on the left (a scrolling row of tabs on phones), the
  chosen one's settings on the right, at /features/<id>. /features alone
  shows the first. Each feature's own component brings its title, line
  and on/off switch; this page puts it in one card so they all look alike.
-->
<script lang="ts">
  import { Card } from '@home-tools/ui/components/card'
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { PageHeader } from '@home-tools/ui/components/page-header'
  import { SettingsSkeleton } from '@/components/settings-skeleton'
  import { BlobMaker } from '@/features/bot/blob-maker'
  import { FEATURES, featureById, type FeatureId } from '@/features/bot/feature-list'
  import { NameList } from '@/features/bot/name-list'
  import { PollSettings } from '@/features/bot/poll-settings'
  import { settings } from '@/features/bot/settings'
  import { StatusBoards } from '@/features/bot/status-boards'
  import { StatusSettings } from '@/features/bot/status-settings'

  let { id }: { id: string | null } = $props()

  const s = settings()

  let current = $derived<FeatureId>(featureById(id ?? '')?.id ?? FEATURES[0].id)

  let tabs = $state<HTMLElement>()

  $effect(() => {
    document.title = `${featureById(current)!.label} · Discord`
  })

  // Phones: keep the chosen tab in view in the sideways-scrolling row.
  $effect(() => {
    void current
    tabs?.querySelector('[aria-current="page"]')?.scrollIntoView({ block: 'nearest', inline: 'nearest' })
  })
</script>

<PageHeader title="Features" />

{#if s.status === 'error'}
  <LoadError what="the bot's settings" onretry={s.load} />
{:else if s.config && s.bot}
  {@const config = s.config}
  <div class="layout">
    <nav class="tabs" aria-label="Features" bind:this={tabs}>
      <ul>
        {#each FEATURES as f (f.id)}
          {@const on = f.on(config)}
          <li>
            <a href="/features/{f.id}" aria-current={f.id === current ? 'page' : undefined}>
              <svg viewBox="0 0 24 24" aria-hidden="true"><path d={f.icon} /></svg>
              <span class="label">{f.label}</span>
              {#if on === false}<span class="off">Off</span>{/if}
            </a>
          </li>
        {/each}
      </ul>
    </nav>

    <div class="panel">
      {#key current}
        <Card padding="lg">
          {#if current === 'game-statuses'}
            <StatusBoards {config} bot={s.bot} save={s.save} />
          {:else if current === 'names'}
            <NameList {config} save={s.save} />
          {:else if current === 'polls'}
            <PollSettings {config} bot={s.bot} save={s.save} onposted={s.refreshBot} />
          {:else if current === 'status'}
            <StatusSettings {config} bot={s.bot} save={s.save} />
          {:else}
            <BlobMaker {config} save={s.save} />
          {/if}
        </Card>
      {/key}
    </div>
  </div>
{:else}
  <SettingsSkeleton sections={1} />
{/if}

<style>
  /* Phones: tabs in a row that scrolls sideways, the panel under them.
   * Wide screens: a list on the left, like a settings app. */
  .layout {
    display: grid;
    gap: var(--space-4);
    align-items: start;
  }

  @media (min-width: 52rem) {
    .layout {
      grid-template-columns: 14rem minmax(0, 1fr);
      gap: var(--space-6);
    }

    .tabs {
      position: sticky;
      top: calc(var(--header-height) + var(--space-5));
    }
  }

  .panel {
    max-width: var(--content-medium);
    min-width: 0;
  }

  ul {
    display: flex;
    gap: var(--space-1);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  @media (max-width: 51.99rem) {
    .tabs {
      margin-inline: calc(-1 * var(--page-gutter));
      padding-inline: var(--page-gutter);
      overflow-x: auto;
      scrollbar-width: none;
    }

    li {
      flex: none;
    }
  }

  @media (min-width: 52rem) {
    ul {
      flex-direction: column;
    }
  }

  a {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border-radius: var(--radius-md);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
    text-decoration: none;
    white-space: nowrap;
    transition:
      background var(--duration-fast) var(--ease-out),
      color var(--duration-fast) var(--ease-out);
  }

  @media (min-width: 52rem) {
    a {
      min-height: 2.5rem;
    }
  }

  @media (hover: hover) {
    a:not([aria-current]):hover {
      background: var(--color-surface-sunken);
      color: var(--color-text);
    }
  }

  a[aria-current='page'] {
    background: var(--color-surface);
    color: var(--color-text);
    box-shadow:
      inset 0 0 0 1px var(--color-border),
      var(--shadow-sm);
  }

  a:focus-visible {
    outline-offset: -2px;
  }

  svg {
    flex: none;
    width: 1.125rem;
    height: 1.125rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.75;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  a[aria-current='page'] svg {
    color: var(--color-accent-text);
  }

  .label {
    flex: 1;
  }

  .off {
    color: var(--color-text-muted);
    font-size: var(--text-xs);
    font-weight: var(--weight-normal);
  }
</style>
