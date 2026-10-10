<!--
  "This week", drawn with tiny copies of the person's blob: filled = worked
  out, outlined = planned, dashed = missed (a past planned day with no
  workout), small dot = rest. Each day's state is also spelled out in
  hidden text, so it never relies on shape or color alone.
-->
<script lang="ts">
  import { isoDate, weekDays, weekdayKey } from '@home-tools/ui/dates'
  import type { FitnessProfile } from '@/features/fitness-profile'
  import { blobPath } from '@home-tools/ui/profiles/blob'
  import type { Profile } from '@home-tools/ui/profiles/types'
  import type { Plan } from '@/features/plans'
  import type { Session } from '@/features/sessions'

  let {
    profile,
    fitness,
    sessions,
    plans,
  }: { profile: Profile; fitness: FitnessProfile; sessions: Session[]; plans: Plan[] } = $props()

  const today = new Date()
  const todayISO = isoDate(today)

  let blob = $derived(blobPath(profile.name))

  let days = $derived(
    weekDays(today).map((d) => {
      const iso = isoDate(d)
      const planId = fitness.schedule?.[weekdayKey(d)]
      const done = sessions.some((s) => isoDate(new Date(s.started_at)) === iso)
      const planned = planId ? (plans.find((r) => r.id === planId)?.name ?? 'Workout') : null
      const state = done ? 'done' : planned ? (iso < todayISO ? 'missed' : 'planned') : 'rest'
      return {
        iso,
        state,
        planned,
        name: d.toLocaleDateString(undefined, { weekday: 'long' }),
        short: d.toLocaleDateString(undefined, { weekday: 'short' }),
        isToday: iso === todayISO,
      }
    }),
  )

  let done = $derived(days.filter((d) => d.state === 'done').length)
  let planned = $derived(days.filter((d) => d.planned).length)
</script>

<section class="week" aria-labelledby="week-title">
  <div class="head">
    <h2 id="week-title">This week</h2>
    <p class="count">
      {#if planned > 0}{done} of {planned} done{:else}{done} {done === 1 ? 'workout' : 'workouts'}{/if}
    </p>
  </div>

  <ol class="days">
    {#each days as day (day.iso)}
      <li class="day {day.state}" class:today={day.isToday}>
        <svg viewBox="0 0 100 100" aria-hidden="true">
          {#if day.state === 'rest'}
            <circle cx="50" cy="50" r="9" />
          {:else}
            <path d={blob} />
          {/if}
        </svg>
        <span class="short" aria-hidden="true">{day.short}</span>
        <span class="visually-hidden">
          {day.name}{day.isToday ? ' (today)' : ''}:
          {#if day.state === 'done'}worked out
          {:else if day.state === 'missed'}missed, {day.planned}
          {:else if day.state === 'planned'}planned, {day.planned}
          {:else}rest{/if}
        </span>
      </li>
    {/each}
  </ol>
</section>

<style>
  .week {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
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

  .count {
    margin: 0;
    color: var(--color-text-muted);
    font-weight: var(--weight-medium);
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

  svg {
    width: 100%;
    max-width: 2.75rem;
    aspect-ratio: 1;
    overflow: visible;
  }

  /* vector-effect keeps outlines 2px however big the blob is drawn. */
  path,
  circle {
    vector-effect: non-scaling-stroke;
    stroke-width: 2px;
  }

  .done path {
    fill: var(--color-accent);
    stroke: var(--color-accent);
  }

  .planned path {
    fill: none;
    stroke: var(--color-accent);
  }

  .missed path {
    fill: none;
    stroke: var(--color-border-strong);
    stroke-dasharray: 4 4;
  }

  .rest circle {
    fill: var(--color-border);
    stroke: none;
  }

  .short {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
  }

  /* Today: bold label in a pill, so it's findable without color. */
  .today .short {
    padding: 0 var(--space-2);
    border-radius: var(--radius-full);
    background: var(--color-text);
    color: var(--color-bg);
    font-weight: var(--weight-bold);
  }
</style>
