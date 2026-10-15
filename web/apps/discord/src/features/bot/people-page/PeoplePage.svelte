<!--
  People (decision #61): who may use /restart and /whitelist. Anyone who
  tried and isn't verified asks first (Verify / Dismiss). To add someone
  before they try, search the bot's servers by name and tap Verify: no
  user IDs to look up. Typing an ID is still there for people the search
  can't find (e.g. not in a server the bot is in).
-->
<script lang="ts">
  import { Badge } from '@home-tools/ui/components/badge'
  import { Card } from '@home-tools/ui/components/card'
  import { EmptyState } from '@home-tools/ui/components/empty-state'
  import { Field } from '@home-tools/ui/components/field'
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { PageHeader } from '@home-tools/ui/components/page-header'
  import { Section } from '@home-tools/ui/components/section'
  import { toast } from '@home-tools/ui/toast'
  import { tick } from 'svelte'
  import { SettingsSkeleton } from '@/components/settings-skeleton'
  import {
    ago,
    dismissRequest,
    isSnowflake,
    searchMembers,
    type AccessRequest,
    type Member,
    type VerifiedUser,
  } from '@/features/bot'
  import { MemberAvatar } from '@/features/bot/member-avatar'
  import { settings } from '@/features/bot/settings'

  const s = settings()

  let busy = $state(false)
  let error = $state('')
  let confirming = $state<string | null>(null) // user ID asking "remove?"

  // Search
  let query = $state('')
  let results = $state<Member[]>([])
  let searching = $state(false)
  let searchError = $state('')
  let searched = $state('') // the query the results are for
  let timer: ReturnType<typeof setTimeout> | undefined
  let latest = 0 // only the newest search's answer counts

  // By ID
  let newName = $state('')
  let newId = $state('')

  let verifiedIds = $derived(new Set(s.config?.verified.map((u) => u.user_id))) // O(1) "already verified?"
  let idProblem = $derived.by(() => {
    const id = newId.trim()
    if (!id) return ''
    if (!isSnowflake(id)) return 'A Discord user ID is 17–20 digits.'
    if (verifiedIds.has(id)) return 'They’re already verified.'
    return ''
  })

  /** Searches 250 ms after the last keystroke, so typing doesn't spam Discord. */
  function onInput() {
    clearTimeout(timer)
    const q = query.trim()
    if (!q) {
      results = []
      searched = searchError = ''
      searching = false
      return
    }
    searching = true
    timer = setTimeout(() => search(q), 250)
  }

  async function search(q: string) {
    const mine = ++latest
    try {
      const found = await searchMembers(q)
      if (mine !== latest) return
      results = found
      searchError = ''
    } catch (err) {
      if (mine !== latest) return
      results = []
      searchError = (err as Error).message
    }
    searched = q
    searching = false
  }

  async function addVerified(user: VerifiedUser) {
    busy = true
    error = await s.save((c) => c.verified.push(user), `${user.name} can now restart servers and edit whitelists`)
    busy = false
    return !error
  }

  const verifyMember = (m: Member) => addVerified({ user_id: m.user_id, name: m.name })
  const verifyRequest = (r: AccessRequest) => addVerified({ user_id: r.user_id, name: r.name })

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

  async function addById(event: SubmitEvent) {
    event.preventDefault()
    const name = newName.trim()
    const id = newId.trim()
    if (!name || !id || idProblem || busy) return
    if (await addVerified({ user_id: id, name })) newName = newId = ''
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
    error = await s.save((c) => (c.verified = c.verified.filter((x) => x.user_id !== u.user_id)), `${u.name} removed`)
    busy = false
    confirming = null
  }
</script>

<div class="page">
  <PageHeader title="People" description="Verified people can restart game servers and edit whitelists from Discord." />

  {#if s.status === 'error'}
    <LoadError what="the bot's settings" onretry={s.load} />
  {:else if s.config && s.bot}
    {#if error}<p class="error" role="alert">{error}</p>{/if}

    {#if s.bot.requests.length > 0}
      <Section title="Asking for access" count={s.bot.requests.length}>
        <Card padding="none">
          <ul class="card-list">
            {#each s.bot.requests as r (r.user_id)}
              <li class="row">
                <MemberAvatar userId={r.user_id} name={r.name} />
                <span class="who">
                  <span class="name">{r.name}</span>
                  <span class="meta">Tried /{r.command} · <time datetime={r.at}>{ago(r.at)}</time></span>
                </span>
                <span class="actions">
                  <button type="button" class="btn btn-quiet btn-sm" disabled={busy} onclick={() => dismiss(r)}>Dismiss</button>
                  <button type="button" class="btn btn-primary btn-sm" disabled={busy} onclick={() => verifyRequest(r)}>Verify</button>
                </span>
              </li>
            {/each}
          </ul>
        </Card>
      </Section>
    {/if}

    <Section title="Add someone">
      <Card padding="none">
        <div class="search">
          <Field label="Search your Discord servers" hideLabel>
            {#snippet children(attrs)}
              <span class="search-box">
                <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M11 18a7 7 0 1 0 0-14 7 7 0 0 0 0 14z M20 20l-4-4" /></svg>
                <input
                  {...attrs}
                  type="search"
                  bind:value={query}
                  oninput={onInput}
                  placeholder={s.bot!.connected ? 'Search by name' : 'Connect the bot to search'}
                  disabled={!s.bot!.connected}
                  autocomplete="off"
                  spellcheck="false"
                  maxlength="32"
                />
              </span>
            {/snippet}
          </Field>
        </div>

        <div aria-live="polite">
          {#if searchError}
            <p class="note error">{searchError}</p>
          {:else if results.length > 0}
            <ul class="card-list results">
              {#each results as m (m.user_id)}
                <li class="row">
                  <MemberAvatar userId={m.user_id} name={m.name} />
                  <span class="who">
                    <span class="name">{m.name}</span>
                    <span class="meta">@{m.username}</span>
                  </span>
                  {#if verifiedIds.has(m.user_id)}
                    <Badge tone="success">Verified</Badge>
                  {:else}
                    <button type="button" class="btn btn-secondary btn-sm" disabled={busy} onclick={() => verifyMember(m)}>
                      Verify<span class="visually-hidden"> {m.name}</span>
                    </button>
                  {/if}
                </li>
              {/each}
            </ul>
          {:else if searched && !searching}
            <p class="note">Nobody whose name starts with “{searched}”.</p>
          {/if}
        </div>
      </Card>

      <details class="by-id">
        <summary>
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M9 6l6 6-6 6" /></svg>
          Add by Discord user ID instead
        </summary>
        <p class="hint">
          For someone the search can't find. In Discord: Settings → Advanced → Developer Mode, then right-click the person
          → Copy User ID.
        </p>
        <form class="add" onsubmit={addById}>
          <Field label="Name">
            {#snippet children(attrs)}
              <input {...attrs} bind:value={newName} autocomplete="off" maxlength="100" required />
            {/snippet}
          </Field>
          <Field label="User ID" error={idProblem || undefined}>
            {#snippet children(attrs)}
              <input {...attrs} bind:value={newId} inputmode="numeric" autocomplete="off" required />
            {/snippet}
          </Field>
          <button type="submit" class="btn btn-secondary" disabled={busy || !newName.trim() || !newId.trim() || !!idProblem}>
            Verify
          </button>
        </form>
      </details>
    </Section>

    <Section title="Verified" count={s.config.verified.length}>
      {#if s.config.verified.length === 0}
        <EmptyState
          size="sm"
          icon="M9 11a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7z M2.5 20a6.5 6.5 0 0 1 13 0 M16 11l2 2 4-4"
          title="Nobody yet"
          description="Search for someone above, or verify them when they try /restart or /whitelist."
        />
      {:else}
        <Card padding="none">
          <ul class="card-list">
            {#each s.config.verified as u (u.user_id)}
              <li class="row">
                {#if confirming === u.user_id}
                  <p class="question">Remove {u.name}? They won't be able to restart servers anymore.</p>
                  <span class="actions">
                    <button type="button" id="keep-{u.user_id}" class="btn btn-secondary btn-sm" onclick={() => keep(u)}>
                      Keep
                    </button>
                    <button type="button" class="btn btn-danger btn-sm" disabled={busy} onclick={() => remove(u)}>
                      {busy ? 'Removing…' : 'Remove'}
                    </button>
                  </span>
                {:else}
                  <MemberAvatar userId={u.user_id} name={u.name} />
                  <span class="who">
                    <span class="name">{u.name}</span>
                  </span>
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
        </Card>
      {/if}
    </Section>
  {:else}
    <SettingsSkeleton sections={1} />
  {/if}
</div>

<style>
  .page {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    max-width: var(--content-medium);
  }

  .page > :global(.page-header) {
    margin-bottom: 0;
  }

  .page > :global(.section) {
    margin-top: var(--space-3);
  }

  p {
    margin: 0;
  }

  .error {
    color: var(--color-danger-text);
  }

  .row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    min-height: 3.5rem;
    padding: var(--space-2) var(--space-3) var(--space-2) var(--space-4);
  }

  .who {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
  }

  .name {
    overflow: hidden;
    font-weight: var(--weight-medium);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .meta {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: var(--space-2);
  }

  .question {
    flex: 1;
    font-size: var(--text-sm);
  }

  .search {
    padding: var(--space-3);
  }

  .search-box {
    position: relative;
    display: flex;
  }

  .search-box svg {
    position: absolute;
    top: 50%;
    left: var(--space-3);
    width: 1.125rem;
    height: 1.125rem;
    translate: 0 -50%;
    fill: none;
    stroke: var(--color-text-muted);
    stroke-width: 2;
    stroke-linecap: round;
    pointer-events: none;
  }

  .search-box input {
    flex: 1;
    padding-left: var(--space-7);
  }

  .results {
    border-top: 1px solid var(--color-border);
  }

  .note {
    padding: var(--space-3) var(--space-4) var(--space-4);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .by-id {
    margin-top: var(--space-2);
  }

  .by-id summary {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    list-style: none;
    min-height: var(--touch-target);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
    cursor: pointer;
  }

  .by-id summary::-webkit-details-marker {
    display: none;
  }

  .by-id summary svg {
    width: 1rem;
    height: 1rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
    transition: rotate var(--duration-fast) var(--ease-out);
  }

  .by-id[open] summary svg {
    rotate: 90deg;
  }

  .hint {
    margin-bottom: var(--space-3);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .add {
    display: grid;
    grid-template-columns: 1fr 1fr auto;
    align-items: start;
    gap: var(--space-3);
  }

  .add > button {
    margin-top: 1.75rem; /* lines up with the inputs, under their labels */
  }

  @media (max-width: 36rem) {
    .add {
      grid-template-columns: 1fr;
    }

    .add > button {
      margin-top: 0;
      justify-self: start;
    }
  }
</style>
