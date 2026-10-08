<script lang="ts">
  import { HomeScreen } from '@/features/home/home-screen'
  import { ProfilePicker } from '@/features/profiles/profile-picker'
  import { forgetProfile } from '@/features/profiles/remembered'
  import type { Profile } from '@/features/profiles/types'

  // No router yet: one profile chosen = home, none = picker.
  let profile = $state<Profile | null>(null)

  // The chosen person's color becomes the accent for the whole app.
  $effect(() => {
    if (profile) document.documentElement.dataset.accent = profile.color
    else delete document.documentElement.dataset.accent
  })

  function switchProfile() {
    forgetProfile()
    profile = null
  }
</script>

{#if profile}
  <HomeScreen {profile} onswitch={switchProfile} />
{:else}
  <main>
    <ProfilePicker onselect={(p) => (profile = p)} />
  </main>
{/if}
