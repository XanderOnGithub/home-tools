// The server list with live state, refreshed every REFRESH_MS while the
// page is visible (a server can stop on its own, or someone else can
// start it). Shared by the list and the server page.
import { onMount } from 'svelte'
import { getServers, type Server } from '@/features/servers'

const REFRESH_MS = 10_000

export function liveServers() {
  const live = $state({
    servers: [] as Server[],
    status: 'loading' as 'loading' | 'ready' | 'error',
  })

  async function load() {
    try {
      live.servers = await getServers()
      live.status = 'ready'
    } catch (err) {
      console.error('Loading servers failed:', err)
      // Keep showing the last good list if a refresh fails.
      if (live.status !== 'ready') live.status = 'error'
    }
  }

  /** Replace one server after an action answered with its new state. */
  function update(s: Server) {
    live.servers = live.servers.map((x) => (x.id === s.id ? s : x))
  }

  onMount(() => {
    load()
    const timer = setInterval(() => document.visibilityState === 'visible' && load(), REFRESH_MS)
    return () => clearInterval(timer)
  })

  return {
    get servers() {
      return live.servers
    },
    get status() {
      return live.status
    },
    load,
    update,
  }
}
