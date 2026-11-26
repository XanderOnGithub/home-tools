<!--
  The last few finished workouts on Home, under Today: what was done,
  when, and how much. "All" opens Progress, which has the full list.
-->
<script lang="ts">
  import type { Plan } from '@/features/plans'
  import type { Session } from '@/features/sessions'

  let { sessions, plans, limit = 3 }: { sessions: Session[]; plans: Plan[]; limit?: number } = $props()

  let recent = $derived(sessions.filter((s) => s.ended_at).slice(0, limit))

  const title = (s: Session) => plans.find((p) => p.id === s.plan_id)?.name ?? 'Workout'
  const sets = (s: Session) => s.entries.reduce((n, e) => n + e.sets.length, 0)
  const minutes = (s: Session) => Math.max(1, Math.round((Date.parse(s.ended_at!) - Date.parse(s.started_at)) / 60_000))
  const plural = (n: number, one: string) => `${n} ${one}${n === 1 ? '' : 's'}`

  function when(iso: string): string {
    const d = new Date(iso)
    const today = new Date()
    const yesterday = new Date(today.getFullYear(), today.getMonth(), today.getDate() - 1)
    if (d.toDateString() === today.toDateString()) return 'Today'
    if (d.toDateString() === yesterday.toDateString()) return 'Yesterday'
    return d.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })
  }
</script>

<section class="recent" aria-labelledby="recent-title">
  <div class="head">
    <h2 id="recent-title">Recent workouts</h2>
    <a class="all" href="/progress">All<span class="visually-hidden"> workouts</span></a>
  </div>

  <ul>
    {#each recent as s (s.id)}
      <li>
        <span class="text">
          <span class="name">{title(s)}</span>
          <span class="meta">
            {plural(s.entries.length, 'exercise')} · {plural(sets(s), 'set')} · {minutes(s)} min
          </span>
        </span>
        <time datetime={s.started_at}>{when(s.started_at)}</time>
      </li>
    {/each}
  </ul>
</section>

<style>
  .recent {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-3);
  }

  .all {
    color: var(--color-accent-text);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
    text-decoration: none;
  }

  .all:hover {
    text-decoration: underline;
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
    padding: var(--space-3) 0;
    border-top: 1px solid var(--color-border);
  }

  .text {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.125rem;
    min-width: 0;
  }

  .name {
    font-weight: var(--weight-semibold);
    overflow-wrap: anywhere;
  }

  .meta,
  time {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  time {
    flex: none;
  }
</style>
