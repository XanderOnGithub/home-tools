<!--
  An on/off switch: a button with role="switch", so screen readers say
  "Daily poll, switch, on". The track fills with the accent when on, and
  the knob slides to the other side (position, not only color, says
  which). `busy` blocks double taps while a save runs. Shared by every
  tool (moved from discord, decision #57).
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
    <span class="knob"></span>
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

  /* Busy (saving): blocks taps without dimming, so a quick save doesn't
   * flash. */
  .switch:disabled {
    cursor: progress;
  }

  .track {
    display: flex;
    align-items: center;
    width: 2.75rem;
    height: 1.625rem;
    padding: 0.1875rem;
    border-radius: var(--radius-full);
    background: var(--color-border-strong);
    transition: background var(--duration-normal) var(--ease-out);
  }

  .knob {
    width: 1.25rem;
    height: 1.25rem;
    border-radius: var(--radius-full);
    background: #fff; /* on the track in both themes */
    box-shadow: 0 1px 2px rgb(0 0 0 / 0.25);
    transition:
      translate var(--duration-normal) var(--ease-out),
      width var(--duration-fast) var(--ease-out);
  }

  [aria-checked='true'] .track {
    background: var(--color-accent);
  }

  [aria-checked='true'] .knob {
    translate: 1.125rem 0;
  }

  @media (hover: hover) {
    .switch[aria-checked='false']:hover:not(:disabled) .track {
      background: var(--color-text-muted);
    }

    .switch[aria-checked='true']:hover:not(:disabled) .track {
      background: var(--color-accent-hover);
    }
  }

  /* Pressed: the knob stretches toward where it's going, like iOS. */
  .switch:active:not(:disabled) .knob {
    width: 1.5rem;
  }

  [aria-checked='true']:active:not(:disabled) .knob {
    translate: 0.875rem 0;
  }
</style>
