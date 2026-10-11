// What to aim for in a set, and how to beat last time (decision #54):
// the "how" next to each exercise in workout mode. Pure functions; the
// longer "why" is the guide page (/guide).
//
//   - Rep range by goal: build muscle 8–12, get stronger 4–6 on big
//     lifts (8–12 on small ones), stay fit 10–15, endurance 15–20. No goal
//     set = build muscle.
//   - Effort: stop 1–2 reps before you couldn't do another (staying fit:
//     2–3). Close enough to count, far enough to recover.
//   - Progress (double progression): add reps until every set reaches the
//     top of the range, then add a little weight and start again at the
//     bottom. Bodyweight: then make it harder (slower, a harder version).
import type { Exercise } from '@/features/exercises'
import type { SetEntry } from '@/features/sessions'

export type RepTarget = { min: number; max: number; effort: string }

/** Big lifts work several muscles hard: 3+ muscles in the activation. */
const isBigLift = (ex: Exercise) => Object.keys(ex.activation).length >= 3

/** The rep range for ex under goal, or null if it isn't counted in reps (holds, cardio). */
export function repTarget(ex: Exercise, goal: string | undefined): RepTarget | null {
  if (!ex.metrics.includes('reps') || ex.category === 'cardio' || ex.category === 'stretching') return null
  switch (goal) {
    case 'strength':
      return isBigLift(ex) ? { min: 4, max: 6, effort: 'stop 1–2 short' } : { min: 8, max: 12, effort: 'stop 1–2 short' }
    case 'general':
      return { min: 10, max: 15, effort: 'stop 2–3 short' }
    case 'endurance':
      return { min: 15, max: 20, effort: 'stop 2–3 short' }
    default:
      return { min: 8, max: 12, effort: 'stop 1–2 short' }
  }
}

/** What to try today, from last time's sets (oldest → newest). */
export function progressTip(ex: Exercise, target: RepTarget, last: SetEntry[]): string {
  const weighted = ex.metrics.includes('weight') && last.some((s) => (s.weight_kg ?? 0) > 0)
  if (last.length === 0) {
    return ex.bodyweight && !ex.metrics.includes('weight')
      ? `First time: do what you can with good form, up to ${target.max} reps.`
      : `First time: pick a ${ex.bodyweight ? 'version' : 'weight'} you could do a few more than ${target.max} times with good form.`
  }
  if (last.every((s) => (s.reps ?? 0) >= target.max)) {
    return weighted || !ex.bodyweight
      ? `Every set reached ${target.max} last time: add a little weight today, and start again around ${target.min}.`
      : `Every set reached ${target.max} last time: make it harder today (slower, or a harder version).`
  }
  return 'Try to beat last time by one rep on a set or two, same weight.'
}
