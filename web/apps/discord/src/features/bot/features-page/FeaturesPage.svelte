<!--
  The bot's optional features (decision #44), each with its own switch:
  off = its commands leave Discord and its jobs stop, at once.
-->
<script lang="ts">
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { PageHeader } from '@home-tools/ui/components/page-header'
  import { SettingsSkeleton } from '@/components/settings-skeleton'
  import { BlobMaker } from '@/features/bot/blob-maker'
  import { PollSettings } from '@/features/bot/poll-settings'
  import { settings } from '@/features/bot/settings'
  import { StatusSettings } from '@/features/bot/status-settings'

  const s = settings()
</script>

<div class="page">
  <PageHeader title="Features" description="Turning one off removes its commands from Discord right away." />

  {#if s.status === 'error'}
    <LoadError what="the bot's settings" onretry={s.load} />
  {:else if s.config && s.bot}
    <PollSettings config={s.config} bot={s.bot} save={s.save} onposted={s.refreshBot} />
    <StatusSettings config={s.config} bot={s.bot} save={s.save} />
    <BlobMaker config={s.config} save={s.save} />
  {:else}
    <SettingsSkeleton sections={0} />
  {/if}
</div>

<style>
  .page {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    max-width: var(--content-medium);
  }

  /* The header's own bottom margin is the gap. */
  .page > :global(.page-header) {
    margin-bottom: var(--space-2);
  }
</style>
