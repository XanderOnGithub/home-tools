// What to aim for in a set, and how to beat last time (decision #54):
// the "how" next to each exercise in workout mode. Pure functions; the
// longer "why" is the guide page (/guide).
//
//   - Rep range and effort by goal: GOALS below (decision #56). Big lifts
//     (3+ muscles in the activation) and small ones can differ.
//   - Effort: stop 1–2 reps before you couldn't do another (staying fit:
//     2–3). Close enough to count, far enough to recover.
//   - Progress (double progression): add reps until every set reaches the
//     top of the range, then add a little weight and start again at the
//     bottom. Bodyweight: then make it harder (slower, a harder version).
import type { Exercise } from '@/features/exercises'
import type { Goal } from '@/features/fitness-profile'
import type { SetEntry } from '@/features/sessions'

export type RepTarget = { min: number; max: number; effort: string }

export type GoalKey = Goal

type Range = [number, number]
export type GoalInfo = {
  label: string
  big: Range // rep range on big lifts
  small: Range // and on small ones
  effort: string
  note: string // one line on what matters for this goal
}

/** Every goal (Go: fitness.AllGoals), in the order pickers show them. */
export const GOALS: Record<GoalKey, GoalInfo> = {
  muscle: {
    label: 'Build muscle',
    big: [8, 12],
    small: [8, 12],
    effort: 'stop 1–2 short',
    note: 'Size comes from hard sets: about 10 or more per muscle a week.',
  },
  strength: {
    label: 'Get stronger',
    big: [4, 6],
    small: [8, 12],
    effort: 'stop 1–2 short',
    note: 'Heavy big lifts with long rests (2–3 minutes); the small ones keep you balanced.',
  },
  strength_size: {
    label: 'Strength + size',
    big: [6, 10],
    small: [10, 15],
    effort: 'stop 1–2 short',
    note: 'Heavy-ish big lifts for strength, more reps on the small ones for size.',
  },
  fat_loss: {
    label: 'Lose fat, keep muscle',
    big: [6, 10],
    small: [10, 15],
    effort: 'stop 1–2 short',
    note: 'Lift heavy enough to keep your muscle; the fat loss itself comes from eating a little less. Walks help.',
  },
  bone: {
    label: 'Bone health',
    big: [8, 12],
    small: [10, 15],
    effort: 'stop 2–3 short',
    note: 'Standing lifts that load the hips and spine, worked up slowly; heavier (6–10) once form is solid. Walking and stairs count too.',
  },
  general: {
    label: 'Stay fit',
    big: [10, 15],
    small: [10, 15],
    effort: 'stop 2–3 short',
    note: 'Comfortable but not easy: you leave each workout with energy.',
  },
  endurance: {
    label: 'Endurance',
    big: [15, 20],
    small: [15, 20],
    effort: 'stop 2–3 short',
    note: 'Lighter weights, more reps, short rests.',
  },
}

export const GOAL_KEYS = Object.keys(GOALS) as GoalKey[]

const goalInfo = (goal: string | undefined) => GOALS[(goal as GoalKey) in GOALS ? (goal as GoalKey) : 'muscle']

/** "8–12 reps · stop 1–2 short", or "4–6 reps on big lifts, 8–12 on small · …". */
export function goalSummary(goal: string | undefined): string {
  const g = goalInfo(goal)
  const reps = g.big[0] === g.small[0] && g.big[1] === g.small[1]
    ? `${g.big[0]}–${g.big[1]} reps`
    : `${g.big[0]}–${g.big[1]} reps on big lifts, ${g.small[0]}–${g.small[1]} on small ones`
  return `${reps} · ${g.effort}`
}

/** Big lifts work several muscles hard: 3+ muscles in the activation. */
const isBigLift = (ex: Exercise) => Object.keys(ex.activation).length >= 3

/**
 * The rep range for ex under goal (no goal = build muscle), or null if it
 * isn't counted in reps (holds, cardio, stretches).
 */
export function repTarget(ex: Exercise, goal: string | undefined): RepTarget | null {
  if (!ex.metrics.includes('reps') || ex.category === 'cardio' || ex.category === 'stretching') return null
  const g = goalInfo(goal)
  const [min, max] = isBigLift(ex) ? g.big : g.small
  return { min, max, effort: g.effort }
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
