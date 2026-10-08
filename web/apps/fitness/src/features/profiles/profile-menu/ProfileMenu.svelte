<!--
  Top-right profile button (blob + caret) and its small menu.

  Uses the native Popover API: the browser closes it on Esc or an outside
  click, puts it above everything (top layer), and the trigger button gets
  aria-expanded automatically. We only position it under the button.
-->
<script lang="ts">
  import { ProfileAvatar } from '@/features/profiles/profile-avatar'
  import type { Profile } from '@/features/profiles/types'

  let { profile, onswitch }: { profile: Profile; onswitch: () => void } = $props()

  let trigger = $state<HTMLButtonElement>()
  let menu = $state<HTMLDivElement>()
  let open = $state(false)

  // Place the menu right under the button, right edges aligned.
  function onBeforeToggle(event: ToggleEvent) {
    open = event.newState === 'open'
    if (!open || !trigger || !menu) return
    const r = trigger.getBoundingClientRect()
    menu.style.top = `${r.bottom + 8}px`
    menu.style.right = `${document.documentElement.clientWidth - r.right}px`
  }

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
  <span class="avatar"><ProfileAvatar id={profile.id} /></span>
  <svg class="caret" class:open viewBox="0 0 24 24" aria-hidden="true">
    <path d="M6 9l6 6 6-6" />
  </svg>
</button>

<div bind:this={menu} id="profile-menu" class="menu" popover="auto" onbeforetoggle={onBeforeToggle}>
  <p class="who">{profile.name}</p>
  <button type="button" class="item" onclick={switchProfile}>
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M16 3l4 4-4 4 M20 7H8 M8 21l-4-4 4-4 M4 17h12" />
    </svg>
    Switch profile
  </button>
</div>

<style>
  .trigger {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    min-height: var(--touch-target);
    padding: var(--space-1) var(--space-2);
    border: none;
    border-radius: var(--radius-full);
    background: none;
    color: var(--color-text-muted);
    cursor: pointer;
    transition: background var(--duration-fast) var(--ease-out);
  }

  @media (hover: hover) {
    .trigger:hover {
      background: var(--color-surface);
      color: var(--color-text);
    }
  }

  .avatar {
    width: 2.25rem;
    height: 2.25rem;
  }

  .caret {
    width: 1.1rem;
    height: 1.1rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.5;
    stroke-linecap: round;
    stroke-linejoin: round;
    transition: rotate var(--duration-fast) var(--ease-out);
  }

  .caret.open {
    rotate: 180deg;
  }

  /* Popovers default to centered on screen; pin it under the trigger
   * (top/right are set in onBeforeToggle). */
  .menu {
    position: fixed;
    inset: auto;
    min-width: 12rem;
    margin: 0;
    padding: var(--space-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface-raised);
    color: var(--color-text);
    box-shadow: var(--shadow-md);
  }

  .who {
    margin: 0;
    padding: var(--space-2) var(--space-3);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
  }

  .item {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: none;
    border-radius: var(--radius-sm);
    background: none;
    text-align: left;
    cursor: pointer;
  }

  @media (hover: hover) {
    .item:hover {
      background: var(--color-accent-subtle);
    }
  }

  .item svg {
    width: 1.25rem;
    height: 1.25rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
</style>
