<!--
  Beside the covers on the servers page: who's playing right now on any
  server, and the latest joins and leaves across all of them (decision
  #40), newest first. Each row names its server; Minecraft rows show the
  player's head. Joins and leaves differ by word and icon, not color.
-->
<script lang="ts">
  import { SvelteSet } from 'svelte/reactivity'
  import { Card } from '@home-tools/ui/components/card'
  import { headUrl, type ActivityEvent, type Server } from '@/features/servers'

  let { servers }: { servers: Server[] } = $props()

  const RECENT = 6
  const noHead = new SvelteSet<string>() // `${server}/${player}` whose head didn't load

  let running = $derived(servers.filter((s) => s.state?.running))
  let playing = $derived(
    running.flatMap((s) => (s.players?.names ?? []).map((name) => ({ server: s, name }))),
  )
  // Players a game counted but didn't name.
  let unnamed = $derived(
    running.reduce((n, s) => n + Math.max(0, (s.players?.online ?? 0) - (s.players?.names.length ?? 0)), 0),
  )
  let recent = $derived(
    servers
      .flatMap((s) => (s.activity ?? []).map((e) => ({ server: s, event: e })))
      .sort((a, b) => Date.parse(b.event.at) - Date.parse(a.event.at))
      .slice(0, RECENT),
  )

  const hasHead = (s: Server, player: string) => s.game === 'minecraft' && !noHead.has(`${s.id}/${player}`)

  function when(iso: string): string {
    const d = new Date(iso)
    const today = new Date()
    const time = d.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' })
    if (d.toDateString() === today.toDateString()) return time
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
  }

  const key = (s: Server, e: ActivityEvent) => `${s.id}/${e.at}/${e.player}/${e.kind}`
</script>

{#snippet face(s: Server, player: string)}
  {#if hasHead(s, player)}
    <img class="head" src={headUrl(s.id, player)} alt="" onerror={() => noHead.add(`${s.id}/${player}`)} />
  {:else}
    <svg class="person" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></svg>
  {/if}
{/snippet}

<div class="now">
  <Card padding="lg">
    <section aria-labelledby="playing-title">
      <div class="head-row">
        <h2 id="playing-title">Playing now</h2>
        {#if playing.length + unnamed > 0}<span class="count">{playing.length + unnamed}</span>{/if}
      </div>
      {#if running.length === 0}
        <p class="muted">No servers are running.</p>
      {:else if playing.length + unnamed === 0}
        <p class="muted">Nobody's playing right now.</p>
      {:else}
        <ul>
          {#each playing as p (p.server.id + '/' + p.name)}
            <li>
              {@render face(p.server, p.name)}
              <span class="text"><span class="name">{p.name}</span></span>
              <a class="where" href="/servers/{encodeURIComponent(p.server.id)}">{p.server.name}</a>
            </li>
          {/each}
        </ul>
        {#if unnamed > 0}
          <p class="muted">{playing.length ? `+${unnamed} more` : `${unnamed} online`}; the game doesn't share names.</p>
        {/if}
      {/if}
    </section>
  </Card>

  <Card padding="lg">
    <section aria-labelledby="recent-title">
      <h2 id="recent-title">Recent activity</h2>
      {#if recent.length === 0}
        <p class="muted">Joins and leaves show up here.</p>
      {:else}
        <ul>
          {#each recent as r (key(r.server, r.event))}
            <li class={r.event.kind}>
              {@render face(r.server, r.event.player)}
              <span class="text">
                <span class="name">{r.event.player}</span>
                {r.event.kind === 'join' ? 'joined' : 'left'}
                <a class="where" href="/servers/{encodeURIComponent(r.server.id)}">{r.server.name}</a>
              </span>
              <time datetime={r.event.at}>{when(r.event.at)}</time>
            </li>
          {/each}
        </ul>
      {/if}
    </section>
  </Card>
</div>

<style>
  .now {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  section {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .head-row {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-3);
  }

  h2 {
    font-size: var(--text-md);
  }

  .count {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-variant-numeric: tabular-nums;
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  ul {
    display: flex;
    flex-direction: column;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    min-height: 2.5rem;
    border-bottom: 1px solid var(--color-border);
    font-size: var(--text-sm);
  }

  li:last-child {
    border-bottom: none;
  }

  .head,
  .person {
    flex: none;
    width: 1.5rem;
    height: 1.5rem;
  }

  .head {
    border-radius: var(--radius-sm);
    image-rendering: pixelated;
  }

  .person {
    fill: none;
    stroke: var(--color-text-muted);
    stroke-width: 2;
    stroke-linecap: round;
  }

  .text {
    flex: 1;
    min-width: 0;
    overflow-wrap: anywhere;
  }

  .name {
    font-weight: var(--weight-medium);
  }

  .leave .text {
    color: var(--color-text-muted);
  }

  .leave .name {
    color: var(--color-text);
  }

  .where {
    color: var(--color-text-muted);
    text-decoration: none;
  }

  .where:hover {
    color: var(--color-text);
    text-decoration: underline;
  }

  li > .where {
    flex: none;
  }

  time {
    flex: none;
    color: var(--color-text-muted);
    font-variant-numeric: tabular-nums;
  }
</style>
