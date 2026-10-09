// The settings and the bot's state, shared by every page: one copy for the
// whole app, so switching pages never refetches or shows a blank page.
//
// While any page uses it, the bot's state is polled every POLL_MS (it's
// served from memory, so polls are cheap) and refreshed when the tab comes
// back into view: new access requests and connection changes show up
// without a reload. The config only changes when someone saves.
import { ApiError } from '@home-tools/ui/api'
import { onMount } from 'svelte'
import { getBot, getConfig, saveConfig, type BotView, type Config } from '@/features/bot'

const POLL_MS = 15_000

const live = $state({
  config: null as Config | null,
  bot: null as BotView | null,
  status: 'loading' as 'loading' | 'ready' | 'error',
})

let users = 0 // mounted pages using it
let timer: ReturnType<typeof setInterval> | undefined

async function load() {
  try {
    ;[live.config, live.bot] = await Promise.all([getConfig(), getBot()])
    live.status = 'ready'
  } catch (err) {
    console.error('Loading settings failed:', err)
    if (live.status !== 'ready') live.status = 'error'
  }
}

async function refreshBot() {
  try {
    live.bot = await getBot()
  } catch (err) {
    console.error('Refreshing the bot failed:', err) // keep showing the last good state
  }
}

/**
 * Applies change to a copy of the config and saves it. Resolves to an
 * error message for people, or '' when saved. If someone else saved in
 * between (409), the page reloads their version, so trying again works.
 */
async function save(change: (c: Config) => void): Promise<string> {
  if (!live.config) return 'The settings haven’t loaded yet.'
  const next = $state.snapshot(live.config) as Config
  change(next)
  try {
    live.config = await saveConfig(next)
    refreshBot() // the persona may change with the names; requests drop when verified
    return ''
  } catch (err) {
    if (err instanceof ApiError && err.status === 409) {
      await load()
      return 'Someone else changed these settings just now. They’re reloaded; try again.'
    }
    return (err as Error).message
  }
}

function onVisibility() {
  if (document.visibilityState === 'visible') refreshBot()
}

export function settings() {
  onMount(() => {
    if (users++ === 0) {
      if (live.status !== 'ready') load()
      else refreshBot()
      timer = setInterval(() => document.visibilityState === 'visible' && refreshBot(), POLL_MS)
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
    get config() {
      return live.config
    },
    get bot() {
      return live.bot
    },
    get status() {
      return live.status
    },
    load,
    refreshBot,
    save,
  }
}
