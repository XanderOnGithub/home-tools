# 0012. The Discord bot is a tool in the binary; it reaches games over HTTP

- Status: Accepted, 2026-10-09 (log #41, #42)

## Context
The old bot (Starport-Assistant) was a separate Go program nobody
configured but its author. The household wants a bot that restarts and
whitelists game servers for trusted people, posts live statuses, and is
configured from a LAN page like every other tool. Everything stays
LAN-only (#2): no port forwarding, no tunnel.

## Decision
- The bot is `tools/discord`, mounted like any tool (`discord.<domain>`
  for its settings) and run inside the same binary. It opens **one
  outbound** websocket to Discord's gateway (`discordgo`, the only new
  dependency); Discord never connects in, so nothing is exposed.
- It reaches game servers through the **games HTTP API** (#41), never by
  importing `tools/games`: the same rules (one action at a time, input
  checks) apply to the web page and the bot, and either tool can be
  removed without touching the other.
- Who may restart or whitelist is the bot's job: a list of verified
  Discord user IDs in its config. Unverified attempts become access
  requests on the settings page.
- The token is an environment variable, never in files, flags or the UI.

## Consequences
- One process still (lowest memory); the bot shares its lifetime: it
  disconnects on shutdown and reconnects on start.
- A localhost HTTP hop per command (sub-millisecond).
- Discord's API is a third party the server talks to (Discord itself
  is the point); browsers never do.
- Changing the games API is now a contract change: the bot is a client.

## Rejected
- **A separate bot process/container:** another app to install and
  update, for no isolation that matters on a trusted LAN.
- **HTTP interactions (webhooks) instead of the gateway:** needs a public
  HTTPS endpoint, which breaks LAN-only.
- **Importing the games package:** couples the tools; removing one breaks
  the other.
- **Writing a gateway client on the standard library:** Go has no
  websocket client in std; reimplementing it and Discord's
  heartbeat/resume logic is a project of its own.
