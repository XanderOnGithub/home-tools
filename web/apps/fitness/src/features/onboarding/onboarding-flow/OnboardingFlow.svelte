<!--
  First-run setup for a profile's fitness: one question per screen
  (height → weight), then saves the workout profile and the first weight
  entry. Both are required (an estimate is fine for weight). Units follow
  the profile: ft + in and lb, or cm and kg.

  Accessibility: each step's heading takes focus when the step appears, so
  screen readers announce the new question; "Step 2 of 3" is real text.
-->
<script lang="ts">
  import { tick } from 'svelte'
  import { prefersReducedMotion } from 'svelte/motion'
  import { fly } from 'svelte/transition'
  import { today } from '@/dates'
  import { saveFitnessProfile, saveWeight, type FitnessProfile } from '@/features/fitness-profile'
  import { ProfileAvatar } from '@/features/profiles/profile-avatar'
  import type { Profile } from '@/features/profiles/types'
  import { cmToM, ftInToM, lbToKg, parseNumber } from '@/units'

  let { profile, oncomplete }: { profile: Profile; oncomplete: (fp: FitnessProfile) => void } =
    $props()

  const STEPS = 2

  let step = $state(0)
  let cm = $state('')
  let ft = $state('')
  let inches = $state('')
  let weight = $state('')
  let error = $state('')
  let saving = $state(false)
  let heading = $state<HTMLHeadingElement>()

  let imperial = $derived(profile.units === 'imperial')

  // Move focus to the new question so it's announced and Tab starts there.
  async function goTo(next: number) {
    error = ''
    step = next
    await tick()
    heading?.focus()
  }

  /** Height in meters from the inputs, or NaN if missing or invalid. */
  function heightM(): number {
    if (!imperial) return cmToM(parseNumber(cm))
    if (!ft.trim()) return NaN
    const f = parseNumber(ft)
    const i = inches.trim() ? parseNumber(inches) : 0 // "5 ft" alone means 5'0"
    return f >= 0 && i >= 0 && i < 12 ? ftInToM(f, i) : NaN
  }

  /** Weight in kg from the input, or NaN if missing or invalid. */
  function weightKg(): number {
    const n = parseNumber(weight)
    return imperial ? lbToKg(n) : n
  }

  function next(event: SubmitEvent) {
    event.preventDefault()
    if (step === 0) {
      const m = heightM()
      if (!(m >= 0.5 && m <= 2.75)) {
        error = imperial ? 'Enter feet and inches, like 5 and 10.' : 'Enter your height in cm, like 175.'
        return
      }
    }
    if (step === 1) {
      const kg = weightKg()
      if (!(kg >= 20 && kg <= 400)) {
        error = imperial ? 'Enter your weight in lb, like 165.' : 'Enter your weight in kg, like 75.'
        return
      }
      finish()
      return
    }
    goTo(step + 1)
  }

  async function finish() {
    saving = true
    error = ''
    const fp: FitnessProfile = {
      user_id: profile.id,
      height_m: Math.round(heightM() * 1000) / 1000, // millimeter precision
    }
    try {
      // Weight first: if the profile saved but the weight failed, a retry
      // would skip onboarding and lose the weight.
      // 0.01 kg precision, so a weight in lb reads back exactly as typed.
      await saveWeight(profile.id, { date: today(), weight_kg: Math.round(weightKg() * 100) / 100 })
      oncomplete(await saveFitnessProfile(fp))
    } catch (err) {
      error = `Couldn't save. ${(err as Error).message}`
    } finally {
      saving = false
    }
  }

  const slide = () => ({ x: 24, duration: prefersReducedMotion.current ? 0 : 200 })
</script>

<main class="onboarding">
  <div class="intro">
    <span class="avatar"><ProfileAvatar id={profile.id} /></span>
    <p class="step-count">Step {step + 1} of {STEPS}</p>
    <div class="progress" aria-hidden="true">
      {#each { length: STEPS } as _, i (i)}
        <span class="progress-dot" class:done={i <= step}></span>
      {/each}
    </div>
  </div>

  <form class="form" onsubmit={next} novalidate>
    {#key step}
      <div class="step" in:fly={slide()}>
        {#if step === 0}
          <h1 bind:this={heading} tabindex="-1">How tall are you?</h1>
          <div class="inputs">
            {#if imperial}
              <label class="unit-input">
                <span class="visually-hidden">Height, feet</span>
                <input class="digits-1" type="text" inputmode="numeric" maxlength="1" bind:value={ft} placeholder="5" autocomplete="off" />
                <span>ft</span>
              </label>
              <label class="unit-input">
                <span class="visually-hidden">Height, inches</span>
                <input class="digits-2" type="text" inputmode="numeric" maxlength="2" bind:value={inches} placeholder="10" autocomplete="off" />
                <span>in</span>
              </label>
            {:else}
              <label class="unit-input">
                <span class="visually-hidden">Height</span>
                <input class="digits-3" type="text" inputmode="numeric" maxlength="3" bind:value={cm} placeholder="175" autocomplete="off" />
                <span>cm</span>
              </label>
            {/if}
          </div>
        {:else}
          <h1 bind:this={heading} tabindex="-1">What do you weigh today?</h1>
          <div class="inputs">
            <label class="unit-input">
              <span class="visually-hidden">Weight</span>
              <input class="digits-5" type="text" inputmode="decimal" maxlength="5" bind:value={weight} placeholder={imperial ? '165' : '75'} autocomplete="off" aria-describedby="weight-hint" />
              <span>{imperial ? 'lb' : 'kg'}</span>
            </label>
          </div>
          <p class="hint" id="weight-hint">Not sure? A rough estimate is fine. You can update it at the weekly check-in.</p>
        {/if}
      </div>
    {/key}

    {#if error}
      <p class="error" role="alert">{error}</p>
    {/if}

    <div class="actions">
      {#if step > 0}
        <button type="button" class="btn btn-quiet" onclick={() => goTo(step - 1)}>Back</button>
      {/if}
      <span class="spacer"></span>
      <button type="submit" class="btn btn-primary" disabled={saving}>
        {#if step < STEPS - 1}Continue{:else if saving}Saving…{:else}Let's go{/if}
      </button>
    </div>
  </form>
</main>

<style>
  /* Phones: content from the top, buttons pinned to the bottom (thumb
   * reach). Wider screens: the whole step is one group, centered, with the
   * buttons right under the inputs (see the media query at the end). */
  .onboarding {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
    width: min(30rem, 100%);
    min-height: 100dvh;
    margin: 0 auto;
    padding: var(--space-6) var(--space-4);
  }

  .intro {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-3);
  }

  .avatar {
    width: 4.5rem;
    height: 4.5rem;
  }

  .step-count {
    margin: 0;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
  }

  .progress {
    display: flex;
    gap: var(--space-2);
  }

  .progress-dot {
    width: 2rem;
    height: 0.375rem;
    border-radius: var(--radius-full);
    background: var(--color-border);
    transition: background var(--duration-normal) var(--ease-out);
  }

  .progress-dot.done {
    background: var(--color-accent);
  }

  .form {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: var(--space-5);
  }

  .step {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
    text-align: center;
  }

  /* The heading takes focus for screen readers; no ring needed on text. */
  h1:focus {
    outline: none;
  }

  .inputs {
    display: flex;
    justify-content: center;
    gap: var(--space-3);
  }

  /* Number + unit inside one box; the whole label is clickable. */
  .unit-input {
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 6.5rem;
    gap: var(--space-2);
    min-height: 3.5rem;
    padding: 0 var(--space-4);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-surface);
    cursor: text;
  }

  .unit-input:focus-within {
    outline: var(--focus-ring);
    outline-offset: var(--focus-offset);
  }

  /* Sized to what they hold (1 to 5 characters), so the boxes stay small
   * and the unit sits right next to the number. */
  .unit-input input {
    padding: 0;
    border: none;
    text-align: center;
    background: none;
    font-size: var(--text-xl);
    font-weight: var(--weight-semibold);
  }

  .unit-input input::placeholder {
    color: var(--color-text-muted);
    font-weight: var(--weight-normal);
  }

  .digits-1 {
    width: 1.5ch;
  }
  .digits-2 {
    width: 2.5ch;
  }
  .digits-3 {
    width: 3.5ch;
  }
  .digits-5 {
    width: 5.5ch;
  }

  /* The box shows the focus ring instead (above). */
  .unit-input input:focus-visible {
    outline: none;
  }

  .unit-input span:not(.visually-hidden) {
    color: var(--color-text-muted);
    font-weight: var(--weight-medium);
  }

  .hint {
    margin: 0;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    text-align: center;
  }

  .error {
    margin: 0;
    color: var(--color-danger-text);
    text-align: center;
  }

  .actions {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-top: auto;
  }

  .spacer {
    flex: 1;
  }

  @media (min-width: 40rem) {
    .onboarding {
      justify-content: center;
    }

    .form {
      flex: none;
    }

    .actions {
      margin-top: var(--space-3);
    }

    .avatar {
      width: 5.5rem;
      height: 5.5rem;
    }

    h1 {
      font-size: var(--text-3xl);
    }
  }
</style>
