// The server list with live state, shared by every page: one copy for the
// whole app, so opening a server shows what the list already loaded (no
// blank page while it fetches again), and both stay in step.
//
// While any page uses it, it polls every POLL_MS (the server answers from
// its snapshot, so polls are cheap) and refreshes at once when the tab
// comes back into view.
import { onMount } from 'svelte'
import { getServers, type Server } from '@/features/servers'

const POLL_MS = 5_000

const live = $state({
  servers: [] as Server[],
  status: 'loading' as 'loading' | 'ready' | 'error',
})

let users = 0 // mounted pages using it
let timer: ReturnType<typeof setInterval> | undefined

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

function onVisibility() {
  if (document.visibilityState === 'visible') load()
}

export function liveServers() {
  onMount(() => {
    if (users++ === 0) {
      load()
      timer = setInterval(() => document.visibilityState === 'visible' && load(), POLL_MS)
      document.addEventListener('visibilitychange', onVisibility)
    }
    return () => {
      if (--users === 0) {
        clearInterval(timer)
        document.removeEventListener('visibilitychange', onVisibility)
      }
    }
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
