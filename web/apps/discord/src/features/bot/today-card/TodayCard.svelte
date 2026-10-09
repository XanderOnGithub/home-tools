<!--
  Who the bot is today (decision #42): its avatar, exactly as Discord
  gets it (the animated GIF; the still image with reduced motion), its
  name, and whether it's connected. The invite link adds it to a Discord
  server with only the permissions it needs.
-->
<script lang="ts">
  import { PERSONA_GIF, PERSONA_PNG, type BotView } from '@/features/bot'

  let { bot }: { bot: BotView } = $props()

  // Changes with the persona, so the browser fetches the new day's image.
  let version = $derived(`${bot.persona.date}-${bot.persona.color}-${encodeURIComponent(bot.persona.name)}`)
  let guilds = $derived(bot.guilds.length)
</script>

<section class="today" aria-labelledby="today-name">
  <picture>
    <source srcset="{PERSONA_PNG}?v={version}" media="(prefers-reduced-motion: reduce)" />
    <img class="avatar" src="{PERSONA_GIF}?v={version}" alt="" width="96" height="96" />
  </picture>
  <div class="body">
    <p class="eyebrow">Today the bot is</p>
    <h2 id="today-name">{bot.persona.name}</h2>
    <p class="muted">A new name and color every day at midnight.</p>
    <p class="status" class:on={bot.connected}>
      <span class="dot" aria-hidden="true"></span>
      {#if bot.connected}
        Connected as {bot.user} · in {guilds} {guilds === 1 ? 'server' : 'servers'}
      {:else}
        Not connected: {bot.error ?? 'unknown reason'}
      {/if}
    </p>
    {#if bot.invite_url}
      <a class="btn btn-secondary invite" href={bot.invite_url} target="_blank" rel="noreferrer">
        Add to a Discord server
      </a>
    {/if}
  </div>
</section>

<style>
  .today {
    display: grid;
    grid-template-columns: auto 1fr;
    align-items: start;
    gap: var(--space-4);
    padding: var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
  }

  .avatar {
    display: block;
    width: 6rem;
    height: 6rem;
  }

  .body {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0;
  }

  p {
    margin: 0;
  }

  .eyebrow,
  .muted {
    color: var(--color-text-muted);
    font-weight: var(--weight-medium);
  }

  .eyebrow {
    font-size: var(--text-sm);
  }

  h2 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
    overflow-wrap: anywhere;
  }

  .status {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-top: var(--space-2);
    font-weight: var(--weight-semibold);
  }

  .status:not(.on) {
    color: var(--color-danger-text);
  }

  /* Same dot as the games tool: filled when connected, hollow when not
   * (the text says it too). */
  .dot {
    flex: none;
    width: 0.7rem;
    height: 0.7rem;
    border: 2px solid currentColor;
    border-radius: var(--radius-full);
  }

  .on .dot {
    border-color: var(--color-accent);
    background: var(--color-accent);
  }

  .invite {
    align-self: flex-start;
    margin-top: var(--space-3);
  }
</style>
