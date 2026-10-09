<!--
  Progress (decision #39): a few numbers, the weight journey, the
  exercises you've logged (each opens its journey), and recent workouts.
  Replaces the old History page; its list lives on at the bottom.
-->
<script lang="ts">
  import { LoadError } from '@home-tools/ui/components/load-error'
  import type { Profile } from '@home-tools/ui/profiles/types'
  import { getCatalog, type Exercise } from '@/features/exercises'
  import { getWeights, type WeightEntry } from '@/features/fitness-profile'
  import { getPlans, type Plan } from '@/features/plans'
  import { getExerciseLog, weekStreak, workoutsThisMonth, type LoggedExercise } from '@/features/progress'
  import { WeightJourney } from '@/features/progress/weight-journey'
  import { getRecentSessions, type Session } from '@/features/sessions'

  let { profile }: { profile: Profile } = $props()

  const FIRST_EXERCISES = 6
  const FIRST_WORKOUTS = 10

  let sessions = $state<Session[]>([])
  let plans = $state<Plan[]>([])
  let weights = $state<WeightEntry[]>([])
  let log = $state<LoggedExercise[]>([])
  let catalog = $state<Exercise[]>([])
  let status = $state<'loading' | 'ready' | 'error'>('loading')
  let allExercises = $state(false)
  let allWorkouts = $state(false)

  async function load() {
    status = 'loading'
    try {
      ;[sessions, plans, weights, log, catalog] = await Promise.all([
        getRecentSessions(profile.id),
        getPlans(),
        getWeights(profile.id),
        getExerciseLog(profile.id),
        getCatalog(),
      ])
      status = 'ready'
    } catch (err) {
      console.error('Loading progress failed:', err)
      status = 'error'
    }
  }
  load()

  let units = $derived({ imperial: profile.units === 'imperial' })
  let names = $derived(new Map(catalog.map((e) => [e.id, e.name])))
  let streak = $derived(weekStreak(sessions))
  let month = $derived(workoutsThisMonth(sessions))
  let exercises = $derived(allExercises ? log : log.slice(0, FIRST_EXERCISES))
  let workouts = $derived(allWorkouts ? sessions : sessions.slice(0, FIRST_WORKOUTS))

  const title = (s: Session) => plans.find((r) => r.id === s.plan_id)?.name ?? 'Workout'
  const when = (iso: string) =>
    new Date(iso).toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })
  const minutes = (s: Session) =>
    s.ended_at ? Math.round((Date.parse(s.ended_at) - Date.parse(s.started_at)) / 60_000) : null
  const plural = (n: number, one: string) => `${n} ${one}${n === 1 ? '' : 's'}`
</script>

<div class="progress">
  <h1 tabindex="-1">Progress</h1>

  {#if status === 'error'}
    <LoadError what="your progress" onretry={load} />
  {:else if status === 'ready'}
    <dl class="stats">
      <div>
        <dt>This month</dt>
        <dd>{month} <span class="unit">{month === 1 ? 'workout' : 'workouts'}</span></dd>
      </div>
      <div>
        <dt>Streak</dt>
        <dd>{streak} <span class="unit">{streak === 1 ? 'week' : 'weeks'}</span></dd>
      </div>
      <div>
        <dt>Exercises</dt>
        <dd>{log.length} <span class="unit">logged</span></dd>
      </div>
    </dl>

    <WeightJourney {weights} {units} />

    <section aria-labelledby="exercises-title">
      <h2 id="exercises-title">Exercises</h2>
      {#if log.length === 0}
        <p class="muted">Finish a workout and your exercises show up here, each with its own progress.</p>
      {:else}
        <ul class="list">
          {#each exercises as e (e.exercise_id)}
            <li>
              <a class="row link" href="/progress/{encodeURIComponent(e.exercise_id)}">
                <span class="text">
                  <span class="name">{names.get(e.exercise_id) ?? e.exercise_id}</span>
                  <span class="muted">
                    {plural(e.workouts, 'workout')} · last <time datetime={e.last_done}>{when(e.last_done)}</time>
                  </span>
                </span>
                <svg class="chevron" viewBox="0 0 24 24" aria-hidden="true"><path d="M9 6l6 6-6 6" /></svg>
              </a>
            </li>
          {/each}
        </ul>
        {#if log.length > FIRST_EXERCISES}
          <button type="button" class="btn btn-quiet more" onclick={() => (allExercises = !allExercises)}>
            {allExercises ? 'Show fewer' : `Show all ${log.length}`}
          </button>
        {/if}
      {/if}
    </section>

    <section aria-labelledby="workouts-title">
      <h2 id="workouts-title">Recent workouts</h2>
      {#if sessions.length === 0}
        <p class="muted">No workouts logged yet.</p>
      {:else}
        <ul class="list">
          {#each workouts as s (s.id)}
            <li class="row">
              <span class="text">
                <span class="name">{title(s)}</span>
                <span class="muted">
                  <time datetime={s.started_at}>{when(s.started_at)}</time>
                  · {plural(s.entries.length, 'exercise')}
                  {#if minutes(s) !== null}· {minutes(s)} min{:else}· in progress{/if}
                </span>
              </span>
            </li>
          {/each}
        </ul>
        {#if sessions.length > FIRST_WORKOUTS}
          <button type="button" class="btn btn-quiet more" onclick={() => (allWorkouts = !allWorkouts)}>
            {allWorkouts ? 'Show fewer' : 'Show more'}
          </button>
        {/if}
      {/if}
    </section>
  {/if}
</div>

<style>
  .progress {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
  }

  h2 {
    margin-bottom: var(--space-3);
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
  }

  .stats {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: var(--space-2);
    margin: 0;
  }

  .stats div {
    display: flex;
    flex-direction: column-reverse; /* number on top, label under it; dt stays first in the markup */
    gap: var(--space-1);
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-md);
    background: var(--color-accent-subtle);
  }

  dt {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
  }

  dd {
    margin: 0;
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  .unit {
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    min-height: var(--touch-target);
    padding: var(--space-3) var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
  }

  .text {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0;
  }

  .name {
    font-weight: var(--weight-semibold);
    overflow-wrap: anywhere;
  }

  .link {
    color: inherit;
    text-decoration: none;
    transition: border-color var(--duration-fast) var(--ease-out);
  }

  .link:active {
    background: var(--color-accent-subtle);
  }

  @media (hover: hover) {
    .link:hover {
      border-color: var(--color-border-strong);
    }
  }

  .chevron {
    flex: none;
    width: 1.25rem;
    height: 1.25rem;
    fill: none;
    stroke: var(--color-text-muted);
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .more {
    margin-top: var(--space-2);
  }
</style>
