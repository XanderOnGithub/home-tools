<!--
  After "Finish" (decision #55): what you just did. Time, sets and
  exercises; new records (each exercise against its own all-time history:
  estimated max for weights, most reps for bodyweight); and the muscles
  it worked, on the map. Rare and earned, so it gets a short entrance
  (none with reduced motion). "Done" goes home.
-->
<script lang="ts">
  import type { Profile } from '@home-tools/ui/profiles/types'
  import { router } from '@home-tools/ui/router'
  import { muscleList, regionShades, type Shade } from '@/features/body'
  import { MuscleMap } from '@/features/body/muscle-map'
  import type { Exercise, Muscle } from '@/features/exercises'
  import { estimatedMax, getExerciseHistory, toWeight, weightUnit } from '@/features/progress'
  import type { Session, SetEntry } from '@/features/sessions'
  import { level, LEVEL_LABELS } from '@/features/training'

  let { profile, session, exercises }: { profile: Profile; session: Session; exercises: Map<string, Exercise> } = $props()

  let units = $derived({ imperial: profile.units === 'imperial' })
  let done = $derived(session.entries.filter((e) => e.sets.length > 0))
  let sets = $derived(done.reduce((n, e) => n + e.sets.length, 0))
  let minutes = $derived(Math.max(1, Math.round((Date.parse(session.ended_at!) - Date.parse(session.started_at)) / 60_000)))

  // Sets per muscle, as in the plan summary: sets × activation.
  let load = $derived.by(() => {
    const load = new Map<Muscle, number>()
    for (const e of done) {
      const ex = exercises.get(e.exercise_id)
      if (!ex || ex.category === 'stretching') continue
      for (const [m, a] of Object.entries(ex.activation) as [Muscle, number][]) {
        load.set(m, (load.get(m) ?? 0) + e.sets.length * a)
      }
    }
    return load
  })
  let shades = $derived(regionShades(Object.fromEntries([...load].map(([m, s]) => [m, level(s) as Shade]))))
  let worked = $derived([...load].filter(([, s]) => s >= 1).toSorted((a, b) => b[1] - a[1]).map(([m]) => m))

  // A set's score, as in tools/fitness/strength.go: bodyweight exercises
  // count reps without added weight, the rest their estimated max.
  const score = (ex: Exercise, s: SetEntry) => (ex.bodyweight ? (s.weight_kg ? 0 : (s.reps ?? 0)) : estimatedMax(s))
  const best = (ex: Exercise, list: SetEntry[]) => Math.max(0, ...list.map((s) => score(ex, s)))

  type Record = { name: string; now: string; before: string }
  let records = $state<Record[]>([])

  // One small request per exercise; records are a bonus, so a failure
  // just means none are shown.
  $effect(() => {
    const id = session.id
    const user = profile.id
    const u = units
    Promise.all(
      done.map(async (e): Promise<Record | null> => {
        const ex = exercises.get(e.exercise_id)
        if (!ex || !ex.metrics.includes('reps')) return null
        const now = best(ex, e.sets)
        if (now <= 0) return null
        const history = await getExerciseHistory(user, ex.id)
        const before = best(ex, history.filter((w) => w.session_id !== id).flatMap((w) => w.sets))
        if (before <= 0 || now <= before) return null // first time, or not beaten
        const show = (v: number) => (ex.bodyweight ? `${v} reps` : `${toWeight(v, u)} ${weightUnit(u)}`)
        return { name: ex.name, now: show(now), before: show(before) }
      }),
    )
      .then((list) => (records = list.filter((r): r is Record => r !== null)))
      .catch((err) => console.error('Checking records failed:', err))
  })
</script>

<section class="summary" aria-labelledby="summary-title">
  <div class="head">
    <h1 id="summary-title" tabindex="-1">Nice work, {profile.name}</h1>
    <p class="muted">Workout saved.</p>
  </div>

  <dl class="stats">
    <div><dt>Time</dt><dd>{minutes} min</dd></div>
    <div><dt>Sets</dt><dd>{sets}</dd></div>
    <div><dt>Exercises</dt><dd>{done.length}</dd></div>
  </dl>

  {#if records.length > 0}
    <section class="records" aria-labelledby="records-title">
      <h2 id="records-title">New {records.length === 1 ? 'record' : 'records'}</h2>
      <ul>
        {#each records as r, i (r.name)}
          <li style:--i={i}>
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 21h8 M12 17v4 M7 4h10v5a5 5 0 0 1-10 0z M17 6h3v2a3 3 0 0 1-3 3 M7 6H4v2a3 3 0 0 0 3 3" /></svg>
            <span><strong>{r.name}</strong><span class="muted">{r.now}{r.now.endsWith('reps') ? '' : ' estimated max'} · was {r.before}</span></span>
          </li>
        {/each}
      </ul>
    </section>
  {/if}

  {#if worked.length > 0}
    <section class="worked" aria-labelledby="worked-title">
      <h2 id="worked-title">What it worked</h2>
      <p>{muscleList(worked)}</p>
      <div class="map">
        <MuscleMap
          {shades}
          legend={[
            { shade: 3, label: LEVEL_LABELS[3] },
            { shade: 2, label: LEVEL_LABELS[2] },
            { shade: 1, label: LEVEL_LABELS[1] },
          ]}
        />
      </div>
    </section>
  {/if}

  <button type="button" class="btn btn-primary done" onclick={() => router.navigate('/')}>Done</button>
</section>

<style>
  .summary {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
    padding: var(--space-6) var(--space-4);
  }

  .head {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  h1 {
    font-size: var(--text-title);
    font-weight: var(--weight-bold);
    line-height: var(--leading-tight);
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
    flex-direction: column-reverse;
    gap: var(--space-1);
    padding: var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
  }

  dt {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  dd {
    margin: 0;
    font-size: var(--text-2xl);
    font-weight: var(--weight-bold);
    font-variant-numeric: tabular-nums;
  }

  .records,
  .worked {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .records ul {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .records li {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
  }

  .records li > span {
    display: flex;
    flex-direction: column;
  }

  .records svg {
    flex: none;
    width: 1.75rem;
    height: 1.75rem;
    fill: none;
    stroke: var(--color-accent-text);
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .map {
    width: 100%;
    max-width: 18rem;
  }

  .map :global(svg) {
    max-height: 14rem;
  }

  .done {
    align-self: stretch;
  }

  /* A short entrance for a rare, earned moment: the parts rise in one
   * after another (50 ms apart). Transform + opacity only. */
  @media (prefers-reduced-motion: no-preference) {
    .summary > *,
    .records li {
      animation: rise 300ms var(--ease-out) both;
    }

    .summary > :nth-child(2) {
      animation-delay: 50ms;
    }

    .summary > :nth-child(3) {
      animation-delay: 100ms;
    }

    .summary > :nth-child(4) {
      animation-delay: 150ms;
    }

    .summary > :nth-child(5) {
      animation-delay: 200ms;
    }

    .records li {
      animation-delay: calc(150ms + var(--i) * 60ms);
    }
  }

  @keyframes rise {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
  }
</style>
