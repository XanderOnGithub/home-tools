<!--
  Workout mode: full screen, one exercise at a time.

  - Each set row is pre-filled with what was done last time for that
    exercise ("Last: 8 × 135 lb" beside it); tapping ✓ completes it.
  - Completing a set saves the whole session (so a refresh or a dead
    battery loses nothing) and starts the rest timer.
  - The screen stays on while this is open (where the browser allows).
  - "Leave" keeps the workout in progress (home offers Resume);
    "Finish" stamps the end time.
-->
<script lang="ts">
  import { onDestroy, onMount, tick } from 'svelte'
  import { getExercise, label, primaryMuscles, type Exercise } from '@/features/exercises'
  import type { Profile } from '@/features/profiles/types'
  import { getRoutines } from '@/features/routines'
  import { RestTimer } from '@/features/workout/rest-timer'
  import {
    getRecentSessions,
    lastSets,
    saveSession,
    type Session,
    type SetEntry,
  } from '@/features/sessions'
  import { router } from '@/router'
  import { kgToLb, lbToKg, parseNumber } from '@/units'
  import { keepScreenOn } from '@/wake-lock'

  let { profile, sessionId }: { profile: Profile; sessionId: string } = $props()

  const DEFAULT_SETS = 3
  const DEFAULT_REST_SEC = 90

  // One editable row per set. Values are kept as typed text; `done` rows
  // are the ones saved to the server. `fromKg` remembers the exact stored
  // weight a pre-filled value came from: lb are shown rounded (84 kg →
  // "185.2"), and converting that text back would drift (→ 84.01 kg).
  type Row = { reps: string; weight: string; duration: string; done: boolean; fromKg?: { kg: number; shown: string } }
  type Step = { exercise: Exercise | null; exerciseId: string; rows: Row[]; last: SetEntry[] }

  let session = $state<Session | null>(null)
  let steps = $state<Step[]>([])
  let step = $state(0)
  let status = $state<'loading' | 'ready' | 'missing' | 'error'>('loading')
  let saveError = $state('')
  let announce = $state('') // screen reader live region
  let confirmingFinish = $state(false)
  let finishing = $state(false)

  let restSec = $state(DEFAULT_REST_SEC) // remembered for this workout
  let restEndsAt = $state<number | null>(null)
  let restTotal = $state(DEFAULT_REST_SEC)
  let now = $state(Date.now())

  let imperial = $derived(profile.units === 'imperial')
  let unit = $derived(imperial ? 'lb' : 'kg')
  const round1 = (n: number) => Math.round(n * 10) / 10
  const showWeight = (kg: number) => String(round1(imperial ? kgToLb(kg) : kg))

  async function load() {
    status = 'loading'
    try {
      const [history, routines] = await Promise.all([getRecentSessions(profile.id), getRoutines()])
      const s = history.find((h) => h.id === sessionId)
      if (!s) {
        status = 'missing'
        return
      }
      const routine = routines.find((r) => r.id === s.routine_id)
      const exercises = await Promise.all(s.entries.map((e) => getExercise(e.exercise_id).catch(() => null)))

      steps = s.entries.map((entry, i) => {
        const last = lastSets(history, entry.exercise_id, s.id)
        const suggested = routine?.exercises.find((r) => r.exercise_id === entry.exercise_id)?.suggested_sets
        const planned = Math.max(entry.sets.length, suggested ?? (last.length || DEFAULT_SETS))
        // Done sets keep their values; the rest are pre-filled from last time.
        const rows = Array.from({ length: planned }, (_, n): Row => {
          const done = entry.sets[n]
          const src = done ?? last[n] ?? last.at(-1)
          const weight = src?.weight_kg ? showWeight(src.weight_kg) : ''
          return {
            reps: src?.reps ? String(src.reps) : '',
            weight,
            duration: src?.duration_sec ? String(src.duration_sec) : '',
            done: !!done,
            fromKg: src?.weight_kg ? { kg: src.weight_kg, shown: weight } : undefined,
          }
        })
        return { exercise: exercises[i], exerciseId: entry.exercise_id, rows, last }
      })
      session = s
      // Resume at the first exercise that still has sets to do.
      const firstOpen = steps.findIndex((st) => st.rows.some((r) => !r.done))
      step = firstOpen === -1 ? steps.length - 1 : firstOpen
      status = 'ready'
    } catch (err) {
      console.error('Loading workout failed:', err)
      status = 'error'
    }
  }
  load()

  let stopWakeLock: () => void = () => {}
  let ticker: ReturnType<typeof setInterval>
  onMount(() => {
    stopWakeLock = keepScreenOn()
    ticker = setInterval(() => (now = Date.now()), 250)
  })
  onDestroy(() => {
    stopWakeLock()
    clearInterval(ticker)
  })

  // Rest over: buzz (where supported), announce, close the timer.
  $effect(() => {
    if (restEndsAt !== null && now >= restEndsAt) {
      restEndsAt = null
      navigator.vibrate?.([200, 100, 200])
      announce = 'Rest over. Next set.'
    }
  })

  let current = $derived(steps[step])
  let elapsed = $derived.by(() => {
    if (!session) return ''
    const sec = Math.max(0, Math.floor((now - Date.parse(session.started_at)) / 1000))
    const h = Math.floor(sec / 3600)
    const m = Math.floor((sec % 3600) / 60)
    const s = String(sec % 60).padStart(2, '0')
    return h ? `${h}:${String(m).padStart(2, '0')}:${s}` : `${m}:${s}`
  })
  let setsDone = $derived(steps.reduce((n, st) => n + st.rows.filter((r) => r.done).length, 0))

  const tracks = (ex: Exercise | null, metric: Exercise['metrics'][number]) =>
    ex ? ex.metrics.includes(metric) : metric === 'reps' || metric === 'weight'

  /** Row → API set, or an error message (mirrors the server's Set rules). */
  function toSet(row: Row, ex: Exercise | null): SetEntry | string {
    const set: SetEntry = {}
    if (tracks(ex, 'reps')) {
      const reps = parseNumber(row.reps)
      if (!(Number.isInteger(reps) && reps > 0)) return 'Enter the reps.'
      set.reps = reps
    }
    if (tracks(ex, 'weight')) {
      const w = parseNumber(row.weight)
      if (row.weight.trim() === '' && ex?.bodyweight) {
        // Bodyweight only: no added weight.
      } else if (!(w > 0)) {
        return ex?.bodyweight ? 'Enter added weight, or leave it empty.' : `Enter the weight in ${unit}.`
      } else if (row.fromKg && row.weight === row.fromKg.shown) {
        set.weight_kg = row.fromKg.kg // untouched pre-fill: keep the exact value
      } else {
        // 0.01 kg precision, so lb reads back exactly as typed.
        set.weight_kg = Math.round((imperial ? lbToKg(w) : w) * 100) / 100
      }
    }
    if (tracks(ex, 'duration')) {
      const d = parseNumber(row.duration)
      if (!(Number.isInteger(d) && d > 0)) return 'Enter the time in seconds.'
      set.duration_sec = d
    }
    return set
  }

  /** Saves every done row; on failure the caller undoes its change. */
  async function persist(): Promise<boolean> {
    if (!session) return false
    const entries = steps.map((st) => ({
      exercise_id: st.exerciseId,
      sets: st.rows.filter((r) => r.done).map((r) => toSet(r, st.exercise) as SetEntry),
    }))
    try {
      session = await saveSession({ ...session, entries })
      saveError = ''
      return true
    } catch (err) {
      saveError = `Couldn't save. ${(err as Error).message}`
      return false
    }
  }

  let rowError = $state<{ step: number; row: number; message: string } | null>(null)

  async function toggleDone(i: number) {
    const row = current.rows[i]
    if (!row.done) {
      const result = toSet(row, current.exercise)
      if (typeof result === 'string') {
        rowError = { step, row: i, message: result }
        return
      }
    }
    rowError = null
    row.done = !row.done
    if (!(await persist())) {
      row.done = !row.done // undo: the server didn't take it
      return
    }
    if (row.done) {
      const allDone = steps.every((st) => st.rows.every((r) => r.done))
      if (!allDone) {
        restTotal = restSec
        restEndsAt = Date.now() + restSec * 1000
        announce = `Set ${i + 1} done. Rest ${restSec} seconds.`
      } else {
        announce = 'All sets done. Finish when ready.'
      }
    }
  }

  function adjustRest(delta: number) {
    if (restEndsAt === null) return
    restSec = Math.max(15, restSec + delta)
    restEndsAt = Math.max(Date.now() + 5000, restEndsAt + delta * 1000)
    restTotal = Math.max(restTotal + delta, 15)
  }

  // The last removed set, so a mis-tap can be undone. Cleared by the
  // next removal or by moving to another exercise.
  let removed = $state<{ step: number; index: number; row: Row } | null>(null)

  async function removeSet(i: number) {
    const [row] = current.rows.splice(i, 1)
    removed = { step, index: i, row }
    rowError = null
    announce = `Set ${i + 1} removed.`
    // A completed set is part of the saved workout: save without it.
    if (row.done && !(await persist())) {
      current.rows.splice(i, 0, row) // put it back: the server didn't take it
      removed = null
      return
    }
    // Keep keyboard focus in the list: the set that moved into this spot,
    // or "Add set" if it was the last one.
    await tick()
    const next = document.querySelectorAll<HTMLButtonElement>('.set .remove')[i]
    ;(next ?? document.querySelector<HTMLButtonElement>('.add-set'))?.focus()
  }

  async function undoRemove() {
    if (!removed) return
    const { step: at, index, row } = removed
    removed = null
    steps[at].rows.splice(index, 0, row)
    announce = `Set ${index + 1} restored.`
    if (row.done && !(await persist())) steps[at].rows.splice(index, 1)
  }

  function addSet() {
    const prev = current.rows.at(-1)
    current.rows.push({
      reps: prev?.reps ?? '',
      weight: prev?.weight ?? '',
      duration: prev?.duration ?? '',
      done: false,
      fromKg: prev?.fromKg,
    })
  }

  let heading = $state<HTMLHeadingElement>()

  async function askFinish() {
    confirmingFinish = true
    await tick()
    document.getElementById('finish-title')?.focus()
  }

  async function goTo(next: number) {
    step = next
    rowError = null
    removed = null
    await tick()
    heading?.focus() // announce the new exercise
  }

  async function finish() {
    if (!session) return
    finishing = true
    const ok = await persist()
    if (ok && session) {
      try {
        await saveSession({ ...session, ended_at: new Date().toISOString() })
        router.navigate('/')
        return
      } catch (err) {
        saveError = `Couldn't finish. ${(err as Error).message}`
      }
    }
    finishing = false
  }

  const lastLabel = (s: SetEntry | undefined) => {
    if (!s) return ''
    const parts = []
    if (s.reps) parts.push(`${s.reps}`)
    if (s.weight_kg) parts.push(`${showWeight(s.weight_kg)} ${unit}`)
    if (s.duration_sec) parts.push(`${s.duration_sec} s`)
    return parts.join(' × ')
  }
</script>

<div class="workout">
  <header class="top">
    <a class="btn btn-quiet leave" href="/">
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M15 6l-6 6 6 6" /></svg>
      Leave
    </a>
    <span class="elapsed" aria-label="Elapsed time">{elapsed}</span>
    {#if status === 'ready'}
      <button type="button" class="btn btn-secondary" onclick={askFinish}>Finish</button>
    {/if}
  </header>

  <p class="visually-hidden" role="status">{announce}</p>

  {#if status === 'missing'}
    <p class="message">This workout doesn't exist anymore. <a href="/">Go home</a></p>
  {:else if status === 'error'}
    <div class="message" role="alert">
      <p>Couldn't load the workout. Check that the server is running.</p>
      <button type="button" class="btn btn-primary" onclick={load}>Try again</button>
    </div>
  {:else if status === 'ready' && current}
    <div class="progress" aria-hidden="true">
      {#each steps as st, i (st.exerciseId)}
        <span class:done={st.rows.every((r) => r.done)} class:current={i === step}></span>
      {/each}
    </div>

    <main class="main">
      {#if confirmingFinish}
        <section class="finish" aria-labelledby="finish-title">
          <h1 id="finish-title" tabindex="-1">Finish workout?</h1>
          <p>
            {setsDone}
            {setsDone === 1 ? 'set' : 'sets'} logged in {elapsed}. Sets you didn't check off aren't saved.
          </p>
          <div class="finish-actions">
            <button type="button" class="btn btn-secondary" onclick={() => (confirmingFinish = false)}>Keep going</button>
            <button type="button" class="btn btn-primary" onclick={finish} disabled={finishing}>
              {finishing ? 'Saving…' : 'Finish'}
            </button>
          </div>
          {#if saveError}<p class="error" role="alert">{saveError}</p>{/if}
        </section>
      {:else}
        <div class="exercise">
          {#if current.exercise?.images?.[0]}
            <img src="/images/{current.exercise.images[0]}" alt="" width="72" height="72" />
          {/if}
          <div>
            <p class="step-count">Exercise {step + 1} of {steps.length}</p>
            <h1 bind:this={heading} tabindex="-1">{current.exercise?.name ?? current.exerciseId}</h1>
            {#if current.exercise}
              <p class="muscles">{primaryMuscles(current.exercise).map(label).join(', ')}</p>
            {/if}
          </div>
        </div>

        <ol class="sets">
          {#each current.rows as row, i (i)}
            {@const last = current.last[i]}
            <li class="set" class:done={row.done}>
              <span class="set-num" aria-hidden="true">{i + 1}</span>
              <span class="inputs">
                {#if tracks(current.exercise, 'reps')}
                  <label class="field">
                    <input type="text" inputmode="numeric" maxlength="3" bind:value={row.reps} disabled={row.done} />
                    <span>reps<span class="visually-hidden">, set {i + 1}</span></span>
                  </label>
                {/if}
                {#if tracks(current.exercise, 'weight')}
                  <label class="field">
                    <input
                      type="text"
                      inputmode="decimal"
                      maxlength="5"
                      bind:value={row.weight}
                      disabled={row.done}
                      placeholder={current.exercise?.bodyweight ? 'BW' : ''}
                    />
                    <span>{unit}<span class="visually-hidden">, set {i + 1}</span></span>
                  </label>
                {/if}
                {#if tracks(current.exercise, 'duration')}
                  <label class="field">
                    <input type="text" inputmode="numeric" maxlength="4" bind:value={row.duration} disabled={row.done} />
                    <span>sec<span class="visually-hidden">, set {i + 1}</span></span>
                  </label>
                {/if}
              </span>
              {#if last}<span class="last">Last: {lastLabel(last)}</span>{/if}
              <button
                type="button"
                class="check"
                aria-pressed={row.done}
                aria-label="Set {i + 1} done"
                onclick={() => toggleDone(i)}
              >
                <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12l5 5 9-10" /></svg>
              </button>
              <button type="button" class="btn btn-icon remove" aria-label="Remove set {i + 1}" onclick={() => removeSet(i)}>
                <svg viewBox="0 0 24 24"><path d="M6 6l12 12 M18 6L6 18" /></svg>
              </button>
              {#if rowError?.step === step && rowError.row === i}
                <p class="error row-error" role="alert">{rowError.message}</p>
              {/if}
            </li>
          {/each}
        </ol>

        {#if removed && removed.step === step}
          <p class="undo">
            Set {removed.index + 1} removed.
            <button type="button" class="btn btn-quiet" onclick={undoRemove}>Undo</button>
          </p>
        {/if}

        <button type="button" class="btn btn-quiet add-set" onclick={addSet}>+ Add set</button>
        {#if saveError}<p class="error" role="alert">{saveError}</p>{/if}
      {/if}
    </main>

    {#if !confirmingFinish}
      <footer class="bottom">
        {#if restEndsAt !== null}
          <RestTimer
            endsAt={restEndsAt}
            total={restTotal}
            {now}
            onadjust={adjustRest}
            onskip={() => (restEndsAt = null)}
          />
        {:else}
          <div class="nav">
            <button type="button" class="btn btn-secondary" disabled={step === 0} onclick={() => goTo(step - 1)}>
              Previous
            </button>
            {#if step < steps.length - 1}
              <button
                type="button"
                class="btn {current.rows.every((r) => r.done) ? 'btn-primary' : 'btn-secondary'}"
                onclick={() => goTo(step + 1)}
              >
                Next exercise
              </button>
            {:else}
              <button type="button" class="btn btn-primary" onclick={askFinish}>Finish workout</button>
            {/if}
          </div>
        {/if}
      </footer>
    {/if}
  {/if}
</div>

<style>
  .workout {
    display: flex;
    flex-direction: column;
    width: min(40rem, 100%);
    min-height: 100dvh;
    margin: 0 auto;
  }

  .top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-3);
  }

  .leave {
    padding-left: var(--space-2);
  }

  .leave svg {
    width: 1.25rem;
    height: 1.25rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .elapsed {
    font-weight: var(--weight-bold);
    font-variant-numeric: tabular-nums;
  }

  /* One segment per exercise: done = filled, current = outlined. */
  .progress {
    display: flex;
    gap: var(--space-1);
    padding: 0 var(--space-4);
  }

  .progress span {
    flex: 1;
    height: 0.375rem;
    border-radius: var(--radius-full);
    background: var(--color-border);
  }

  .progress .done {
    background: var(--color-accent);
  }

  .progress .current:not(.done) {
    background: var(--color-text);
  }

  .main {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: var(--space-4);
    padding: var(--space-5) var(--space-4);
  }

  .message {
    padding: var(--space-6) var(--space-4);
  }

  .exercise {
    display: flex;
    align-items: center;
    gap: var(--space-4);
  }

  .exercise img {
    flex: none;
    width: 4.5rem;
    height: 4.5rem;
    border-radius: var(--radius-md);
    object-fit: cover;
  }

  .step-count,
  .muscles {
    margin: 0;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
    line-height: var(--leading-tight);
  }

  h1:focus {
    outline: none;
  }

  .sets {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .set {
    display: grid;
    grid-template-columns: auto 1fr auto auto;
    grid-template-areas: 'num inputs check remove' 'num last check remove' 'err err err err';
    align-items: center;
    column-gap: var(--space-2);
    padding: var(--space-3) var(--space-1) var(--space-3) var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
    transition: background var(--duration-fast) var(--ease-out);
  }

  .set.done {
    border-color: transparent;
    background: var(--color-accent-subtle);
  }

  .set-num {
    grid-area: num;
    width: 1.75rem;
    color: var(--color-text-muted);
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
    text-align: center;
  }

  .inputs {
    grid-area: inputs;
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }

  .field {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  /* Narrow enough that reps + weight stay on one line on a phone, wide
   * enough for "185.5" at this size. */
  .field input {
    width: 3.75rem;
    min-height: var(--touch-target);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-bg);
    color: var(--color-text);
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
    text-align: center;
  }

  .field input:disabled {
    border-color: transparent;
    background: none;
    opacity: 1;
  }

  .last {
    grid-area: last;
    color: var(--color-text-muted);
    font-size: var(--text-xs);
  }

  /* Big round check: outlined until done, then filled with a check. */
  .check {
    grid-area: check;
    display: grid;
    place-items: center;
    width: 3rem;
    height: 3rem;
    border: 2px solid var(--color-border-strong);
    border-radius: var(--radius-full);
    background: none;
    color: var(--color-text-muted);
    cursor: pointer;
    transition:
      background var(--duration-fast) var(--ease-out),
      scale var(--duration-fast) var(--ease-out);
  }

  .check:active {
    scale: 0.92;
  }

  .check[aria-pressed='true'] {
    border-color: var(--color-accent);
    background: var(--color-accent);
    color: var(--color-on-accent);
  }

  .check svg {
    width: 1.5rem;
    height: 1.5rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 3;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  /* Remove sits beside the check, quieter, so it isn't hit by accident. */
  .remove {
    grid-area: remove;
  }

  .undo {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin: 0;
    color: var(--color-text-muted);
  }

  .row-error {
    grid-area: err;
    margin-top: var(--space-2);
  }

  .error {
    margin: 0;
    color: var(--color-danger-text);
  }

  .add-set {
    align-self: flex-start;
  }

  /* Sticky bottom bar: exercise navigation, or the rest timer. */
  .bottom {
    position: sticky;
    bottom: 0;
    padding: var(--space-4) var(--space-4) calc(var(--space-4) + env(safe-area-inset-bottom));
    border-top: 1px solid var(--color-border);
    background: var(--color-bg);
  }

  .nav {
    display: flex;
    justify-content: space-between;
    gap: var(--space-3);
  }

  .nav .btn {
    flex: 1;
  }

  .finish {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .finish p {
    margin: 0;
  }

  .finish-actions {
    display: flex;
    gap: var(--space-3);
  }
</style>
