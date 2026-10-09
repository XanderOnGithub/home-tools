<script lang="ts">
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { getFitnessProfile, type FitnessProfile } from '@/features/fitness-profile'
  import { HistoryScreen } from '@/features/history/history-screen'
  import { HomeScreen } from '@/features/home/home-screen'
  import { OnboardingFlow } from '@/features/onboarding/onboarding-flow'
  import { ProfilePicker } from '@home-tools/ui/profiles/profile-picker'
  import { forgetProfile } from '@home-tools/ui/profiles/remembered'
  import type { Profile } from '@home-tools/ui/profiles/types'
  import { PlanEditor } from '@/features/plans/plan-editor'
  import { PlansScreen } from '@/features/plans/plans-screen'
  import { AppShell } from '@home-tools/ui/app-shell'
  import { NotFound } from '@/features/shell/not-found'
  import { WorkoutMode } from '@/features/workout/workout-mode'
  import { router } from '@home-tools/ui/router'

  const LINKS = [
    { href: '/', label: 'Home', icon: 'M3 11l9-8 9 8 M5 9.5V20h5v-6h4v6h5V9.5' },
    { href: '/plans', label: 'Plans', icon: 'M8 6h12 M8 12h12 M8 18h12 M4 6h.01 M4 12h.01 M4 18h.01' },
    { href: '/history', label: 'History', icon: 'M12 7v5l3 2 M3.05 11a9 9 0 1 1 .5 4 M3 4v5h5' },
  ]

  // Gates, in order:
  //   no profile chosen      → picker
  //   chosen, not onboarded  → onboarding
  //   chosen and onboarded   → the app (pages chosen by URL, see router)
  let profile = $state<Profile | null>(null)
  let fitness = $state<FitnessProfile | null>(null)
  let status = $state<'idle' | 'loading' | 'ready' | 'error'>('idle')

  // /plans/new → editor for a new plan; /plans/<id> → edit it.
  let planId = $derived(router.path.match(/^\/plans\/([^/]+)$/)?.[1] ?? null)
  // /workout/<session id> → workout mode (full screen, no navigation).
  let workoutId = $derived(router.path.match(/^\/workout\/([^/]+)$/)?.[1] ?? null)

  // "Page · Fitness" (most specific first: tabs cut off the end). Home is
  // the tool's front page, so it names the whole set instead.
  const TITLES: Record<string, string> = { '/plans': 'Plans', '/history': 'History' }
  $effect(() => {
    if (router.path === '/') {
      document.title = 'Fitness · Home Tools'
      return
    }
    const title = workoutId
      ? 'Workout'
      : planId
        ? planId === 'new'
          ? 'New plan'
          : 'Edit plan'
        : TITLES[router.path]
    document.title = `${title ?? 'Not found'} · Fitness`
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
    <ProfilePicker onselect={choose} title="Who's working out?" />
  </main>
{:else if status === 'error'}
  <main class="center">
    <LoadError what="your fitness profile" onretry={() => profile && choose(profile)} align="center">
      <button type="button" class="btn btn-quiet" onclick={switchProfile}>Switch profile</button>
    </LoadError>
  </main>
{:else if status === 'ready' && !fitness}
  <OnboardingFlow {profile} oncomplete={(fp) => (fitness = fp)} />
{:else if status === 'ready' && fitness && workoutId}
  {#key workoutId}
    <WorkoutMode {profile} sessionId={decodeURIComponent(workoutId)} />
  {/key}
{:else if status === 'ready' && fitness}
  <AppShell links={LINKS} {profile} onswitch={switchProfile}>
    {#if router.path === '/'}
      <HomeScreen {profile} bind:fitness />
    {:else if router.path === '/plans'}
      <PlansScreen bind:fitness />
    {:else if planId}
      {#key planId}
        <PlanEditor {profile} id={planId === 'new' ? null : decodeURIComponent(planId)} />
      {/key}
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
</style>
