<!--
  A profile's avatar: their blob, in their color, with eyes that blink and
  glance around (and the blob leans a little when they do). Shape, face
  and timing all come from the ID, so each person is consistent everywhere.
  Decorative: the surrounding UI always shows the name, so it's aria-hidden.
  Color comes from the nearest [data-accent] (--color-accent).

  When `id` changes (e.g. live preview while typing a name), the shape and
  eyes morph smoothly to the new ones instead of jumping.
-->
<script lang="ts">
  import { untrack } from 'svelte'
  import { cubicOut } from 'svelte/easing'
  import { prefersReducedMotion, Tween } from 'svelte/motion'
  import { blobFace, blobPath } from '@/features/profiles/blob'

  let { id }: { id: string } = $props()

  const MORPH_MS = 400

  // Every blob path has the same structure (M + 24 curves), so morphing is
  // just tweening its numbers. Tween interpolates arrays and objects of
  // numbers element by element.
  const numbers = (d: string) => d.match(/-?\d+(?:\.\d+)?/g)!.map(Number)

  let face = $derived(blobFace(id)) // timing uses this directly: no tweening
  // Start at the first ID's look (untrack: only the starting value is
  // wanted here); the effect below morphs to every later ID.
  const shape = new Tween(untrack(() => numbers(blobPath(id))))
  const eyes = new Tween(untrack(() => blobFace(id)))

  $effect(() => {
    const duration = prefersReducedMotion.current ? 0 : MORPH_MS
    shape.set(numbers(blobPath(id)), { duration, easing: cubicOut })
    eyes.set(blobFace(id), { duration, easing: cubicOut })
  })

  // Rebuild "M x y C x y x y x y ... Z" from the tweened numbers.
  let path = $derived.by(() => {
    const n = shape.current.map((v) => v.toFixed(2))
    let d = `M ${n[0]} ${n[1]}`
    for (let i = 2; i < n.length; i += 6) {
      d += ` C ${n.slice(i, i + 6).join(' ')}`
    }
    return d + ' Z'
  })
</script>

<svg
  class="avatar"
  viewBox="0 0 100 100"
  aria-hidden="true"
  style:--blink-sec="{face.blinkSec}s"
  style:--glance-sec="{face.glanceSec}s"
  style:--offset-sec="-{face.offsetSec}s"
>
  <g class="body">
    <path class="skin" d={path} />
    <g transform="translate({eyes.current.x} {eyes.current.y}) rotate({eyes.current.tilt})">
      <g class="look">
        {#each [-1, 1] as side (side)}
          <g transform="translate({side * eyes.current.gap} 0)">
            <ellipse class="eye" rx={eyes.current.rx} ry={eyes.current.ry} />
          </g>
        {/each}
      </g>
    </g>
  </g>
</svg>

<style>
  .avatar {
    display: block;
    width: 100%;
    height: 100%;
    overflow: visible; /* the wobble may nudge a few units past the box */
  }

  .skin {
    fill: var(--color-accent);
  }

  /* Simple dark ovals. Same in light and dark mode: they're the
   * character, not UI chrome, so they don't follow the theme. */
  .eye {
    fill: var(--color-avatar-eye);
  }

  /* Transforms on SVG parts pivot around their own center. */
  .body,
  .look,
  .eye {
    transform-box: fill-box;
    transform-origin: center;
  }

  /* All motion is opt-out: with reduced motion the face simply stays still. */
  @media (prefers-reduced-motion: no-preference) {
    .eye {
      animation: blink var(--blink-sec) var(--offset-sec) infinite;
    }

    /* The eyes glance as a pair; glances and wobble share one clock, so
     * the body leans with the eyes. */
    .look {
      animation: glance var(--glance-sec) var(--offset-sec) infinite ease-in-out;
    }
    .body {
      animation: wobble var(--glance-sec) var(--offset-sec) infinite ease-in-out;
    }
  }

  /* A quick close-open near the end of each loop. */
  @keyframes blink {
    0%,
    94%,
    100% {
      scale: 1 1;
    }
    97% {
      scale: 1 0.1;
    }
  }

  /* Mostly still: one glance left, one up-right, long rests between.
   * Holds dominate so the eyes rest far more than they move. */
  @keyframes glance {
    0%,
    30%,
    100% {
      translate: 0 0;
    }
    34%,
    44% {
      translate: -2.6px 0.4px;
    }
    48%,
    76% {
      translate: 0 0;
    }
    80%,
    88% {
      translate: 2.2px -2px;
    }
    92% {
      translate: 0 0;
    }
  }

  /* A small lean toward wherever the eyes look. */
  @keyframes wobble {
    0%,
    30%,
    48%,
    76%,
    92%,
    100% {
      rotate: 0deg;
      translate: 0 0;
    }
    34%,
    44% {
      rotate: -2deg;
      translate: -0.6px 0;
    }
    80%,
    88% {
      rotate: 1.5deg;
      translate: 0.5px -0.5px;
    }
  }
</style>
