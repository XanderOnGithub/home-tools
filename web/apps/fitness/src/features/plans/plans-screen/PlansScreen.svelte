<!--
  Plans: this person's routine (which plan on which weekday), then the
  household's shared list of plans (decisions #30, #33). Anyone can create
  or edit a plan; each person's routine decides which plan falls on which day.
-->
<script lang="ts">
  import { api } from '@/api'
  import type { FitnessProfile } from '@/features/fitness-profile'
  import type { Profile } from '@/features/profiles/types'
  import { getPlans, type Plan } from '@/features/plans'
  import { WeekPlanner } from '@/features/plans/week-planner'

  let { fitness = $bindable() }: { fitness: FitnessProfile } = $props()

  let plans = $state<Plan[]>([])
  let people = $state<Profile[]>([])
  let status = $state<'loading' | 'ready' | 'error'>('loading')

  async function load() {
    status = 'loading'
    try {
      ;[plans, people] = await Promise.all([getPlans(), api.get<Profile[]>('/api/users')])
      status = 'ready'
    } catch (err) {
      console.error('Loading plans failed:', err)
      status = 'error'
    }
  }
  load()

  let active = $derived(plans.filter((r) => !r.archived))

  const author = (r: Plan) => people.find((p) => p.id === r.created_by)?.name
  // Days of this person's week that use r, in week order ("Tue, Sat").
  const WEEK = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'] as const
  const myDays = (r: Plan) =>
    WEEK.filter((day) => fitness.schedule?.[day] === r.id).map((day) => day[0].toUpperCase() + day.slice(1, 3))
</script>

<div class="page">
  <div class="title-row">
    <h1 tabindex="-1">Plans</h1>
    <a class="btn btn-primary" href="/plans/new">New plan</a>
  </div>

  {#if status === 'error'}
    <div role="alert">
      <p>Couldn't load plans. Check that the server is running.</p>
      <button type="button" class="btn btn-primary" onclick={load}>Try again</button>
    </div>
  {:else if status === 'ready'}
    <div class="grid">
      <section class="list" aria-labelledby="list-title">
        <h2 id="list-title">All plans</h2>
        {#if active.length === 0}
          <p class="muted">No plans yet. Create one, then add it to your routine.</p>
        {:else}
          <ul>
            {#each active as r (r.id)}
              <li>
                <a class="plan" href="/plans/{r.id}">
                  <span class="name">{r.name}</span>
                  <span class="meta">
                    {[`${r.exercises.length} ${r.exercises.length === 1 ? 'exercise' : 'exercises'}`, author(r) && `by ${author(r)}`]
                      .filter(Boolean)
                      .join(' · ')}
                  </span>
                  {#if myDays(r).length > 0}
                    <span class="days">On your {myDays(r).join(', ')}</span>
                  {/if}
                </a>
              </li>
            {/each}
          </ul>
        {/if}
      </section>

      <WeekPlanner bind:fitness {plans} />
    </div>
  {/if}
</div>

<style>
  .page {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  .title-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
  }

  h2 {
    margin-bottom: var(--space-3);
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
  }

  .grid {
    display: grid;
    gap: var(--space-6);
    align-items: start;
  }

  /* Wide screens: the list, with your routine beside it. */
  @media (min-width: 56rem) {
    .grid {
      grid-template-columns: 3fr 2fr;
      gap: var(--space-7);
    }
  }

  ul {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  /* The whole row is the link to the editor. */
  .plan {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
    color: inherit;
    text-decoration: none;
    transition: border-color var(--duration-fast) var(--ease-out);
  }

  @media (hover: hover) {
    .plan:hover {
      border-color: var(--color-border-strong);
    }
  }

  .name {
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
  }

  .meta,
  .muted {
    color: var(--color-text-muted);
  }

  .days {
    color: var(--color-accent-text);
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
  }

  p {
    margin: 0 0 var(--space-3);
  }
</style>
