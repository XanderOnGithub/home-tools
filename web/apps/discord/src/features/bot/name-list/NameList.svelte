<!--
  The names the bot picks from, one a day in a shuffled order (every name
  gets a day before any repeats). Plain, friendly names: Jim, Larry, Pam.
  At least one must stay; Discord allows 32 characters.
-->
<script lang="ts">
  import { tick } from 'svelte'
  import type { Config } from '@/features/bot'

  let { config, save }: { config: Config; save: (change: (c: Config) => void) => Promise<string> } = $props()

  const MAX = 32

  let name = $state('')
  let busy = $state(false)
  let error = $state('')
  let input = $state<HTMLInputElement>()

  let lower = $derived(new Set(config.names.map((n) => n.toLowerCase()))) // O(1) duplicate check
  let problem = $derived.by(() => {
    const n = name.trim()
    if (n.length > MAX) return `At most ${MAX} characters.`
    if (lower.has(n.toLowerCase())) return `${n} is already on the list.`
    return ''
  })

  async function add(event: SubmitEvent) {
    event.preventDefault()
    const n = name.trim()
    if (!n || problem || busy) return
    busy = true
    error = await save((c) => c.names.push(n))
    busy = false
    if (!error) {
      name = ''
      input?.focus()
    }
  }

  async function remove(n: string, i: number) {
    busy = true
    error = await save((c) => (c.names = c.names.filter((x) => x !== n)))
    busy = false
    // Keep focus in the list: on the name that took this one's place.
    await tick()
    const buttons = document.querySelectorAll<HTMLButtonElement>('.names .remove')
    ;(buttons[Math.min(i, buttons.length - 1)] ?? input)?.focus()
  }
</script>

<section class="section" aria-labelledby="names-title">
  <div class="head">
    <h2 id="names-title">Names</h2>
    <span class="muted count">{config.names.length}</span>
  </div>
  <p class="muted">One a day, shuffled, so everyone gets a turn. The color cycles through green, blue, orange and purple.</p>

  <ul class="names">
    {#each config.names as n, i (n)}
      <li class="chip">
        <span>{n}</span>
        {#if config.names.length > 1}
          <button type="button" class="remove" aria-label="Remove {n}" disabled={busy} onclick={() => remove(n, i)}>
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M7 7l10 10 M17 7L7 17" /></svg>
          </button>
        {/if}
      </li>
    {/each}
  </ul>

  <form class="add" onsubmit={add}>
    <label class="visually-hidden" for="new-name">New name</label>
    <input
      id="new-name"
      bind:this={input}
      bind:value={name}
      placeholder="Add a name, e.g. Norm"
      autocomplete="off"
      enterkeyhint="done"
      aria-invalid={problem ? 'true' : undefined}
      aria-describedby={problem ? 'name-problem' : undefined}
    />
    <button type="submit" class="btn btn-secondary" disabled={busy || !name.trim() || !!problem}>Add</button>
  </form>
  {#if problem}<p class="error" id="name-problem">{problem}</p>{/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
</section>

<style>
  .section {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .head {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
  }

  h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
  }

  .count {
    font-weight: var(--weight-semibold);
  }

  .error {
    color: var(--color-danger-text);
  }

  .names {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .chip {
    display: inline-flex;
    align-items: center;
    min-height: var(--touch-target);
    padding-left: var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-full);
    background: var(--color-surface);
    font-weight: var(--weight-medium);
  }

  /* A chip without a remove button (the last name) keeps even padding. */
  .chip:not(:has(.remove)) {
    padding-right: var(--space-4);
  }

  .remove {
    display: grid;
    place-items: center;
    width: var(--touch-target);
    height: var(--touch-target);
    padding: 0;
    border: none;
    border-radius: var(--radius-full);
    background: none;
    color: var(--color-text-muted);
    cursor: pointer;
  }

  @media (hover: hover) {
    .remove:hover {
      color: var(--color-text);
    }
  }

  .remove:active {
    color: var(--color-danger-text);
  }

  .remove:disabled {
    cursor: progress;
    opacity: 0.7;
  }

  .remove svg {
    width: 1rem;
    height: 1rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.5;
    stroke-linecap: round;
  }

  .add {
    display: flex;
    gap: var(--space-2);
  }

  input {
    flex: 1;
    min-width: 0;
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-bg);
  }
</style>
