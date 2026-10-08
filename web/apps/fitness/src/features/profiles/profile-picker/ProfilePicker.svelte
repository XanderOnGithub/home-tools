<script lang="ts">
  import { blobPath } from '@/features/profiles/blob'
  import { CreateProfileDialog } from '@/features/profiles/create-profile-dialog'
  import { ProfileAvatar } from '@/features/profiles/profile-avatar'
  import { rememberedProfileId, rememberProfile } from '@/features/profiles/remembered'
  import type { Profile } from '@/features/profiles/types'

  let { onselect }: { onselect: (profile: Profile) => void } = $props()

  // Only show "Loading…" if the request is actually slow. On the LAN it
  // usually answers in milliseconds, and a flash of loading text looks broken.
  const LOADING_DELAY_MS = 400

  let users = $state<Profile[]>([])
  let status = $state<'loading' | 'ready' | 'error'>('loading')
  let slow = $state(false)
  let creating = $state(false)

  // If this browser already picked someone, stay blank while loading and
  // skip straight past the picker, instead of flashing it first.
  const remembered = rememberedProfileId()

  // Archived profiles stay in the API (history needs their names) but
  // never appear in the picker.
  let active = $derived(users.filter((u) => !u.archived))

  async function load() {
    status = 'loading'
    slow = false
    const timer = setTimeout(() => (slow = true), LOADING_DELAY_MS)
    try {
      const res = await fetch('/api/users')
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      users = await res.json()
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

  load()
</script>

<!--
  "Who's working out?" Each profile is one button (avatar + name inside, so
  the whole tile is the tap target and the button's accessible name is the
  person's name). The avatar is aria-hidden: the name already says it.
-->
<section class="profile-picker" aria-labelledby="profile-picker-title">
  <h1 id="profile-picker-title" class:hidden={remembered && status === 'loading'}>
    Who's working out?
  </h1>

  <!-- One live region for every status message: screen readers announce
       loading, empty and error without the user having to look for them. -->
  <div class="profile-status" role="status">
    {#if status === 'loading' && slow}
      <p>Loading profiles…</p>
    {:else if status === 'ready' && active.length === 0}
      <p>No profiles yet. Add one to get started.</p>
    {/if}
  </div>

  {#if status === 'error'}
    <div class="profile-error" role="alert">
      <p>Couldn't load profiles. Check that the server is running.</p>
      <button type="button" class="retry-button" onclick={load}>Try again</button>
    </div>
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
            onclick={() => select(user)}
          >
            <span class="profile-avatar">
              <ProfileAvatar id={user.id} />
            </span>
            <span class="profile-name">{user.name}</span>
          </button>
        </li>
      {/each}

      <li class="profile-list-item">
        <button type="button" class="profile-button profile-button-create" onclick={() => (creating = true)}>
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
  {/if}
</section>

<CreateProfileDialog
  bind:open={creating}
  existing={users}
  oncreated={(profile) => users.push(profile)}
/>

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
  .profile-error p {
    margin: 0;
    color: var(--color-text-muted);
    font-size: var(--text-lg);
  }

  .profile-error {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-4);
  }

  .retry-button {
    min-height: var(--touch-target);
    padding: 0 var(--space-5);
    border: none;
    border-radius: var(--radius-full);
    background: var(--color-accent);
    color: var(--color-on-accent);
    font-weight: var(--weight-semibold);
    cursor: pointer;
    transition: background var(--duration-fast) var(--ease-out);
  }

  @media (hover: hover) {
    .retry-button:hover {
      background: var(--color-accent-hover);
    }
  }

  .retry-button:active {
    background: var(--color-accent-pressed);
  }

  /* Kept in the DOM (the section still needs its label), just not shown. */
  .hidden {
    visibility: hidden;
  }
</style>