# Architecture decision records

One file per decision that shapes the system: why it was made, what it
costs, what was rejected. Smaller decisions live only in the decision log
(`AGENTS.md` §4); each ADR names its log number(s).

| ADR | Decision | Log |
|---|---|---|
| [0001](0001-go-and-svelte.md) | Go backend, Svelte + Vite SPAs, no SSR | #3 |
| [0002](0002-json-files-not-a-database.md) | JSON files, not a database | #4, #15 |
| [0003](0003-one-binary-tools-as-packages.md) | One binary; tools are packages; host-based routing | #5 |
| [0004](0004-one-spa-per-tool.md) | One SPA per tool | #6 |
| [0005](0005-https-on-the-lan.md) | HTTPS on the LAN: Let's Encrypt via Cloudflare DNS-01, Caddy in front | #8, #9 |
| [0006](0006-no-authentication.md) | No authentication; a profile picker | #10 |
| [0007](0007-shared-profiles.md) | Profiles are shared by all tools | #24 |
| [0008](0008-docker-with-bind-mounted-data.md) | One Docker image; data in a bind-mounted host folder | #31 |
| [0009](0009-ci-images-on-ghcr.md) | CI publishes public multi-arch images; ZimaOS pulls them | #32 |
| [0010](0010-vocabulary-plan-routine-workout.md) | Vocabulary: Plan, Routine, Workout | #33 |
| [0011](0011-docker-access-via-socket-proxy.md) | Game servers via the Docker API, through a filtered socket proxy | #37 |

## Writing a new one
Copy this, number it next, keep it under a page:

    # NNNN. Title (a decision, not a topic)

    - Status: Accepted, YYYY-MM-DD (log #N)

    ## Context
    What forced a choice; the constraints that mattered.

    ## Decision
    What we do, stated plainly.

    ## Consequences
    What gets easier, what gets harder, what we now must keep doing.

    ## Rejected
    The serious alternatives and the one reason each lost.

A decision that changes later gets a new ADR that says "Supersedes NNNN";
the old one's status becomes "Superseded by NNNN". Never rewrite history.
