<!--
  Home: today's date and "Summary" (no greeting: decision #57). Wide
  screens: a main column (today's plan, the weekly weight check-in when
  due, recent workouts, the guide) beside this week and the body's
  muscle map (#49). Phones: one column, the week and body right after
  the check-in.
-->
<script lang="ts">
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { isoDate, isoWeek, startOfWeek, weekDays } from '@home-tools/ui/dates'
  import { getRecovery, getStrength, type MuscleRecovery, type MuscleStrength } from '@/features/body'
  import { BodyCard } from '@/features/body/body-card'
  import { getWeights, type FitnessProfile, type WeightEntry } from '@/features/fitness-profile'
  import { ThisWeek } from '@/features/home/this-week'
  import { RecentWorkouts } from '@/features/home/recent-workouts'
  import { UpNext } from '@/features/home/up-next'
  import { WeightCheckIn } from '@/features/home/weight-check-in'
  import { Card } from '@home-tools/ui/components/card'
  import { PageHeader } from '@home-tools/ui/components/page-header'
  import { Skeleton, SkeletonGroup } from '@home-tools/ui/components/skeleton'
  import type { Profile } from '@home-tools/ui/profiles/types'
  import { getPlans, type Plan } from '@/features/plans'
  import { getRecentSessions, inProgress, type Session } from '@/features/sessions'

  let { profile, fitness = $bindable() }: { profile: Profile; fitness: FitnessProfile } = $props()

  const now = new Date()
  const date = now.toLocaleDateString(undefined, { weekday: 'long', month: 'long', day: 'numeric' })

  let sessions = $state<Session[]>([])
  let plans = $state<Plan[]>([])
  let weights = $state<WeightEntry[]>([])
  let recovery = $state<MuscleRecovery[]>([])
  let strength = $state<MuscleStrength[]>([])
  let status = $state<'loading' | 'ready' | 'error'>('loading')

  async function load() {
    status = 'loading'
    try {
      // Independent requests: fetch in parallel.
      let s
      ;[sessions, plans, weights, recovery, s] = await Promise.all([
        getRecentSessions(profile.id),
        getPlans(),
        getWeights(profile.id),
        getRecovery(profile.id),
        getStrength(profile.id),
      ])
      strength = s.muscles
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

<div class="page">
  <PageHeader title="Summary">
    {#snippet eyebrow()}<time datetime={isoDate(now)}>{date}</time>{/snippet}
  </PageHeader>

  {#if status === 'error'}
    <LoadError what="your week" onretry={load} />
  {:else if status === 'loading'}
    <SkeletonGroup label="your week">
      <div class="grid">
        <div class="main">
          <Skeleton height="18rem" radius="lg" />
          <Skeleton height="12rem" radius="lg" />
        </div>
        <div class="side">
          <Skeleton height="8rem" radius="lg" />
          <Skeleton height="18rem" radius="lg" />
        </div>
      </div>
    </SkeletonGroup>
  {:else}
    <div class="grid">
      <div class="main">
        <div class="today">
          <UpNext {profile} {fitness} sessions={thisWeek} {plans} {recovery} active={inProgress(sessions)}
            ondiscarded={(id) => (sessions = sessions.filter((s) => s.id !== id))}
          />
        </div>
        {#if checkInDue}
          <div class="weight">
            <WeightCheckIn {profile} bind:fitness {weights} onsaved={(w) => (weights = [...weights, w])} />
          </div>
        {/if}
        {#if sessions.some((s) => s.ended_at)}
          <div class="recent"><Card padding="lg"><RecentWorkouts {sessions} {plans} /></Card></div>
        {/if}
        <div class="guide-link">
          <Card as="a" href="/guide" padding="md">
            <span class="guide">
              <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 5a2 2 0 0 1 2-2h13v16H6a2 2 0 0 0-2 2z M4 21V5 M8 7h7 M8 11h5" /></svg>
              <span class="guide-text">
                <strong>How training works</strong>
                <span class="guide-hint">Effort, reps, rest and progress, in five minutes</span>
              </span>
              <svg class="chevron" viewBox="0 0 24 24" aria-hidden="true"><path d="M9 6l6 6-6 6" /></svg>
            </span>
          </Card>
        </div>
      </div>
      <div class="side">
        <div class="week"><Card padding="lg"><ThisWeek {profile} {fitness} sessions={thisWeek} {plans} /></Card></div>
        <div class="body"><Card padding="lg"><BodyCard {recovery} {strength} /></Card></div>
      </div>
    </div>
  {/if}
</div>

<style>
  .page {
    --page-gap: var(--space-4);
  }

  /* Phones: one column, in reading order (the column wrappers step aside
   * so their cards can be ordered): today, check-in, week, body, then the
   * rest. Wide screens: Today's column (the main thing) wider, the week
   * and body beside it. */
  .grid {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .main,
  .side {
    display: contents;
  }

  .today {
    order: 1;
  }

  .weight {
    order: 2;
  }

  .week {
    order: 3;
  }

  .body {
    order: 4;
  }

  .recent {
    order: 5;
  }

  .guide-link {
    order: 6;
  }

  @media (min-width: 56rem) {
    .grid {
      display: grid;
      grid-template-columns: 3fr 2fr;
      align-items: start;
      gap: var(--space-5);
    }

    .main,
    .side {
      display: flex;
      flex-direction: column;
      gap: var(--space-4);
    }
  }

  .guide {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .guide svg {
    flex: none;
    width: 1.25rem;
    height: 1.25rem;
    fill: none;
    stroke: var(--color-text-muted);
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .guide-text {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.125rem;
    color: var(--color-text);
    font-size: var(--text-sm);
  }

  .guide-text strong {
    font-weight: var(--weight-semibold);
  }

  .guide-hint {
    color: var(--color-text-muted);
  }
</style>
