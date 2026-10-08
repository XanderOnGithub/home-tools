// Routines (Go: fitness.Routine). Shared by everyone; `created_by` = user ID.
import { api } from '@/api'

export type Routine = {
  id: string
  name: string
  created_by: string
  exercises: { exercise_id: string; suggested_sets?: number }[]
  archived?: boolean
}

export const getRoutines = () => api.get<Routine[]>('/api/routines')

export const saveRoutine = (r: Routine) => api.put<Routine>(`/api/routines/${r.id}`, r)
