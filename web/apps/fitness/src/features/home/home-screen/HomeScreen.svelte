<!--
  Fitness home, for the chosen profile. Placeholder until the real home
  (start a workout, recent sessions) is designed.
-->
<script lang="ts">
  import { ProfileAvatar } from '@/features/profiles/profile-avatar'
  import type { Profile } from '@/features/profiles/types'

  let { profile, onswitch }: { profile: Profile; onswitch: () => void } = $props()
</script>

<header class="header">
  <div class="who">
    <span class="avatar"><ProfileAvatar id={profile.id} /></span>
    <span class="name">{profile.name}</span>
  </div>
  <button type="button" class="switch" onclick={onswitch}>Switch profile</button>
</header>

<main class="main">
  <h1>Hi, {profile.name}</h1>
  <p>Your workouts will live here.</p>
</main>

<style>
  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--color-border);
  }

  .who {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    min-width: 0;
  }

  .avatar {
    flex: none;
    width: 2.5rem;
    height: 2.5rem;
  }

  .name {
    overflow: hidden;
    font-weight: var(--weight-semibold);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .switch {
    flex: none;
    min-height: var(--touch-target);
    padding: 0 var(--space-4);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-full);
    background: none;
    font-weight: var(--weight-medium);
    cursor: pointer;
    transition: border-color var(--duration-fast) var(--ease-out);
  }

  @media (hover: hover) {
    .switch:hover {
      border-color: var(--color-text);
    }
  }

  .main {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-6) var(--space-4);
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
  }

  p {
    margin: 0;
    color: var(--color-text-muted);
  }
</style>
