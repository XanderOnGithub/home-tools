# Fitness tool: agent guide

Workout logging for the household at `fitness.<domain>`. Read the root
`AGENTS.md` first (role, principles, decision log); this file is what's
specific to fitness. Words follow ADR 0010: **Plan** (shared list of
exercises), **Routine** (a person's weekly schedule of plans), **Workout**
in the UI = **Session** in the code.

## Where things are
| Go file | What |
|---|---|
| `model.go` | Types and enums: Exercise, Plan, Session, Set, Profile, WeightEntry. Start here. |
| `validate.go` | Each type's own rules (`Validate`), with `ErrInvalid` / `ErrConflict`. |
| `store.go` | `Store`: loads everything at `Open`, serves reads from memory, writes through to JSON. Exercises, plans, sessions. |
| `progress.go` | Progress reads (#39): exercise log + one exercise's all-time history. |
| `body.go` | Fitness profile (`fitness.json`, incl. the routine) and the weight log. |
| `handlers.go` | HTTP API (`Register`). |
| `fedb.go` | free-exercise-db → Exercise mapping (`fedbMetrics`, holds, distance cardio). |
| `../../cmd/fitness-import` | Imports the catalog + photos; `-fix-metrics` re-applies metric rules. |

The UI is `web/apps/fitness` (Svelte); its `src/features/<feature>/`
folders mirror these concepts: `exercises` (catalog, picker, photos),
`plans` (list, editor, routine planner), `sessions` (API + helpers),
`workout` (workout mode, start dialog, rest timer, chime), `home`,
`progress` (summary, weight + exercise charts, workout list), `onboarding`, `fitness-profile`, `profiles` (shared picker,
avatar blob), `shell` (layout, nav).

## Data on disk (`data/fitness/`)
    exercises/<id>.json             catalog (876 imported + any added)
    images/<Exercise_Id>/0.jpg, 1.jpg   photos, start and end position
    plans/<id>.json                 shared plans
    users/<user id>/fitness.json    profile + routine ("schedule"); missing = not onboarded
    users/<user id>/weights.json    [{date, weight_kg}], ≤ 1 per day
    users/<user id>/sessions/<start>.json   one file per workout, named by start time

Every `<user id>` must exist in the shared profile store (`internal/users`,
ADR 0007). Units are metric with the unit in the key (`weight_kg`,
`duration_sec`, `distance_m`); the UI converts per profile.

## Rules and who enforces them
- **A type's own rules** → its `Validate()` in `validate.go`, table-tested.
  - A set must fill every metric its exercise tracks, and nothing else.
    Exceptions: weight on bodyweight exercises (0 = bodyweight only) and
    distance (0 = not measured).
  - Plan hints: `suggested_sets ≥ 0`, `rest_sec` 0–3600 (0 = 60 s
    default, the constant lives in the UI: `DEFAULT_REST_SEC`).
- **References** → the store (it has the data): exercises a session or
  plan names exist, users exist, routine days name existing plans.
- **History stays valid** → `SaveExercise` refuses a metrics change that
  would make logged sets invalid; `SaveSession` with an empty ID
  (create) refuses to overwrite a session from the same second (409).
- **Nothing is deleted** → `archived: true` on Exercise, Plan, Session.
  Lists return archived items (history needs their names); the UI hides
  them from pickers. "Discard workout" = archive.
- **Files are trusted no more than requests** → `Open` decodes strictly,
  validates every file, checks its ID matches its filename and its
  references resolve. One bad file stops startup with its path (ADR 0002).

## API
All JSON; bodies are decoded strictly (unknown fields → 400).

| Method + path | Does |
|---|---|
| `GET /api/exercises`, `GET /api/exercises/{id}` | Catalog (archived included) / one exercise |
| `PUT /api/exercises/{id}` | Create or replace (`httpx.PutByID`) |
| `GET /api/plans`, `PUT /api/plans/{id}` | Plans by name (archived included) / create or replace |
| `GET /api/users/{user}/sessions?limit=n` | Newest first, archived skipped; `limit` 1–100, default 20 |
| `POST /api/users/{user}/sessions` | Start a workout; server assigns the ID (201; 409 same second) |
| `PUT /api/users/{user}/sessions/{id}` | Save sets, finish (`ended_at`), archive |
| `GET /api/users/{user}/exercise-log` | Progress (#39): exercises with logged sets, `{exercise_id, last_done, workouts}`, most recent first |
| `GET /api/users/{user}/exercise-log/{exercise}` | One exercise's history, all time, newest first: `{session_id, started_at, sets}` per workout. Both count only finished, non-archived workouts |
| `GET`/`PUT /api/users/{user}/fitness` | Fitness profile + routine (404 = not onboarded) |
| `GET /api/users/{user}/weights`, `PUT …/weights/{date}` | Weight log (upsert by date) |
| `GET /images/<path>` | Exercise photos (no listings, 1-day cache) |

Profiles themselves (`/api/users`) come from `internal/users`.

## Adding a feature
1. **Decide first.** New fields, files or screens are Xander's call:
   propose, then record it in the root decision log (§4).
2. **Model → rule → test:** add the field in `model.go` (snake_case JSON,
   unit in the name, `omitempty`/`omitzero` if optional), its rule in
   `validate.go`, a table-test row.
3. **Store:** references or invariants across files go in `store.go` /
   `body.go`, with a test. Keep writes atomic (`jsonfile.Write`) under the
   store's lock.
4. **API:** a route in `Register`; reuse `httpx` helpers (`PutByID`,
   `DecodeJSON`, `WriteError`, `ServerError`). `ErrInvalid` → 400 with
   its message; anything else → logged 500.
5. **UI:** mirror the type in the feature's `index.ts`, build the screen
   from the shared pieces in `web/DESIGN.md` §7, check every state
   (loading, empty, error, long content) at phone width.
6. **Check:** `make check` and `pnpm --dir web check`; update this file
   and the root `AGENTS.md` status in the same change.

## Gotchas
- **Renaming or removing a stored field is a migration.** Strict loading
  rejects the old name, and the server won't start. Either migrate the
  files or (if there's nothing worth keeping) say so and delete them.
- **Changing how imported exercises are tracked** = change `fedbMetrics`,
  then on each server run `fitness-import -fix-metrics` with the app
  stopped (`deploy/README.md`).
- **The server reads files only at startup.** Hand edits and one-off tools
  need the app stopped, then started.
- **Session IDs are the start time to the second** (`2026-10-08T18-00-00Z`).
- **Weights are stored to 0.01 kg**; the UI remembers the exact stored
  value behind a rounded lb pre-fill so re-saving doesn't drift.
- **Phone-only behavior:** Screen Wake Lock needs HTTPS; the rest chime
  needs a tap first (iOS); vibration is Android-only. Test on a phone.
