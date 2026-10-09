<!-- Past workouts, newest first. -->
<script lang="ts">
  import { LoadError } from '@/components/load-error'
  import type { Profile } from '@/features/profiles/types'
  import { getPlans, type Plan } from '@/features/plans'
  import { getRecentSessions, type Session } from '@/features/sessions'

  let { profile }: { profile: Profile } = $props()

  let sessions = $state<Session[]>([])
  let plans = $state<Plan[]>([])
  let status = $state<'loading' | 'ready' | 'error'>('loading')

  async function load() {
    status = 'loading'
    try {
      ;[sessions, plans] = await Promise.all([getRecentSessions(profile.id), getPlans()])
      status = 'ready'
    } catch (err) {
      console.error('Loading history failed:', err)
      status = 'error'
    }
  }
  load()

  const title = (s: Session) => plans.find((r) => r.id === s.plan_id)?.name ?? 'Workout'
  const when = (s: Session) =>
    new Date(s.started_at).toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })
  const minutes = (s: Session) =>
    s.ended_at ? Math.round((Date.parse(s.ended_at) - Date.parse(s.started_at)) / 60_000) : null
</script>

<h1 tabindex="-1">History</h1>

{#if status === 'error'}
  <LoadError what="your workouts" onretry={load} />
{:else if status === 'ready' && sessions.length === 0}
  <p class="muted">No workouts logged yet.</p>
{:else if status === 'ready'}
  <ul class="list">
    {#each sessions as s (s.id)}
      <li class="row">
        <span class="name">{title(s)}</span>
        <span class="muted">
          <time datetime={s.started_at}>{when(s)}</time>
          · {s.entries.length} {s.entries.length === 1 ? 'exercise' : 'exercises'}
          {#if minutes(s) !== null}· {minutes(s)} min{:else}· in progress{/if}
        </span>
      </li>
    {/each}
  </ul>
{/if}

<style>
  h1 {
    margin-bottom: var(--space-4);
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
  }

  p {
    margin: 0 0 var(--space-3);
  }

  .muted {
    color: var(--color-text-muted);
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .row {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
  }

  .name {
    font-weight: var(--weight-semibold);
  }
</style>
