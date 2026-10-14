<!--
  A form field: label, the control, a hint and an error, wired together
  (decision #57). The control comes in as a snippet that receives the
  attributes to spread on it, so the label points at it and the hint and
  error are read with it:

    <Field label="Channel" hint="Where the poll is posted">
      {#snippet children(attrs)}<select {...attrs} bind:value>…</select>{/snippet}
    </Field>
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  type Attrs = { id: string; 'aria-describedby'?: string; 'aria-invalid'?: 'true' }

  let {
    label,
    hint,
    error,
    optional = false,
    hideLabel = false,
    children,
  }: {
    label: string
    hint?: string
    error?: string
    optional?: boolean
    hideLabel?: boolean // still read by screen readers
    children: Snippet<[Attrs]>
  } = $props()

  const id = `field-${Math.random().toString(36).slice(2, 9)}`

  let attrs = $derived<Attrs>({
    id,
    'aria-describedby': [hint && `${id}-hint`, error && `${id}-error`].filter(Boolean).join(' ') || undefined,
    'aria-invalid': error ? 'true' : undefined,
  })
</script>

<div class="field">
  <label for={id} class:visually-hidden={hideLabel}>
    {label}{#if optional}<span class="optional">Optional</span>{/if}
  </label>
  {@render children(attrs)}
  {#if error}
    <p class="error" id="{id}-error">
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 8v5 M12 16.5h.01 M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18z" /></svg>
      {error}
    </p>
  {:else if hint}
    <p class="hint" id="{id}-hint">{hint}</p>
  {/if}
</div>

<style>
  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    min-width: 0;
  }

  label {
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
  }

  .optional {
    margin-left: var(--space-2);
    color: var(--color-text-muted);
    font-weight: var(--weight-normal);
  }

  .hint,
  .error {
    margin: 0;
    font-size: var(--text-sm);
  }

  .hint {
    color: var(--color-text-muted);
  }

  .error {
    display: flex;
    align-items: flex-start;
    gap: var(--space-1);
    color: var(--color-danger-text);
  }

  .error svg {
    flex: none;
    width: 1rem;
    height: 1rem;
    margin-top: 0.15em;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
  }
</style>
