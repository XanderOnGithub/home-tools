<!--
  Add or edit a profile (edit when `editing` is set). Uses the native
  <dialog> with showModal(): the browser handles focus (moves in, stays in,
  returns on close), Esc to close, and hides the page behind it from
  screen readers.

  The whole dialog takes the chosen color (data-accent). The preview
  avatar morphs as the name is typed, when adding and when renaming (the
  shape comes from the name, decision #22); the ID never changes.
  "Remove" archives (nothing is deleted, decision #16) after a confirm.
-->
<script lang="ts">
  import { tick } from 'svelte'
  import { api } from '../../api'
  import { today } from '../../dates'
  import { blobPath } from '../blob'
  import { ProfileAvatar } from '../profile-avatar'
  import type { Profile, ProfileColor } from '../types'
  import { COLORS, firstFreeColor, idFromName } from './utils'

  let {
    open = $bindable(false),
    existing,
    editing = null,
    onsaved,
  }: {
    open?: boolean
    existing: Profile[] // all profiles, archived too (their IDs are taken)
    editing?: Profile | null // null = adding a new profile
    onsaved: (profile: Profile) => void
  } = $props()

  let dialog: HTMLDialogElement
  let name = $state('')
  let color = $state<ProfileColor>('green')
  let birthday = $state('') // "YYYY-MM-DD" from <input type="date">, or ''
  let units = $state<Profile['units']>('imperial')
  let nameError = $state('')
  let formError = $state('')
  let saving = $state(false)
  let confirmingRemove = $state(false)
  let keepButton = $state<HTMLButtonElement>()
  let removeButton = $state<HTMLButtonElement>()

  // The button that was pressed disappears, so move focus deliberately
  // (otherwise keyboard and screen reader users are dropped to the top).
  async function askRemove() {
    confirmingRemove = true
    await tick()
    keepButton?.focus()
  }

  async function keep() {
    confirmingRemove = false
    await tick()
    removeButton?.focus()
  }

  let taken = $derived(new Set(existing.map((p) => p.id)))
  let id = $derived(editing ? editing.id : idFromName(name, taken))

  // Parent flips `open`; the effect drives the real dialog.
  $effect(() => {
    if (open && !dialog.open) {
      name = editing?.name ?? ''
      birthday = editing?.birthday ?? ''
      units = editing?.units ?? 'imperial'
      color = editing?.color ?? firstFreeColor(existing.filter((p) => !p.archived).map((p) => p.color))
      nameError = ''
      formError = ''
      confirmingRemove = false
      dialog.showModal()
    } else if (!open && dialog.open) {
      dialog.close()
    }
  })

  async function submit(event: SubmitEvent) {
    event.preventDefault()
    nameError = ''
    formError = ''
    if (!name.trim()) {
      nameError = 'Enter a name.'
      return
    }
    if (!id) {
      nameError = 'Use at least one letter or number.'
      return
    }

    const profile: Profile = { id, name: name.trim(), color, units }
    if (birthday) profile.birthday = birthday
    await save(profile, editing ? "Couldn't save the profile." : "Couldn't create the profile.")
  }

  /** Archives the profile being edited: hidden everywhere, history kept. */
  async function remove() {
    if (!editing) return
    await save({ ...editing, archived: true }, "Couldn't remove the profile.")
  }

  async function save(profile: Profile, failure: string) {
    saving = true
    formError = ''
    try {
      onsaved(await api.put<Profile>(`/api/users/${profile.id}`, profile))
      open = false
    } catch (err) {
      formError = `${failure} ${(err as Error).message}`
    } finally {
      saving = false
    }
  }
</script>

<dialog
  bind:this={dialog}
  class="dialog"
  aria-labelledby="profile-dialog-title"
  data-accent={color}
  onclose={() => (open = false)}
>
  <form class="form" onsubmit={submit} novalidate>
    <h2 id="profile-dialog-title">{editing ? 'Edit profile' : 'New profile'}</h2>

    <div class="preview">
      <ProfileAvatar name={name.trim() || 'new-profile'} />
    </div>

    <div class="field">
      <label for="profile-name">Name</label>
      <input
        id="profile-name"
        type="text"
        bind:value={name}
        maxlength="30"
        autocomplete="off"
        autocapitalize="words"
        spellcheck="false"
        aria-invalid={nameError ? 'true' : undefined}
        aria-describedby={nameError ? 'profile-name-error' : undefined}
      />
      {#if nameError}
        <p id="profile-name-error" class="field-error">{nameError}</p>
      {/if}
    </div>

    <div class="field">
      <label for="profile-birthday">Birthday <span class="optional">(optional)</span></label>
      <input id="profile-birthday" type="date" bind:value={birthday} max={today()} />
    </div>

    <!-- Two radios styled as a segmented control; arrow keys switch. -->
    <fieldset class="field">
      <legend>Units</legend>
      <div class="segmented">
        <label>
          <input class="visually-hidden" type="radio" name="units" value="imperial" bind:group={units} />
          <span>Imperial <span class="unit-detail">lb, ft</span></span>
        </label>
        <label>
          <input class="visually-hidden" type="radio" name="units" value="metric" bind:group={units} />
          <span>Metric <span class="unit-detail">kg, cm</span></span>
        </label>
      </div>
    </fieldset>

    <!-- Radios, visually replaced by blobs: arrow keys move between
         colors and screen readers say "Green, radio button, 1 of 4". -->
    <fieldset class="field">
      <legend>Color</legend>
      <div class="swatches">
        {#each COLORS as c (c)}
          <label class="swatch" data-accent={c}>
            <input class="visually-hidden" type="radio" name="color" value={c} bind:group={color} />
            <svg viewBox="0 0 100 100" aria-hidden="true">
              <path d={blobPath(`swatch-${c}`)} />
            </svg>
            <span class="visually-hidden">{c[0].toUpperCase() + c.slice(1)}</span>
          </label>
        {/each}
      </div>
    </fieldset>

    {#if formError}
      <p class="form-error" role="alert">{formError}</p>
    {/if}

    {#if confirmingRemove && editing}
      <!-- Replaces the normal buttons, so a stray tap can't remove. -->
      <div class="confirm" role="group" aria-labelledby="confirm-remove-text">
        <p id="confirm-remove-text">
          Remove {editing.name}? Their workouts stay saved, and you can restore the profile from
          Archived.
        </p>
        <div class="actions">
          <button type="button" class="btn btn-secondary" bind:this={keepButton} onclick={keep}>
            Keep
          </button>
          <button type="button" class="btn btn-danger" onclick={remove} disabled={saving}>
            {saving ? 'Removing…' : 'Remove'}
          </button>
        </div>
      </div>
    {:else}
      <div class="actions">
        {#if editing}
          <button type="button" class="btn btn-text-danger" bind:this={removeButton} onclick={askRemove}>
            Remove
          </button>
          <span class="spacer"></span>
        {/if}
        <button type="button" class="btn btn-secondary" onclick={() => (open = false)}>
          Cancel
        </button>
        <button type="submit" class="btn btn-primary" disabled={saving}>
          {#if saving}Saving…{:else if editing}Save{:else}Create profile{/if}
        </button>
      </div>
    {/if}
  </form>
</dialog>

<style>
  .dialog {
    width: min(26rem, calc(100vw - 2 * var(--space-4)));
    max-height: calc(100dvh - 2 * var(--space-4));
    overflow-y: auto;
    padding: var(--space-6);
  }

  .form {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
    text-align: center;
  }

  .preview {
    width: 7rem;
    height: 7rem;
    align-self: center;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    border: none;
    min-width: 0;
  }

  label,
  legend {
    padding: 0;
    font-weight: var(--weight-medium);
  }

  .optional {
    color: var(--color-text-muted);
    font-weight: var(--weight-normal);
  }

  input[type='text'],
  input[type='date'] {
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-surface);
    font-size: var(--text-md);
  }

  input[aria-invalid='true'] {
    border-color: var(--color-danger);
  }

  .field-error,
  .form-error {
    margin: 0;
    color: var(--color-danger-text);
    font-size: var(--text-sm);
  }

  .segmented {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-1);
    padding: var(--space-1);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-full);
  }

  .segmented label {
    display: grid;
    place-items: center;
    min-height: calc(var(--touch-target) - 2 * var(--space-1));
    border-radius: var(--radius-full);
    font-weight: var(--weight-medium);
    cursor: pointer;
    transition: background var(--duration-fast) var(--ease-out);
  }

  .segmented label:has(input:checked) {
    background: var(--color-accent);
    color: var(--color-on-accent);
  }

  .segmented label:has(input:focus-visible) {
    outline: var(--focus-ring);
    outline-offset: var(--focus-offset);
  }

  @media (hover: hover) {
    .segmented label:not(:has(input:checked)):hover {
      background: var(--color-accent-subtle);
    }
  }

  .unit-detail {
    font-size: var(--text-sm);
    font-weight: var(--weight-normal);
  }

  .swatches {
    display: flex;
    gap: var(--space-3);
  }

  .swatch {
    position: relative;
    display: grid;
    place-items: center;
    width: var(--touch-target);
    height: var(--touch-target);
    padding: var(--space-1);
    border-radius: var(--radius-full);
    cursor: pointer;
  }

  .swatch svg {
    width: 100%;
    height: 100%;
    transition: scale var(--duration-fast) var(--ease-out);
  }

  .swatch path {
    fill: var(--color-accent);
  }

  /* Selected: a ring around the blob, not just color (DESIGN.md §2). */
  .swatch:has(input:checked) {
    box-shadow: 0 0 0 2px var(--color-surface-raised), 0 0 0 4px var(--color-text);
  }

  /* The radio is hidden, so its keyboard focus shows on the swatch. */
  .swatch:has(input:focus-visible) {
    outline: var(--focus-ring);
    outline-offset: 4px;
  }

  @media (hover: hover) {
    .swatch:hover svg {
      scale: 1.1;
    }
  }

  .swatch:active svg {
    scale: 0.95;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-3);
  }

  .spacer {
    flex: 1;
  }

  .confirm {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    padding: var(--space-4);
    border-radius: var(--radius-md);
    background: var(--color-danger-subtle);
  }

  .confirm p {
    margin: 0;
  }
</style>
