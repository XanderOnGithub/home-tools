// Plans (Go: fitness.Plan). Shared by everyone; `created_by` = user ID.
import { api } from '@/api'

export type Plan = {
  id: string
  name: string
  created_by: string
  exercises: { exercise_id: string; suggested_sets?: number; rest_sec?: number }[] // sets, rest: hints
  archived?: boolean
}

/** Rest between sets when a plan exercise has no rest_sec (decision #35). */
export const DEFAULT_REST_SEC = 60

export const getPlans = () => api.get<Plan[]>('/api/plans')

export const savePlan = (r: Plan) => api.put<Plan>(`/api/plans/${r.id}`, r)
