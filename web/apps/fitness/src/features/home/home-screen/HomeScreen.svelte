<!--
  Home: a greeting with the person's blob, today's date under it, then the
  weekly weight check-in (only when due), today's plan, and this week.
-->
<script lang="ts">
  import { isoDate, isoWeek, startOfWeek, weekDays } from '@/dates'
  import { getWeights, type FitnessProfile, type WeightEntry } from '@/features/fitness-profile'
  import { ThisWeek } from '@/features/home/this-week'
  import { UpNext } from '@/features/home/up-next'
  import { WeightCheckIn } from '@/features/home/weight-check-in'
  import { ProfileAvatar } from '@/features/profiles/profile-avatar'
  import type { Profile } from '@/features/profiles/types'
  import { getPlans, type Plan } from '@/features/plans'
  import { getRecentSessions, inProgress, type Session } from '@/features/sessions'

  let { profile, fitness = $bindable() }: { profile: Profile; fitness: FitnessProfile } = $props()

  const now = new Date()
  const date = now.toLocaleDateString(undefined, { weekday: 'long', month: 'long', day: 'numeric' })

  // 5–11 morning, 12–16 afternoon, otherwise evening ("Good morning" at
  // 3 a.m. reads oddly, so late night counts as evening).
  const hour = now.getHours()
  const greeting = hour >= 5 && hour < 12 ? 'Good morning' : hour >= 12 && hour < 17 ? 'Good afternoon' : 'Good evening'

  let sessions = $state<Session[]>([])
  let plans = $state<Plan[]>([])
  let weights = $state<WeightEntry[]>([])
  let status = $state<'loading' | 'ready' | 'error'>('loading')

  async function load() {
    status = 'loading'
    try {
      // Independent requests: fetch in parallel.
      ;[sessions, plans, weights] = await Promise.all([
        getRecentSessions(profile.id),
        getPlans(),
        getWeights(profile.id),
      ])
      status = 'ready'
    } catch (err) {
      console.error('Loading home failed:', err)
      status = 'error'
    }
  }
  load()

  // Only this week's sessions matter here.
  let weekStart = isoDate(startOfWeek(now))
  let weekEnd = isoDate(weekDays(now)[6])
  let thisWeek = $derived(
    sessions.filter((s) => {
      const day = isoDate(new Date(s.started_at))
      return day >= weekStart && day <= weekEnd
    }),
  )

  // The check-in is due if nothing was logged this week and it wasn't skipped.
  let checkInDue = $derived(
    fitness.weight_prompt_skipped !== isoWeek(now) &&
      !weights.some((w) => w.date >= weekStart && w.date <= weekEnd),
  )
</script>

<div class="home">
  <div class="greeting">
    <span class="avatar"><ProfileAvatar id={profile.id} /></span>
    <div>
      <h1 tabindex="-1">{greeting}, {profile.name}</h1>
      <p class="date"><time datetime={isoDate(now)}>{date}</time></p>
    </div>
  </div>

  {#if status === 'error'}
    <div class="error" role="alert">
      <p>Couldn't load your week. Check that the server is running.</p>
      <button type="button" class="btn btn-primary" onclick={load}>Try again</button>
    </div>
  {:else if status === 'ready'}
    {#if checkInDue}
      <WeightCheckIn {profile} bind:fitness {weights} onsaved={(w) => (weights = [...weights, w])} />
    {/if}
    <div class="grid">
      <UpNext {profile} {fitness} sessions={thisWeek} {plans} active={inProgress(sessions)}
        ondiscarded={(id) => (sessions = sessions.filter((s) => s.id !== id))}
      />
      <ThisWeek {profile} {fitness} sessions={thisWeek} {plans} />
    </div>
  {/if}
</div>

<style>
  .home {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  /* Phones: one column. Wide screens: Today (the hero) wider, the week
   * beside it. */
  .grid {
    display: grid;
    gap: var(--space-6);
    align-items: start;
  }

  @media (min-width: 56rem) {
    .grid {
      grid-template-columns: 3fr 2fr;
      gap: var(--space-7);
    }
  }

  .greeting {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    margin-bottom: var(--space-3);
  }

  .avatar {
    flex: none;
    width: 4rem;
    height: 4rem;
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
  }

  .date {
    margin: var(--space-1) 0 0;
    color: var(--color-text-muted);
    font-weight: var(--weight-medium);
  }

  @media (min-width: 40rem) {
    .avatar {
      width: 5rem;
      height: 5rem;
    }

    h1 {
      font-size: var(--text-3xl);
    }
  }

  .error {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-3);
  }

  .error p {
    margin: 0;
  }
</style>
