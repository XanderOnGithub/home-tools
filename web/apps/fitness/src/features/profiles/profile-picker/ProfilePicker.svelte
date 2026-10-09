<script lang="ts">
  import { LoadError } from '@/components/load-error'
  import { api } from '@/api'
  import { blobPath } from '@/features/profiles/blob'
  import { ProfileAvatar } from '@/features/profiles/profile-avatar'
  import { ProfileDialog } from '@/features/profiles/profile-dialog'
  import { rememberedProfileId, rememberProfile } from '@/features/profiles/remembered'
  import type { Profile } from '@/features/profiles/types'

  let { onselect }: { onselect: (profile: Profile) => void } = $props()

  // Only show "Loading…" if the request is actually slow. On the LAN it
  // usually answers in milliseconds, and a flash of loading text looks broken.
  const LOADING_DELAY_MS = 400

  let users = $state<Profile[]>([])
  let status = $state<'loading' | 'ready' | 'error'>('loading')
  let slow = $state(false)
  let dialogOpen = $state(false)
  let editing = $state<Profile | null>(null) // null = the dialog adds
  let managing = $state(false)
  let restoreError = $state('')

  // If this browser already picked someone, stay blank while loading and
  // skip straight past the picker, instead of flashing it first.
  const remembered = rememberedProfileId()

  // Archived profiles stay in the API (history needs their names) but
  // never appear in the picker; manage mode lists them for restoring.
  let active = $derived(users.filter((u) => !u.archived))
  let archived = $derived(users.filter((u) => u.archived))

  async function load() {
    status = 'loading'
    slow = false
    const timer = setTimeout(() => (slow = true), LOADING_DELAY_MS)
    try {
      users = await api.get<Profile[]>('/api/users')
      // Remembered profile still exists (and isn't archived)? Go straight in.
      const match = users.find((u) => u.id === remembered && !u.archived)
      if (match) {
        onselect(match)
        return
      }
      status = 'ready'
    } catch (err) {
      console.error('Loading profiles failed:', err)
      status = 'error'
    } finally {
      clearTimeout(timer)
    }
  }

  function select(profile: Profile) {
    rememberProfile(profile.id)
    onselect(profile)
  }

  function openAdd() {
    editing = null
    dialogOpen = true
  }

  function openEdit(profile: Profile) {
    editing = profile
    dialogOpen = true
  }

  /** Puts a saved profile into the list (replacing the old copy, if any). */
  function upsert(profile: Profile) {
    const i = users.findIndex((u) => u.id === profile.id)
    if (i === -1) users.push(profile)
    else users[i] = profile
  }

  async function restore(profile: Profile) {
    restoreError = ''
    try {
      upsert(await api.put<Profile>(`/api/users/${profile.id}`, { ...profile, archived: false }))
    } catch (err) {
      restoreError = `Couldn't restore ${profile.name}. ${(err as Error).message}`
    }
  }

  load()
</script>

<!--
  "Who's working out?" Each profile is one button (avatar + name inside, so
  the whole tile is the tap target and the button's accessible name is the
  person's name). The avatar is aria-hidden: the name already says it.

  Manage mode: the same tiles open the edit dialog instead (each shows a
  pencil, and its name becomes "Edit Xander"), and archived profiles are
  listed underneath with Restore.
-->
<section class="profile-picker" aria-labelledby="profile-picker-title">
  <h1 id="profile-picker-title" class:hidden={remembered && status === 'loading'}>
    {managing ? 'Manage profiles' : "Who's working out?"}
  </h1>

  <!-- One live region for every status message: screen readers announce
       loading, empty and error without the user having to look for them. -->
  <div class="profile-status" role="status">
    {#if status === 'loading' && slow}
      <p>Loading profiles…</p>
    {:else if status === 'ready' && active.length === 0 && archived.length > 0}
      <p>Everyone's archived. Add a profile, or restore one under Manage profiles.</p>
    {:else if status === 'ready' && active.length === 0}
      <p>No profiles yet. Add one to get started.</p>
    {/if}
  </div>

  {#if status === 'error'}
    <LoadError what="profiles" onretry={load} align="center" />
  {:else if status === 'ready'}
    <ul class="profile-list">
      {#each active as user (user.id)}
        <li class="profile-list-item">
          <!-- data-accent switches the accent tokens to this person's color
               for everything inside the button. -->
          <button
            type="button"
            class="profile-button"
            data-accent={user.color}
            onclick={() => (managing ? openEdit(user) : select(user))}
          >
            <span class="profile-avatar">
              <ProfileAvatar id={user.id} />
              {#if managing}
                <span class="edit-badge" aria-hidden="true">
                  <svg viewBox="0 0 24 24">
                    <path d="M4 20h4L19 9l-4-4L4 16v4z M13.5 6.5l4 4" />
                  </svg>
                </span>
              {/if}
            </span>
            <span class="profile-name">
              {#if managing}<span class="visually-hidden">Edit</span>{/if}
              {user.name}
            </span>
          </button>
        </li>
      {/each}

      <li class="profile-list-item">
        <button type="button" class="profile-button profile-button-create" onclick={openAdd}>
          <!-- An empty blob (dashed outline) waiting for a person. Its shape
               comes from a fixed seed, so it never changes. -->
          <span class="profile-avatar" aria-hidden="true">
            <svg viewBox="0 0 100 100">
              <path class="create-blob" d={blobPath('add-profile')} />
              <path class="create-plus" d="M 50 38 V 62 M 38 50 H 62" />
            </svg>
          </span>
          <span class="profile-name">Add profile</span>
        </button>
      </li>
    </ul>

    {#if managing && archived.length > 0}
      <section class="archived" aria-labelledby="archived-title">
        <h2 id="archived-title">Archived</h2>
        <ul>
          {#each archived as user (user.id)}
            <li class="archived-row" data-accent={user.color}>
              <span class="archived-avatar"><ProfileAvatar id={user.id} /></span>
              <span class="archived-name">{user.name}</span>
              <button type="button" class="btn btn-quiet" onclick={() => restore(user)}>
                Restore<span class="visually-hidden"> {user.name}</span>
              </button>
            </li>
          {/each}
        </ul>
        {#if restoreError}
          <p class="restore-error" role="alert">{restoreError}</p>
        {/if}
      </section>
    {/if}

    <!-- One button whose label flips, so keyboard focus stays on it. -->
    {#if users.length > 0}
      <button type="button" class="btn btn-quiet manage-toggle" onclick={() => (managing = !managing)}>
        {managing ? 'Done' : 'Manage profiles'}
      </button>
    {/if}
  {/if}
</section>

<ProfileDialog bind:open={dialogOpen} existing={users} {editing} onsaved={upsert} />

<style>
  /* Centered in the viewport: this is the whole first screen. */
  .profile-picker {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-6);
    min-height: 100dvh;
    padding: var(--space-5) var(--space-4);
    text-align: center;
  }

  h1 {
    font-size: var(--text-3xl);
    font-weight: var(--weight-extrabold);
  }

  .profile-list {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: var(--space-6);
    list-style: none;
    padding: 0;
    margin: 0;
  }

  /* The whole tile is the button: avatar on top, name below. */
  .profile-button {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2); /* room between the avatar and the focus ring */
    background: none;
    border: none;
    border-radius: var(--radius-lg); /* rounds the focus ring */
    cursor: pointer;
  }

  .profile-avatar {
    position: relative;
    display: grid;
    place-items: center;
    width: 8rem;
    height: 8rem;
    color: var(--color-on-accent);
    font-size: var(--text-3xl);
    font-weight: var(--weight-semibold);
    transition: scale var(--duration-fast) var(--ease-out);
  }

  /* The blob and the letter share one grid cell, so they stack. */
  .profile-avatar > * {
    grid-area: 1 / 1;
  }

  .profile-avatar svg {
    width: 100%;
    height: 100%;
  }

  /* Names are quiet until you point at them, so the colors lead. */
  .profile-name {
    color: var(--color-text-muted);
    font-size: var(--text-lg);
    font-weight: var(--weight-medium);
    transition: color var(--duration-fast) var(--ease-out);
  }

  /* Hover only where a pointer exists; on touch, :hover sticks after a tap. */
  @media (hover: hover) {
    .profile-button:hover .profile-avatar {
      scale: 1.05;
    }
    .profile-button:hover .profile-name {
      color: var(--color-text);
    }
  }

  .profile-button:focus-visible .profile-name {
    color: var(--color-text);
  }

  /* Pressed beats hover: instant feedback on tap. */
  .profile-button:active .profile-avatar {
    scale: 0.97;
  }

  /* "Add profile": the same blob language, but empty: a dashed outline and
   * a drawn plus. non-scaling-stroke keeps lines 2px at any avatar size. */
  .profile-button-create path {
    fill: none;
    stroke: var(--color-border-strong);
    stroke-width: 2px;
    stroke-linecap: round;
    vector-effect: non-scaling-stroke;
    transition:
      fill var(--duration-fast) var(--ease-out),
      stroke var(--duration-fast) var(--ease-out);
  }

  .profile-button-create .create-blob {
    stroke-dasharray: 6 6;
  }

  .profile-button-create .create-plus {
    stroke: var(--color-text-muted);
    stroke-width: 3px;
  }

  @media (hover: hover) {
    /* Hover stays neutral (no accent: no profile is chosen yet). The lines
     * just darken, alongside the grow every avatar gets. */
    .profile-button-create:hover .create-blob,
    .profile-button-create:hover .create-plus {
      stroke: var(--color-text);
    }
  }

  /* Status text sits where the list will be, so nothing jumps around. */
  .profile-status p,
  /* Pencil in manage mode: a small badge on the avatar's lower right. */
  .edit-badge {
    position: absolute;
    right: 0;
    bottom: var(--space-1);
    display: grid;
    place-items: center;
    width: 2.25rem;
    height: 2.25rem;
    border: 3px solid var(--color-bg);
    border-radius: var(--radius-full);
    background: var(--color-text);
    color: var(--color-bg);
  }

  .edit-badge svg {
    width: 1.1rem;
    height: 1.1rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .manage-toggle {
    font-size: var(--text-lg);
  }

  .archived {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    width: min(24rem, 100%);
    text-align: left;
  }

  .archived h2 {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .archived ul {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .archived-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
  }

  .archived-avatar {
    flex: none;
    width: 2.5rem;
    height: 2.5rem;
    opacity: 0.6; /* archived: present, but faded */
  }

  .archived-name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .restore-error {
    margin: 0;
    color: var(--color-danger-text);
  }

  /* Kept in the DOM (the section still needs its label), just not shown. */
  .hidden {
    visibility: hidden;
  }
</style>