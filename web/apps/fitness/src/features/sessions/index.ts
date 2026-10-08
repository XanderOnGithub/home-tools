// Workout sessions (Go: fitness.Session).
import { api } from '@/api'

export type Session = {
  id: string
  user_id: string
  routine_id?: string
  started_at: string // RFC 3339 timestamp
  ended_at?: string // missing = still in progress
  entries: { exercise_id: string; sets: SetEntry[] }[]
  archived?: boolean
}

export type SetEntry = { reps?: number; weight_kg?: number; duration_sec?: number; distance_m?: number }

/** The user's most recent non-archived sessions, newest first (max 100). */
export const getRecentSessions = (userId: string, limit = 100) =>
  api.get<Session[]>(`/api/users/${userId}/sessions?limit=${limit}`)
