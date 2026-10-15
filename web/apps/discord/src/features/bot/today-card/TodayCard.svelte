<!--
  The bot at a glance (top of the overview): today's avatar and name,
  whether it's connected, how many Discord servers it's in, its status
  line right now, and the invite link.
-->
<script lang="ts">
  import { Badge } from '@home-tools/ui/components/badge'
  import { PERSONA_GIF, PERSONA_PNG, type BotView } from '@/features/bot'

  let { bot }: { bot: BotView } = $props()

  // Changes with the persona, so the browser fetches the new day's image.
  let version = $derived(`${bot.persona.date}-${bot.persona.color}-${encodeURIComponent(bot.persona.name)}`)
  let guilds = $derived(bot.guilds.length)
</script>

<section class="today" aria-labelledby="today-name">
  <picture>
    <source srcset="{PERSONA_PNG}?v={version}" media="(prefers-reduced-motion: reduce)" />
    <img class="avatar" src="{PERSONA_GIF}?v={version}" alt="" width="80" height="80" />
  </picture>

  <div class="who">
    <p class="eyebrow">Today the bot is</p>
    <h2 id="today-name">{bot.persona.name}</h2>
    <div class="state">
      {#if bot.connected}
        <Badge tone="success" dot pulse>Online</Badge>
        {#if bot.user}<span class="user">{bot.user}</span>{/if}
      {:else}
        <Badge tone="danger" dot>Offline</Badge>
        <span class="user">{bot.error ?? 'Not connected'}</span>
      {/if}
    </div>
  </div>

  <dl class="facts">
    <div>
      <dt>Servers</dt>
      <dd>{guilds}</dd>
    </div>
    <div>
      <dt>Status</dt>
      <dd class="status">{bot.status || 'None'}</dd>
    </div>
  </dl>

  {#if bot.invite_url}
    <a class="btn btn-secondary invite" href={bot.invite_url} target="_blank" rel="noreferrer">
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5v14 M5 12h14" /></svg>
      Add to a server
    </a>
  {/if}
</section>

<style>
  .today {
    display: grid;
    grid-template-columns: auto 1fr;
    grid-template-areas: 'avatar who' 'facts facts' 'invite invite';
    align-items: center;
    gap: var(--space-4) var(--space-4);
    padding: var(--space-5);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
  }

  @media (min-width: 44rem) {
    .today {
      grid-template-columns: auto minmax(0, 1fr) auto auto;
      grid-template-areas: 'avatar who facts invite';
      gap: var(--space-6);
    }
  }

  picture {
    grid-area: avatar;
  }

  .avatar {
    display: block;
    width: 5rem;
    height: 5rem;
  }

  .who {
    grid-area: who;
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0;
  }

  p {
    margin: 0;
  }

  .eyebrow {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
    overflow-wrap: anywhere;
  }

  .state {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    margin-top: var(--space-1);
  }

  .user {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .facts {
    grid-area: facts;
    display: flex;
    gap: var(--space-6);
    margin: 0;
  }

  @media (max-width: 43.99rem) {
    .facts {
      padding-top: var(--space-4);
      border-top: 1px solid var(--color-border);
    }
  }

  .facts div {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
    min-width: 0;
  }

  dt {
    color: var(--color-text-muted);
    font-size: var(--text-xs);
  }

  dd {
    margin: 0;
    font-weight: var(--weight-semibold);
    font-variant-numeric: tabular-nums;
  }

  .status {
    max-width: 14rem;
    overflow: hidden;
    font-weight: var(--weight-medium);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .invite {
    grid-area: invite;
    justify-self: start;
  }

  .invite svg {
    width: 1rem;
    height: 1rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
  }
</style>
