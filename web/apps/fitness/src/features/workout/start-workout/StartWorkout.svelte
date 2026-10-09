<!--
  "Start a workout": any plan, or an empty workout that gets its exercises
  as you go. For days the routine doesn't cover (rest days, a second
  workout, trying another plan). A native modal <dialog>, like the
  exercise picker.
-->
<script lang="ts">
  import type { Profile } from '@/features/profiles/types'
  import type { Plan } from '@/features/plans'
  import { startSession } from '@/features/sessions'
  import { router } from '@/router'

  let {
    open = $bindable(false),
    profile,
    plans,
    todayPlanId,
  }: {
    open?: boolean
    profile: Profile
    plans: Plan[]
    todayPlanId?: string // marked "Today" in the list
  } = $props()

  let dialog: HTMLDialogElement
  let starting = $state<string | null>(null) // plan ID, or '' for empty
  let error = $state('')

  $effect(() => {
    if (open && !dialog.open) {
      error = ''
      starting = null
      dialog.showModal()
    } else if (!open && dialog.open) {
      dialog.close()
    }
  })

  let choices = $derived(plans.filter((p) => !p.archived))

  async function start(plan: Plan | null) {
    starting = plan?.id ?? ''
    error = ''
    try {
      const s = await startSession(profile.id, plan?.id, plan?.exercises.map((e) => e.exercise_id) ?? [])
      open = false
      router.navigate(`/workout/${s.id}`)
    } catch (err) {
      error = `Couldn't start. ${(err as Error).message}`
      starting = null
    }
  }
</script>

<dialog bind:this={dialog} class="dialog" aria-labelledby="start-title" onclose={() => (open = false)}>
  <div class="top">
    <h2 id="start-title">Start a workout</h2>
    <button type="button" class="btn btn-icon" aria-label="Close" onclick={() => (open = false)}>
      <svg viewBox="0 0 24 24"><path d="M6 6l12 12 M18 6L6 18" /></svg>
    </button>
  </div>

  <ul class="choices">
    <li>
      <button type="button" class="choice" onclick={() => start(null)} disabled={starting !== null}>
        <span class="plus" aria-hidden="true">
          <svg viewBox="0 0 24 24"><path d="M12 5v14 M5 12h14" /></svg>
        </span>
        <span class="text">
          <span class="name">{starting === '' ? 'Starting…' : 'Empty workout'}</span>
          <span class="meta">Add exercises as you go</span>
        </span>
      </button>
    </li>
    {#each choices as plan (plan.id)}
      <li>
        <button type="button" class="choice" onclick={() => start(plan)} disabled={starting !== null}>
          <span class="text">
            <span class="name">{starting === plan.id ? 'Starting…' : plan.name}</span>
            <span class="meta">
              {plan.exercises.length}
              {plan.exercises.length === 1 ? 'exercise' : 'exercises'}
            </span>
          </span>
          {#if plan.id === todayPlanId}<span class="today">Today</span>{/if}
        </button>
      </li>
    {/each}
  </ul>

  {#if choices.length === 0}
    <p class="hint">No plans yet. <a href="/plans/new" onclick={() => (open = false)}>Create a plan</a> to reuse a list of exercises.</p>
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
</dialog>

<style>
  .dialog {
    width: min(28rem, calc(100vw - 2 * var(--space-3)));
    max-height: min(40rem, calc(100dvh - 2 * var(--space-3)));
  }

  .top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-2) var(--space-2) var(--space-5);
  }

  h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  .choices {
    flex: 1;
    overflow-y: auto;
    margin: 0;
    padding: 0 var(--space-2) var(--space-3);
    list-style: none;
  }

  .choice {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    min-height: var(--touch-target);
    padding: var(--space-3);
    border: none;
    border-radius: var(--radius-md);
    background: none;
    color: inherit;
    text-align: left;
    cursor: pointer;
  }

  .choice:disabled {
    cursor: progress;
  }

  @media (hover: hover) {
    .choice:hover:not(:disabled) {
      background: var(--color-surface);
    }
  }

  .choice:active:not(:disabled) {
    background: var(--color-accent-subtle);
  }

  .plus {
    display: grid;
    flex: none;
    place-items: center;
    width: 2.5rem;
    height: 2.5rem;
    border-radius: var(--radius-full);
    background: var(--color-accent-subtle);
    color: var(--color-accent-text);
  }

  .plus svg {
    width: 1.25rem;
    height: 1.25rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.5;
    stroke-linecap: round;
  }

  .text {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
  }

  .name {
    font-weight: var(--weight-semibold);
  }

  .meta {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .today {
    flex: none;
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-full);
    background: var(--color-accent-subtle);
    color: var(--color-accent-text);
    font-size: var(--text-xs);
    font-weight: var(--weight-bold);
  }

  .hint,
  .error {
    margin: 0;
    padding: 0 var(--space-5) var(--space-4);
  }

  .hint {
    color: var(--color-text-muted);
  }

  .hint a {
    color: var(--color-accent-text);
    font-weight: var(--weight-semibold);
  }

  .error {
    color: var(--color-danger-text);
  }
</style>
