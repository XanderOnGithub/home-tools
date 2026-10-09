<!--
  "Your routine": one row per weekday, each with a plan picker (or Rest).
  Saves on change into this person's fitness profile (decision #27, #30);
  a live region announces "Saved" so it's clear nothing else is needed.
-->
<script lang="ts">
  import { saveFitnessProfile, type FitnessProfile, type Weekday } from '@/features/fitness-profile'
  import type { Plan } from '@/features/plans'

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

  // Archived plans can't be newly picked, but stay listed on the days
  // that already use them, so the select doesn't silently show "Rest".
  const choicesFor = (day: Weekday) =>
    plans.filter((r) => !r.archived || r.id === fitness.schedule?.[day])

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
</section>

<style>
  .week {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
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
