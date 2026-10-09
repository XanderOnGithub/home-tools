<!--
  "Today": the hero of home. What the person's routine (weekly schedule) says for today
  (decision #27) in large type on the person's color, with the first few
  exercises; a rest day shows what's next instead.
-->
<script lang="ts">
  import { isoDate, weekdayKey } from '@/dates'
  import { getExercise } from '@/features/exercises'
  import type { FitnessProfile } from '@/features/fitness-profile'
  import type { Profile } from '@/features/profiles/types'
  import type { Plan } from '@/features/plans'
  import { startSession, type Session } from '@/features/sessions'
  import { router } from '@/router'

  let {
    profile,
    fitness,
    sessions,
    plans,
    active,
  }: {
    profile: Profile
    fitness: FitnessProfile
    sessions: Session[]
    plans: Plan[]
    active: Session | undefined // an unfinished workout, if any
  } = $props()

  let starting = $state(false)
  let startError = $state('')

  async function start(plan: Plan) {
    starting = true
    startError = ''
    try {
      const s = await startSession(profile.id, plan.id, plan.exercises.map((e) => e.exercise_id))
      router.navigate(`/workout/${s.id}`)
    } catch (err) {
      startError = `Couldn't start. ${(err as Error).message}`
      starting = false
    }
  }

  const PREVIEW = 4 // exercises listed before "+N more"

  const now = new Date()
  const planOn = (d: Date) => {
    const id = fitness.schedule?.[weekdayKey(d)]
    return id ? plans.find((r) => r.id === id) : undefined
  }

  let hasSchedule = $derived(Object.keys(fitness.schedule ?? {}).length > 0)
  let todays = $derived(planOn(now))
  let doneToday = $derived(sessions.some((s) => isoDate(new Date(s.started_at)) === isoDate(now)))

  // Next planned day after today, within the coming week.
  let next = $derived.by(() => {
    for (let i = 1; i <= 7; i++) {
      const d = new Date(now.getFullYear(), now.getMonth(), now.getDate() + i)
      const r = planOn(d)
      if (r) return { plan: r, day: i === 1 ? 'tomorrow' : d.toLocaleDateString(undefined, { weekday: 'long' }) }
    }
    return null
  })

  // Names for the preview list: a few small requests instead of
  // downloading the whole 876-exercise catalog.
  let names = $state<string[]>([])
  $effect(() => {
    const ids = todays?.exercises.slice(0, PREVIEW).map((e) => e.exercise_id) ?? []
    Promise.all(ids.map((id) => getExercise(id).then((ex) => ex.name).catch(() => null))).then(
      (list) => (names = list.filter((n): n is string => n !== null)),
    )
  })
</script>

<section class="today" aria-labelledby="today-label">
  <p id="today-label" class="label">Today</p>

  {#if active}
    <!-- An unfinished workout always comes first. -->
    <h2 class="title">Workout in progress</h2>
    <p class="meta">{plans.find((r) => r.id === active.plan_id)?.name ?? 'Workout'}</p>
    <a class="btn btn-primary start" href="/workout/{active.id}">Resume workout</a>
  {:else}

  {#if !hasSchedule}
    <h2 class="title">No routine yet</h2>
    <p class="meta">Set up your routine in <a href="/plans">Plans</a>.</p>
  {:else if todays}
    <h2 class="title">{todays.name}</h2>
    <p class="meta">
      {todays.exercises.length}
      {todays.exercises.length === 1 ? 'exercise' : 'exercises'}
      {#if doneToday}<span class="done">· Done</span>{/if}
    </p>
    {#if names.length > 0}
      <ul class="exercises">
        {#each names as name, i (i)}
          <li>{name}</li>
        {/each}
        {#if todays.exercises.length > PREVIEW}
          <li class="more">+{todays.exercises.length - PREVIEW} more</li>
        {/if}
      </ul>
    {/if}
  {:else}
    <h2 class="title">Rest day</h2>
    {#if next}
      <p class="meta">Next: {next.plan.name}, {next.day}</p>
    {/if}
  {/if}
  {#if todays && !doneToday}
    <button type="button" class="btn btn-primary start" onclick={() => todays && start(todays)} disabled={starting}>
      {starting ? 'Starting…' : 'Start workout'}
    </button>
  {/if}
  {#if startError}<p class="error" role="alert">{startError}</p>{/if}
  {/if}
</section>

<style>
  .today {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-6);
    border-radius: var(--radius-lg);
    background: var(--color-accent-subtle);
  }

  p {
    margin: 0;
  }

  .label {
    color: var(--color-accent-text);
    font-weight: var(--weight-bold);
  }

  .title {
    font-size: var(--text-3xl);
    font-weight: var(--weight-extrabold);
    line-height: var(--leading-tight);
  }

  .meta {
    color: var(--color-text-muted);
    font-size: var(--text-lg);
  }

  .done {
    color: var(--color-accent-text);
    font-weight: var(--weight-semibold);
  }

  .exercises {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: var(--space-3) 0 0;
    padding: 0;
    list-style: none;
  }

  .exercises li {
    padding-left: var(--space-4);
    border-left: 3px solid var(--color-accent);
    font-weight: var(--weight-medium);
  }

  .exercises .more {
    border-left-color: transparent;
    color: var(--color-text-muted);
  }

  .meta a {
    color: var(--color-accent-text);
    font-weight: var(--weight-semibold);
  }

  .start {
    align-self: flex-start;
    margin-top: var(--space-3);
    padding-inline: var(--space-6);
    font-size: var(--text-lg);
  }

  .error {
    color: var(--color-danger-text);
  }
</style>
