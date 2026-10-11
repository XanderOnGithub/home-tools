<!--
  What a workout or a week works (decision #51): the muscle map shaded by
  sets, the same in words (grouped by how much), and the notes (`checks`).
  One workout: light 1–2 sets, solid 3–5, a lot 6+. A week: 1–4, 5–9,
  10+ sets (about 10 a week builds a muscle).
-->
<script lang="ts">
  import { MUSCLE_NAMES, regionShades, type LegendItem, type Shade } from '@/features/body'
  import { MuscleMap } from '@/features/body/muscle-map'
  import type { Muscle } from '@/features/exercises'
  import { level, type Check, type Load } from '@/features/training'

  let { load, checks, mode }: { load: Load; checks: Check[]; mode: 'workout' | 'week' } = $props()

  const GROUPS: Record<'workout' | 'week', { shade: Exclude<Shade, 0>; label: string }[]> = {
    workout: [
      { shade: 3, label: 'A lot' },
      { shade: 2, label: 'Solid' },
      { shade: 1, label: 'Light' },
    ],
    week: [
      { shade: 3, label: '10+ sets' },
      { shade: 2, label: '5–9 sets' },
      { shade: 1, label: '1–4 sets' },
    ],
  }

  const weekShade = (sets: number): Shade => (sets >= 10 ? 3 : sets >= 5 ? 2 : sets > 0 ? 1 : 0)
  let shadeOf = $derived(mode === 'week' ? weekShade : (sets: number) => level(sets) as Shade)

  let muscleShades = $derived(
    Object.fromEntries([...load].map(([m, sets]) => [m, shadeOf(sets)])) as Partial<Record<Muscle, Shade>>,
  )
  let legend = $derived<LegendItem[]>([...GROUPS[mode], { shade: 0, label: 'Not worked' }])
  // Most worked first within a group.
  let groups = $derived(
    GROUPS[mode]
      .map((g) => ({
        ...g,
        muscles: [...load]
          .filter(([, sets]) => shadeOf(sets) === g.shade)
          .toSorted((a, b) => b[1] - a[1])
          .map(([m]) => MUSCLE_NAMES[m]),
      }))
      .filter((g) => g.muscles.length > 0),
  )
</script>

<div class="frame">
<div class="summary">
  <div class="map"><MuscleMap shades={regionShades(muscleShades)} {legend} /></div>

  <div class="words">
    {#if groups.length === 0}
      <p class="muted">Add exercises to see which muscles they work.</p>
    {:else}
      <dl class="groups">
        {#each groups as g (g.shade)}
          <div>
            <dt><span class="dot shade-{g.shade}" aria-hidden="true"></span>{g.label}</dt>
            <dd>{g.muscles.join(', ')}</dd>
          </div>
        {/each}
      </dl>
    {/if}

    {#if checks.length > 0}
      <ul class="checks">
        {#each checks as c (c.text)}
          <li class={c.tone}>
            <svg viewBox="0 0 24 24" aria-hidden="true">
              {#if c.tone === 'warn'}
                <path d="M12 4l9 16H3z M12 10v4 M12 17h.01" />
              {:else}
                <path d="M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18z M12 11v5 M12 8h.01" />
              {/if}
            </svg>
            <span><span class="visually-hidden">{c.tone === 'warn' ? 'Worth fixing: ' : 'Good to know: '}</span>{c.text}</span>
          </li>
        {/each}
      </ul>
      <a class="why" href="/guide#sets">How many sets is enough?</a>
    {/if}
  </div>
</div>
</div>

<style>
  /* Side by side only when the summary itself has room (it sits in a
   * narrow column on the Plans page), not when the screen does. */
  .frame {
    container-type: inline-size;
  }

  .summary {
    display: grid;
    gap: var(--space-5);
    align-items: start;
  }

  @container (min-width: 34rem) {
    .summary {
      grid-template-columns: minmax(12rem, 2fr) 3fr;
    }
  }

  .map {
    width: 100%;
    max-width: 18rem;
    margin: 0 auto;
  }

  .map :global(svg) {
    max-height: 14rem;
  }

  .words {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
  }

  .groups {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
  }

  .groups div {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1) var(--space-3);
  }

  dt {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-width: 6rem;
    font-weight: var(--weight-semibold);
  }

  dd {
    margin: 0;
    color: var(--color-text-muted);
  }

  .dot {
    width: 0.75rem;
    height: 0.75rem;
    border-radius: var(--radius-full);
  }

  .shade-1 {
    background: var(--color-muscle-light);
  }

  .shade-2 {
    background: var(--color-muscle-medium);
  }

  .shade-3 {
    background: var(--color-muscle-strong);
  }

  .checks {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .checks li {
    display: flex;
    gap: var(--space-2);
    font-size: var(--text-sm);
    line-height: var(--leading-normal);
  }

  .checks svg {
    flex: none;
    width: 1.25rem;
    height: 1.25rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  /* Warnings get the accent (red is only for errors, DESIGN.md §2); the
   * icon and the hidden "Worth fixing" say it without color too. */
  .warn svg {
    color: var(--color-accent-text);
  }

  .info {
    color: var(--color-text-muted);
  }

  .why {
    align-self: flex-start;
    color: var(--color-accent-text);
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
  }
</style>
