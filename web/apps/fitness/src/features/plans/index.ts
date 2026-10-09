// Plans (Go: fitness.Plan). Shared by everyone; `created_by` = user ID.
import { api } from '@/api'

export type Plan = {
  id: string
  name: string
  created_by: string
  exercises: { exercise_id: string; suggested_sets?: number }[]
  archived?: boolean
}

export const getPlans = () => api.get<Plan[]>('/api/plans')

export const savePlan = (r: Plan) => api.put<Plan>(`/api/plans/${r.id}`, r)
