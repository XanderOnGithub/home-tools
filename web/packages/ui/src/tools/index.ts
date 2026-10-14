// Every home tool in one list (decision #46): the tools home page
// (tools.<domain>) and the tool switcher in every app's header show these,
// and the favicon script draws each one's icon from the same glyph and
// color, so they always match. Adding a tool = one entry here.
import type { ProfileColor } from '../profiles/types.ts' // .ts: the Node favicon script imports this file too

export type Tool = {
  id: string // its subdomain: "fitness" → fitness.<domain>
  name: string
  description: string // one short line under the name
  icon: string // glyph: SVG path in a 24×24 box, stroked (same as NavLink icons)
  color: ProfileColor // the icon tile's color
}

/** The tools, in the order the home page shows them. */
export const TOOLS: Tool[] = [
  {
    id: 'fitness',
    name: 'Fitness',
    description: 'Workouts, plans and progress',
    icon: 'M6.5 7v10 M17.5 7v10 M3.5 9.5v5 M20.5 9.5v5 M6.5 12h11', // a dumbbell
    color: 'green',
  },
  {
    id: 'games',
    name: 'Games',
    description: "Game servers and who's online",
    // A controller: body, d-pad, two buttons.
    icon: 'M7 8h10a4.5 4.5 0 0 1 4.5 4.5v.5a3 3 0 0 1-5.4 1.8L15 13.5H9l-1.1 1.3A3 3 0 0 1 2.5 13v-.5A4.5 4.5 0 0 1 7 8z M7.5 10v3 M6 11.5h3 M15.5 11h.01 M17.5 12.5h.01',
    color: 'orange',
  },
  {
    id: 'discord',
    name: 'Discord',
    description: 'The bot and its features',
    icon: 'M4 5h16v11H9l-5 4z M9 10.5h.01 M15 10.5h.01', // a chat bubble with a face: the bot
    color: 'blue',
  },
]

/** The home page itself: not a shortcut, but it has an icon too. */
export const HOME: Tool = {
  id: 'tools',
  name: 'Home Tools',
  description: 'Every tool in one place',
  icon: 'M4.5 4.5h6v6h-6z M13.5 4.5h6v6h-6z M4.5 13.5h6v6h-6z M13.5 13.5h6v6h-6z', // four tiles
  color: 'purple',
}

/** A tool by its id (the subdomain), HOME included. */
export const toolById = (id: string): Tool | undefined => [HOME, ...TOOLS].find((t) => t.id === id)

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
