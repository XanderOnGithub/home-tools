// Exercise catalog (Go: fitness.Exercise). 876 imported from
// free-exercise-db; photos are served at /images/<path>.
import { api } from '@/api'

export type Exercise = {
  id: string
  name: string
  activation: Partial<Record<Muscle, number>> // muscle → (0, 1]; 1 = primary
  bodyweight?: boolean
  metrics: ('reps' | 'weight' | 'duration' | 'distance')[]
  equipment?: Equipment[]
  category?: string
  level?: string
  instructions?: string[]
  images?: string[]
  archived?: boolean
}

// Mirror the Go enums (tools/fitness/model.go).
export const MUSCLES = [
  'abdominals', 'abductors', 'adductors', 'biceps', 'calves', 'chest', 'forearms', 'glutes',
  'hamstrings', 'lats', 'lower_back', 'middle_back', 'neck', 'quadriceps', 'shoulders', 'traps', 'triceps',
] as const
export type Muscle = (typeof MUSCLES)[number]

export const EQUIPMENT = [
  'barbell', 'dumbbell', 'kettlebell', 'cable', 'machine', 'bench', 'pull_up_bar', 'band',
  'treadmill', 'ez_bar', 'medicine_ball', 'exercise_ball', 'foam_roller',
] as const
export type Equipment = (typeof EQUIPMENT)[number]

/** "lower_back" → "Lower back", "ez_bar" → "EZ bar". */
export function label(value: string): string {
  if (value === 'ez_bar') return 'EZ bar'
  const words = value.replace(/_/g, ' ')
  return words[0].toUpperCase() + words.slice(1)
}

/** Muscles the exercise works most (activation 1), for one-line summaries. */
export const primaryMuscles = (ex: Exercise) =>
  (Object.entries(ex.activation) as [Muscle, number][]).filter(([, v]) => v >= 1).map(([m]) => m)

export const getExercise = (id: string) => api.get<Exercise>(`/api/exercises/${id}`)

// The whole catalog (~1 MB) is fetched at most once per page load and
// shared by everyone who asks; a failed fetch is forgotten so it can retry.
let catalog: Promise<Exercise[]> | null = null
export function getCatalog(): Promise<Exercise[]> {
  catalog ??= api.get<Exercise[]>('/api/exercises').catch((err) => {
    catalog = null
    throw err
  })
  return catalog
}
