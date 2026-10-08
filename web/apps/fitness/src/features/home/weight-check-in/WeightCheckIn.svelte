<!--
  Weekly weight check-in (decision #28). Pre-filled with the last weight,
  so "no change" is one tap. Skip records nothing (no invented data point)
  and just hides the card until next week.
-->
<script lang="ts">
  import { untrack } from 'svelte'
  import { Card } from '@/components/card'
  import { isoWeek, today } from '@/dates'
  import {
    saveFitnessProfile,
    saveWeight,
    type FitnessProfile,
    type WeightEntry,
  } from '@/features/fitness-profile'
  import type { Profile } from '@/features/profiles/types'
  import { kgToLb, lbToKg, parseNumber } from '@/units'

  let {
    profile,
    fitness = $bindable(),
    weights,
    onsaved,
  }: {
    profile: Profile
    fitness: FitnessProfile
    weights: WeightEntry[]
    onsaved: (entry: WeightEntry) => void
  } = $props()

  let imperial = $derived(profile.units === 'imperial')
  let unit = $derived(imperial ? 'lb' : 'kg')
  let last = $derived(weights.at(-1))
  const toDisplay = (kg: number) => String(Math.round((imperial ? kgToLb(kg) : kg) * 10) / 10)

  // Pre-filled once with the last weight; after that it's the user's input.
  let value = $state(untrack(() => (last ? toDisplay(last.weight_kg) : '')))
  let error = $state('')
  let busy = $state(false)

  async function save(event: SubmitEvent) {
    event.preventDefault()
    const n = parseNumber(value)
    const kg = imperial ? lbToKg(n) : n
    if (!(kg >= 20 && kg <= 400)) {
      error = `Enter your weight in ${unit}.`
      return
    }
    busy = true
    error = ''
    try {
      // 0.01 kg precision, so a weight in lb reads back exactly as typed.
      onsaved(await saveWeight(profile.id, { date: today(), weight_kg: Math.round(kg * 100) / 100 }))
    } catch (err) {
      error = `Couldn't save. ${(err as Error).message}`
    } finally {
      busy = false
    }
  }

  async function skip() {
    busy = true
    error = ''
    try {
      fitness = await saveFitnessProfile({ ...fitness, weight_prompt_skipped: isoWeek(new Date()) })
    } catch (err) {
      error = `Couldn't skip. ${(err as Error).message}`
    } finally {
      busy = false
    }
  }
</script>

<Card title="Weekly check-in">
  <form class="form" onsubmit={save} novalidate>
    <label class="field">
      <span>Your weight this week</span>
      <span class="input">
        <input type="text" inputmode="decimal" maxlength="5" bind:value autocomplete="off" />
        <span class="unit">{unit}</span>
      </span>
    </label>
    {#if last}
      <p class="muted">Last: {toDisplay(last.weight_kg)} {unit}</p>
    {/if}
    {#if error}
      <p class="error" role="alert">{error}</p>
    {/if}
    <div class="actions">
      <button type="button" class="button quiet" onclick={skip} disabled={busy}>Skip this week</button>
      <button type="submit" class="button primary" disabled={busy}>Save</button>
    </div>
  </form>
</Card>

<style>
  .form {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    font-weight: var(--weight-medium);
  }

  .input {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: fit-content;
    min-height: var(--touch-target);
    padding: 0 var(--space-4);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-bg);
  }

  .input:focus-within {
    outline: var(--focus-ring);
    outline-offset: var(--focus-offset);
  }

  .input input {
    width: 5ch;
    padding: 0;
    border: none;
    background: none;
    font-size: var(--text-lg);
    font-weight: var(--weight-semibold);
  }

  .input input:focus-visible {
    outline: none;
  }

  .unit,
  .muted {
    color: var(--color-text-muted);
  }

  p {
    margin: 0;
  }

  .error {
    color: var(--color-danger-text);
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
  }

  .button {
    min-height: var(--touch-target);
    padding: 0 var(--space-5);
    border: none;
    border-radius: var(--radius-full);
    font-weight: var(--weight-semibold);
    cursor: pointer;
    transition: background var(--duration-fast) var(--ease-out);
  }

  .primary {
    background: var(--color-accent);
    color: var(--color-on-accent);
  }

  .quiet {
    background: none;
    color: var(--color-text-muted);
  }

  @media (hover: hover) {
    .primary:hover {
      background: var(--color-accent-hover);
    }
    .quiet:hover {
      color: var(--color-text);
    }
  }

  .primary:active {
    background: var(--color-accent-pressed);
  }

  .button:disabled {
    cursor: progress;
    opacity: 0.7;
  }
</style>
