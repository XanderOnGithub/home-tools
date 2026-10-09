<!--
  Games: pick a profile (shared with every tool, ADR 0007; the choice is
  remembered across subdomains), then the shared app shell (same header,
  nav and spacing as fitness) around the server list ("/") or one server's
  page ("/servers/<id>"). The profile gives the accent color; who may do
  what is a later decision.
-->
<script lang="ts">
  import { AppShell } from '@home-tools/ui/app-shell'
  import { ProfilePicker } from '@home-tools/ui/profiles/profile-picker'
  import { forgetProfile } from '@home-tools/ui/profiles/remembered'
  import type { Profile } from '@home-tools/ui/profiles/types'
  import { router } from '@home-tools/ui/router'
  import { ServerList } from '@/features/servers/server-list'
  import { ServerPage } from '@/features/servers/server-page'

  const LINKS = [
    // A rack server: two stacked units with status lights.
    { href: '/', label: 'Servers', icon: 'M4 4h16v7H4z M4 13h16v7H4z M8 7.5h.01 M8 16.5h.01' },
  ]

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
    <ProfilePicker onselect={(p) => (profile = p)} title="Who's playing?" />
  </main>
{:else}
  <AppShell links={LINKS} {profile} onswitch={switchProfile}>
    {#if router.path === '/'}
      <ServerList />
    {:else if serverId}
      {#key serverId}
        <ServerPage id={decodeURIComponent(serverId)} />
      {/key}
    {:else}
      <h1 tabindex="-1">Not found</h1>
      <p>There's no page here. <a href="/">All servers</a></p>
    {/if}
  </AppShell>
{/if}
