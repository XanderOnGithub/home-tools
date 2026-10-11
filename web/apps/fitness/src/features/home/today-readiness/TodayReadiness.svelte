<!--
  Today's plan against the body's recovery (decision #52), on the Today
  card. With a plan: what it works hard that still needs rest (and a
  plan to do instead, if one fits), what's still recovering, or that
  everything's ready. Without a routine: which muscles are ready and a
  plan that suits them. Rest days get nothing: rest is the plan.
  Needs the catalog for the plans' muscles; shows nothing until it's
  loaded (or if it fails: the card works without it).
-->
<script lang="ts">
  import { muscleList, readyWhen, type MuscleRecovery } from '@/features/body'
  import { getCatalog, type Exercise } from '@/features/exercises'
  import type { Plan } from '@/features/plans'
  import { bestPlan, listNames, planLoad, planReadiness, readyGroups } from '@/features/training'

  let {
    plan,
    plans,
    recovery,
    starting,
    onstart,
  }: {
    plan: Plan | undefined // today's; undefined = suggest one (no routine)
    plans: Plan[]
    recovery: MuscleRecovery[]
    starting: boolean
    onstart: (plan: Plan) => void
  } = $props()

  let catalog = $state<Exercise[] | null>(null)
  getCatalog()
    .then((c) => (catalog = c))
    .catch((err) => console.error('Loading the catalog for today failed:', err))
  let byId = $derived(new Map((catalog ?? []).map((e) => [e.id, e])))

  let readiness = $derived(plan && catalog ? planReadiness(planLoad(plan.exercises, byId), recovery) : null)
  let instead = $derived(catalog ? bestPlan(plans, byId, recovery, plan?.id) : undefined)
  // When everything that needs rest is ready again: the latest of them.
  let readyBy = $derived(
    readiness?.rest.map((r) => r.ready_at ?? '').reduce((a, b) => (a > b ? a : b), '') ?? '',
  )
  let trained = $derived(recovery.some((r) => r.last_worked))
  const verb = (list: MuscleRecovery[], one: string, many: string) => (list.length === 1 ? one : many)
</script>

{#if catalog && plan && readiness}
  {#if readiness.rest.length > 0}
    <div class="note warn">
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 4l9 16H3z M12 10v4 M12 17h.01" /></svg>
      <div>
        <p>
          {muscleList(readiness.rest.map((r) => r.muscle))}
          {verb(readiness.rest, 'still needs', 'still need')} rest{#if readyBy}: ready {readyWhen(readyBy)}{/if}.
        </p>
        {#if instead}
          <button type="button" class="btn btn-secondary" disabled={starting} onclick={() => instead && onstart(instead)}>
            Start {instead.name} instead
          </button>
        {:else}
          <p class="muted">Go lighter on {readiness.rest.length === 1 ? 'it' : 'them'} today: fewer sets, or stop a few reps short.</p>
        {/if}
      </div>
    </div>
  {:else if readiness.recovering.length > 0}
    <p class="note">
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18z M12 11v5 M12 8h.01" /></svg>
      <span>
        {muscleList(readiness.recovering.map((r) => r.muscle))}
        {verb(readiness.recovering, 'is', 'are')} still recovering: fine to train, maybe a set less.
      </span>
    </p>
  {:else if trained}
    <p class="note ok">
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12.5l4.5 4.5L19 7" /></svg>
      <span>Everything this plan works is ready.</span>
    </p>
  {/if}
{:else if catalog && !plan && trained}
  {@const groups = readyGroups(recovery)}
  {#if groups.length > 0}
    <div class="note">
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12.5l4.5 4.5L19 7" /></svg>
      <div>
        <p>Ready to train: {listNames(groups)}.</p>
        {#if instead}
          <button type="button" class="btn btn-secondary" disabled={starting} onclick={() => instead && onstart(instead)}>
            Start {instead.name}
          </button>
        {/if}
      </div>
    </div>
  {/if}
{/if}

<style>
  .note {
    display: flex;
    align-items: flex-start;
    gap: var(--space-2);
    margin: var(--space-2) 0 0;
    line-height: var(--leading-normal);
  }

  .note > div {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-3);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
  }

  svg {
    flex: none;
    width: 1.25rem;
    height: 1.25rem;
    margin-top: 0.1rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .warn svg,
  .ok svg {
    color: var(--color-accent-text);
  }
</style>
