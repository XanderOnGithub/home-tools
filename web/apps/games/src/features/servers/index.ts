// Game servers (Go: games.Server + its live Docker state, decision #37).
import { api } from '@home-tools/ui/api'

export type Game = 'minecraft' | 'valheim'

export type ServerState = {
  running: boolean
  status: string // Docker's word: running, exited, restarting, …
  started_at?: string // while running
}

/** Who's online. max is 0 when the game didn't say (Valheim's query often doesn't answer). */
export type Players = { online: number; max: number; names: string[] }

/** A join or leave, read from the server's log (decision #40). */
export type ActivityEvent = { at: string; player: string; kind: 'join' | 'leave' }

export type Server = {
  id: string
  name: string
  game: Game
  container: string
  archived?: boolean
  state?: ServerState // missing when Docker couldn't be asked; see error
  error?: string
  players?: Players // only while running and if a query address is set
  players_error?: string
  activity?: ActivityEvent[] // newest first, ≤ 50; since Home Tools started reading
}

export type Action = 'start' | 'stop' | 'restart'

export const GAMES: Record<Game, string> = { minecraft: 'Minecraft', valheim: 'Valheim' }

/** Every server with its live state, archived included. */
export const getServers = () => api.get<Server[]>('/api/servers')

/** Starts/stops/restarts; answers with the new state. Stop can take a minute. */
export const runAction = (id: string, action: Action) =>
  api.post<Server>(`/api/servers/${encodeURIComponent(id)}/${action}`, undefined)

/** A Minecraft player's face, 8×8 PNG (scale it up pixelated); 404 when there's none. */
export const headUrl = (id: string, player: string) =>
  `/api/servers/${encodeURIComponent(id)}/heads/${encodeURIComponent(player)}`

/** Server-Sent Events: the last 200 lines, then live (see LogView). */
export const logsUrl = (id: string) => `/api/servers/${encodeURIComponent(id)}/logs`

/** "2 h 5 min", "12 min", "under a minute", from a start time. */
export function uptime(startedAt: string, now = Date.now()): string {
  const min = Math.floor((now - Date.parse(startedAt)) / 60000)
  if (min < 1) return 'under a minute'
  const days = Math.floor(min / 1440)
  const h = Math.floor((min % 1440) / 60)
  const m = min % 60
  if (days) return `${days} d ${h} h`
  if (h) return `${h} h ${m} min`
  return `${m} min`
}

/** "2 of 20 online", "Nobody online · 20 slots", "2 online" (slots unknown), "Players unknown", or null
 * when there's nothing to say (stopped, or no query address configured). */
export function playersSummary(s: Server): string | null {
  if (!s.state?.running) return null
  if (s.players) {
    const { online, max } = s.players
    if (!max) return online === 0 ? 'Nobody online' : `${online} online`
    return online === 0 ? `Nobody online · ${max} slots` : `${online} of ${max} online`
  }
  return s.players_error ? 'Players unknown' : null
}

/** The line under the status: "Up 2 h 5 min · 2 of 20 online". */
export function details(s: Server): string {
  const parts: string[] = []
  if (s.state?.running && s.state.started_at) parts.push(`Up ${uptime(s.state.started_at)}`)
  const players = playersSummary(s)
  if (players) parts.push(players)
  return parts.join(' · ')
}
