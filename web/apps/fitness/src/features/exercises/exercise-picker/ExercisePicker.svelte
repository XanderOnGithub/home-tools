<!--
  Search the catalog and add exercises. A native modal <dialog> (focus,
  Esc, background hidden from screen readers), nearly full-screen on
  phones. Each result is a toggle button (aria-pressed): pressed = in the
  routine. It stays open, so several exercises can be added in a row.

  Searching 876 names on each keystroke is O(n) and instant; only the
  first PAGE results are rendered, with "Show more", to keep the DOM light.
-->
<script lang="ts">
  import {
    EQUIPMENT,
    MUSCLES,
    label,
    primaryMuscles,
    type Equipment,
    type Exercise,
    type Muscle,
  } from '@/features/exercises'

  let {
    open = $bindable(false),
    catalog,
    selected,
    ontoggle,
  }: {
    open?: boolean
    catalog: Exercise[]
    selected: Set<string> // exercise IDs already in the routine
    ontoggle: (exercise: Exercise) => void
  } = $props()

  const PAGE = 40

  let dialog: HTMLDialogElement
  let query = $state('')
  let muscle = $state<Muscle | ''>('')
  let equipment = $state<Equipment | ''>('')
  let shown = $state(PAGE)

  $effect(() => {
    if (open && !dialog.open) {
      query = ''
      muscle = ''
      equipment = ''
      shown = PAGE
      dialog.showModal()
    } else if (!open && dialog.open) {
      dialog.close()
    }
  })

  // Every word must appear somewhere in the name: "bench db" finds
  // "Dumbbell Bench Press" only if both match, so "db" doesn't (no alias
  // magic; keep it predictable).
  let results = $derived.by(() => {
    const words = query.toLowerCase().split(/\s+/).filter(Boolean)
    return catalog.filter(
      (ex) =>
        !ex.archived &&
        words.every((w) => ex.name.toLowerCase().includes(w)) &&
        (!muscle || (ex.activation[muscle] ?? 0) >= 1) && // primary muscle only
        (!equipment || !!ex.equipment?.includes(equipment)),
    )
  })

  // New search or filter: start from the top of the results again.
  $effect(() => {
    void query, muscle, equipment
    shown = PAGE
  })
</script>

<dialog bind:this={dialog} class="dialog" aria-labelledby="picker-title" onclose={() => (open = false)}>
  <div class="top">
    <div class="title-row">
      <h2 id="picker-title">Add exercises</h2>
      <button type="button" class="btn btn-primary" onclick={() => (open = false)}>
        Done{selected.size ? ` (${selected.size})` : ''}
      </button>
    </div>

    <label class="search">
      <span class="visually-hidden">Search exercises</span>
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M11 18a7 7 0 1 1 0-14 7 7 0 0 1 0 14z M20 20l-4-4" /></svg>
      <input type="search" bind:value={query} placeholder="Search exercises" autocomplete="off" />
    </label>

    <div class="filters">
      <label>
        <span class="visually-hidden">Muscle</span>
        <select bind:value={muscle}>
          <option value="">All muscles</option>
          {#each MUSCLES as m (m)}<option value={m}>{label(m)}</option>{/each}
        </select>
      </label>
      <label>
        <span class="visually-hidden">Equipment</span>
        <select bind:value={equipment}>
          <option value="">All equipment</option>
          {#each EQUIPMENT as e (e)}<option value={e}>{label(e)}</option>{/each}
        </select>
      </label>
    </div>

    <p class="count" role="status">{results.length} {results.length === 1 ? 'exercise' : 'exercises'}</p>
  </div>

  <ul class="results">
    {#each results.slice(0, shown) as ex (ex.id)}
      {@const added = selected.has(ex.id)}
      <li>
        <button type="button" class="result" class:added aria-pressed={added} onclick={() => ontoggle(ex)}>
          {#if ex.images?.[0]}
            <img src="/images/{ex.images[0]}" alt="" loading="lazy" width="56" height="56" />
          {:else}
            <span class="img-placeholder" aria-hidden="true"></span>
          {/if}
          <span class="text">
            <span class="name">{ex.name}</span>
            <span class="muscles">{primaryMuscles(ex).map(label).join(', ')}</span>
          </span>
          <span class="state" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d={added ? 'M5 12l5 5 9-10' : 'M12 5v14 M5 12h14'} /></svg>
          </span>
        </button>
      </li>
    {/each}
  </ul>

  {#if results.length > shown}
    <div class="more">
      <button type="button" class="btn btn-secondary" onclick={() => (shown += PAGE)}>
        Show more ({results.length - shown} left)
      </button>
    </div>
  {:else if results.length === 0}
    <p class="empty">No exercises match. Try fewer words or clear a filter.</p>
  {/if}
</dialog>

<style>
  .dialog {
    width: min(40rem, calc(100vw - 2 * var(--space-3)));
    height: min(48rem, calc(100dvh - 2 * var(--space-3)));
    padding: 0;
    border: none;
    border-radius: var(--radius-lg);
    background: var(--color-surface-raised);
    color: var(--color-text);
    box-shadow: var(--shadow-md);
  }

  .dialog[open] {
    display: flex;
    flex-direction: column;
  }

  .dialog::backdrop {
    background: rgb(0 0 0 / 0.45);
  }

  /* Search and filters stay put; only the results scroll. */
  .top {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-4) var(--space-4) var(--space-2);
    border-bottom: 1px solid var(--color-border);
  }

  .title-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
  }

  h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  .search {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-full);
    background: var(--color-bg);
  }

  .search:focus-within {
    outline: var(--focus-ring);
    outline-offset: var(--focus-offset);
  }

  .search svg {
    flex: none;
    width: 1.25rem;
    height: 1.25rem;
    fill: none;
    stroke: var(--color-text-muted);
    stroke-width: 2;
    stroke-linecap: round;
  }

  .search input {
    flex: 1;
    min-width: 0;
    border: none;
    background: none;
  }

  .search input:focus-visible {
    outline: none;
  }

  .filters {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-2);
  }

  select {
    width: 100%;
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-bg);
    cursor: pointer;
  }

  .count {
    margin: 0;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .results {
    flex: 1;
    overflow-y: auto;
    margin: 0;
    padding: var(--space-2);
    list-style: none;
  }

  .result {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    padding: var(--space-2);
    border: none;
    border-radius: var(--radius-md);
    background: none;
    text-align: left;
    cursor: pointer;
  }

  @media (hover: hover) {
    .result:hover {
      background: var(--color-surface);
    }
  }

  .result.added {
    background: var(--color-accent-subtle);
  }

  img,
  .img-placeholder {
    flex: none;
    width: 3.5rem;
    height: 3.5rem;
    border-radius: var(--radius-sm);
    background: var(--color-border);
    object-fit: cover;
  }

  .text {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
  }

  .name {
    font-weight: var(--weight-semibold);
  }

  .muscles {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  /* Plus to add, check when added: shape changes, not just color. */
  .state {
    display: grid;
    flex: none;
    place-items: center;
    width: 2rem;
    height: 2rem;
    border: 2px solid var(--color-border-strong);
    border-radius: var(--radius-full);
    color: var(--color-text-muted);
  }

  .added .state {
    border-color: var(--color-accent);
    background: var(--color-accent);
    color: var(--color-on-accent);
  }

  .state svg {
    width: 1.1rem;
    height: 1.1rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.5;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .more {
    display: flex;
    justify-content: center;
    padding: var(--space-3);
  }

  .empty {
    margin: 0;
    padding: var(--space-5);
    color: var(--color-text-muted);
    text-align: center;
  }
</style>
