// Every home tool in one list (decision #46): the tools home page
// (tools.<domain>) shows these as shortcuts, and the favicon script draws
// each one's icon from the same seed and color, so they always match.
// Adding a tool = one entry here.
import type { ProfileColor } from '../profiles/types.ts' // .ts: the Node favicon script imports this file too

export type Tool = {
  id: string // its subdomain: "fitness" → fitness.<domain>
  name: string
  description: string // one short line under the name
  seed: string // blob shape and face (profiles/blob), same as its favicon
  color: ProfileColor
}

/** The tools, in the order the home page shows them. */
export const TOOLS: Tool[] = [
  { id: 'fitness', name: 'Fitness', description: 'Workouts, plans and progress', seed: 'gym', color: 'green' },
  { id: 'games', name: 'Games', description: "Game servers and who's online", seed: 'games', color: 'orange' },
  { id: 'discord', name: 'Discord', description: 'The bot and its features', seed: 'discord', color: 'blue' },
]

/** The home page itself: not a shortcut, but it has an icon too. */
export const HOME: Tool = { id: 'tools', name: 'Home Tools', description: 'Every tool in one place', seed: 'tools', color: 'purple' }

/**
 * The tool's address, built from the current one by swapping the first
 * label (tools.example.com → fitness.example.com), so the real domain never
 * has to be in git. A host without a subdomain (localhost) gets one added:
 * fitness.localhost, which browsers resolve to this machine. `here` is
 * usually `location`.
 */
export function toolUrl(id: string, here: { protocol: string; hostname: string; port: string }): string {
  const host = here.hostname
  const isIP = /^[\d.]+$/.test(host) || host.includes(':')
  const labels = host.split('.')
  const target = isIP ? host : labels.length > 1 ? [id, ...labels.slice(1)].join('.') : `${id}.${host}`
  const port = here.port ? `:${here.port}` : ''
  return `${here.protocol}//${target}${port}/`
}
