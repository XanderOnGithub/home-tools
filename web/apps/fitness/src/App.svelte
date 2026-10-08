<script lang="ts">
  import { getFitnessProfile, type FitnessProfile } from '@/features/fitness-profile'
  import { HistoryScreen } from '@/features/history/history-screen'
  import { HomeScreen } from '@/features/home/home-screen'
  import { OnboardingFlow } from '@/features/onboarding/onboarding-flow'
  import { ProfilePicker } from '@/features/profiles/profile-picker'
  import { forgetProfile } from '@/features/profiles/remembered'
  import type { Profile } from '@/features/profiles/types'
  import { RoutinesScreen } from '@/features/routines/routines-screen'
  import { AppShell } from '@/features/shell/app-shell'
  import { NotFound } from '@/features/shell/not-found'
  import { router } from '@/router'

  // Gates, in order:
  //   no profile chosen      → picker
  //   chosen, not onboarded  → onboarding
  //   chosen and onboarded   → the app (pages chosen by URL, see router)
  let profile = $state<Profile | null>(null)
  let fitness = $state<FitnessProfile | null>(null)
  let status = $state<'idle' | 'loading' | 'ready' | 'error'>('idle')

  const TITLES: Record<string, string> = { '/': 'Home', '/routines': 'Routines', '/history': 'History' }
  $effect(() => {
    document.title = `${TITLES[router.path] ?? 'Not found'} · Fitness`
  })

  // The chosen person's color becomes the accent for the whole app.
  $effect(() => {
    if (profile) document.documentElement.dataset.accent = profile.color
    else delete document.documentElement.dataset.accent
  })

  async function choose(p: Profile) {
    profile = p
    fitness = null
    status = 'loading'
    try {
      fitness = await getFitnessProfile(p.id) // null = not onboarded
      status = 'ready'
    } catch (err) {
      console.error('Loading fitness profile failed:', err)
      status = 'error'
    }
  }

  function switchProfile() {
    forgetProfile()
    profile = null
    fitness = null
    status = 'idle'
  }
</script>

{#if !profile}
  <main>
    <ProfilePicker onselect={choose} />
  </main>
{:else if status === 'error'}
  <main class="center">
    <p role="alert">Couldn't load your fitness profile. Check that the server is running.</p>
    <button type="button" class="retry" onclick={() => profile && choose(profile)}>Try again</button>
    <button type="button" class="retry quiet" onclick={switchProfile}>Switch profile</button>
  </main>
{:else if status === 'ready' && !fitness}
  <OnboardingFlow {profile} oncomplete={(fp) => (fitness = fp)} />
{:else if status === 'ready' && fitness}
  <AppShell {profile} onswitch={switchProfile}>
    {#if router.path === '/'}
      <HomeScreen {profile} bind:fitness />
    {:else if router.path === '/routines'}
      <RoutinesScreen />
    {:else if router.path === '/history'}
      <HistoryScreen {profile} />
    {:else}
      <NotFound />
    {/if}
  </AppShell>
{/if}

<style>
  .center {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-3);
    min-height: 100dvh;
    padding: var(--space-4);
    text-align: center;
  }

  .retry {
    min-height: var(--touch-target);
    padding: 0 var(--space-5);
    border: none;
    border-radius: var(--radius-full);
    background: var(--color-accent);
    color: var(--color-on-accent);
    font-weight: var(--weight-semibold);
    cursor: pointer;
  }

  .retry.quiet {
    background: none;
    color: var(--color-text-muted);
  }
</style>
