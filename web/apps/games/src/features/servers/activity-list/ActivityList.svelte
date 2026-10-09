<!--
  Who joined and left, newest first (decision #40), read from the
  server's log. Grouped by day ("Today", "Yesterday", "Mon, Oct 6"), each
  row "Steve joined · 6:05 PM". Joins and leaves differ by icon and
  word, never color alone. Minecraft rows show the player's head.
-->
<script lang="ts">
  import { SvelteSet } from 'svelte/reactivity'
  import { headUrl, type ActivityEvent, type Server } from '@/features/servers'

  let { server }: { server: Server } = $props()

  const FIRST = 10
  let all = $state(false)
  const noHead = new SvelteSet<string>()

  let events = $derived(server.activity ?? [])
  let shown = $derived(all ? events : events.slice(0, FIRST))

  const dayKey = (iso: string) => new Date(iso).toDateString()
  function dayLabel(iso: string): string {
    const d = new Date(iso)
    const today = new Date()
    const yesterday = new Date(today.getFullYear(), today.getMonth(), today.getDate() - 1)
    if (d.toDateString() === today.toDateString()) return 'Today'
    if (d.toDateString() === yesterday.toDateString()) return 'Yesterday'
    return d.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })
  }
  const time = (iso: string) => new Date(iso).toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' })

  // Consecutive events on the same day share a heading.
  let days = $derived.by(() => {
    const out: { key: string; label: string; events: ActivityEvent[] }[] = []
    for (const e of shown) {
      const key = dayKey(e.at)
      if (out.at(-1)?.key !== key) out.push({ key, label: dayLabel(e.at), events: [] })
      out.at(-1)!.events.push(e)
    }
    return out
  })
</script>

<section class="activity" aria-labelledby="activity-title">
  <h2 id="activity-title">Activity</h2>

  {#if events.length === 0}
    <p class="muted">No one has joined since the server started.</p>
  {:else}
    {#each days as day (day.key)}
      <h3>{day.label}</h3>
      <ul>
        {#each day.events as e (e.at + e.player + e.kind)}
          <li class={e.kind}>
            {#if server.game === 'minecraft' && !noHead.has(e.player)}
              <img class="head" src={headUrl(server.id, e.player)} alt="" onerror={() => noHead.add(e.player)} />
            {:else}
              <svg class="person" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></svg>
            {/if}
            <span class="text">
              <span class="name">{e.player}</span>
              {e.kind === 'join' ? 'joined' : 'left'}
            </span>
            <svg class="arrow" viewBox="0 0 24 24" aria-hidden="true">
              {#if e.kind === 'join'}<path d="M5 12h14 M13 6l6 6-6 6" />{:else}<path d="M19 12H5 M11 6l-6 6 6 6" />{/if}
            </svg>
            <time datetime={e.at}>{time(e.at)}</time>
          </li>
        {/each}
      </ul>
    {/each}
    {#if events.length > FIRST}
      <button type="button" class="btn btn-quiet" onclick={() => (all = !all)}>
        {all ? 'Show fewer' : `Show all ${events.length}`}
      </button>
    {/if}
  {/if}
</section>

<style>
  .activity {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  h3 {
    margin-top: var(--space-2);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
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
    min-height: var(--touch-target);
    border-bottom: 1px solid var(--color-border);
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
    font-weight: var(--weight-semibold);
  }

  .leave .text {
    color: var(--color-text-muted);
  }

  .leave .name {
    color: var(--color-text);
  }

  .arrow {
    flex: none;
    width: 1rem;
    height: 1rem;
    fill: none;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .join .arrow {
    stroke: var(--color-accent-text);
  }

  .leave .arrow {
    stroke: var(--color-text-muted);
  }

  time {
    flex: none;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-variant-numeric: tabular-nums;
  }
</style>
