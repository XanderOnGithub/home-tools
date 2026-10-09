# 0006. No authentication; a profile picker

- Status: Accepted, 2026-10-08 (log #10)

## Context
The users are one household on a LAN-only server (ADR 0005). Family
members are not a security boundary from each other, and logging in at
the gym is friction for nothing.

## Decision
- No accounts, passwords or sessions. The app opens on a Netflix-style
  "Who's working out?" picker.
- The chosen profile is remembered in a cookie on `.<domain>` (shared by
  every tool's subdomain) until "Switch profile".
- The API takes the user ID in the URL (`/api/users/{id}/…`); the server
  never reads the cookie.

## Consequences
- Anyone on the LAN can read and change anyone's data. Acceptable at
  home; a reason never to expose the server to the internet.
- A per-profile PIN can be added later as a field without changing URLs.

## Rejected
- Real auth (passwords, OAuth): security theatre inside a home, and a lot
  of code to maintain.
