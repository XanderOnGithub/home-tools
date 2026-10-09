<!--
  Weekly weight check-in (decision #28): one slim row. Pre-filled with the
  last weight, so "no change" is one tap. Skip records nothing (no
  invented data point) and just hides it until next week.
-->
<script lang="ts">
  import { untrack } from 'svelte'
  import { isoWeek, today } from '@home-tools/ui/dates'
  import {
    saveFitnessProfile,
    saveWeight,
    type FitnessProfile,
    type WeightEntry,
  } from '@/features/fitness-profile'
  import type { Profile } from '@home-tools/ui/profiles/types'
  import { kgToLb, lbToKg, parseNumber } from '@home-tools/ui/units'

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

<form class="check-in" onsubmit={save} novalidate aria-labelledby="check-in-label">
  <label class="label" id="check-in-label" for="check-in-weight">Weekly weight</label>
  <!-- Input and buttons wrap as one unit, never apart. -->
  <span class="controls">
    <span class="input">
      <input id="check-in-weight" type="text" inputmode="decimal" maxlength="5" bind:value autocomplete="off" />
      <span class="unit">{unit}</span>
    </span>
    <button type="submit" class="btn btn-primary" disabled={busy}>Save</button>
    <button type="button" class="btn btn-quiet" onclick={skip} disabled={busy}>Skip</button>
  </span>
  {#if error}
    <p class="error" role="alert">{error}</p>
  {/if}
</form>

<style>
  .check-in {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2) var(--space-3);
    padding: var(--space-3) var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
  }

  .label {
    flex: 1;
    min-width: 8rem;
    font-weight: var(--weight-semibold);
  }

  .input {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
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
    font-weight: var(--weight-semibold);
    text-align: right;
  }

  .input input:focus-visible {
    outline: none;
  }

  .unit {
    color: var(--color-text-muted);
  }

  .controls {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .error {
    flex-basis: 100%;
    margin: 0;
    color: var(--color-danger-text);
  }
</style>
