# Discord tool: agent guide

The household's Discord bot and its settings page at `discord.<domain>`
(decision #42, ADR 0012). "Discord" is a working name; the tool will be
renamed. Read the root `AGENTS.md` first; this file is what's specific
to the bot.

## What it does
- `/sens`: mouse sensitivity between 17 games (+ DPI, cm/360°). Anyone.
- `/restart <server>`: verified people only; one at a time per server
  (games answers 409), and not again within 2 minutes (`restartCooldown`).
- `/whitelist add|remove|list <server> [player]`: verified people only.
  Minecraft names over RCON; Valheim SteamID64s (or a
  `steamcommunity.com/profiles/…` link) on its permitted list, applied at
  the next restart (the reply says so).
- **Status boards:** per (server, channel) one message the bot keeps
  editing. Set on the settings page only.
- **Persona:** a new name (from the config's list) and color (the 4
  profile colors) every local midnight, as a per-server nickname + avatar.
  The avatar is the household blob (decision #22) for that name, animated.
- Unverified people who try a verified-only command become **access
  requests** on the People page (Verify / Dismiss).

### Features (#44): each with an on/off switch
- **Poll:** a native Discord poll every `every_days` (1–7, default 2) at
  `post_at` (local time) in one channel, open `duration_hours`.
  Questions are edited on the Features page; `rotate` shuffles them so
  none repeat before all were asked.
- **`/blob name color [animated]`:** anyone; PNG 512 px (transparent,
  with margin) or the blinking GIF 256 px. The page has a maker.
- **Status:** the bot's custom status, a new phrase every local hour
  (`rotate` keyed by the hour), `{name}` = the day's name. With
  `live_games`, an hour when someone's playing may say "Watching Steve
  play Minecraft" instead, never two hours running. Resent on reconnect
  (Discord forgets it).
- **Adding a feature:** a struct in `Features` (`model.go`, zero value =
  off, defaults in `fillDefaults`, rules in `Validate`); its commands in
  `commands(f)` behind its switch (and a check in `handle`, since Discord
  may lag a moment); its job as a loop in `Run` that re-reads the config
  and wakes on `Changed`; a card on the Features page with
  `ToggleSwitch`.

## Where things are
| File | What |
|---|---|
| `model.go` | `Config` (`revision`, `status_boards`, `verified`, `names`), limits, `Validate`. |
| `store.go` | `config.json` (people edit) + `state.json` (bot writes: board message IDs, access requests ≤ 20). Verified IDs in a set (O(1) per command). |
| `bot.go` | Gateway connection (guilds intent only), command registration, persona loop, caches (games list, avatars), cooldowns, `View` for the page. |
| `commands.go` | Slash command definitions; handlers on plain values (`invocation` → `responder`), so they're tested without Discord; autocomplete. |
| `boards.go` | Status board sync: one games request per tick, edit only when an embed's hash changed. |
| `games.go` | Client for the games HTTP API (`Games` interface; fakes in tests). |
| `sens.go` | The conversion table (from Starport-Assistant) and math. |
| `persona.go` | `PersonaFor(date, names)` and `rotate` (shuffled turns, no repeats back to back), both pure. Shares the blob's seeded generator. |
| `poll.go` | Poll schedule (`nextPollSlot`, pure), the posting loop, "post now". |
| `status.go` | Hourly status: `pickStatus` + `gameStatus` (pure), the loop. |
| `blob.go` | Go port of the frontend blob (shape + face); tested against the TypeScript's numbers. |
| `raster.go` | Scanline polygon fill with anti-aliasing (edge table + active list). |
| `avatar.go` | The animation script (rest → blink ×2 → look left → right → back) → GIF, plus a still PNG. |
| `handlers.go` | Settings API. |

UI: `web/apps/discord` (Svelte, shared `AppShell`): `features/bot/`:
`settings` (shared state, polls the bot every 15 s while visible; saves
with the revision, reloads on 409), `today-card`, `status-boards`,
`name-list`, `commands-help`, `overview-page` (`/`), `features-page`
(`/features`: `poll-settings` + `poll-questions`, `status-settings`,
`blob-maker`),
`people-page` (`/people`); `components/toggle-switch`.

## Configuration
- **Token:** `DISCORD_TOKEN` env var only (never in files, flags or the
  UI). Without it the bot stays offline; the page says so.
- `data/discord/config.json`: hand-editable, loaded strictly; the bot
  reads the store on every use, so saves apply at once.
- **Time zone:** `TZ` env (e.g. `America/New_York`); "midnight" (persona)
  and poll times are local to it. Unset = UTC. The zone data is built
  into the binary (`time/tzdata` in `main.go`): the image has none.
- `-games-url`: where the games API is. Default: this same server
  (`http://127.0.0.1:<port>`, `Host: games.internal`; host routing only
  looks at the first label).
- Discord developer portal: a bot with **no privileged intents**.
  Invite with the link on the page (view channels, send messages, embed
  links, attach files, read history, change nickname, send polls; no
  admin).

## API
| Method + path | Does |
|---|---|
| `GET /api/config` | The config. |
| `PUT /api/config` | Replace it. Body's `revision` must be current: 409 otherwise (someone else saved). 400 = a rule broke. Answers with the new revision. |
| `GET /api/bot` | Connection (`connected`, `error`, `user`, `invite_url`), `guilds` with postable text channels, today's `persona`, `requests`, games `servers` (+ `servers_error`). |
| `DELETE /api/requests/{user}` | Dismiss an access request. |
| `POST /api/poll` | Post the next poll now (takes the next slot). 409 = off, no questions, offline, or Discord refused. |
| `GET /api/blob?name=&color=[&animated=1]` | The blob `/blob` would send (immutable cache). |
| `GET /api/persona.gif` / `.png` | Today's avatar, animated / still (ETag per persona). |

## Gotchas
- **Discord answers within 3 s or the command fails.** Slow work
  (restart, whitelist) replies first, then edits that reply (valid 15 min).
- **Commands are registered globally on every connect** (bulk overwrite:
  idempotent; removed commands disappear). New options can take a moment
  to show in Discord's client.
- **Autocomplete isn't enforced by Discord:** people can type any server;
  `pickServer` checks it.
- **Per-server avatars:** Modify Current Member with `avatar`. If an
  animated one is refused, the still PNG is tried, then nickname only.
  Unverified at the time of writing (2026-10-09): check the bot log on
  first deploy.
- **Status boards survive restarts** through `state.json`; a message
  deleted in Discord is posted again; a board removed in the UI has its
  message deleted.
- **Local dev:** `make run TOOL=discord` serves only Discord, so its games
  calls 404. Run games separately (`go run ./cmd/home-tools -tool games
  -addr :8081`) and start Discord with `-games-url http://127.0.0.1:8081`.
  A test bot (own token, own test server) avoids touching the real one.
