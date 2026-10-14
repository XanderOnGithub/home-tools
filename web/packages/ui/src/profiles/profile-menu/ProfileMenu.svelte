<!--
  Top-right profile button (avatar) and its menu: who's signed in, switch
  profile. Going to another tool lives in the tool switcher (app-shell).

  Native Popover API: the browser closes it on Esc or an outside click,
  puts it above everything and sets aria-expanded on the trigger.
  Looks: .menu in app.css.
-->
<script lang="ts">
  import { placeUnder } from '../../menu'
  import { ProfileAvatar } from '../profile-avatar'
  import type { Profile } from '../types'

  let { profile, onswitch }: { profile: Profile; onswitch: () => void } = $props()

  let trigger = $state<HTMLButtonElement>()
  let menu = $state<HTMLDivElement>()

  function switchProfile() {
    menu?.hidePopover()
    onswitch()
  }
</script>

<button
  bind:this={trigger}
  type="button"
  class="trigger"
  popovertarget="profile-menu"
  aria-label="Profile menu, {profile.name}"
>
  <span class="avatar"><ProfileAvatar name={profile.name} /></span>
</button>

<div
  bind:this={menu}
  id="profile-menu"
  class="menu"
  popover="auto"
  onbeforetoggle={(e) => placeUnder(e, trigger, menu, 'end')}
>
  <div class="who">
    <span class="who-avatar"><ProfileAvatar name={profile.name} /></span>
    <span class="who-name">{profile.name}</span>
  </div>
  <hr class="menu-separator" />
  <button type="button" class="menu-item" onclick={switchProfile}>
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M16 3l4 4-4 4 M20 7H8 M8 21l-4-4 4-4 M4 17h12" />
    </svg>
    Switch profile
  </button>
</div>

<style>
  .trigger {
    display: grid;
    place-items: center;
    width: var(--touch-target);
    height: var(--touch-target);
    margin-right: calc(-1 * var(--space-1));
    padding: 0;
    border: none;
    border-radius: var(--radius-full);
    background: none;
    cursor: pointer;
    transition: background var(--duration-fast) var(--ease-out);
  }

  @media (hover: hover) {
    .trigger:hover {
      background: var(--color-surface-sunken);
    }
  }

  .avatar {
    width: 2rem;
    height: 2rem;
  }

  .who {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-3);
  }

  .who-avatar {
    flex: none;
    width: 2rem;
    height: 2rem;
  }

  .who-name {
    font-weight: var(--weight-semibold);
  }
</style>
