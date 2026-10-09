// Workout sessions (Go: fitness.Session). In progress = no ended_at.
import { api } from '@/api'

export type Session = {
  id: string
  user_id: string
  plan_id?: string
  started_at: string // RFC 3339 timestamp
  ended_at?: string // missing = still in progress
  entries: Entry[]
  archived?: boolean
}

export type Entry = { exercise_id: string; sets: SetEntry[] }

export type SetEntry = { reps?: number; weight_kg?: number; duration_sec?: number; distance_m?: number }

/** The user's most recent non-archived sessions, newest first (max 100). */
export const getRecentSessions = (userId: string, limit = 100) =>
  api.get<Session[]>(`/api/users/${userId}/sessions?limit=${limit}`)

/** Starts a session now; the server assigns its ID from started_at. */
export function startSession(userId: string, planId: string | undefined, exerciseIds: string[]) {
  const session: Omit<Session, 'id'> = {
    user_id: userId,
    // Whole seconds: the ID is derived from this, at one-second precision.
    started_at: new Date(Math.floor(Date.now() / 1000) * 1000).toISOString(),
    entries: exerciseIds.map((exercise_id) => ({ exercise_id, sets: [] })),
  }
  if (planId) session.plan_id = planId
  return api.post<Session>(`/api/users/${userId}/sessions`, session)
}

export const saveSession = (s: Session) => api.put<Session>(`/api/users/${s.user_id}/sessions/${s.id}`, s)

/** The sets from the most recent finished session that did this exercise. */
export function lastSets(history: Session[], exerciseId: string, excludeSessionId?: string): SetEntry[] {
  for (const s of history) {
    if (s.id === excludeSessionId || !s.ended_at) continue
    const entry = s.entries.find((e) => e.exercise_id === exerciseId && e.sets.length > 0)
    if (entry) return entry.sets
  }
  return []
}

/** The in-progress session, if any (history is newest first). */
export const inProgress = (history: Session[]) => history.find((s) => !s.ended_at)
