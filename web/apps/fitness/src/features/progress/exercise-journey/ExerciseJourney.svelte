<!--
  One exercise's journey (/progress/<exercise id>): the chart (what it
  shows depends on what the exercise tracks, see `journey`), records,
  then every workout's sets, newest first.
-->
<script lang="ts">
  import { LoadError } from '@home-tools/ui/components/load-error'
  import type { Profile } from '@home-tools/ui/profiles/types'
  import { LineChart } from '@/components/line-chart'
  import { getExercise, type Exercise } from '@/features/exercises'
  import { getExerciseHistory, journey, setLabel, type ExerciseWorkout } from '@/features/progress'

  let { profile, id }: { profile: Profile; id: string } = $props()

  let exercise = $state<Exercise | null>(null)
  let history = $state<ExerciseWorkout[]>([])
  let status = $state<'loading' | 'ready' | 'error'>('loading')

  async function load() {
    status = 'loading'
    try {
      ;[exercise, history] = await Promise.all([getExercise(id), getExerciseHistory(profile.id, id)])
      status = 'ready'
    } catch (err) {
      console.error('Loading exercise progress failed:', err)
      status = 'error'
    }
  }
  load()

  $effect(() => {
    if (exercise) document.title = `${exercise.name} · Fitness`
  })

  let units = $derived({ imperial: profile.units === 'imperial' })
  let j = $derived(exercise && history.length > 0 ? journey(exercise, history, units) : null)

  const when = (iso: string) =>
    new Date(iso).toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' })
</script>

<div class="journey">
  <a class="back" href="/progress">
    <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M15 6l-6 6 6 6" /></svg>
    Progress
  </a>

  {#if status === 'error'}
    <h1 tabindex="-1">Exercise</h1>
    <LoadError what="this exercise" onretry={load} />
  {:else if status === 'ready' && exercise}
    <header>
      <h1 tabindex="-1">{exercise.name}</h1>
      <p class="muted">
        {history.length}
        {history.length === 1 ? 'workout' : 'workouts'}{#if history.length > 0}&nbsp;· since
          <time datetime={history.at(-1)!.started_at}>{when(history.at(-1)!.started_at)}</time>{/if}
      </p>
    </header>

    {#if !j}
      <p class="muted">You haven't finished a workout with this exercise yet.</p>
    {:else}
      <section class="card" aria-labelledby="chart-title">
        <h2 id="chart-title">{j.measure}</h2>
        {#if j.points.length > 0}
          <LineChart points={j.points} format={j.format} label={`${exercise.name}: ${j.measure.toLowerCase()}`} />
        {:else}
          <p class="muted">Nothing to chart yet.</p>
        {/if}
      </section>

      {#if j.highlights.length > 0}
        <section aria-labelledby="records-title">
          <h2 id="records-title">Records</h2>
          <dl class="records">
            {#each j.highlights as h (h.label)}
              <div>
                <dt>{h.label}</dt>
                <dd>
                  <span class="value">{h.value}</span>
                  {#if h.detail}<span class="detail">{h.detail}</span>{/if}
                </dd>
              </div>
            {/each}
          </dl>
        </section>
      {/if}

      <section aria-labelledby="workouts-title">
        <h2 id="workouts-title">Workouts</h2>
        <ol class="list">
          {#each history as w (w.session_id)}
            <li class="row">
              <time class="date" datetime={w.started_at}>{when(w.started_at)}</time>
              <ul class="sets">
                {#each w.sets as s, i (i)}
                  <li>{setLabel(s, units)}</li>
                {/each}
              </ul>
            </li>
          {/each}
        </ol>
      </section>
    {/if}
  {/if}
</div>

<style>
  .journey {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
  }

  .back {
    display: inline-flex;
    align-items: center;
    align-self: flex-start;
    gap: var(--space-1);
    min-height: var(--touch-target);
    margin: calc(-1 * var(--space-3)) 0 calc(-1 * var(--space-4)) calc(-1 * var(--space-2));
    padding: 0 var(--space-2);
    border-radius: var(--radius-md);
    color: var(--color-accent-text);
    font-weight: var(--weight-semibold);
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

  @media (hover: hover) {
    .back:hover {
      background: var(--color-accent-subtle);
    }
  }

  header {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
    line-height: var(--leading-tight);
    overflow-wrap: anywhere;
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

  .card {
    padding: var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
  }

  .records {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(9rem, 1fr));
    gap: var(--space-2);
    margin: 0;
  }

  .records div {
    display: flex;
    flex-direction: column;
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
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    margin: 0;
  }

  .value {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  .detail {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
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
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-3) var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
  }

  .date {
    font-weight: var(--weight-semibold);
  }

  .sets {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .sets li {
    padding: var(--space-1) var(--space-3);
    border-radius: var(--radius-full);
    background: var(--color-bg);
    border: 1px solid var(--color-border);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
  }
</style>
