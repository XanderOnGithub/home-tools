<!--
  An on/off switch: a button with role="switch", so screen readers say
  "Daily poll, switch, on". The track fills with the accent when on, and
  the knob moves (with its own check mark, so color isn't the only
  signal). `busy` blocks double taps while a save runs.
-->
<script lang="ts">
  let {
    checked,
    label,
    busy = false,
    onchange,
  }: { checked: boolean; label: string; busy?: boolean; onchange: (on: boolean) => void } = $props()
</script>

<button
  type="button"
  class="switch"
  role="switch"
  aria-checked={checked}
  aria-label={label}
  aria-busy={busy}
  disabled={busy}
  onclick={() => onchange(!checked)}
>
  <span class="track" aria-hidden="true">
    <span class="knob">
      <svg viewBox="0 0 24 24"><path d="M6 12.5l4 4 8-9" /></svg>
    </span>
  </span>
</button>

<style>
  /* The button is a full touch target; the track inside is smaller. */
  .switch {
    display: grid;
    place-items: center;
    min-width: var(--touch-target);
    min-height: var(--touch-target);
    padding: 0;
    border: none;
    border-radius: var(--radius-full);
    background: none;
    cursor: pointer;
  }

  .switch:disabled {
    cursor: progress;
    opacity: 0.7;
  }

  .track {
    display: flex;
    align-items: center;
    width: 3.25rem;
    height: 2rem;
    padding: 0.1875rem;
    border-radius: var(--radius-full);
    background: var(--color-border-strong);
    transition: background var(--duration-fast) var(--ease-out);
  }

  .knob {
    display: grid;
    place-items: center;
    width: 1.625rem;
    height: 1.625rem;
    border-radius: var(--radius-full);
    background: var(--color-bg);
    transition: translate var(--duration-fast) var(--ease-out);
  }

  .knob svg {
    width: 1rem;
    height: 1rem;
    fill: none;
    stroke: var(--color-accent);
    stroke-width: 3;
    stroke-linecap: round;
    stroke-linejoin: round;
    opacity: 0;
    transition: opacity var(--duration-fast) var(--ease-out);
  }

  [aria-checked='true'] .track {
    background: var(--color-accent);
  }

  [aria-checked='true'] .knob {
    translate: 1.25rem 0;
  }

  [aria-checked='true'] .knob svg {
    opacity: 1;
  }

  @media (hover: hover) {
    .switch[aria-checked='false']:hover:not(:disabled) .track {
      background: var(--color-text-muted);
    }

    .switch[aria-checked='true']:hover:not(:disabled) .track {
      background: var(--color-accent-hover);
    }
  }

  .switch:active:not(:disabled) .knob {
    scale: 0.9;
  }
</style>
