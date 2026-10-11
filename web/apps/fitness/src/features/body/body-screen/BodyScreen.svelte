<!--
  The Body page (/body, decision #49): the muscle map, big, and every
  muscle grouped by how ready it is: when it's ready again, when it was
  last worked, this week's sets, which exercises did it, and whether it's
  getting stronger (#50). Tapping a
  muscle on the map jumps to its row (and outlines it); the list is the
  same information for keyboards and screen readers.
-->
<script lang="ts">
  import { tick } from 'svelte'
  import { LoadError } from '@home-tools/ui/components/load-error'
  import type { Profile } from '@home-tools/ui/profiles/types'
  import {
    changeLabel,
    getRecovery,
    getStrength,
    isStronger,
    MUSCLE_NAMES,
    MUSCLE_REGIONS,
    readyWhen,
    RECOVERY_LEGEND,
    recoveryShades,
    STATE_LABELS,
    workedWhen,
    type MuscleRecovery,
    type MuscleStrength,
    type RecoveryState,
  } from '@/features/body'
  import { MuscleMap } from '@/features/body/muscle-map'
  import type { Region } from '@/features/body/muscle-map/body-model'
  import { getCatalog, type Muscle } from '@/features/exercises'

  let { profile }: { profile: Profile } = $props()

  let recovery = $state<MuscleRecovery[]>([])
  let names = $state(new Map<string, string>()) // exercise ID → name
  let strength = $state(new Map<Muscle, MuscleStrength>())
  let status = $state<'loading' | 'ready' | 'error'>('loading')
  let selected = $state<Muscle | null>(null)

  async function load() {
    status = 'loading'
    try {
      const [r, s, catalog] = await Promise.all([getRecovery(profile.id), getStrength(profile.id), getCatalog()])
      recovery = r
      strength = new Map(s.muscles.map((m) => [m.muscle, m]))
      names = new Map(catalog.map((e) => [e.id, e.name]))
      status = 'ready'
    } catch (err) {
      console.error('Loading recovery failed:', err)
      status = 'error'
    }
  }
  load()

  const GROUPS: RecoveryState[] = ['rest', 'recovering', 'ready']
  // Within a group: the soonest ready first, then the most recently worked,
  // then by name, so the order means something.
  let groups = $derived(
    GROUPS.map((state) => ({
      state,
      // Untrained muscles get one line under Ready instead of a row each.
      muscles: recovery
        .filter((m) => m.state === state && (state !== 'ready' || m.last_worked))
        .toSorted(
          (a, b) =>
            (a.ready_at ?? '').localeCompare(b.ready_at ?? '') ||
            (b.last_worked ?? '').localeCompare(a.last_worked ?? '') ||
            MUSCLE_NAMES[a.muscle].localeCompare(MUSCLE_NAMES[b.muscle]),
        ),
    })).filter((g) => g.muscles.length > 0),
  )
  let untrained = $derived(
    recovery
      .filter((m) => !m.last_worked)
      .map((m) => MUSCLE_NAMES[m.muscle])
      .toSorted((a, b) => a.localeCompare(b)),
  )

  let selectedRegions = $derived<Region[]>(selected ? MUSCLE_REGIONS[selected] : [])

  /** A tap on the map: the first muscle drawn there, its row into view. */
  async function pick(region: Region) {
    const muscle = (Object.keys(MUSCLE_REGIONS) as Muscle[]).find((m) => MUSCLE_REGIONS[m].includes(region))
    if (!muscle) return
    selected = muscle
    await tick()
    const row = document.getElementById(`muscle-${muscle}`)
    row?.focus({ preventScroll: true })
    row?.scrollIntoView({ behavior: matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth', block: 'center' })
  }

  // Secondary muscles count half a set, so totals can be fractional:
  // people think in whole sets.
  function sets(n: number) {
    const whole = Math.max(1, Math.round(n))
    return `${whole} ${whole === 1 ? 'set' : 'sets'} this week`
  }
</script>

<div class="body-page">
  <div class="intro">
    <h1 tabindex="-1">Your body</h1>
    <p class="muted">
      An estimate from your last week of workouts: every set tires the muscles it works, and they recover over two to three
      days. A guide, not a rule: how you feel counts too.
    </p>
  </div>

  {#if status === 'error'}
    <LoadError what="your muscles" onretry={load} />
  {:else if status === 'loading'}
    <p class="muted" role="status">Loading…</p>
  {:else}
    <div class="layout">
      <div class="map">
        <MuscleMap shades={recoveryShades(recovery)} legend={RECOVERY_LEGEND} selected={selectedRegions} onpick={pick} />
      </div>

      <div class="groups">
        {#each groups as group (group.state)}
          <section aria-labelledby="group-{group.state}">
            <h2 id="group-{group.state}" class="group-title">
              <span class="dot {group.state}" aria-hidden="true"></span>
              {STATE_LABELS[group.state]}
              <span class="count">{group.muscles.length}</span>
            </h2>
            <ul class="list">
              {#each group.muscles as m (m.muscle)}
                {@const trend = strength.get(m.muscle)}
                <li id="muscle-{m.muscle}" class="row" class:selected={selected === m.muscle} tabindex="-1">
                  <div class="name-line">
                    <span class="name">{MUSCLE_NAMES[m.muscle]}</span>
                    {#if m.ready_at}<span class="ready-at">Ready {readyWhen(m.ready_at)}</span>{/if}
                  </div>
                  <p class="muted detail">
                    {#if m.last_worked}
                      Worked {workedWhen(m.last_worked)} · {sets(m.sets_7d)}
                    {:else}
                      Not worked this week
                    {/if}
                  </p>
                  {#if trend}
                    <p class="detail strength" class:up={isStronger(trend.change)}>
                      Strength {changeLabel(trend.change, trend.basis).replace(/^Holding/, 'holding')}
                    </p>
                  {/if}
                  {#if m.exercises.length > 0}
                    <p class="muted detail">
                      From {m.exercises.map((id) => names.get(id) ?? id).join(', ')}
                    </p>
                  {/if}
                </li>
              {/each}
            </ul>
          </section>
        {/each}
        {#if untrained.length > 0}
          <section aria-labelledby="group-untrained">
            <h2 id="group-untrained" class="group-title">
              <span class="dot ready" aria-hidden="true"></span>
              Not worked this week
              <span class="count">{untrained.length}</span>
            </h2>
            <p class="muted">{untrained.join(', ')}</p>
          </section>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .body-page {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
  }

  .intro {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    max-width: 40rem;
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
  }

  /* Phones: map on top, list below. Wide: side by side, the map stays put
   * while the list scrolls. */
  .layout {
    display: grid;
    gap: var(--space-6);
    align-items: start;
  }

  .map {
    width: 100%;
    max-width: 24rem;
    margin: 0 auto;
  }

  @media (min-width: 56rem) {
    .layout {
      grid-template-columns: 2fr 3fr;
      gap: var(--space-7);
    }

    .map {
      position: sticky;
      top: var(--space-5);
    }
  }

  .groups {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
  }

  .group-title {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
  }

  .count {
    color: var(--color-text-muted);
    font-weight: var(--weight-semibold);
  }

  .dot {
    width: 0.875rem;
    height: 0.875rem;
    border-radius: var(--radius-full);
  }

  .dot.rest {
    background: var(--color-muscle-strong);
  }

  .dot.recovering {
    background: var(--color-muscle-medium);
  }

  .dot.ready {
    background: var(--color-body);
  }

  .list {
    display: flex;
    flex-direction: column;
    margin: 0;
    padding: 0;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
    list-style: none;
  }

  .row {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-3) var(--space-4);
    scroll-margin: var(--space-6);
    transition: background-color var(--duration-normal) var(--ease-out);
  }

  .row + .row {
    border-top: 1px solid var(--color-border);
  }

  .row.selected {
    background: var(--color-accent-subtle);
  }

  .row:focus-visible {
    outline-offset: -2px;
  }

  .name-line {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-1) var(--space-3);
  }

  .name {
    font-weight: var(--weight-semibold);
  }

  .ready-at {
    color: var(--color-accent-text);
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
  }

  .detail {
    font-size: var(--text-sm);
  }

  .strength {
    font-weight: var(--weight-semibold);
  }

  .strength.up {
    color: var(--color-accent-text);
  }
</style>
