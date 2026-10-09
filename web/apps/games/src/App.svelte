<!--
  Games: pick a profile (shared with every tool, ADR 0007; the choice is
  remembered across subdomains), then the server list ("/") or one
  server's page ("/servers/<id>"). The profile gives the accent color;
  who may do what is a later decision.
-->
<script lang="ts">
  import { ProfileMenu } from '@home-tools/ui/profiles/profile-menu'
  import { ProfilePicker } from '@home-tools/ui/profiles/profile-picker'
  import { forgetProfile } from '@home-tools/ui/profiles/remembered'
  import type { Profile } from '@home-tools/ui/profiles/types'
  import { router } from '@home-tools/ui/router'
  import { ServerList } from '@/features/servers/server-list'
  import { ServerPage } from '@/features/servers/server-page'

  let profile = $state<Profile | null>(null)

  let serverId = $derived(router.path.match(/^\/servers\/([^/]+)$/)?.[1] ?? null)

  $effect(() => {
    if (profile) document.documentElement.dataset.accent = profile.color
    else delete document.documentElement.dataset.accent
  })

  $effect(() => {
    document.title = serverId ? `${decodeURIComponent(serverId)} · Games` : 'Games · Home Tools'
  })

  function switchProfile() {
    forgetProfile()
    profile = null
  }
</script>

{#if !profile}
  <main>
    <ProfilePicker onselect={(p) => (profile = p)} />
  </main>
{:else}
  <header class="top">
    <a class="brand" href="/">Games</a>
    <ProfileMenu {profile} onswitch={switchProfile} />
  </header>
  <main class="content">
    {#if router.path === '/'}
      <ServerList />
    {:else if serverId}
      {#key serverId}
        <ServerPage id={decodeURIComponent(serverId)} />
      {/key}
    {:else}
      <p>There's no page here. <a href="/">All servers</a></p>
    {/if}
  </main>
{/if}

<style>
  .top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: min(var(--page-width), 100%);
    margin: 0 auto;
    padding: var(--space-2) var(--space-4);
    border-bottom: 1px solid var(--color-border);
  }

  .brand {
    display: inline-flex;
    align-items: center;
    min-height: var(--touch-target);
    color: var(--color-text);
    font-size: var(--text-lg);
    font-weight: var(--weight-extrabold);
    text-decoration: none;
  }

  .content {
    width: min(var(--page-width), 100%);
    margin: 0 auto;
    padding: var(--space-6) var(--space-4) var(--space-8);
  }
</style>
