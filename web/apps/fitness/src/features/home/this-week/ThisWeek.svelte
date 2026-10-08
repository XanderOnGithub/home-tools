<!--
  "This week": workouts done vs. planned, and a Monday-to-Sunday row.
  Each day shows done (filled), planned (ring), missed (faded ring: a past
  planned day with no workout) or rest (small dot); the same state is
  spelled out in hidden text, so it never relies on color.
-->
<script lang="ts">
  import { Card } from '@/components/card'
  import { isoDate, weekDays, weekdayKey } from '@/dates'
  import type { FitnessProfile } from '@/features/fitness-profile'
  import type { Routine } from '@/features/routines'
  import type { Session } from '@/features/sessions'

  let {
    fitness,
    sessions,
    routines,
  }: { fitness: FitnessProfile; sessions: Session[]; routines: Routine[] } = $props()

  const today = new Date()
  const todayISO = isoDate(today)

  let days = $derived(
    weekDays(today).map((d) => {
      const iso = isoDate(d)
      const routineId = fitness.schedule?.[weekdayKey(d)]
      return {
        iso,
        name: d.toLocaleDateString(undefined, { weekday: 'long' }),
        initial: d.toLocaleDateString(undefined, { weekday: 'narrow' }),
        done: sessions.some((s) => isoDate(new Date(s.started_at)) === iso),
        planned: routineId ? (routines.find((r) => r.id === routineId)?.name ?? 'Workout') : null,
        isToday: iso === todayISO,
        isPast: iso < todayISO, // "YYYY-MM-DD" strings compare like dates
      }
    }),
  )

  let done = $derived(days.filter((d) => d.done).length)
  let planned = $derived(days.filter((d) => d.planned).length)
</script>

<Card title="This week">
  <p class="summary">
    {#if planned > 0}
      <strong>{done}</strong> of {planned} workouts
    {:else}
      <strong>{done}</strong> {done === 1 ? 'workout' : 'workouts'}
    {/if}
  </p>

  <ol class="days">
    {#each days as day (day.iso)}
      <li class="day" class:today={day.isToday}>
        <span class="initial" aria-hidden="true">{day.initial}</span>
        <span
          class="mark"
          class:done={day.done}
          class:planned={!day.done && day.planned}
          class:missed={!day.done && day.planned && day.isPast}
          aria-hidden="true"
        ></span>
        <span class="visually-hidden">
          {day.name}{day.isToday ? ' (today)' : ''}:
          {#if day.done}worked out
          {:else if day.planned && day.isPast}missed, {day.planned}
          {:else if day.planned}planned, {day.planned}
          {:else}rest{/if}
        </span>
      </li>
    {/each}
  </ol>
</Card>

<style>
  .summary {
    margin: 0;
    font-size: var(--text-lg);
  }

  .summary strong {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
  }

  .days {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: var(--space-1);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .day {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
  }

  .initial {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
  }

  .today .initial {
    color: var(--color-text);
  }

  /* Rest: small dot. Planned: ring. Done: filled. Today: underline bar. */
  .mark {
    display: block;
    width: 0.5rem;
    height: 0.5rem;
    margin: 0.5rem 0;
    border-radius: var(--radius-full);
    background: var(--color-border);
  }

  .mark.planned,
  .mark.done {
    width: 1.5rem;
    height: 1.5rem;
    margin: 0;
  }

  .mark.planned {
    background: none;
    border: 2px solid var(--color-accent);
  }

  .mark.done {
    background: var(--color-accent);
  }

  /* Missed: same ring, faded and dashed, so it reads "didn't happen"
   * without looking like an error (no red: nothing is wrong). */
  .mark.missed {
    border-color: var(--color-border-strong);
    border-style: dashed;
  }

  .today::after {
    content: '';
    width: 1.25rem;
    height: 3px;
    border-radius: var(--radius-full);
    background: var(--color-text);
  }
</style>
