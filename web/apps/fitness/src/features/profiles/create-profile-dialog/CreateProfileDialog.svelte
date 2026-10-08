<!--
  "New profile" dialog. Uses the native <dialog> with showModal(): the
  browser handles focus (moves in, stays in, returns on close), Esc to
  close, and hides the page behind it from screen readers.

  The whole dialog takes the chosen color (data-accent), and the preview
  avatar morphs as the name changes, since the shape comes from the ID.
-->
<script lang="ts">
  import { blobPath } from '@/features/profiles/blob'
  import { ProfileAvatar } from '@/features/profiles/profile-avatar'
  import type { Profile, ProfileColor } from '@/features/profiles/types'
  import { COLORS, firstFreeColor, idFromName } from './utils'

  let {
    open = $bindable(false),
    existing,
    oncreated,
  }: {
    open?: boolean
    existing: Profile[] // all profiles, archived too (their IDs are taken)
    oncreated: (profile: Profile) => void
  } = $props()

  let dialog: HTMLDialogElement
  let name = $state('')
  let color = $state<ProfileColor>('green')
  let nameError = $state('')
  let formError = $state('')
  let saving = $state(false)

  let taken = $derived(new Set(existing.map((p) => p.id)))
  let id = $derived(idFromName(name, taken))

  // Parent flips `open`; the effect drives the real dialog.
  $effect(() => {
    if (open && !dialog.open) {
      name = ''
      color = firstFreeColor(existing.filter((p) => !p.archived).map((p) => p.color))
      nameError = ''
      formError = ''
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

    saving = true
    const profile: Profile = { id, name: name.trim(), color, units: 'metric' }
    try {
      const res = await fetch(`/api/users/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(profile),
      })
      if (!res.ok) {
        // 400s carry a message written for people; anything else doesn't.
        const body = await res.json().catch(() => null)
        throw new Error(res.status === 400 && body?.error ? body.error : 'Something went wrong.')
      }
      oncreated(await res.json())
      open = false
    } catch (err) {
      formError =
        err instanceof TypeError
          ? "Couldn't reach the server. Check that it's running."
          : `Couldn't create the profile: ${(err as Error).message}`
    } finally {
      saving = false
    }
  }
</script>

<dialog
  bind:this={dialog}
  class="dialog"
  aria-labelledby="create-profile-title"
  data-accent={color}
  onclose={() => (open = false)}
>
  <form class="form" onsubmit={submit} novalidate>
    <h2 id="create-profile-title">New profile</h2>

    <div class="preview">
      <ProfileAvatar id={id || 'new-profile'} />
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

    <div class="actions">
      <button type="button" class="button button-secondary" onclick={() => (open = false)}>
        Cancel
      </button>
      <button type="submit" class="button button-primary" disabled={saving}>
        {saving ? 'Creating…' : 'Create profile'}
      </button>
    </div>
  </form>
</dialog>

<style>
  .dialog {
    width: min(26rem, calc(100vw - 2 * var(--space-4)));
    padding: var(--space-6);
    border: none;
    border-radius: var(--radius-lg);
    background: var(--color-surface-raised);
    color: var(--color-text);
    box-shadow: var(--shadow-md);
  }

  .dialog::backdrop {
    background: rgb(0 0 0 / 0.45);
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

  input[type='text'] {
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

  .button {
    min-height: var(--touch-target);
    padding: 0 var(--space-5);
    border-radius: var(--radius-full);
    font-weight: var(--weight-semibold);
    cursor: pointer;
    transition:
      background var(--duration-fast) var(--ease-out),
      border-color var(--duration-fast) var(--ease-out);
  }

  .button-primary {
    border: none;
    background: var(--color-accent);
    color: var(--color-on-accent);
  }

  .button-secondary {
    border: 1px solid var(--color-border-strong);
    background: none;
  }

  @media (hover: hover) {
    .button-primary:hover {
      background: var(--color-accent-hover);
    }
    .button-secondary:hover {
      border-color: var(--color-text);
    }
  }

  .button-primary:active {
    background: var(--color-accent-pressed);
  }

  .button:disabled {
    cursor: progress;
    opacity: 0.7;
  }
</style>
