<!--
  Discord: pick a profile (shared with every tool, ADR 0007; the choice is
  remembered across subdomains), then the shared app shell (same header,
  nav and spacing as fitness and games) around the bot page ("/"), the
  features page ("/features") or the people page ("/people"). The profile gives the accent color.
-->
<script lang="ts">
  import { AppShell } from '@home-tools/ui/app-shell'
  import { ProfilePicker } from '@home-tools/ui/profiles/profile-picker'
  import { forgetProfile } from '@home-tools/ui/profiles/remembered'
  import type { Profile } from '@home-tools/ui/profiles/types'
  import { router } from '@home-tools/ui/router'
  import { FeaturesPage } from '@/features/bot/features-page'
  import { OverviewPage } from '@/features/bot/overview-page'
  import { PeoplePage } from '@/features/bot/people-page'

  const LINKS = [
    // A chat bubble with a smile: the bot.
    { href: '/', label: 'Bot', icon: 'M4 5h16v11H9l-5 4z M9 10.5h.01 M15 10.5h.01' },
    // A sparkle: the optional extras.
    { href: '/features', label: 'Features', icon: 'M12 3l2 6 6 2-6 2-2 6-2-6-6-2 6-2z M19 3v4 M21 5h-4' },
    // Two people.
    {
      href: '/people',
      label: 'People',
      icon: 'M9 11a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7z M2.5 20a6.5 6.5 0 0 1 13 0 M16 4.5a3.5 3.5 0 0 1 0 6.5 M18 14.5a6.5 6.5 0 0 1 3.5 5.5',
    },
  ]

  let profile = $state<Profile | null>(null)

  $effect(() => {
    if (profile) document.documentElement.dataset.accent = profile.color
    else delete document.documentElement.dataset.accent
  })

  $effect(() => {
    const titles: Record<string, string> = { '/people': 'People · Discord', '/features': 'Features · Discord' }
    document.title = titles[router.path] ?? 'Discord · Home Tools'
  })

  function switchProfile() {
    forgetProfile()
    profile = null
  }
</script>

{#if !profile}
  <main>
    <ProfilePicker onselect={(p) => (profile = p)} title="Who's this?" />
  </main>
{:else}
  <AppShell links={LINKS} {profile} onswitch={switchProfile}>
    {#if router.path === '/'}
      <OverviewPage />
    {:else if router.path === '/features'}
      <FeaturesPage />
    {:else if router.path === '/people'}
      <PeoplePage />
    {:else}
      <h1 tabindex="-1">Not found</h1>
      <p>There's no page here. <a href="/">The bot</a></p>
    {/if}
  </AppShell>
{/if}
