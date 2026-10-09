# 0007. Profiles are shared by all tools

- Status: Accepted, 2026-10-08 (log #24, #25)

## Context
Every tool needs to know who is using it. Picking yourself again in each
tool, or keeping names and units in several places, would drift.

## Decision
- One profile store in `internal/users`:
  `User{id, name, color, units, birthday?, archived}`, one file per person
  (`data/users/<id>.json`), API `GET /api/users`, `PUT /api/users/{id}`
  mounted on every host.
- Tools keep their own data keyed by the profile ID (fitness:
  `data/fitness/users/<id>/…`) and check it against this store; a tool's
  `Open` fails if its data names an unknown profile.
- Things every tool may need live here (units, color, birthday);
  tool-specific settings live in the tool (fitness: `fitness.json`).

## Consequences
- Pick once, same person and accent color everywhere.
- Profiles are never deleted, only archived, because tools' history
  points at them.
- Loading order matters: users first, then each tool.

## Rejected
- Profiles per tool: duplicate data that drifts.
