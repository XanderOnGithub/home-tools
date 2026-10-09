// The Discord tool's settings and the bot's state (Go: tools/discord,
// decision #42).
import { api } from '@home-tools/ui/api'
import type { ProfileColor } from '@home-tools/ui/profiles/types'

/** A game server's live status, kept in one Discord message. */
export type StatusBoard = { server_id: string; channel_id: string }

/** May restart servers and edit whitelists. name is just a label. */
export type VerifiedUser = { user_id: string; name: string }

/** config.json: everything this page edits. */
export type Config = {
  revision: number // a save must send the revision it was based on (409 otherwise)
  status_boards: StatusBoard[]
  verified: VerifiedUser[]
  names: string[]
}

export type Persona = { name: string; color: ProfileColor; date: string }

export type Channel = { id: string; name: string; can_post: boolean }
export type Guild = { id: string; name: string; channels: Channel[] }

/** Someone unverified who tried /restart or /whitelist. */
export type AccessRequest = { user_id: string; name: string; command: string; at: string }

export type GameServer = { id: string; name: string; game: string; whitelist: boolean }

/** The bot's connection, today's persona, and lists for the pickers. */
export type BotView = {
  connected: boolean
  error?: string
  user?: string
  invite_url?: string
  guilds: Guild[]
  persona: Persona
  requests: AccessRequest[]
  servers: GameServer[]
  servers_error?: string
}

export const getConfig = () => api.get<Config>('/api/config')
export const saveConfig = (c: Config) => api.put<Config>('/api/config', c)
export const getBot = () => api.get<BotView>('/api/bot')
export const dismissRequest = (userId: string) => api.del(`/api/requests/${encodeURIComponent(userId)}`)

/** Today's avatar, exactly as Discord gets it. */
export const PERSONA_GIF = '/api/persona.gif'

/** A Discord ID (snowflake): 17–20 digits. */
export const isSnowflake = (s: string) => /^[0-9]{17,20}$/.test(s)

/** Today's avatar at rest (for reduced motion). */
export const PERSONA_PNG = '/api/persona.png'

/** "just now", "5 min ago", "3 h ago", "2 days ago". */
export function ago(iso: string, now = Date.now()): string {
  const min = Math.floor((now - Date.parse(iso)) / 60000)
  if (min < 1) return 'just now'
  if (min < 60) return `${min} min ago`
  const h = Math.floor(min / 60)
  if (h < 24) return `${h} h ago`
  const days = Math.floor(h / 24)
  return days === 1 ? 'yesterday' : `${days} days ago`
}
