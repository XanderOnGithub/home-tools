<!--
  Create (id = null) or edit a shared plan: a name and an ordered list
  of exercises, each with an optional suggested number of sets (a hint,
  never enforced; decision #7). Reorder with up/down buttons (keyboard and
  touch friendly; the moved row keeps focus). Archive instead of delete.
-->
<script lang="ts">
  import { tick } from 'svelte'
  import { getCatalog, label, primaryMuscles, type Exercise } from '@/features/exercises'
  import { ExercisePicker } from '@/features/exercises/exercise-picker'
  import type { Profile } from '@/features/profiles/types'
  import { getPlans, savePlan, type Plan } from '@/features/plans'
  import { idFromName } from '@/ids'
  import { router } from '@/router'

  let { profile, id }: { profile: Profile; id: string | null } = $props()

  type Item = { exercise_id: string; sets: string } // sets as typed ('' = no hint)

  let plans = $state<Plan[]>([])
  let catalog = $state<Exercise[]>([])
  let original = $state<Plan | null>(null)
  let name = $state('')
  let items = $state<Item[]>([])
  let status = $state<'loading' | 'ready' | 'missing' | 'error'>('loading')
  let picking = $state(false)
  let saving = $state(false)
  let nameError = $state('')
  let formError = $state('')
  let confirmingArchive = $state(false)

  async function load() {
    status = 'loading'
    try {
      ;[plans, catalog] = await Promise.all([getPlans(), getCatalog()])
      if (id) {
        original = plans.find((r) => r.id === id) ?? null
        if (!original) {
          status = 'missing'
          return
        }
        name = original.name
        items = original.exercises.map((e) => ({
          exercise_id: e.exercise_id,
          sets: e.suggested_sets ? String(e.suggested_sets) : '',
        }))
      }
      status = 'ready'
    } catch (err) {
      console.error('Loading plan failed:', err)
      status = 'error'
    }
  }
  load()

  // id → exercise, so each row's lookup is O(1) instead of scanning 876.
  let byId = $derived(new Map(catalog.map((ex) => [ex.id, ex])))
  let selected = $derived(new Set(items.map((i) => i.exercise_id)))

  function toggle(ex: Exercise) {
    const i = items.findIndex((it) => it.exercise_id === ex.id)
    if (i === -1) items.push({ exercise_id: ex.id, sets: '' })
    else items.splice(i, 1)
  }

  async function move(index: number, by: -1 | 1) {
    const [item] = items.splice(index, 1)
    items.splice(index + by, 0, item)
    // Keyed rows move in the DOM, so the pressed button keeps focus;
    // if it hit the end (now disabled), focus the row's other arrow.
    await tick()
    const row = document.getElementById(`row-${item.exercise_id}`)
    const btn = row?.querySelector<HTMLButtonElement>(by === -1 ? '.up' : '.down')
    if (btn?.disabled) row?.querySelector<HTMLButtonElement>(by === -1 ? '.down' : '.up')?.focus()
  }

  async function save(event: SubmitEvent) {
    event.preventDefault()
    nameError = ''
    formError = ''
    if (!name.trim()) {
      nameError = 'Give the plan a name.'
      return
    }
    if (items.length === 0) {
      formError = 'Add at least one exercise.'
      return
    }
    const bad = items.find((i) => i.sets.trim() && !(Number(i.sets) >= 1 && Number(i.sets) <= 20))
    if (bad) {
      formError = `Sets for ${byId.get(bad.exercise_id)?.name ?? 'an exercise'} should be 1 to 20, or empty.`
      return
    }

    // "new" is reserved: /plans/new is the create page, not a plan.
    const taken = new Set([...plans.map((r) => r.id), 'new'])
    const plan: Plan = {
      id: original?.id ?? idFromName(name, taken),
      name: name.trim(),
      created_by: original?.created_by ?? profile.id,
      exercises: items.map((i) =>
        i.sets.trim() ? { exercise_id: i.exercise_id, suggested_sets: Number(i.sets) } : { exercise_id: i.exercise_id },
      ),
    }
    if (!plan.id) {
      nameError = 'Use at least one letter or number.'
      return
    }
    saving = true
    try {
      await savePlan(plan)
      router.navigate('/plans')
    } catch (err) {
      formError = `Couldn't save. ${(err as Error).message}`
    } finally {
      saving = false
    }
  }

  async function archive() {
    if (!original) return
    saving = true
    try {
      await savePlan({ ...original, archived: true })
      router.navigate('/plans')
    } catch (err) {
      formError = `Couldn't archive. ${(err as Error).message}`
      saving = false
    }
  }
</script>

<div class="page">
  <a class="back" href="/plans">
    <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M15 6l-6 6 6 6" /></svg>
    Plans
  </a>

  <h1 tabindex="-1">{id ? 'Edit plan' : 'New plan'}</h1>

  {#if status === 'missing'}
    <p>That plan doesn't exist. <a href="/plans">Back to plans</a></p>
  {:else if status === 'error'}
    <div role="alert">
      <p>Couldn't load. Check that the server is running.</p>
      <button type="button" class="btn btn-primary" onclick={load}>Try again</button>
    </div>
  {:else if status === 'ready'}
    <form class="form" onsubmit={save} novalidate>
      <div class="field">
        <label for="plan-name">Name</label>
        <input
          id="plan-name"
          type="text"
          bind:value={name}
          maxlength="40"
          placeholder="e.g. Upper body"
          autocomplete="off"
          aria-invalid={nameError ? 'true' : undefined}
          aria-describedby={nameError ? 'plan-name-error' : undefined}
        />
        {#if nameError}<p id="plan-name-error" class="error">{nameError}</p>{/if}
      </div>

      <section class="exercises" aria-labelledby="exercises-title">
        <div class="section-head">
          <h2 id="exercises-title">Exercises</h2>
          <span class="hint">Sets are a suggestion</span>
        </div>

        {#if items.length === 0}
          <p class="empty">No exercises yet.</p>
        {:else}
          <ol class="items">
            {#each items as item, i (item.exercise_id)}
              {@const ex = byId.get(item.exercise_id)}
              <li class="item" id="row-{item.exercise_id}">
                <span class="num" aria-hidden="true">{i + 1}</span>
                {#if ex?.images?.[0]}
                  <img src="/images/{ex.images[0]}" alt="" loading="lazy" width="48" height="48" />
                {/if}
                <span class="text">
                  <span class="name">{ex?.name ?? item.exercise_id}</span>
                  {#if ex}<span class="muscles">{primaryMuscles(ex).map(label).join(', ')}</span>{/if}
                </span>
                <label class="sets">
                  <input type="text" inputmode="numeric" maxlength="2" placeholder="–" bind:value={item.sets} />
                  <span>sets<span class="visually-hidden"> for {ex?.name}</span></span>
                </label>
                <span class="tools">
                  <button
                    type="button"
                    class="btn btn-icon up"
                    aria-label="Move {ex?.name} up"
                    disabled={i === 0}
                    onclick={() => move(i, -1)}
                  >
                    <svg viewBox="0 0 24 24"><path d="M6 15l6-6 6 6" /></svg>
                  </button>
                  <button
                    type="button"
                    class="btn btn-icon down"
                    aria-label="Move {ex?.name} down"
                    disabled={i === items.length - 1}
                    onclick={() => move(i, 1)}
                  >
                    <svg viewBox="0 0 24 24"><path d="M6 9l6 6 6-6" /></svg>
                  </button>
                  <button
                    type="button"
                    class="btn btn-icon"
                    aria-label="Remove {ex?.name}"
                    onclick={() => items.splice(i, 1)}
                  >
                    <svg viewBox="0 0 24 24"><path d="M6 6l12 12 M18 6L6 18" /></svg>
                  </button>
                </span>
              </li>
            {/each}
          </ol>
        {/if}

        <button type="button" class="btn btn-secondary add" onclick={() => (picking = true)}>
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5v14 M5 12h14" /></svg>
          Add exercises
        </button>
      </section>

      {#if formError}<p class="error" role="alert">{formError}</p>{/if}

      {#if confirmingArchive}
        <div class="confirm" role="group" aria-labelledby="archive-text">
          <p id="archive-text">
            Archive “{original?.name}”? It disappears from the list; past workouts and anyone's week that
            uses it keep working.
          </p>
          <div class="actions">
            <button type="button" class="btn btn-secondary" onclick={() => (confirmingArchive = false)}>Keep</button>
            <button type="button" class="btn btn-danger" onclick={archive} disabled={saving}>Archive</button>
          </div>
        </div>
      {:else}
        <div class="actions">
          {#if original}
            <button type="button" class="btn btn-text-danger" onclick={() => (confirmingArchive = true)}>
              Archive
            </button>
            <span class="spacer"></span>
          {/if}
          <a class="btn btn-quiet" href="/plans">Cancel</a>
          <button type="submit" class="btn btn-primary" disabled={saving}>
            {saving ? 'Saving…' : id ? 'Save' : 'Create plan'}
          </button>
        </div>
      {/if}
    </form>

    <ExercisePicker bind:open={picking} {catalog} {selected} ontoggle={toggle} />
  {/if}
</div>

<style>
  .page {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    max-width: 44rem;
  }

  .back {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    align-self: flex-start;
    min-height: var(--touch-target);
    color: var(--color-text-muted);
    font-weight: var(--weight-semibold);
    text-decoration: none;
  }

  @media (hover: hover) {
    .back:hover {
      color: var(--color-text);
    }
  }

  .back svg,
  .add svg {
    width: 1.25rem;
    height: 1.25rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
  }

  h2 {
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
  }

  .form {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .field label {
    font-weight: var(--weight-semibold);
  }

  .field input {
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-surface);
    font-size: var(--text-lg);
  }

  .field input[aria-invalid='true'] {
    border-color: var(--color-danger);
  }

  .exercises {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .section-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
  }

  .hint,
  .empty,
  .muscles {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .empty {
    margin: 0;
  }

  .items {
    display: flex;
    flex-direction: column;
    margin: 0;
    padding: 0;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
    list-style: none;
  }

  .item {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2) var(--space-3);
    padding: var(--space-3);
  }

  .item + .item {
    border-top: 1px solid var(--color-border);
  }

  .num {
    width: 1.5rem;
    color: var(--color-text-muted);
    font-weight: var(--weight-bold);
    text-align: center;
  }

  img {
    flex: none;
    width: 3rem;
    height: 3rem;
    border-radius: var(--radius-sm);
    object-fit: cover;
  }

  .text {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 10rem;
  }

  .name {
    font-weight: var(--weight-semibold);
  }

  .sets {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .sets input {
    width: 2.75rem;
    min-height: var(--touch-target);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-bg);
    font-weight: var(--weight-semibold);
    text-align: center;
  }

  .tools {
    display: flex;
    margin-left: auto;
  }

  .add {
    align-self: flex-start;
  }

  .error {
    margin: 0;
    color: var(--color-danger-text);
  }

  .actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: var(--space-2);
  }

  .spacer {
    flex: 1;
  }

  .confirm {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-4);
    border-radius: var(--radius-md);
    background: var(--color-danger-subtle);
  }

  .confirm p {
    margin: 0;
  }
</style>
