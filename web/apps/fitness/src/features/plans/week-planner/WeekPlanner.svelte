<!--
  "Your routine": one row per weekday, each with a plan picker (or Rest).
  Saves on change into this person's fitness profile (decision #27, #30);
  a live region announces "Saved" so it's clear nothing else is needed.
-->
<script lang="ts">
  import { getCatalog, type Exercise } from '@/features/exercises'
  import { saveFitnessProfile, type FitnessProfile, type Weekday } from '@/features/fitness-profile'
  import type { Plan } from '@/features/plans'
  import { weekChecks, weekLoad } from '@/features/training'
  import { GOAL_KEYS, GOALS, goalSummary } from '@/features/training/guidance'
  import type { Goal } from '@/features/fitness-profile'
  import { LoadSummary } from '@/features/training/load-summary'

  let { fitness = $bindable(), plans }: { fitness: FitnessProfile; plans: Plan[] } = $props()

  const DAYS: { key: Weekday; label: string }[] = [
    { key: 'monday', label: 'Monday' },
    { key: 'tuesday', label: 'Tuesday' },
    { key: 'wednesday', label: 'Wednesday' },
    { key: 'thursday', label: 'Thursday' },
    { key: 'friday', label: 'Friday' },
    { key: 'saturday', label: 'Saturday' },
    { key: 'sunday', label: 'Sunday' },
  ]

  let status = $state('')
  let hasRoutine = $derived(Object.keys(fitness.schedule ?? {}).length > 0)

  // The week's muscles (#51): needs the catalog; until it's here (or if
  // it fails) the planner simply works without the summary.
  let catalog = $state<Exercise[] | null>(null)
  getCatalog()
    .then((c) => (catalog = c))
    .catch((err) => console.error('Loading the catalog for the week summary failed:', err))
  let week = $derived(catalog ? weekLoad(fitness.schedule, plans, new Map(catalog.map((e) => [e.id, e]))) : null)

  // Archived plans can't be newly picked, but stay listed on the days
  // that already use them, so the select doesn't silently show "Rest".
  const choicesFor = (day: Weekday) =>
    plans.filter((r) => !r.archived || r.id === fitness.schedule?.[day])

  async function changeGoal(goal: Goal) {
    status = 'Saving…'
    try {
      fitness = await saveFitnessProfile({ ...fitness, goal })
      status = 'Saved'
    } catch (err) {
      status = `Couldn't save. ${(err as Error).message}`
    }
  }

  async function change(day: Weekday, planId: string) {
    const schedule = { ...fitness.schedule }
    if (planId) schedule[day] = planId
    else delete schedule[day] // missing = rest day
    status = 'Saving…'
    try {
      fitness = await saveFitnessProfile({ ...fitness, schedule })
      status = 'Saved'
    } catch (err) {
      status = `Couldn't save. ${(err as Error).message}`
    }
  }
</script>

<section class="week" aria-labelledby="week-title">
  <div class="head">
    <h2 id="week-title">Your routine</h2>
    <p class="status" role="status">{status}</p>
  </div>

  <div class="goal">
    <label for="goal">Your goal</label>
    <select id="goal" value={fitness.goal ?? 'muscle'} onchange={(e) => changeGoal(e.currentTarget.value as Goal)} aria-describedby="goal-hint">
      {#each GOAL_KEYS as key (key)}
        <option value={key}>{GOALS[key].label}</option>
      {/each}
    </select>
    <p id="goal-hint" class="goal-hint">
      {goalSummary(fitness.goal)}. {GOALS[fitness.goal ?? 'muscle'].note} <a href="/guide#reps">More</a>
    </p>
  </div>

  <ul class="days">
    {#each DAYS as day (day.key)}
      <li class="day">
        <label for="day-{day.key}">{day.label}</label>
        <select
          id="day-{day.key}"
          value={fitness.schedule?.[day.key] ?? ''}
          onchange={(e) => change(day.key, e.currentTarget.value)}
          class:rest={!fitness.schedule?.[day.key]}
        >
          <option value="">Rest</option>
          {#each choicesFor(day.key) as r (r.id)}
            <option value={r.id}>{r.name}</option>
          {/each}
        </select>
      </li>
    {/each}
  </ul>

  <a class="btn {hasRoutine ? 'btn-quiet' : 'btn-primary'} builder" href="/plans/build">
    {hasRoutine ? 'Rebuild it for me' : 'Build a routine for me'}
  </a>

  {#if week && week.days.size > 0}
    <section class="works" aria-labelledby="week-works-title">
      <h3 id="week-works-title">What your week works</h3>
      <LoadSummary load={week.total} checks={weekChecks(week, fitness.goal)} mode="week" />
    </section>
  {/if}
</section>

<style>
  .week {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .works {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    margin-top: var(--space-4);
    padding: var(--space-5);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
  }

  .goal {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .goal label {
    font-weight: var(--weight-semibold);
  }

  .goal select {
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-surface);
    font-weight: var(--weight-medium);
  }

  .goal-hint {
    margin: 0;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    line-height: var(--leading-normal);
  }

  .goal-hint a {
    color: var(--color-accent-text);
    font-weight: var(--weight-semibold);
  }

  .builder {
    align-self: flex-start;
  }

  h3 {
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
  }

  .head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-3);
  }

  h2 {
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
  }

  .status {
    margin: 0;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .days {
    display: flex;
    flex-direction: column;
    margin: 0;
    padding: 0;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
    list-style: none;
  }

  .day {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-2) var(--space-2) var(--space-4);
  }

  .day + .day {
    border-top: 1px solid var(--color-border);
  }

  label {
    font-weight: var(--weight-semibold);
  }

  select {
    min-width: 0;
    max-width: 60%;
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-bg);
    font-weight: var(--weight-medium);
    cursor: pointer;
  }

  select.rest {
    color: var(--color-text-muted);
  }
</style>
