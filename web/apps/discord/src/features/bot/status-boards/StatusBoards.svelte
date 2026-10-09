<!--
  Game statuses in Discord (decision #42): each row puts one game server's
  live status in one channel, as a message the bot keeps editing. Set here
  only (no command). Removing a row deletes its message in Discord, so it
  asks first (inline confirm, focus on the safe choice).
-->
<script lang="ts">
  import { tick } from 'svelte'
  import type { BotView, Config, StatusBoard } from '@/features/bot'

  let {
    config,
    bot,
    save,
  }: { config: Config; bot: BotView; save: (change: (c: Config) => void) => Promise<string> } = $props()

  let serverId = $state('')
  let channelId = $state('')
  let busy = $state(false)
  let error = $state('')
  let confirming = $state<string | null>(null) // key of the row asking "remove?"

  const key = (b: StatusBoard) => `${b.server_id}:${b.channel_id}`

  // O(1) lookups for each row's labels.
  let servers = $derived(new Map(bot.servers.map((s) => [s.id, s.name])))
  let channels = $derived(
    new Map(bot.guilds.flatMap((g) => g.channels.map((c) => [c.id, { name: c.name, guild: g.name }] as const))),
  )
  let taken = $derived(new Set(config.status_boards.map(key)))
  let duplicate = $derived(taken.has(`${serverId}:${channelId}`))

  async function add(event: SubmitEvent) {
    event.preventDefault()
    if (!serverId || !channelId || duplicate || busy) return
    busy = true
    error = await save((c) => c.status_boards.push({ server_id: serverId, channel_id: channelId }))
    busy = false
    if (!error) serverId = channelId = ''
  }

  async function askRemove(b: StatusBoard) {
    confirming = key(b)
    await tick()
    document.getElementById(`keep-${confirming}`)?.focus()
  }

  async function keep(b: StatusBoard) {
    confirming = null
    await tick()
    document.getElementById(`remove-${key(b)}`)?.focus()
  }

  async function remove(b: StatusBoard) {
    busy = true
    error = await save((c) => (c.status_boards = c.status_boards.filter((x) => key(x) !== key(b))))
    busy = false
    confirming = null
  }
</script>

<section class="boards" aria-labelledby="boards-title">
  <h2 id="boards-title">Game statuses</h2>
  <p class="muted">
    A live status message for a game server in a Discord channel: running or not, uptime, who's online. The bot
    edits the same message as things change.
  </p>

  {#if config.status_boards.length > 0}
    <ul class="list">
      {#each config.status_boards as b (key(b))}
        {@const ch = channels.get(b.channel_id)}
        <li class="row">
          {#if confirming === key(b)}
            <p class="question" id="question-{key(b)}">Remove this status? Its message in Discord is deleted too.</p>
            <div class="actions">
              <button type="button" id="keep-{key(b)}" class="btn btn-secondary" onclick={() => keep(b)}>Keep it</button>
              <button type="button" class="btn btn-danger" onclick={() => remove(b)} disabled={busy}>
                {busy ? 'Removing…' : 'Remove'}
              </button>
            </div>
          {:else}
            <div class="label">
              <span class="name">{servers.get(b.server_id) ?? `Unknown server “${b.server_id}”`}</span>
              <span class="muted">
                in {ch ? `#${ch.name}` : 'a channel the bot can’t see'}{ch && bot.guilds.length > 1 ? ` · ${ch.guild}` : ''}
              </span>
            </div>
            <button
              type="button"
              id="remove-{key(b)}"
              class="btn btn-icon"
              aria-label="Remove {servers.get(b.server_id) ?? b.server_id} status"
              onclick={() => askRemove(b)}
            >
              <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12 M18 6L6 18" /></svg>
            </button>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}

  {#if !bot.connected}
    <p class="muted">Connect the bot to pick a channel.</p>
  {:else if bot.servers_error}
    <p class="muted">{bot.servers_error}</p>
  {:else if bot.servers.length === 0}
    <p class="muted">No game servers yet: add them in the games tool first.</p>
  {:else}
    <form class="add" onsubmit={add}>
      <div class="field">
        <label for="board-server">Server</label>
        <select id="board-server" bind:value={serverId} required>
          <option value="" disabled>Pick a server</option>
          {#each bot.servers as s (s.id)}
            <option value={s.id}>{s.name}</option>
          {/each}
        </select>
      </div>
      <div class="field">
        <label for="board-channel">Channel</label>
        <select id="board-channel" bind:value={channelId} required aria-describedby="board-error">
          <option value="" disabled>Pick a channel</option>
          {#each bot.guilds as g (g.id)}
            <optgroup label={g.name}>
              {#each g.channels as c (c.id)}
                <option value={c.id} disabled={!c.can_post}>#{c.name}{c.can_post ? '' : ' (no permission to post)'}</option>
              {/each}
            </optgroup>
          {/each}
        </select>
      </div>
      <button type="submit" class="btn btn-primary" disabled={busy || !serverId || !channelId || duplicate}>
        {busy && !confirming ? 'Adding…' : 'Add status'}
      </button>
    </form>
    {#if duplicate}<p class="muted" id="board-error">That server is already shown in that channel.</p>{/if}
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
</section>

<style>
  .boards {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
  }

  .error {
    color: var(--color-danger-text);
  }

  .list {
    display: flex;
    flex-direction: column;
    margin: 0;
    padding: 0;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
    list-style: none;
  }

  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    min-height: calc(var(--touch-target) + var(--space-4));
    padding: var(--space-2) var(--space-2) var(--space-2) var(--space-4);
  }

  .row + .row {
    border-top: 1px solid var(--color-border);
  }

  .label {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .name {
    font-weight: var(--weight-semibold);
    overflow-wrap: anywhere;
  }

  .label .muted {
    font-size: var(--text-sm);
    overflow-wrap: anywhere;
  }

  .question {
    flex: 1 1 14rem;
  }

  .actions {
    display: flex;
    gap: var(--space-2);
  }

  .add {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    gap: var(--space-3);
  }

  .field {
    display: flex;
    flex: 1 1 12rem;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0;
  }

  label {
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
  }

  select {
    min-width: 0;
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-bg);
    font-weight: var(--weight-medium);
    cursor: pointer;
  }
</style>
