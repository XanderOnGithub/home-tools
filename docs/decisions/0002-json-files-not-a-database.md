# 0002. JSON files, not a database

- Status: Accepted, 2026-10-07 (log #4, #15)

## Context
A household's data is small (thousands of records, not millions). The owner
wants data he can read, edit by hand, back up by copying a folder, and keep
forever without a database to administer. Principle 5: never corrupt on
crash.

## Decision
- Every record is one human-readable JSON file under the data folder
  (`data/users/<id>.json`, `data/fitness/plans/<id>.json`,
  `data/fitness/users/<id>/sessions/<start>.json`, …). snake_case keys,
  units in the key (`weight_kg`, `duration_sec`).
- Each store loads everything into memory at startup and serves reads from
  memory, with in-memory indexes (maps, sorted slices) instead of queries.
- Writes are atomic (`internal/jsonfile`: temp file → fsync → rename →
  fsync dir) and fully serialized: the store holds its write lock across
  the disk write (#15).
- Files are trusted no more than API input: on load each must decode
  strictly (unknown fields rejected), pass `Validate`, match its filename
  and resolve its references. Any bad file stops startup with its path.

## Consequences
- Backup = copy the folder. Hand edits are possible (with the server
  stopped: it reads files only at startup).
- Renaming a field is a data migration: strict decoding rejects the old
  name. Plan for it (see the Routine → Plan rename, ADR 0010).
- Writes block reads for a few ms. Fine for household traffic; revisit
  only if measured.
- No ad-hoc queries; every new view needs an index or a scan.
- Memory grows with data. A household's history stays small for decades.

## Rejected
- SQLite: robust, but opaque to hand edits and backups, and a cgo or
  pure-Go driver dependency.
- Postgres: a second service to run and back up.
