<!--
  "This week" as seven circles: filled with a check = worked
  out, outlined = planned, dashed = missed (a past planned day with no
  workout), small dot = rest. Each day's state is also spelled out in
  hidden text, so it never relies on shape or color alone.
-->
<script lang="ts">
  import { isoDate, weekDays, weekdayKey } from '@home-tools/ui/dates'
  import type { FitnessProfile } from '@/features/fitness-profile'
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
            <circle class="rest-dot" cx="50" cy="50" r="8" />
          {:else}
            <circle class="ring" cx="50" cy="50" r="42" />
            {#if day.state === 'done'}<path class="check" d="M32 51l12 12 24-26" />{/if}
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

  .count {
    margin: 0;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
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
    max-width: 2.25rem;
    aspect-ratio: 1;
    overflow: visible;
  }

  /* vector-effect keeps outlines 2px however big the circle is drawn.
   * Done = filled with a check, planned = accent ring, missed = dashed
   * gray ring, rest = a dot: the shape says it, not only the color. */
  path,
  circle {
    vector-effect: non-scaling-stroke;
    stroke-width: 2px;
  }

  .done .ring {
    fill: var(--color-accent);
    stroke: var(--color-accent);
  }

  .check {
    fill: none;
    stroke: var(--color-on-accent);
    stroke-width: 2.5px;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .planned .ring {
    fill: none;
    stroke: var(--color-accent);
  }

  .missed .ring {
    fill: none;
    stroke: var(--color-border-strong);
    stroke-dasharray: 3 4;
  }

  .rest-dot {
    fill: var(--color-border-strong);
    stroke: none;
    opacity: 0.5;
  }

  .short {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
  }

  /* Today: bold label in a pill, so it's findable without color. */
  .today .short {
    padding: 0 var(--space-2);
    border-radius: var(--radius-sm);
    background: var(--color-text);
    color: var(--color-surface);
    font-weight: var(--weight-semibold);
  }
</style>
