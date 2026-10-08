<!--
  Rest countdown between sets. Time is computed from a fixed end
  timestamp, not by counting ticks, so it stays right even if the phone
  sleeps or the tab is in the background. Screen readers get "Rest
  started"/"Rest over" from the parent's live region rather than a
  per-second announcement.
-->
<script lang="ts">
  let {
    endsAt,
    total,
    now,
    onadjust,
    onskip,
  }: {
    endsAt: number // epoch ms
    total: number // seconds the rest started with (for the progress ring)
    now: number // epoch ms, ticked by the parent
    onadjust: (deltaSec: number) => void
    onskip: () => void
  } = $props()

  let left = $derived(Math.max(0, Math.ceil((endsAt - now) / 1000)))
  let clock = $derived(`${Math.floor(left / 60)}:${String(left % 60).padStart(2, '0')}`)
  // Ring: full at the start, empty at zero. r=45 → circumference ≈ 282.7.
  const CIRC = 2 * Math.PI * 45
  let progress = $derived(total > 0 ? left / total : 0)
</script>

<div class="rest" role="timer" aria-label="Rest timer, {clock} left">
  <svg class="ring" viewBox="0 0 100 100" aria-hidden="true">
    <circle class="track" cx="50" cy="50" r="45" />
    <circle class="bar" cx="50" cy="50" r="45" stroke-dasharray={CIRC} stroke-dashoffset={CIRC * (1 - progress)} />
  </svg>
  <div class="center">
    <span class="label">Rest</span>
    <span class="clock" aria-hidden="true">{clock}</span>
  </div>
  <div class="controls">
    <button type="button" class="btn btn-secondary" onclick={() => onadjust(-15)}>−15 s</button>
    <button type="button" class="btn btn-primary" onclick={onskip}>Skip rest</button>
    <button type="button" class="btn btn-secondary" onclick={() => onadjust(15)}>+15 s</button>
  </div>
</div>

<style>
  .rest {
    display: grid;
    grid-template-areas: 'ring' 'controls';
    justify-items: center;
    gap: var(--space-4);
  }

  .ring,
  .center {
    grid-area: ring;
    width: 11rem;
    height: 11rem;
  }

  .ring {
    rotate: -90deg; /* start the ring at 12 o'clock */
  }

  circle {
    fill: none;
    stroke-width: 6;
  }

  .track {
    stroke: var(--color-border);
  }

  .bar {
    stroke: var(--color-accent);
    stroke-linecap: round;
    transition: stroke-dashoffset 250ms linear;
  }

  .center {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }

  .label {
    color: var(--color-text-muted);
    font-weight: var(--weight-semibold);
  }

  .clock {
    font-size: var(--text-3xl);
    font-weight: var(--weight-extrabold);
    font-variant-numeric: tabular-nums;
  }

  .controls {
    grid-area: controls;
    display: flex;
    gap: var(--space-2);
  }

  @media (prefers-reduced-motion: reduce) {
    .bar {
      transition: none;
    }
  }
</style>
