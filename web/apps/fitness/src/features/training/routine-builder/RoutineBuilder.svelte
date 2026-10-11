<!--
  "Build a routine for me" (/plans/build, decision #53): three questions
  (days a week, equipment, goal), then a live preview of the week (which
  plan on which day, its exercises, what the week works) and one button
  that makes it real: the starter plans are saved as normal shared plans
  (reused if the household already has them) and become this person's
  routine. Replacing an existing routine asks first; their plans stay.
-->
<script lang="ts">
  import { LoadError } from '@home-tools/ui/components/load-error'
  import type { Profile } from '@home-tools/ui/profiles/types'
  import { router } from '@home-tools/ui/router'
  import { getCatalog, type Exercise } from '@/features/exercises'
  import { saveFitnessProfile, type FitnessProfile, type Weekday } from '@/features/fitness-profile'
  import { getPlans, savePlan, type Plan } from '@/features/plans'
  import { WEEKDAYS, weekChecks, weekLoad } from '@/features/training'
  import { LoadSummary } from '@/features/training/load-summary'
  import { buildRoutine, EQUIPMENT_LABELS, GOAL_LABELS, type Days, type Equipment, type Goal } from './templates'

  let { profile, fitness = $bindable() }: { profile: Profile; fitness: FitnessProfile } = $props()

  let days = $state<Days>(3)
  let equipment = $state<Equipment>('gym')
  let goal = $state<Goal>(fitness.goal === 'strength' || fitness.goal === 'general' ? fitness.goal : 'muscle')

  let catalog = $state<Exercise[]>([])
  let existing = $state<Plan[]>([])
  let status = $state<'loading' | 'ready' | 'error'>('loading')
  let saving = $state(false)
  let error = $state('')
  let confirming = $state(false)

  async function load() {
    status = 'loading'
    try {
      ;[catalog, existing] = await Promise.all([getCatalog(), getPlans()])
      status = 'ready'
    } catch (err) {
      console.error('Loading the routine builder failed:', err)
      status = 'error'
    }
  }
  load()

  let byId = $derived(new Map(catalog.map((e) => [e.id, e])))
  let routine = $derived(buildRoutine(days, equipment, goal, profile.id))
  let week = $derived(weekLoad(routine.schedule, routine.plans, byId))
  let hasRoutine = $derived(Object.keys(fitness.schedule ?? {}).length > 0)

  const DAY_NAMES: Record<Weekday, string> = {
    monday: 'Monday', tuesday: 'Tuesday', wednesday: 'Wednesday', thursday: 'Thursday',
    friday: 'Friday', saturday: 'Saturday', sunday: 'Sunday',
  }
  const planFor = (day: Weekday) => routine.plans.find((p) => p.id === routine.schedule[day])

  function use() {
    if (hasRoutine && !confirming) {
      confirming = true
      return
    }
    save()
  }

  async function save() {
    saving = true
    error = ''
    try {
      // Reuse plans the household already has (someone made the same
      // choices, maybe edited since); create the rest.
      const have = new Map(existing.map((p) => [p.id, p]))
      for (const plan of routine.plans) {
        const current = have.get(plan.id)
        if (!current) await savePlan(plan)
        else if (current.archived) await savePlan({ ...current, archived: false })
      }
      fitness = await saveFitnessProfile({ ...fitness, schedule: routine.schedule, goal })
      router.navigate('/plans')
    } catch (err) {
      error = `Couldn't save the routine. ${(err as Error).message}`
      saving = false
    }
  }
</script>

<div class="page">
  <a class="back" href="/plans">
    <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M15 6l-6 6 6 6" /></svg>
    Plans
  </a>

  <div class="intro">
    <h1 tabindex="-1">Build a routine</h1>
    <p class="muted">Three questions, and you get a week of workouts to start with. They're normal plans: change anything later.</p>
  </div>

  {#if status === 'error'}
    <LoadError what="the exercises" onretry={load} />
  {:else if status === 'ready'}
    <div class="questions">
      <fieldset>
        <legend>How many days a week?</legend>
        <div class="options">
          {#each [2, 3, 4] as const as n (n)}
            <label class="option">
              <input type="radio" name="days" value={n} bind:group={days} />
              <span class="option-label">{n} days</span>
              <span class="option-hint">{n === 4 ? 'Upper and lower body' : 'Full body'}</span>
            </label>
          {/each}
        </div>
      </fieldset>

      <fieldset>
        <legend>What do you have?</legend>
        <div class="options">
          {#each Object.entries(EQUIPMENT_LABELS) as [key, e] (key)}
            <label class="option">
              <input type="radio" name="equipment" value={key} bind:group={equipment} />
              <span class="option-label">{e.label}</span>
              <span class="option-hint">{e.hint}</span>
            </label>
          {/each}
        </div>
      </fieldset>

      <fieldset>
        <legend>What's the goal?</legend>
        <div class="options">
          {#each Object.entries(GOAL_LABELS) as [key, g] (key)}
            <label class="option">
              <input type="radio" name="goal" value={key} bind:group={goal} />
              <span class="option-label">{g.label}</span>
              <span class="option-hint">{g.hint}</span>
            </label>
          {/each}
        </div>
      </fieldset>
    </div>

    <section class="preview" aria-labelledby="preview-title" aria-live="polite">
      <h2 id="preview-title">Your week</h2>
      <ol class="week">
        {#each WEEKDAYS as day (day)}
          {@const plan = planFor(day)}
          <li class="day" class:rest={!plan}>
            <span class="day-name">{DAY_NAMES[day]}</span>
            {#if plan}
              <span class="plan-name">{plan.name}</span>
              <span class="exercises">
                {plan.exercises.map((e) => byId.get(e.exercise_id)?.name ?? e.exercise_id).join(' · ')}
              </span>
            {:else}
              <span class="plan-name">Rest</span>
            {/if}
          </li>
        {/each}
      </ol>
      <p class="reps"><strong>How many reps:</strong> {GOAL_LABELS[goal].reps}</p>
      <LoadSummary load={week.total} checks={weekChecks(week, goal)} mode="week" />
    </section>

    {#if confirming}
      <div class="confirm" role="group" aria-labelledby="replace-text">
        <p id="replace-text">This replaces your current routine. Your plans stay in the list.</p>
        <div class="actions">
          <button type="button" class="btn btn-secondary" onclick={() => (confirming = false)}>Keep mine</button>
          <button type="button" class="btn btn-primary" disabled={saving} onclick={save}>
            {saving ? 'Saving…' : 'Replace it'}
          </button>
        </div>
      </div>
    {:else}
      <div class="actions">
        <a class="btn btn-quiet" href="/plans">Cancel</a>
        <button type="button" class="btn btn-primary" disabled={saving} onclick={use}>
          {saving ? 'Saving…' : 'Use this routine'}
        </button>
      </div>
    {/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
  {/if}
</div>

<style>
  .page {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
    max-width: 48rem;
  }

  .back {
    display: inline-flex;
    align-items: center;
    align-self: flex-start;
    gap: var(--space-1);
    min-height: var(--touch-target);
    color: var(--color-text-muted);
    font-weight: var(--weight-medium);
    text-decoration: none;
  }

  .back svg {
    width: 1.25rem;
    height: 1.25rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .intro {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
  }

  h2 {
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
  }

  .questions {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  fieldset {
    margin: 0;
    padding: 0;
    border: none;
  }

  legend {
    margin-bottom: var(--space-2);
    font-weight: var(--weight-bold);
  }

  .options {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
    gap: var(--space-2);
  }

  /* A whole card is the radio's label: big targets, clear selected state
   * (accent border + tint + the native radio, so not color alone). */
  .option {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0 var(--space-2);
    align-items: center;
    min-height: var(--touch-target);
    padding: var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-surface);
    cursor: pointer;
    transition: background-color var(--duration-fast) var(--ease-out), border-color var(--duration-fast) var(--ease-out);
  }

  .option:has(input:checked) {
    border-color: var(--color-accent);
    background: var(--color-accent-subtle);
  }

  .option:has(input:focus-visible) {
    outline: 2px solid var(--color-accent);
    outline-offset: 2px;
  }

  .option input {
    width: 1.1rem;
    height: 1.1rem;
    margin: 0;
    accent-color: var(--color-accent);
  }

  .option-label {
    font-weight: var(--weight-semibold);
  }

  .option-hint {
    grid-column: 2;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  @media (hover: hover) {
    .option:hover {
      border-color: var(--color-accent);
    }
  }

  .preview {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    padding: var(--space-5);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
  }

  .week {
    display: flex;
    flex-direction: column;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .day {
    display: grid;
    grid-template-columns: 6.5rem 1fr;
    gap: var(--space-1) var(--space-3);
    padding: var(--space-2) 0;
  }

  .day + .day {
    border-top: 1px solid var(--color-border);
  }

  .day-name {
    color: var(--color-text-muted);
    font-weight: var(--weight-medium);
  }

  .plan-name {
    font-weight: var(--weight-semibold);
  }

  .rest .plan-name {
    color: var(--color-text-muted);
    font-weight: var(--weight-normal);
  }

  .exercises {
    grid-column: 2;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .reps {
    font-size: var(--text-sm);
    line-height: var(--leading-normal);
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: var(--space-3);
  }

  .confirm {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .error {
    color: var(--color-danger-text);
  }
</style>
