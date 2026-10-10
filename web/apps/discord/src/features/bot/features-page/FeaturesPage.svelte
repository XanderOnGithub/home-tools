<!--
  The bot's optional features (decision #44), each with its own switch:
  off = its commands leave Discord and its jobs stop, at once.
-->
<script lang="ts">
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { BlobMaker } from '@/features/bot/blob-maker'
  import { PollSettings } from '@/features/bot/poll-settings'
  import { settings } from '@/features/bot/settings'
  import { StatusSettings } from '@/features/bot/status-settings'

  const s = settings()
</script>

<div class="page">
  <div class="intro">
    <h1 tabindex="-1">Features</h1>
    <p class="muted">Extras for the bot. Turning one off removes its commands from Discord right away.</p>
  </div>

  {#if s.status === 'error'}
    <LoadError what="the bot's settings" onretry={s.load} />
  {:else if s.config && s.bot}
    <PollSettings config={s.config} bot={s.bot} save={s.save} onposted={s.refreshBot} />
    <StatusSettings config={s.config} bot={s.bot} save={s.save} />
    <BlobMaker config={s.config} save={s.save} />
  {/if}
</div>

<style>
  .page {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  .intro {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
  }
</style>
