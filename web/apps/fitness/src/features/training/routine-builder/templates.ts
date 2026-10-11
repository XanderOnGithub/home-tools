// Starter routines (decision #53): built from three answers (days per
// week, equipment, goal) into real, editable plans + a weekly schedule.
// A plan is a list of movement slots (squat, hinge, push…), each filled
// with a beginner-friendly free-exercise-db exercise for the equipment;
// a slot with no exercise for it is skipped. Nothing here is stored
// until the person uses the routine.
import type { Weekday } from '@/features/fitness-profile'
import type { Plan } from '@/features/plans'

export type Days = 2 | 3 | 4
export type Equipment = 'gym' | 'dumbbells' | 'bodyweight'
export type Goal = 'muscle' | 'strength' | 'general'

export const EQUIPMENT_LABELS: Record<Equipment, { label: string; hint: string }> = {
  gym: { label: 'A gym', hint: 'Barbells, machines, cables' },
  dumbbells: { label: 'Dumbbells', hint: 'At home, no bench needed' },
  bodyweight: { label: 'Just my body', hint: 'A sturdy table for rows; arms get less work' },
}

export const GOAL_LABELS: Record<Goal, { label: string; hint: string; reps: string }> = {
  muscle: {
    label: 'Build muscle',
    hint: '3–4 sets, 90 s rest',
    reps: '8–12 reps a set. Stop 1–2 reps before you couldn’t do another: close enough to count, far enough to recover.',
  },
  strength: {
    label: 'Get stronger',
    hint: 'Heavier, 4 sets, longer rests',
    reps: 'Main lifts: 4–6 heavy reps with 2–3 minutes of rest; the smaller exercises 8–12. Stop 1–2 reps short.',
  },
  general: {
    label: 'Stay fit',
    hint: '2 sets, quick workouts',
    reps: '10–15 reps a set, comfortable but not easy: leave 2–3 reps in the tank.',
  },
}

type Slot =
  | 'squat' | 'squat_b' | 'hinge' | 'lunge' | 'leg_curl' | 'calves'
  | 'push' | 'push_b' | 'press' | 'row' | 'row_b' | 'pulldown'
  | 'biceps' | 'triceps' | 'raise' | 'core'

/** Small exercises: fewer sets and less rest than the main lifts. */
const ACCESSORY = new Set<Slot>(['leg_curl', 'calves', 'biceps', 'triceps', 'raise', 'core'])

const EXERCISES: Record<Equipment, Partial<Record<Slot, string>>> = {
  gym: {
    squat: 'Barbell_Squat', squat_b: 'Leg_Press', hinge: 'Romanian_Deadlift', lunge: 'Dumbbell_Lunges',
    leg_curl: 'Lying_Leg_Curls', calves: 'Standing_Calf_Raises',
    push: 'Barbell_Bench_Press_-_Medium_Grip', push_b: 'Incline_Dumbbell_Press', press: 'Dumbbell_Shoulder_Press',
    row: 'Seated_Cable_Rows', row_b: 'One-Arm_Dumbbell_Row', pulldown: 'Wide-Grip_Lat_Pulldown',
    biceps: 'Dumbbell_Bicep_Curl', triceps: 'Triceps_Pushdown', raise: 'Side_Lateral_Raise', core: 'Plank',
  },
  dumbbells: {
    squat: 'Goblet_Squat', squat_b: 'Dumbbell_Step_Ups', hinge: 'Stiff-Legged_Dumbbell_Deadlift', lunge: 'Dumbbell_Lunges',
    leg_curl: 'Single_Leg_Glute_Bridge', calves: 'Calf_Raise_On_A_Dumbbell',
    push: 'Dumbbell_Floor_Press', push_b: 'Pushups', press: 'Dumbbell_Shoulder_Press',
    row: 'One-Arm_Dumbbell_Row', row_b: 'Bent_Over_Two-Dumbbell_Row', pulldown: 'Bent_Over_Two-Dumbbell_Row',
    biceps: 'Dumbbell_Bicep_Curl', triceps: 'Standing_Dumbbell_Triceps_Extension', raise: 'Side_Lateral_Raise', core: 'Plank',
  },
  bodyweight: {
    squat: 'Bodyweight_Squat', hinge: 'Single_Leg_Glute_Bridge', lunge: 'Bodyweight_Walking_Lunge', leg_curl: 'Butt_Lift_Bridge',
    push: 'Pushups', push_b: 'Incline_Push-Up', press: 'Bench_Dips',
    row: 'Inverted_Row', row_b: 'Superman', pulldown: 'Bodyweight_Mid_Row',
    triceps: 'Bench_Dips', core: 'Plank', raise: 'Dead_Bug',
  },
}

type Template = { key: string; name: string; slots: Slot[] }

const TEMPLATES: Record<string, Template> = {
  full_a: { key: 'full-a', name: 'Full body A', slots: ['squat', 'push', 'row', 'press', 'core'] },
  full_b: { key: 'full-b', name: 'Full body B', slots: ['hinge', 'pulldown', 'push_b', 'lunge', 'biceps', 'triceps'] },
  upper_a: { key: 'upper-a', name: 'Upper A', slots: ['push', 'row', 'press', 'biceps', 'triceps'] },
  upper_b: { key: 'upper-b', name: 'Upper B', slots: ['pulldown', 'push_b', 'row_b', 'raise', 'core'] },
  lower_a: { key: 'lower-a', name: 'Lower A', slots: ['squat', 'hinge', 'lunge', 'calves', 'core'] },
  lower_b: { key: 'lower-b', name: 'Lower B', slots: ['squat_b', 'hinge', 'leg_curl', 'calves', 'core'] },
}

/** Which plan on which day: hard days never back to back for the same muscles. */
const WEEKS: Record<Days, [Weekday, Template][]> = {
  2: [['monday', TEMPLATES.full_a], ['thursday', TEMPLATES.full_b]],
  3: [['monday', TEMPLATES.full_a], ['wednesday', TEMPLATES.full_b], ['friday', TEMPLATES.full_a]],
  4: [['monday', TEMPLATES.upper_a], ['tuesday', TEMPLATES.lower_a], ['thursday', TEMPLATES.upper_b], ['friday', TEMPLATES.lower_b]],
}

function hints(goal: Goal, slot: Slot): { suggested_sets: number; rest_sec: number } {
  const small = ACCESSORY.has(slot)
  switch (goal) {
    case 'strength':
      return small ? { suggested_sets: 3, rest_sec: 90 } : { suggested_sets: 4, rest_sec: 150 }
    case 'muscle':
      // 4 sets on the main lifts: what gets most muscles to 10+ sets a week.
      return small ? { suggested_sets: 3, rest_sec: 60 } : { suggested_sets: 4, rest_sec: 90 }
    case 'general':
      return { suggested_sets: 2, rest_sec: 60 }
  }
}

/** "Full body A", "Full body A (dumbbells)", "Full body A (dumbbells, strength)". */
function planName(t: Template, equipment: Equipment, goal: Goal) {
  const extra = [equipment !== 'gym' ? equipment : '', goal !== 'muscle' ? goal : ''].filter(Boolean)
  return extra.length === 0 ? t.name : `${t.name} (${extra.join(', ')})`
}

export type Routine = { plans: Plan[]; schedule: Partial<Record<Weekday, string>> }

/**
 * The routine for these answers. Plan IDs are fixed per template,
 * equipment and goal (starter-full-a-gym-muscle), so the household shares
 * one copy of each: someone else picking the same answers reuses it.
 */
export function buildRoutine(days: Days, equipment: Equipment, goal: Goal, createdBy: string): Routine {
  const plans = new Map<string, Plan>()
  const schedule: Partial<Record<Weekday, string>> = {}
  for (const [day, t] of WEEKS[days]) {
    const id = `starter-${t.key}-${equipment}-${goal}`
    if (!plans.has(id)) {
      const seen = new Set<string>()
      const exercises: Plan['exercises'] = []
      for (const slot of t.slots) {
        const ex = EXERCISES[equipment][slot]
        if (!ex || seen.has(ex)) continue // no exercise for it here, or already in the plan
        seen.add(ex)
        exercises.push({ exercise_id: ex, ...hints(goal, slot) })
      }
      plans.set(id, { id, name: planName(t, equipment, goal), created_by: createdBy, exercises })
    }
    schedule[day] = id
  }
  return { plans: [...plans.values()], schedule }
}
