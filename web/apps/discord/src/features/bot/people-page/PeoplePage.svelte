<!--
  Who may restart servers and edit whitelists (decision #42). Anyone who
  tries without being verified shows up under "Asked for access", so
  verifying is one tap and nobody has to find a Discord user ID. Adding
  by ID is there for everyone else. Removing someone asks first.
-->
<script lang="ts">
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { tick } from 'svelte'
  import { ago, dismissRequest, isSnowflake, type AccessRequest, type VerifiedUser } from '@/features/bot'
  import { settings } from '@/features/bot/settings'

  const s = settings()

  let busy = $state(false)
  let error = $state('')
  let confirming = $state<string | null>(null) // user ID asking "remove?"
  let newName = $state('')
  let newId = $state('')

  let verifiedIds = $derived(new Set(s.config?.verified.map((u) => u.user_id))) // O(1) duplicate check
  let idProblem = $derived.by(() => {
    const id = newId.trim()
    if (!id) return ''
    if (!isSnowflake(id)) return 'A Discord user ID is 17–20 digits.'
    if (verifiedIds.has(id)) return 'They’re already verified.'
    return ''
  })

  async function verify(r: AccessRequest) {
    busy = true
    error = await s.save((c) => c.verified.push({ user_id: r.user_id, name: r.name }))
    busy = false
  }

  async function dismiss(r: AccessRequest) {
    busy = true
    try {
      await dismissRequest(r.user_id)
      await s.refreshBot()
      error = ''
    } catch (err) {
      error = (err as Error).message
    }
    busy = false
  }

  async function add(event: SubmitEvent) {
    event.preventDefault()
    const name = newName.trim()
    const id = newId.trim()
    if (!name || !id || idProblem || busy) return
    busy = true
    error = await s.save((c) => c.verified.push({ user_id: id, name }))
    busy = false
    if (!error) newName = newId = ''
  }

  async function askRemove(u: VerifiedUser) {
    confirming = u.user_id
    await tick()
    document.getElementById(`keep-${u.user_id}`)?.focus()
  }

  async function keep(u: VerifiedUser) {
    confirming = null
    await tick()
    document.getElementById(`remove-${u.user_id}`)?.focus()
  }

  async function remove(u: VerifiedUser) {
    busy = true
    error = await s.save((c) => (c.verified = c.verified.filter((x) => x.user_id !== u.user_id)))
    busy = false
    confirming = null
  }
</script>

<div class="page">
  <h1 tabindex="-1">People</h1>

  {#if s.status === 'error'}
    <LoadError what="the bot's settings" onretry={s.load} />
  {:else if s.config && s.bot}
    {#if error}<p class="error" role="alert">{error}</p>{/if}

    {#if s.bot.requests.length > 0}
      <section class="section" aria-labelledby="requests-title">
        <h2 id="requests-title">Asked for access</h2>
        <ul class="list">
          {#each s.bot.requests as r (r.user_id)}
            <li class="row">
              <div class="label">
                <span class="name">{r.name}</span>
                <span class="muted">Tried /{r.command} · <time datetime={r.at}>{ago(r.at)}</time></span>
              </div>
              <div class="actions">
                <button type="button" class="btn btn-quiet" disabled={busy} onclick={() => dismiss(r)}>Dismiss</button>
                <button type="button" class="btn btn-primary" disabled={busy} onclick={() => verify(r)}>Verify</button>
              </div>
            </li>
          {/each}
        </ul>
      </section>
    {/if}

    <section class="section" aria-labelledby="verified-title">
      <h2 id="verified-title">Verified</h2>
      <p class="muted">They can restart game servers and edit whitelists from Discord.</p>
      {#if s.config.verified.length === 0}
        <p class="muted">
          Nobody yet. When someone tries <code>/restart</code> or <code>/whitelist</code>, they show up here to verify
          with one tap.
        </p>
      {:else}
        <ul class="list">
          {#each s.config.verified as u (u.user_id)}
            <li class="row">
              {#if confirming === u.user_id}
                <p class="question">Remove {u.name}? They won't be able to restart servers anymore.</p>
                <div class="actions">
                  <button type="button" id="keep-{u.user_id}" class="btn btn-secondary" onclick={() => keep(u)}>
                    Keep
                  </button>
                  <button type="button" class="btn btn-danger" disabled={busy} onclick={() => remove(u)}>
                    {busy ? 'Removing…' : 'Remove'}
                  </button>
                </div>
              {:else}
                <div class="label">
                  <span class="name">{u.name}</span>
                  <span class="muted id">{u.user_id}</span>
                </div>
                <button
                  type="button"
                  id="remove-{u.user_id}"
                  class="btn btn-icon"
                  aria-label="Remove {u.name}"
                  onclick={() => askRemove(u)}
                >
                  <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12 M18 6L6 18" /></svg>
                </button>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}

      <details class="by-id">
        <summary>Add someone by Discord user ID</summary>
        <p class="muted">
          In Discord: Settings → Advanced → turn on Developer Mode, then right-click the person → Copy User ID.
        </p>
        <form class="add" onsubmit={add}>
          <div class="field">
            <label for="person-name">Name</label>
            <input id="person-name" bind:value={newName} autocomplete="off" maxlength="100" required />
          </div>
          <div class="field">
            <label for="person-id">User ID</label>
            <input
              id="person-id"
              bind:value={newId}
              inputmode="numeric"
              autocomplete="off"
              required
              aria-invalid={idProblem ? 'true' : undefined}
              aria-describedby={idProblem ? 'person-id-problem' : undefined}
            />
          </div>
          <button type="submit" class="btn btn-secondary" disabled={busy || !newName.trim() || !newId.trim() || !!idProblem}>
            Verify
          </button>
        </form>
        {#if idProblem}<p class="error" id="person-id-problem">{idProblem}</p>{/if}
      </details>
    </section>
  {/if}
</div>

<style>
  .page {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-extrabold);
  }

  h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  .section {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .section + .section {
    margin-top: var(--space-2);
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

  code {
    font-family: var(--font-mono);
    font-size: 0.9em;
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
  }

  .id {
    font-family: var(--font-mono);
  }

  .question {
    flex: 1 1 14rem;
  }

  .actions {
    display: flex;
    gap: var(--space-2);
  }

  /* Keeps the browser's ▸ marker; the padding makes a full touch target. */
  summary {
    padding: var(--space-3) 0;
    color: var(--color-accent-text);
    font-weight: var(--weight-semibold);
    cursor: pointer;
  }

  .by-id > * + * {
    margin-top: var(--space-3);
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

  input {
    min-width: 0;
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-bg);
  }
</style>
