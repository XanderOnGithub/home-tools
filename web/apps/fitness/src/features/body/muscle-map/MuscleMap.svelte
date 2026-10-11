<!--
  The body, front and back, with each muscle colored by how worked it is
  (decision #49): needs rest = the person's accent, recovering = half of
  it, ready = the plain figure. Decorative for screen readers: whoever
  uses it always shows the same states as text next to it.

  `onpick` (optional) makes the muscles clickable with a pointer; the
  text list beside the map is the keyboard and screen-reader way to the
  same details.
-->
<script lang="ts">
  import { STATE_LABELS, type RecoveryState } from '@/features/body'
  import { BACK, FRONT, type Region } from './body-model'

  let {
    states,
    selected = [],
    onpick,
  }: {
    states: Partial<Record<Region, RecoveryState>>
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
              class={muscle ? (states[shape.muscle] ?? 'ready') : 'figure'}
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
  {#each ['rest', 'recovering', 'ready'] as const as state (state)}
    <li><span class="swatch {state}"></span>{STATE_LABELS[state]}</li>
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

  .figure,
  .ready {
    fill: var(--color-body);
    background: var(--color-body);
  }

  .recovering {
    fill: var(--color-muscle-recovering);
    background: var(--color-muscle-recovering);
  }

  .rest {
    fill: var(--color-muscle-rest);
    background: var(--color-muscle-rest);
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
