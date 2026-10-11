// Muscle recovery (decision #49): how worked each muscle is, estimated by
// the server from the last week of workouts (Go: fitness.Recovery).
// Nothing is stored; this module fetches it and words it for people.
import { api } from '@home-tools/ui/api'
import type { Muscle } from '@/features/exercises'
import type { Region } from './muscle-map/body-model'

export type RecoveryState = 'ready' | 'recovering' | 'rest'

/** One muscle's estimate (Go: fitness.MuscleRecovery). */
export type MuscleRecovery = {
  muscle: Muscle
  state: RecoveryState
  ready_at?: string // when it's ready again; missing = it is
  last_worked?: string // start of the latest workout that worked it (last 7 days)
  sets_7d: number // this week's sets; a secondary muscle's set counts half
  exercises: string[] // exercise IDs that worked it this week, most recent first
}

/** Every muscle, in a fixed order. */
export const getRecovery = (userId: string) => api.get<MuscleRecovery[]>(`/api/users/${userId}/recovery`)

/** Plain names: what people call them, not the textbook. */
export const MUSCLE_NAMES: Record<Muscle, string> = {
  abdominals: 'Abs',
  abductors: 'Outer thighs',
  adductors: 'Inner thighs',
  biceps: 'Biceps',
  calves: 'Calves',
  chest: 'Chest',
  forearms: 'Forearms',
  glutes: 'Glutes',
  hamstrings: 'Hamstrings',
  lats: 'Lats',
  lower_back: 'Lower back',
  middle_back: 'Middle back',
  neck: 'Neck',
  quadriceps: 'Quads',
  shoulders: 'Shoulders',
  traps: 'Traps',
  triceps: 'Triceps',
}

/** Where each muscle is on the drawing. Lats and middle back share one region. */
export const MUSCLE_REGIONS: Record<Muscle, Region[]> = {
  abdominals: ['abs', 'obliques'],
  abductors: ['abductors'],
  adductors: ['adductor'],
  biceps: ['biceps'],
  calves: ['calves', 'left-soleus', 'right-soleus'],
  chest: ['chest'],
  forearms: ['forearm'],
  glutes: ['gluteal'],
  hamstrings: ['hamstring'],
  lats: ['upper-back'],
  lower_back: ['lower-back'],
  middle_back: ['upper-back'],
  neck: ['neck'],
  quadriceps: ['quadriceps'],
  shoulders: ['front-deltoids', 'back-deltoids'],
  traps: ['trapezius'],
  triceps: ['triceps'],
}

const SEVERITY: Record<RecoveryState, number> = { ready: 0, recovering: 1, rest: 2 }

/** Each region's state: the most worked of the muscles drawn there. */
export function regionStates(list: MuscleRecovery[]): Partial<Record<Region, RecoveryState>> {
  const out: Partial<Record<Region, RecoveryState>> = {}
  for (const m of list) {
    for (const r of MUSCLE_REGIONS[m.muscle]) {
      const was = out[r]
      if (!was || SEVERITY[m.state] > SEVERITY[was]) out[r] = m.state
    }
  }
  return out
}

export const STATE_LABELS: Record<RecoveryState, string> = {
  rest: 'Needs rest',
  recovering: 'Recovering',
  ready: 'Ready',
}

/** The parts of a day, as people say them. */
function partOfDay(d: Date) {
  const h = d.getHours()
  return h < 5 ? 'night' : h < 12 ? 'morning' : h < 17 ? 'afternoon' : h < 22 ? 'evening' : 'night'
}

/** Calendar days from a to b (both local), ignoring the time. */
function daysBetween(a: Date, b: Date) {
  const day = (d: Date) => Date.UTC(d.getFullYear(), d.getMonth(), d.getDate()) / 86_400_000
  return day(b) - day(a)
}

/** "later today", "tomorrow morning", "Thursday". */
export function readyWhen(iso: string, now = new Date()): string {
  const d = new Date(iso)
  const days = daysBetween(now, d)
  if (days <= 0) return 'later today'
  if (days === 1) return `tomorrow ${partOfDay(d)}`
  return d.toLocaleDateString(undefined, { weekday: 'long' })
}

/** "today", "yesterday", "Monday", "Oct 2" (more than a week ago). */
export function workedWhen(iso: string, now = new Date()): string {
  const d = new Date(iso)
  const days = daysBetween(d, now)
  if (days <= 0) return 'today'
  if (days === 1) return 'yesterday'
  if (days < 7) return d.toLocaleDateString(undefined, { weekday: 'long' })
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

/** "Chest", "Chest and triceps", "Chest, triceps and 2 more". */
export function muscleList(muscles: Muscle[]): string {
  const names = muscles.map((m, i) => (i === 0 ? MUSCLE_NAMES[m] : MUSCLE_NAMES[m].toLowerCase()))
  if (names.length <= 2) return names.join(' and ')
  if (names.length === 3) return `${names[0]}, ${names[1]} and ${names[2]}`
  return `${names[0]}, ${names[1]} and ${names.length - 2} more`
}
