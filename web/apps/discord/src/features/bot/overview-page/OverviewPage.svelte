<!--
  Home of the Discord tool (decision #60): the bot at a glance (who it is
  today, online or not, its servers and status), anyone waiting for
  access, every feature with what it's doing (each card opens its
  settings under Features), and the slash commands that are live.
-->
<script lang="ts">
  import { Badge } from '@home-tools/ui/components/badge'
  import { Card } from '@home-tools/ui/components/card'
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { PageHeader } from '@home-tools/ui/components/page-header'
  import { Section } from '@home-tools/ui/components/section'
  import { SettingsSkeleton } from '@/components/settings-skeleton'
  import { CommandsHelp } from '@/features/bot/commands-help'
  import { FEATURES } from '@/features/bot/feature-list'
  import { settings } from '@/features/bot/settings'
  import { TodayCard } from '@/features/bot/today-card'

  const s = settings()

  let requests = $derived(s.bot?.requests.length ?? 0)
</script>

<div class="page">
  <PageHeader title="Bot" />

  {#if s.status === 'error'}
    <LoadError what="the bot's settings" onretry={s.load} />
  {:else if s.config && s.bot}
    {@const config = s.config}
    {@const bot = s.bot}
    <TodayCard {bot} />

    {#if requests > 0}
      <a class="requests" href="/people">
        <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 8v5 M12 16.5h.01 M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18z" /></svg>
        <span>
          {requests === 1 ? `${bot.requests[0].name} is` : `${requests} people are`} asking to use
          {requests === 1 ? `/${bot.requests[0].command}` : 'verified commands'}
        </span>
        <span class="go">Review</span>
      </a>
    {/if}

    <Section title="Features">
      {#snippet actions()}<a class="btn btn-quiet btn-sm" href="/features">Manage</a>{/snippet}
      <ul class="features">
        {#each FEATURES as f (f.id)}
          {@const on = f.on(config)}
          <li>
            <Card as="a" href="/features/{f.id}" padding="md">
              <span class="feature">
                <span class="icon" class:dim={on === false} aria-hidden="true">
                  <svg viewBox="0 0 24 24"><path d={f.icon} /></svg>
                </span>
                <span class="text">
                  <span class="name">{f.label}</span>
                  <span class="summary">{f.summary(config, bot)}</span>
                </span>
                {#if on !== null}
                  <Badge tone={on ? 'success' : 'neutral'}>{on ? 'On' : 'Off'}</Badge>
                {/if}
              </span>
            </Card>
          </li>
        {/each}
      </ul>
    </Section>

    <Section title="Commands">
      <CommandsHelp features={config.features} />
    </Section>
  {:else}
    <SettingsSkeleton sections={2} />
  {/if}
</div>

<style>
  .page {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    max-width: var(--content-medium);
  }

  .page > :global(.page-header) {
    margin-bottom: 0;
  }

  .page > :global(.section + .section),
  .page > :global(.section) {
    margin-top: var(--space-4);
  }

  .requests {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border: 1px solid color-mix(in oklab, var(--color-warning-text) 35%, transparent);
    border-radius: var(--radius-lg);
    background: var(--color-warning-subtle);
    color: var(--color-text);
    font-size: var(--text-sm);
    text-decoration: none;
  }

  .requests svg {
    flex: none;
    width: 1.125rem;
    height: 1.125rem;
    fill: none;
    stroke: var(--color-warning-text);
    stroke-width: 2;
    stroke-linecap: round;
  }

  .requests span:not(.go) {
    flex: 1;
  }

  .go {
    color: var(--color-accent-text);
    font-weight: var(--weight-semibold);
  }

  .features {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(16rem, 1fr));
    gap: var(--space-3);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .feature {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .icon {
    display: grid;
    flex: none;
    place-items: center;
    width: 2.25rem;
    height: 2.25rem;
    border-radius: var(--radius-md);
    background: var(--color-accent-subtle);
    color: var(--color-accent-text);
  }

  .icon.dim {
    background: var(--color-surface-sunken);
    color: var(--color-text-muted);
  }

  .icon svg {
    width: 1.125rem;
    height: 1.125rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.75;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .text {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.125rem;
    min-width: 0;
  }

  .name {
    color: var(--color-text);
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
  }

  .summary {
    overflow: hidden;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
