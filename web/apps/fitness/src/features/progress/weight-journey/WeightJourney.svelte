<!-- The weight log as a chart, with a time range (1M / 3M / 1Y / All). -->
<script lang="ts">
  import { LineChart } from '@/components/line-chart'
  import type { WeightEntry } from '@/features/fitness-profile'
  import { toWeight, weightUnit, type Units } from '@/features/progress'

  let { weights, units }: { weights: WeightEntry[]; units: Units } = $props()

  const RANGES = [
    { id: '1m', label: '1M', name: 'Last month', months: 1 },
    { id: '3m', label: '3M', name: 'Last 3 months', months: 3 },
    { id: '1y', label: '1Y', name: 'Last year', months: 12 },
    { id: 'all', label: 'All', name: 'All time', months: Infinity },
  ] as const
  let range = $state<(typeof RANGES)[number]['id']>('3m')

  // Dates are local days ("2026-10-09"); noon keeps them on that day anywhere.
  let all = $derived(
    weights
      .map((w) => ({ t: Date.parse(`${w.date}T12:00:00`), v: toWeight(w.weight_kg, units) }))
      .sort((a, b) => a.t - b.t),
  )
  let points = $derived.by(() => {
    const months = RANGES.find((r) => r.id === range)!.months
    if (months === Infinity) return all
    const since = new Date()
    since.setMonth(since.getMonth() - months)
    return all.filter((p) => p.t >= since.getTime())
  })
  const format = (v: number) => `${Math.round(v * 10) / 10} ${weightUnit(units)}`
</script>

<section class="card" aria-labelledby="weight-title">
  <div class="head">
    <h2 id="weight-title">Weight</h2>
    {#if all.length > 0}
      <fieldset class="ranges">
        <legend class="visually-hidden">Time range</legend>
        {#each RANGES as r (r.id)}
          <label>
            <input class="visually-hidden" type="radio" name="weight-range" value={r.id} bind:group={range} />
            <span aria-hidden="true">{r.label}</span>
            <span class="visually-hidden">{r.name}</span>
          </label>
        {/each}
      </fieldset>
    {/if}
  </div>

  {#if all.length === 0}
    <p class="muted">No weights logged yet. Home asks once a week.</p>
  {:else if points.length === 0}
    <p class="muted">Nothing logged in this time. Last: {format(all.at(-1)!.v)}.</p>
  {:else}
    <LineChart {points} {format} label="Body weight" />
  {/if}
</section>

<style>
  .card {
    padding: var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
  }

  h2 {
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
  }

  .muted {
    margin: 0;
    color: var(--color-text-muted);
  }

  .head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
  }

  .ranges {
    display: flex;
    gap: var(--space-1);
    margin: 0;
    padding: var(--space-1);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-full);
  }

  .ranges label {
    display: grid;
    place-items: center;
    min-width: var(--touch-target);
    min-height: calc(var(--touch-target) - 2 * var(--space-1));
    padding: 0 var(--space-2);
    border-radius: var(--radius-full);
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
    cursor: pointer;
    transition: background var(--duration-fast) var(--ease-out);
  }

  .ranges label:has(input:checked) {
    background: var(--color-accent);
    color: var(--color-on-accent);
  }

  .ranges label:has(input:focus-visible) {
    outline: var(--focus-ring);
    outline-offset: var(--focus-offset);
  }

  @media (hover: hover) {
    .ranges label:not(:has(input:checked)):hover {
      background: var(--color-accent-subtle);
    }
  }
</style>
