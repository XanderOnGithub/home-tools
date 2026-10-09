<!--
  One game server: name, game, live status, and what you can do with it.
  Stopped → Start. Running → Restart / Stop, each behind an inline confirm
  (both kick out anyone playing). An action can take up to a minute (the
  game saves its world first), so the buttons say what's happening and
  stay disabled until Docker answers.
-->
<script lang="ts">
  import { tick } from 'svelte'
  import { GAMES, runAction, uptime, type Action, type Server } from '@/features/servers'

  let {
    server,
    onchange,
    logsLink = true,
  }: {
    server: Server
    onchange: (updated: Server) => void
    logsLink?: boolean // the server page shows its log itself
  } = $props()

  const DOING: Record<Action, string> = { start: 'Starting…', stop: 'Stopping…', restart: 'Restarting…' }
  const CONFIRM: Record<'stop' | 'restart', { question: string; button: string }> = {
    stop: { question: 'Stop it?', button: 'Stop' },
    restart: { question: 'Restart it?', button: 'Restart' },
  }

  let pending = $state<Action | null>(null)
  let confirming = $state<'stop' | 'restart' | null>(null)
  let error = $state('')

  let running = $derived(server.state?.running ?? false)

  async function run(action: Action) {
    confirming = null
    pending = action
    error = ''
    try {
      onchange(await runAction(server.id, action))
    } catch (err) {
      error = (err as Error).message
    }
    pending = null
  }

  async function ask(action: 'stop' | 'restart') {
    confirming = action
    await tick()
    document.getElementById(`keep-${server.id}`)?.focus()
  }

  async function keep() {
    const was = confirming
    confirming = null
    await tick()
    document.getElementById(`${was}-${server.id}`)?.focus()
  }
</script>

<article class="card" aria-labelledby="name-{server.id}">
  <div class="head">
    <div>
      <h2 id="name-{server.id}">{server.name}</h2>
      <!-- Only when it adds something ("Old world" → Minecraft). -->
      {#if GAMES[server.game] !== server.name}
        <p class="game">{GAMES[server.game] ?? server.game}</p>
      {/if}
    </div>
    {#if logsLink && server.state}
      <a class="btn btn-quiet" href="/servers/{server.id}">Log</a>
    {/if}
  </div>

  <p class="status" class:on={running} class:unknown={!server.state}>
    <span class="dot" aria-hidden="true"></span>
    {#if pending}
      {DOING[pending]}
    {:else if !server.state}
      Unavailable: {server.error ?? 'unknown error'}
    {:else if running}
      Running{#if server.state.started_at}<span class="muted">&nbsp;· up {uptime(server.state.started_at)}</span>{/if}
    {:else}
      Stopped
    {/if}
  </p>

  {#if server.state}
    {#if confirming}
      <div class="confirm" role="group" aria-labelledby="confirm-{server.id}">
        <p id="confirm-{server.id}">
          {CONFIRM[confirming].question} Anyone playing is disconnected; the world is saved first.
        </p>
        <div class="actions">
          <button type="button" id="keep-{server.id}" class="btn btn-secondary" onclick={keep}>Keep running</button>
          <button type="button" class="btn btn-danger" onclick={() => confirming && run(confirming)}>
            {CONFIRM[confirming].button}
          </button>
        </div>
      </div>
    {:else if running}
      <div class="actions">
        <button type="button" id="restart-{server.id}" class="btn btn-secondary" disabled={!!pending} onclick={() => ask('restart')}>
          Restart
        </button>
        <button type="button" id="stop-{server.id}" class="btn btn-secondary" disabled={!!pending} onclick={() => ask('stop')}>
          Stop
        </button>
      </div>
    {:else}
      <div class="actions">
        <button type="button" class="btn btn-primary" disabled={!!pending} onclick={() => run('start')}>
          {pending === 'start' ? 'Starting…' : 'Start'}
        </button>
      </div>
    {/if}
  {/if}

  {#if error}<p class="error" role="alert">{error}</p>{/if}
</article>

<style>
  .card {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-5);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
  }

  .head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
  }

  h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  p {
    margin: 0;
  }

  .game,
  .muted {
    color: var(--color-text-muted);
    font-weight: var(--weight-medium);
  }

  .game {
    font-size: var(--text-sm);
  }

  /* Status: a dot plus words (color is never the only signal). Running =
   * filled accent dot, stopped = hollow, unavailable = danger text. */
  .status {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    font-weight: var(--weight-semibold);
  }

  .dot {
    width: 0.75rem;
    height: 0.75rem;
    border: 2px solid var(--color-border-strong);
    border-radius: var(--radius-full);
  }

  .on .dot {
    border-color: var(--color-accent);
    background: var(--color-accent);
  }

  .unknown {
    color: var(--color-danger-text);
    font-weight: var(--weight-medium);
  }

  .unknown .dot {
    border-color: var(--color-danger-text);
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }

  .confirm {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-4);
    border-radius: var(--radius-md);
    background: var(--color-danger-subtle);
  }

  .error {
    color: var(--color-danger-text);
  }
</style>
