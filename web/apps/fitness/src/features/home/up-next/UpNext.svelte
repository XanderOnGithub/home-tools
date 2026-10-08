<!--
  "Today": what the weekly schedule says for today (decision #27), or the
  next planned day if today is a rest day.
-->
<script lang="ts">
  import { Card } from '@/components/card'
  import { isoDate, weekdayKey } from '@/dates'
  import type { FitnessProfile } from '@/features/fitness-profile'
  import type { Routine } from '@/features/routines'
  import type { Session } from '@/features/sessions'

  let {
    fitness,
    sessions,
    routines,
  }: { fitness: FitnessProfile; sessions: Session[]; routines: Routine[] } = $props()

  const now = new Date()
  const routineOn = (d: Date) => {
    const id = fitness.schedule?.[weekdayKey(d)]
    return id ? routines.find((r) => r.id === id) : undefined
  }

  let hasSchedule = $derived(Object.keys(fitness.schedule ?? {}).length > 0)
  let todays = $derived(routineOn(now))
  let doneToday = $derived(sessions.some((s) => isoDate(new Date(s.started_at)) === isoDate(now)))

  // Next planned day after today, within the coming week.
  let next = $derived.by(() => {
    for (let i = 1; i <= 7; i++) {
      const d = new Date(now.getFullYear(), now.getMonth(), now.getDate() + i)
      const r = routineOn(d)
      if (r) return { routine: r, day: i === 1 ? 'tomorrow' : d.toLocaleDateString(undefined, { weekday: 'long' }) }
    }
    return null
  })
</script>

<Card title="Today">
  {#if !hasSchedule}
    <p class="big">No weekly schedule yet</p>
    <p class="muted">Plan which routine goes on which day in <a href="/routines">Routines</a>.</p>
  {:else if todays}
    <p class="big">{todays.name}</p>
    <p class="muted">
      {todays.exercises.length}
      {todays.exercises.length === 1 ? 'exercise' : 'exercises'}{doneToday ? ' · done today' : ''}
    </p>
  {:else}
    <p class="big">Rest day</p>
    {#if next}
      <p class="muted">Next: {next.routine.name}, {next.day}</p>
    {/if}
  {/if}
</Card>

<style>
  p {
    margin: 0;
  }

  .big {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  .muted {
    color: var(--color-text-muted);
  }

  a {
    color: var(--color-accent-text);
    font-weight: var(--weight-semibold);
  }
</style>
