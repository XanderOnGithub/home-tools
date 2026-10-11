<!--
  The body, front and back, with each muscle shaded by how worked it is:
  0 = the plain figure up to 3 = the person's full accent (recovery #49,
  a plan's load #51), with a legend saying what the shades mean.
  Decorative for screen readers: whoever uses it always shows the same
  thing as text next to it.

  `onpick` (optional) makes the muscles clickable with a pointer; the
  text list beside the map is the keyboard and screen-reader way to the
  same details.
-->
<script lang="ts">
  import type { LegendItem, Shade } from '@/features/body'
  import { BACK, FRONT, type Region } from './body-model'

  let {
    shades,
    legend,
    selected = [],
    onpick,
  }: {
    shades: Partial<Record<Region, Shade>>
    legend: LegendItem[]
    selected?: Region[] // outlined, e.g. the muscle whose details are open
    onpick?: (region: Region) => void
  } = $props()

  // The figure's own parts: drawn, never muscles.
  const NOT_MUSCLES = new Set<Region>(['head', 'knees'])
  const VIEWS = [
    { label: 'Front', shapes: FRONT },
    { label: 'Back', shapes: BACK },
  ]
</script>

<div class="wrap" aria-hidden="true">
<div class="map">
  {#each VIEWS as view (view.label)}
    <figure>
      <svg viewBox="0 0 100 200">
        {#each view.shapes as shape, i (`${shape.muscle}-${i}`)}
          {@const muscle = !NOT_MUSCLES.has(shape.muscle)}
          {#each shape.svgPoints as points, j (j)}
            <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions:
                 pointer shortcut only; the list next to the map is the accessible way -->
            <polygon
              {points}
              class={muscle ? `shade-${shades[shape.muscle] ?? 0}` : 'shade-0'}
              class:selected={selected.includes(shape.muscle)}
              class:pickable={muscle && !!onpick}
              onclick={muscle && onpick ? () => onpick(shape.muscle) : undefined}
            />
          {/each}
        {/each}
      </svg>
      <figcaption>{view.label}</figcaption>
    </figure>
  {/each}
</div>
<ul class="legend">
  {#each legend as item (item.label)}
    <li><span class="swatch shade-{item.shade}"></span>{item.label}</li>
  {/each}
</ul>
</div>

<style>
  .wrap {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .map {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3);
  }

  figure {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-1);
    margin: 0;
  }

  svg {
    width: 100%;
    max-height: 22rem;
    overflow: visible;
  }

  figcaption {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
  }

  polygon {
    stroke: var(--color-bg);
    stroke-width: 0.4;
    stroke-linejoin: round;
    transition: fill var(--duration-normal) var(--ease-out);
  }

  .shade-0 {
    fill: var(--color-body);
    background: var(--color-body);
  }

  .shade-1 {
    fill: var(--color-muscle-light);
    background: var(--color-muscle-light);
  }

  .shade-2 {
    fill: var(--color-muscle-medium);
    background: var(--color-muscle-medium);
  }

  .shade-3 {
    fill: var(--color-muscle-strong);
    background: var(--color-muscle-strong);
  }

  .legend {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: var(--space-2) var(--space-4);
    margin: 0;
    padding: 0;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    list-style: none;
  }

  .legend li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .swatch {
    width: 0.875rem;
    height: 0.875rem;
    border-radius: var(--radius-full);
  }

  .selected {
    stroke: var(--color-text);
    stroke-width: 1;
  }

  .pickable {
    cursor: pointer;
  }

  @media (hover: hover) {
    .pickable:hover {
      stroke: var(--color-text);
      stroke-width: 0.8;
    }
  }
</style>
