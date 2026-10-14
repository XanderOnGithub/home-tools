<!--
  Home of the Discord tool: who the bot is today and whether it's
  connected, the game statuses it keeps in Discord, the names it picks
  from, and what its commands do.
-->
<script lang="ts">
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { PageHeader } from '@home-tools/ui/components/page-header'
  import { SettingsSkeleton } from '@/components/settings-skeleton'
  import { CommandsHelp } from '@/features/bot/commands-help'
  import { NameList } from '@/features/bot/name-list'
  import { settings } from '@/features/bot/settings'
  import { StatusBoards } from '@/features/bot/status-boards'
  import { TodayCard } from '@/features/bot/today-card'

  const s = settings()
</script>

<div class="page">
  <PageHeader title="Bot" />

  {#if s.status === 'error'}
    <LoadError what="the bot's settings" onretry={s.load} />
  {:else if s.config && s.bot}
    <TodayCard bot={s.bot} />
    <div class="sections">
      <StatusBoards config={s.config} bot={s.bot} save={s.save} />
      <NameList config={s.config} save={s.save} />
      <CommandsHelp features={s.config.features} />
    </div>
  {:else}
    <SettingsSkeleton sections={3} />
  {/if}
</div>

<style>
  .page {
    max-width: var(--content-medium);
  }

  .sections {
    display: flex;
    flex-direction: column;
    gap: var(--space-7);
    margin-top: var(--space-7);
  }
</style>
