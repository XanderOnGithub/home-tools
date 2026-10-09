// The fitness "workout profile" and weight log (Go: tools/fitness/body.go).
import { api, ApiError } from '@/api'

// Mirrors fitness.Profile in Go (snake_case keys, decision #14).
export type FitnessProfile = {
  user_id: string
  goal?: Goal // optional; not asked in onboarding
  height_m?: number
  schedule?: Partial<Record<Weekday, string>> // weekday → plan ID; missing = rest
  weight_prompt_skipped?: string // ISO week, e.g. "2026-W41"
}

export type Goal = 'strength' | 'muscle' | 'endurance' | 'general'

export type Weekday =
  | 'monday'
  | 'tuesday'
  | 'wednesday'
  | 'thursday'
  | 'friday'
  | 'saturday'
  | 'sunday'

export type WeightEntry = { date: string; weight_kg: number }

/** The user's fitness profile, or null if they haven't onboarded (404). */
export async function getFitnessProfile(userId: string): Promise<FitnessProfile | null> {
  try {
    return await api.get<FitnessProfile>(`/api/users/${userId}/fitness`)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return null
    throw err
  }
}

export const saveFitnessProfile = (p: FitnessProfile) =>
  api.put<FitnessProfile>(`/api/users/${p.user_id}/fitness`, p)

export const getWeights = (userId: string) => api.get<WeightEntry[]>(`/api/users/${userId}/weights`)

export const saveWeight = (userId: string, entry: WeightEntry) =>
  api.put<WeightEntry>(`/api/users/${userId}/weights/${entry.date}`, entry)
