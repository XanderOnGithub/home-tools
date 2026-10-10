<!--
  The bot's custom status (decision #45): a new phrase every hour, and
  now and then who's playing. Phrases are long, so they're a list (not
  chips like the names); "{name}" becomes the day's name. Removing one
  needs no confirm: it's a line of text, quick to add back.
-->
<script lang="ts">
  import { STATUS_MAX, type BotView, type Config, type StatusFeature } from '@/features/bot'
  import { ToggleSwitch } from '@/components/toggle-switch'

  let {
    config,
    bot,
    save,
  }: { config: Config; bot: BotView; save: (change: (c: Config) => void) => Promise<string> } = $props()

  let busy = $state(false)
  let error = $state('')
  let phrase = $state('')
  let input = $state<HTMLInputElement>()

  let status = $derived(config.features.status)
  let lower = $derived(new Set(status.phrases.map((p) => p.toLowerCase()))) // O(1) duplicate check
  let problem = $derived.by(() => {
    const p = phrase.trim()
    if (p.length > STATUS_MAX) return `At most ${STATUS_MAX} characters.`
    if (lower.has(p.toLowerCase())) return 'That one is already on the list.'
    return ''
  })

  async function change(f: (s: StatusFeature) => void) {
    busy = true
    error = await save((c) => f(c.features.status))
    busy = false
  }

  async function add(event: SubmitEvent) {
    event.preventDefault()
    const p = phrase.trim()
    if (!p || problem || busy) return
    await change((s) => s.phrases.push(p))
    if (!error) {
      phrase = ''
      input?.focus()
    }
  }

  async function remove(p: string, i: number) {
    await change((s) => (s.phrases = s.phrases.filter((x) => x !== p)))
    const buttons = document.querySelectorAll<HTMLButtonElement>('.phrases .remove')
    ;(buttons[Math.min(i, buttons.length - 1)] ?? input)?.focus()
  }
</script>

<section class="feature" aria-labelledby="status-title">
  <div class="head">
    <div>
      <h2 id="status-title">Status</h2>
      <p class="muted">The line under the bot's name in Discord. A new one every hour, on the hour.</p>
    </div>
    <ToggleSwitch checked={status.enabled} label="Status" {busy} onchange={(on) => change((s) => (s.enabled = on))} />
  </div>

  {#if status.enabled && bot.status}
    <p class="now"><span class="muted">Showing now:</span> {bot.status}</p>
  {/if}

  <label class="check">
    <input type="checkbox" checked={status.live_games} disabled={busy} onchange={(e) => change((s) => (s.live_games = e.currentTarget.checked))} />
    <span>
      Now and then, show who's playing
      <span class="muted hint">e.g. “Watching Steve play Minecraft”; never two hours in a row</span>
    </span>
  </label>

  <div class="phrases-block">
    <div class="title">
      <h3>Phrases</h3>
      <span class="muted count">{status.phrases.length}</span>
    </div>
    <p class="muted">Shuffled, so each one shows before any repeats. <code>{'{name}'}</code> becomes today's name.</p>

    {#if status.phrases.length > 0}
      <ul class="phrases">
        {#each status.phrases as p, i (p)}
          <li class="row">
            <span class="text">{p}</span>
            <button type="button" class="btn btn-icon remove" aria-label="Remove “{p}”" disabled={busy} onclick={() => remove(p, i)}>
              <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12 M18 6L6 18" /></svg>
            </button>
          </li>
        {/each}
      </ul>
    {/if}

    <form class="add" onsubmit={add}>
      <label class="visually-hidden" for="new-phrase">New phrase</label>
      <input
        id="new-phrase"
        bind:this={input}
        bind:value={phrase}
        placeholder="Add a phrase, e.g. Blobbing on main"
        autocomplete="off"
        maxlength={STATUS_MAX + 10}
        aria-invalid={problem ? 'true' : undefined}
        aria-describedby={problem ? 'phrase-problem' : undefined}
      />
      <button type="submit" class="btn btn-secondary" disabled={busy || !phrase.trim() || !!problem}>Add</button>
    </form>
    {#if problem}<p class="error" id="phrase-problem">{problem}</p>{/if}
  </div>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
</section>

<style>
  .feature {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    padding: var(--space-4);
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

  .head > div {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  h3 {
    font-size: var(--text-lg);
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

  code {
    font-family: var(--font-mono);
    font-size: 0.9em;
  }

  .now {
    font-weight: var(--weight-semibold);
    overflow-wrap: anywhere;
  }

  .check {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    min-height: var(--touch-target);
    padding-top: var(--space-2);
    font-weight: var(--weight-medium);
    cursor: pointer;
  }

  .check > span {
    display: flex;
    flex-direction: column;
  }

  .check input {
    flex: none;
    width: 1.25rem;
    height: 1.25rem;
    margin: 0.125rem 0 0;
    accent-color: var(--color-accent);
  }

  .hint {
    font-size: var(--text-sm);
    font-weight: var(--weight-normal);
  }

  .phrases-block {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding-top: var(--space-4);
    border-top: 1px solid var(--color-border);
  }

  .title {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
  }

  .count {
    font-weight: var(--weight-semibold);
  }

  .phrases {
    display: flex;
    flex-direction: column;
    max-height: 22rem;
    margin: 0;
    padding: 0;
    overflow-y: auto;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-bg);
    list-style: none;
  }

  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-1) var(--space-1) var(--space-1) var(--space-4);
  }

  .row + .row {
    border-top: 1px solid var(--color-border);
  }

  .text {
    min-width: 0;
    overflow-wrap: anywhere;
  }

  .add {
    display: flex;
    gap: var(--space-2);
  }

  input:not([type='checkbox']) {
    flex: 1;
    min-width: 0;
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-bg);
  }
</style>
