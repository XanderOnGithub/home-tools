<!--
  A small, hand-rolled line chart over time (decision #39: no chart
  library until one is needed). Drawn at the box's real pixel width, so
  text and lines stay crisp at any size. The readout above it shows one
  point: the latest by default; tap or drag on the chart (or focus it and
  use the arrow keys) to look at another. It's a slider to screen readers,
  each value read as "Oct 9, 2026: 110 kg".
-->
<script lang="ts">
  type Point = { t: number; v: number } // t: ms since epoch; v: display units

  let {
    points,
    format,
    label,
  }: {
    points: Point[] // oldest → newest
    format: (v: number) => string
    label: string // what's measured, e.g. "Heaviest set"
  } = $props()

  const HEIGHT = 176
  const PAD = { top: 12, right: 12, bottom: 28, left: 52 } // room for axis labels

  let width = $state(0)
  let picked = $state<number | null>(null) // index into points; null = latest
  let index = $derived(Math.min(picked ?? points.length - 1, points.length - 1))
  // New data (another range, another exercise): back to the latest point.
  $effect.pre(() => {
    void points
    picked = null
  })

  // Value axis: about three "nice" steps (1, 2, 2.5 or 5 × 10^n) around the data.
  let ticks = $derived.by(() => {
    const vs = points.map((p) => p.v)
    let lo = Math.min(...vs)
    let hi = Math.max(...vs)
    if (lo === hi) [lo, hi] = [lo - Math.max(1, lo * 0.1), hi + Math.max(1, hi * 0.1)]
    const raw = (hi - lo) / 3
    const mag = 10 ** Math.floor(Math.log10(raw))
    const step = [1, 2, 2.5, 5, 10].map((m) => m * mag).find((s) => s >= raw)!
    const out: number[] = []
    for (let v = Math.floor(lo / step) * step; v <= hi + step * 0.001; v += step) out.push(v)
    if (out.at(-1)! < hi) out.push(out.at(-1)! + step)
    return out.map((v) => Math.max(0, v)).filter((v, i, a) => a.indexOf(v) === i)
  })

  let x = $derived.by(() => {
    const t0 = points[0].t
    const span = points.at(-1)!.t - t0
    const w = width - PAD.left - PAD.right
    // One point (or all on one day): in the middle.
    return (t: number) => PAD.left + (span > 0 ? ((t - t0) / span) * w : w / 2)
  })
  let y = $derived.by(() => {
    const lo = ticks[0]
    const hi = ticks.at(-1)!
    const h = HEIGHT - PAD.top - PAD.bottom
    return (v: number) => PAD.top + h - ((v - lo) / (hi - lo || 1)) * h
  })

  let line = $derived(points.map((p, i) => `${i ? 'L' : 'M'}${x(p.t).toFixed(1)},${y(p.v).toFixed(1)}`).join(' '))

  const date = (t: number, year = true) =>
    new Date(t).toLocaleDateString(undefined, { month: 'short', day: 'numeric', ...(year && { year: 'numeric' }) })
  let current = $derived(points[index])
  let readout = $derived(`${date(current.t)}: ${format(current.v)}`)

  // The point nearest to where the finger or pointer is, along time.
  function pick(event: PointerEvent) {
    const box = (event.currentTarget as SVGElement).getBoundingClientRect()
    const px = event.clientX - box.left
    let nearest = 0
    points.forEach((p, i) => {
      if (Math.abs(x(p.t) - px) < Math.abs(x(points[nearest].t) - px)) nearest = i
    })
    picked = nearest
  }

  function onpointerdown(event: PointerEvent) {
    ;(event.currentTarget as SVGElement).setPointerCapture(event.pointerId)
    pick(event)
  }

  function onpointermove(event: PointerEvent) {
    // Mouse: follow hover. Touch/pen: only while pressed (dragging).
    if (event.pointerType === 'mouse' || event.buttons) pick(event)
  }

  function onkeydown(event: KeyboardEvent) {
    const last = points.length - 1
    const next = { ArrowLeft: index - 1, ArrowDown: index - 1, ArrowRight: index + 1, ArrowUp: index + 1, Home: 0, End: last }[
      event.key
    ]
    if (next === undefined) return
    event.preventDefault()
    picked = Math.max(0, Math.min(last, next))
  }
</script>

<div class="chart">
  <p class="readout">
    <span class="value">{format(current.v)}</span>
    <span class="when">{date(current.t)}</span>
  </p>
  <div class="plot" bind:clientWidth={width}>
    {#if width > 0}
      <svg
        width={width}
        height={HEIGHT}
        role="slider"
        tabindex="0"
        aria-label={label}
        aria-valuemin={0}
        aria-valuemax={points.length - 1}
        aria-valuenow={index}
        aria-valuetext={readout}
        {onpointerdown}
        {onpointermove}
        {onkeydown}
      >
        {#each ticks as v (v)}
          <line class="grid" x1={PAD.left} x2={width - PAD.right} y1={y(v)} y2={y(v)} />
          <text class="axis" x={PAD.left - 8} y={y(v)} text-anchor="end" dominant-baseline="middle">{format(v)}</text>
        {/each}
        <text class="axis" x={PAD.left} y={HEIGHT - 6} text-anchor="start">{date(points[0].t, false)}</text>
        {#if points.length > 1}
          <text class="axis" x={width - PAD.right} y={HEIGHT - 6} text-anchor="end">{date(points.at(-1)!.t, false)}</text>
        {/if}

        <line class="cursor" x1={x(current.t)} x2={x(current.t)} y1={PAD.top} y2={HEIGHT - PAD.bottom} />
        <path class="line" d={line} />
        {#each points as p, i (p.t)}
          <circle class="dot" class:current={i === index} cx={x(p.t)} cy={y(p.v)} r={i === index ? 6 : 3.5} />
        {/each}
      </svg>
    {/if}
  </div>
</div>

<style>
  .chart {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .readout {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    margin: 0;
  }

  .value {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  .when {
    color: var(--color-text-muted);
    font-weight: var(--weight-medium);
  }

  .plot {
    min-height: 11rem; /* the chart's height, so nothing jumps while it measures */
  }

  svg {
    display: block;
    overflow: visible;
    touch-action: pan-y; /* a horizontal drag picks points; vertical still scrolls */
    cursor: crosshair;
    border-radius: var(--radius-sm);
  }

  svg:focus-visible {
    outline: var(--focus-ring);
    outline-offset: var(--focus-offset);
  }

  .grid {
    stroke: var(--color-border);
    stroke-width: 1;
  }

  .axis {
    fill: var(--color-text-muted);
    font-size: var(--text-xs);
  }

  .cursor {
    stroke: var(--color-border-strong);
    stroke-width: 1;
    stroke-dasharray: 3 3;
  }

  .line {
    fill: none;
    stroke: var(--color-accent);
    stroke-width: 2.5;
    stroke-linejoin: round;
    stroke-linecap: round;
  }

  .dot {
    fill: var(--color-accent);
    stroke: var(--color-surface);
    stroke-width: 2;
  }

  .dot.current {
    stroke-width: 3;
  }
</style>
