# 0010. Vocabulary: Plan, Routine, Workout

- Status: Accepted, 2026-10-08 (log #33)

## Context
The first version called a named list of exercises a "routine". In use,
"routine" meant something else: *your* weekly schedule. Mismatched words
between the UI, the code and how people talk make every feature harder to
discuss.

## Decision
- **Plan:** a named, shared list of exercises with hints (suggested sets,
  rest). Go `Plan`, API `/api/plans`, files `fitness/plans/<id>.json`,
  `plan_id` on sessions.
- **Routine:** one person's weekly schedule, which plan on which weekday
  (JSON key `schedule` in their `fitness.json`).
- **Workout** in the UI = **Session** in the code: one logged workout.
- Same words in UI, code, API, files and docs.

## Consequences
- The rename was done everywhere at once, with no data migration (there
  was no data worth keeping). Old files (`routines/`, `routine_id`) stop
  the server at startup until removed (ADR 0002, strict loading).
- New features use these words; no synonyms.

## Rejected
- Renaming only the UI: the code would say the opposite of the screen.
