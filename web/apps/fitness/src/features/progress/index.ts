// Progress (decision #39): all-time reads per exercise, and the numbers
// the charts and records are made of. Everything is derived from logged
// sets; nothing here is stored.
import { api } from '@home-tools/ui/api'
import { startOfWeek, isoDate } from '@home-tools/ui/dates'
import { formatDuration, kgToLb, mToKm, mToMi } from '@home-tools/ui/units'
import type { Exercise } from '@/features/exercises'
import type { Session, SetEntry } from '@/features/sessions'

/** An exercise the user has logged sets for (Go: fitness.LoggedExercise). */
export type LoggedExercise = { exercise_id: string; last_done: string; workouts: number }

/** What one workout logged for one exercise (Go: fitness.ExerciseWorkout). */
export type ExerciseWorkout = { session_id: string; started_at: string; sets: SetEntry[] }

/** Exercises with logged sets, most recently done first (finished workouts only). */
export const getExerciseLog = (userId: string) => api.get<LoggedExercise[]>(`/api/users/${userId}/exercise-log`)

/** Every workout with this exercise, all time, newest first. */
export const getExerciseHistory = (userId: string, exerciseId: string) =>
  api.get<ExerciseWorkout[]>(`/api/users/${userId}/exercise-log/${encodeURIComponent(exerciseId)}`)

// ---- Display units (the API is metric, #13) ----

export type Units = { imperial: boolean }

const round = (n: number, places: number) => Math.round(n * 10 ** places) / 10 ** places

export const weightUnit = ({ imperial }: Units) => (imperial ? 'lb' : 'kg')
export const distanceUnit = ({ imperial }: Units) => (imperial ? 'mi' : 'km')
export const toWeight = (kg: number, { imperial }: Units) => round(imperial ? kgToLb(kg) : kg, 1)
export const toDistance = (m: number, { imperial }: Units) => round(imperial ? mToMi(m) : mToKm(m), 2)

/** One set as people say it: "8 × 135 lb", "1:30", "5 km in 25:00". */
export function setLabel(s: SetEntry, u: Units): string {
  const parts: string[] = []
  if (s.reps && s.weight_kg) parts.push(`${s.reps} × ${toWeight(s.weight_kg, u)} ${weightUnit(u)}`)
  else if (s.reps) parts.push(`${s.reps} reps`)
  else if (s.weight_kg) parts.push(`${toWeight(s.weight_kg, u)} ${weightUnit(u)}`)
  if (s.distance_m) parts.push(`${toDistance(s.distance_m, u)} ${distanceUnit(u)}`)
  if (s.duration_sec) parts.push(s.distance_m ? `in ${formatDuration(s.duration_sec)}` : formatDuration(s.duration_sec))
  return parts.join(' ')
}

// ---- An exercise's journey ----

/** One point on a chart: when (ms) and the value in display units. */
export type Point = { t: number; v: number }

export type Highlight = { label: string; value: string; detail?: string }

export type Journey = {
  measure: string // what the chart shows, e.g. "Heaviest set"
  points: Point[] // oldest → newest
  format: (v: number) => string // a chart value with its unit
  highlights: Highlight[] // records, change since the first workout
}

const day = (iso: string) =>
  new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })

const sum = (sets: SetEntry[], key: keyof SetEntry) => sets.reduce((n, s) => n + (s[key] ?? 0), 0)

/** Picks the best item by score (higher wins; the earliest wins ties). */
function best<T>(items: T[], score: (x: T) => number): T | undefined {
  let top: T | undefined
  for (const x of items) if (top === undefined || score(x) > score(top)) top = x
  return top
}

/**
 * What to chart and which records to show for an exercise, from its
 * history (newest first, as the API sends it). What counts as "better"
 * follows what the exercise tracks:
 *   weight (+ reps) → heaviest set; records: heaviest, most reps
 *   reps only (or bodyweight never loaded) → most reps in a set
 *   distance cardio → distance per workout; records: longest, best pace
 *   time only (holds, stretches, cardio without distance) → longest
 */
export function journey(ex: Exercise, history: ExerciseWorkout[], u: Units): Journey {
  const workouts = [...history].reverse() // oldest → newest
  const sets = workouts.flatMap((w) => w.sets.map((s) => ({ s, at: w.started_at })))
  const has = (key: keyof SetEntry) => sets.some(({ s }) => (s[key] ?? 0) > 0)
  const points = (value: (w: ExerciseWorkout) => number) =>
    workouts.map((w) => ({ t: Date.parse(w.started_at), v: value(w) })).filter((p) => p.v > 0)
  const change = (pts: Point[], format: (v: number) => string): Highlight[] => {
    if (pts.length < 2) return []
    const d = pts.at(-1)!.v - pts[0].v
    const sign = d > 0 ? '+' : d < 0 ? '−' : '±'
    return [{ label: 'Since first', value: `${sign}${format(Math.abs(d))}`, detail: `from ${format(pts[0].v)}` }]
  }

  if (ex.metrics.includes('weight') && has('weight_kg')) {
    const unit = weightUnit(u)
    const format = (v: number) => `${round(v, 1)} ${unit}`
    const pts = points((w) => Math.max(...w.sets.map((s) => toWeight(s.weight_kg ?? 0, u))))
    // Heaviest: by weight, then reps. Most reps: by reps, then weight.
    const heavy = best(sets, ({ s }) => (s.weight_kg ?? 0) * 1000 + (s.reps ?? 0))!
    const reps = best(sets, ({ s }) => (s.reps ?? 0) * 10000 + (s.weight_kg ?? 0))!
    const highlights: Highlight[] = [
      { label: 'Heaviest', value: format(toWeight(heavy.s.weight_kg!, u)), detail: `${heavy.s.reps ?? 0} reps · ${day(heavy.at)}` },
    ]
    if (reps.s.reps) {
      highlights.push({ label: 'Most reps', value: `${reps.s.reps}`, detail: `at ${setLabel({ weight_kg: reps.s.weight_kg }, u) || 'bodyweight'} · ${day(reps.at)}` })
    }
    return { measure: 'Heaviest set', points: pts, format, highlights: [...highlights, ...change(pts, format)] }
  }

  if (has('distance_m')) {
    const unit = distanceUnit(u)
    const format = (v: number) => `${round(v, 2)} ${unit}`
    const pts = points((w) => toDistance(sum(w.sets, 'distance_m'), u))
    const far = best(workouts, (w) => sum(w.sets, 'distance_m'))!
    const highlights: Highlight[] = [
      { label: 'Longest', value: format(toDistance(sum(far.sets, 'distance_m'), u)), detail: day(far.started_at) },
    ]
    // Pace only where a workout logged both time and distance.
    const paced = workouts.filter((w) => sum(w.sets, 'distance_m') > 0 && sum(w.sets, 'duration_sec') > 0)
    const perUnit = (w: ExerciseWorkout) => sum(w.sets, 'duration_sec') / toDistance(sum(w.sets, 'distance_m'), u)
    const fast = best(paced, (w) => -perUnit(w))
    if (fast) {
      highlights.push({ label: 'Best pace', value: `${formatDuration(Math.round(perUnit(fast)))} /${unit}`, detail: day(fast.started_at) })
    }
    return { measure: 'Distance', points: pts, format, highlights: [...highlights, ...change(pts, format)] }
  }

  if (has('duration_sec')) {
    const format = (v: number) => formatDuration(Math.round(v))
    const timed = sets.filter(({ s }) => s.duration_sec)
    const pts = points((w) => Math.max(...w.sets.map((s) => s.duration_sec ?? 0)))
    const long = best(timed, ({ s }) => s.duration_sec!)!
    return {
      measure: 'Longest',
      points: pts,
      format,
      highlights: [{ label: 'Longest', value: format(long.s.duration_sec!), detail: day(long.at) }, ...change(pts, format)],
    }
  }

  const format = (v: number) => `${round(v, 0)} reps`
  const pts = points((w) => Math.max(...w.sets.map((s) => s.reps ?? 0)))
  const most = best(sets, ({ s }) => s.reps ?? 0)
  const highlights: Highlight[] = most?.s.reps
    ? [{ label: 'Most reps', value: `${most.s.reps}`, detail: day(most.at) }]
    : []
  return { measure: 'Most reps', points: pts, format, highlights: [...highlights, ...change(pts, format)] }
}

// ---- Summary ----

/**
 * Weeks in a row (Monday–Sunday) with at least one finished workout,
 * counting back from this week. A week not over yet doesn't break it:
 * with nothing logged this week, the streak counts from last week.
 */
export function weekStreak(sessions: Session[], now = new Date()): number {
  const weeks = new Set(sessions.filter((s) => s.ended_at).map((s) => isoDate(startOfWeek(new Date(s.started_at)))))
  const week = startOfWeek(now)
  if (!weeks.has(isoDate(week))) week.setDate(week.getDate() - 7)
  let n = 0
  while (weeks.has(isoDate(week))) {
    n++
    week.setDate(week.getDate() - 7)
  }
  return n
}

/** Finished workouts started in now's calendar month. */
export const workoutsThisMonth = (sessions: Session[], now = new Date()) =>
  sessions.filter((s) => {
    const d = new Date(s.started_at)
    return s.ended_at && d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth()
  }).length
