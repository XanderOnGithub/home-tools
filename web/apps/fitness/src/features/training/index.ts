// What a plan or a routine works (decision #51): sets per muscle and
// plain-language checks, so you can tell while building it whether it's
// a good one. Pure functions of the plans and the catalog; nothing is
// stored or fetched here.
//
// Sets per muscle = suggested sets × activation (a primary muscle counts
// the whole set, a secondary one half), like the recovery model (#49).
// Cardio and stretches don't count: they aren't strength sets.
//
// Thresholds, from docs/research/muscle-recovery.md and the common
// guideline for building muscle:
//   - one workout: 1–2 sets light, 3–5 solid, 6+ a lot; 10+ gets a note
//     (6–8 hard sets already take 3–4 days to recover from)
//   - one week: about 10 hard sets per muscle or more to build it
//   - the same muscle worked hard (3+ sets) on back-to-back days won't
//     have recovered (2–3 days)
import type { Exercise, Muscle } from '@/features/exercises'
import { MUSCLE_NAMES, type MuscleRecovery } from '@/features/body'
import type { Weekday } from '@/features/fitness-profile'
import type { Plan } from '@/features/plans'

/** Sets a plan exercise counts for when it has no suggestion. */
export const DEFAULT_PLAN_SETS = 3

/** 0 = not worked, 1 = light, 2 = solid, 3 = a lot (one workout). */
export type Level = 0 | 1 | 2 | 3

export const LEVEL_LABELS: Record<Exclude<Level, 0>, string> = { 3: 'A lot', 2: 'Solid', 1: 'Light' }

export function level(sets: number): Level {
  if (sets <= 0) return 0
  return sets < 3 ? 1 : sets < 6 ? 2 : 3
}

/** A note about a plan or week: info (worth knowing) or warn (worth fixing). */
export type Check = { tone: 'info' | 'warn'; text: string }

export type Load = Map<Muscle, number> // muscle → sets

/** Sets per muscle for one workout's exercises. O(exercises × muscles). */
export function planLoad(exercises: Plan['exercises'], byId: Map<string, Exercise>): Load {
  const load: Load = new Map()
  for (const item of exercises) {
    const ex = byId.get(item.exercise_id)
    if (!ex || ex.category === 'cardio' || ex.category === 'stretching') continue
    const sets = item.suggested_sets || DEFAULT_PLAN_SETS
    for (const [m, a] of Object.entries(ex.activation) as [Muscle, number][]) {
      load.set(m, (load.get(m) ?? 0) + sets * a)
    }
  }
  return load
}

const get = (load: Load, m: Muscle) => load.get(m) ?? 0
// Lats and middle back are one "back" to people (and on the map).
const back = (load: Load) => Math.max(get(load, 'lats'), get(load, 'middle_back'))

/** Opposite muscle pairs, with what to add for the weaker side. */
const PAIRS: { a: string; b: string; sets: (l: Load) => [number, number]; addA: string; addB: string }[] = [
  { a: 'chest', b: 'back', sets: (l) => [get(l, 'chest'), back(l)], addA: 'a push-up or press', addB: 'a row or pull-up' },
  { a: 'quads', b: 'hamstrings', sets: (l) => [get(l, 'quadriceps'), get(l, 'hamstrings')], addA: 'a squat or lunge', addB: 'a Romanian deadlift or leg curl' },
  { a: 'biceps', b: 'triceps', sets: (l) => [get(l, 'biceps'), get(l, 'triceps')], addA: 'a curl', addB: 'a dip or push-down' },
]

/** Pairs where one side has 3+ sets and the other under a third of that. */
function imbalances(load: Load) {
  const out: { strong: string; weak: string; none: boolean; add: string }[] = []
  for (const p of PAIRS) {
    const [a, b] = p.sets(load)
    if (a >= 3 && b < a / 3) out.push({ strong: p.a, weak: p.b, none: b === 0, add: p.addB })
    else if (b >= 3 && a < b / 3) out.push({ strong: p.b, weak: p.a, none: a === 0, add: p.addA })
  }
  return out
}

const rounded = (n: number) => Math.max(1, Math.round(n))

/** Notes for one workout: a lot for one muscle, and what it pairs well with. */
export function planChecks(load: Load): Check[] {
  const checks: Check[] = []
  for (const [m, sets] of load) {
    if (sets >= 10) {
      checks.push({
        tone: 'warn',
        text: `${MUSCLE_NAMES[m]}: about ${rounded(sets)} sets is a lot for one workout. 6–8 hard sets already take 3–4 days to recover from; splitting them over two days gets more out of them.`,
      })
    }
  }
  // One plan can be a push or a leg day on purpose: say what it pairs
  // with instead of calling it wrong. The week view checks the balance.
  for (const i of imbalances(load)) {
    checks.push({
      tone: 'info',
      text: `Lots of ${i.strong}, ${i.none ? 'nothing' : 'little'} for the ${i.weak}: it pairs well with a day that works the ${i.weak} (${i.add}).`,
    })
  }
  return checks
}

// ---- A week (the routine) ----

export const WEEKDAYS: Weekday[] = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday']
const DAY_NAMES: Record<Weekday, string> = {
  monday: 'Monday', tuesday: 'Tuesday', wednesday: 'Wednesday', thursday: 'Thursday',
  friday: 'Friday', saturday: 'Saturday', sunday: 'Sunday',
}

/** The muscles a week should cover, as people name them (back = lats or middle back). */
const MAIN: { name: string; sets: (l: Load) => number }[] = [
  { name: 'chest', sets: (l) => get(l, 'chest') },
  { name: 'back', sets: back },
  { name: 'shoulders', sets: (l) => get(l, 'shoulders') },
  { name: 'quads', sets: (l) => get(l, 'quadriceps') },
  { name: 'hamstrings', sets: (l) => get(l, 'hamstrings') },
  { name: 'glutes', sets: (l) => get(l, 'glutes') },
  { name: 'biceps', sets: (l) => get(l, 'biceps') },
  { name: 'triceps', sets: (l) => get(l, 'triceps') },
]

export type WeekLoad = { days: Map<Weekday, Load>; total: Load }

/** Sets per muscle per scheduled day, and for the whole week. */
export function weekLoad(schedule: Partial<Record<Weekday, string>> | undefined, plans: Plan[], byId: Map<string, Exercise>): WeekLoad {
  const days = new Map<Weekday, Load>()
  const total: Load = new Map()
  for (const day of WEEKDAYS) {
    const plan = plans.find((p) => p.id === schedule?.[day])
    if (!plan) continue
    const load = planLoad(plan.exercises, byId)
    days.set(day, load)
    for (const [m, s] of load) total.set(m, (total.get(m) ?? 0) + s)
  }
  return { days, total }
}

/** "chest", "chest and back", "chest, back and quads". */
function list(names: string[]) {
  return names.length <= 1 ? names.join('') : `${names.slice(0, -1).join(', ')} and ${names.at(-1)}`
}

/** Notes for the week: muscles under 10 sets, imbalances, back-to-back days. */
export function weekChecks(week: WeekLoad): Check[] {
  if (week.days.size === 0) return []
  const checks: Check[] = []

  const low = MAIN.filter((g) => g.sets(week.total) < 10).map((g) => g.name)
  if (low.length > 0) {
    checks.push({
      tone: 'warn',
      text: `Under 10 sets a week: ${list(low)}. About 10 or more hard sets a week is what builds a muscle.`,
    })
  }

  for (const i of imbalances(week.total)) {
    checks.push({ tone: 'warn', text: `Lots of ${i.strong}, ${i.none ? 'nothing' : 'little'} for the ${i.weak} this week: add ${i.add}.` })
  }

  // Back-to-back days, Sunday → Monday included (the week repeats).
  for (const [i, day] of WEEKDAYS.entries()) {
    const next = WEEKDAYS[(i + 1) % 7]
    const a = week.days.get(day)
    const b = week.days.get(next)
    if (!a || !b || day === next) continue
    const both = [...a.keys()].filter((m) => get(a, m) >= 3 && get(b, m) >= 3).map((m) => MUSCLE_NAMES[m].toLowerCase())
    if (both.length > 0) {
      const what = list(both)
      checks.push({
        tone: 'warn',
        text: `${what[0].toUpperCase()}${what.slice(1)} worked hard on ${DAY_NAMES[day]} and ${DAY_NAMES[next]}: they need 2–3 days to recover. Try a rest day or a different plan in between.`,
      })
    }
  }
  return checks
}

// ---- Today: a plan against the body's recovery (#52) ----

/** A plan works a muscle hard from this many sets on (Solid or more). */
const HARD = 3

/** What a plan works hard that isn't ready yet: needs rest, or still recovering. */
export function planReadiness(load: Load, recovery: MuscleRecovery[]) {
  const hard = recovery.filter((r) => get(load, r.muscle) >= HARD)
  return {
    rest: hard.filter((r) => r.state === 'rest'),
    recovering: hard.filter((r) => r.state === 'recovering'),
  }
}

/**
 * The plan that suits today best: nothing it works hard needs rest, and
 * as much of its work as possible lands on ready muscles. undefined when
 * no plan fits. O(plans × exercises).
 */
export function bestPlan(plans: Plan[], byId: Map<string, Exercise>, recovery: MuscleRecovery[], except?: string): Plan | undefined {
  const state = new Map(recovery.map((r) => [r.muscle, r.state]))
  let best: { plan: Plan; score: number } | undefined
  for (const plan of plans) {
    if (plan.archived || plan.id === except || plan.exercises.length === 0) continue
    const load = planLoad(plan.exercises, byId)
    if (planReadiness(load, recovery).rest.length > 0) continue
    let score = 0
    for (const [m, sets] of load) if (state.get(m) === 'ready') score += sets
    if (score > 0 && (!best || score > best.score)) best = { plan, score }
  }
  return best?.plan
}

/** The main muscles (as people name them) that are ready to train. */
export function readyGroups(recovery: MuscleRecovery[]): string[] {
  const ready = new Set(recovery.filter((r) => r.state === 'ready').map((r) => r.muscle))
  const groups: [string, Muscle[]][] = [
    ['chest', ['chest']],
    ['back', ['lats', 'middle_back']],
    ['shoulders', ['shoulders']],
    ['arms', ['biceps', 'triceps']],
    ['legs', ['quadriceps', 'hamstrings', 'glutes']],
    ['core', ['abdominals']],
  ]
  return groups.filter(([, ms]) => ms.every((m) => ready.has(m))).map(([name]) => name)
}

export { list as listNames }
