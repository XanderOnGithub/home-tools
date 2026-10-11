<!--
  The poll's packs (decision #47): named sets of questions, each with its
  own switch. Polls are drawn from every pack that's on, each question
  once per round. A pack shows how much of this round it has used; "Start
  over" makes all of its questions fair game again. One pack is open at a
  time, for its questions and its name. Removing asks first.
-->
<script lang="ts">
  import { tick } from 'svelte'
  import { POLL_LIMITS, pollKey, resetPollPack, type BotView, type Config } from '@/features/bot'
  import { PollQuestions } from '@/features/bot/poll-questions'
  import { ToggleSwitch } from '@/components/toggle-switch'

  let {
    config,
    bot,
    save,
    onreset,
  }: {
    config: Config
    bot: BotView
    save: (change: (c: Config) => void) => Promise<string>
    onreset: () => void
  } = $props()

  const NAME_MAX = 100 // maxLabelLen in tools/discord/model.go

  let packs = $derived(config.features.poll.packs)
  let open = $state<number | null>(null) // the pack whose questions are shown
  let busy = $state(false)
  let error = $state('')
  let confirming = $state<number | null>(null)

  let newName = $state('')
  let rename = $state('')

  /** How many of the pack's questions were asked this round. */
  const askedIn = (i: number) => packs[i].questions.filter((q) => bot.poll_asked[pollKey(q)]).length

  function nameProblem(name: string, except: number | null) {
    const n = name.trim()
    if (n.length > NAME_MAX) return `At most ${NAME_MAX} characters.`
    if (packs.some((p, i) => i !== except && p.name.toLowerCase() === n.toLowerCase())) return 'There’s a pack with that name already.'
    return ''
  }
  let newProblem = $derived(nameProblem(newName, null))
  let renameProblem = $derived(open === null ? '' : nameProblem(rename, open))

  async function change(f: (c: Config) => void) {
    busy = true
    error = await save(f)
    busy = false
    return !error
  }

  async function add(event: SubmitEvent) {
    event.preventDefault()
    const name = newName.trim()
    if (!name || newProblem || busy) return
    if (await change((c) => c.features.poll.packs.push({ name, enabled: true, questions: [] }))) {
      newName = ''
      toggle(packs.length - 1) // open it, ready for its first question
    }
  }

  function toggle(i: number) {
    open = open === i ? null : i
    rename = open === null ? '' : packs[open].name
  }

  async function saveName(event: SubmitEvent) {
    event.preventDefault()
    const i = open
    const name = rename.trim()
    if (i === null || !name || renameProblem || name === packs[i].name) return
    await change((c) => (c.features.poll.packs[i].name = name))
  }

  async function startOver(i: number) {
    busy = true
    error = ''
    try {
      await resetPollPack(packs[i].name)
      onreset()
    } catch (err) {
      error = (err as Error).message
    }
    busy = false
  }

  async function askRemove(i: number) {
    confirming = i
    await tick()
    document.getElementById(`keep-pack-${i}`)?.focus()
  }

  async function keep(i: number) {
    confirming = null
    await tick()
    document.getElementById(`remove-pack-${i}`)?.focus()
  }

  async function remove(i: number) {
    if (await change((c) => c.features.poll.packs.splice(i, 1))) {
      if (open === i) open = null
      else if (open !== null && open > i) open-- // indexes shift up
    }
    confirming = null
  }
</script>

<div class="packs">
  <div class="title">
    <h3>Packs</h3>
    <span class="muted count">{packs.length}</span>
  </div>
  <p class="muted">Polls come from the packs that are on: every question once, in a random order, then a new round.</p>

  {#if packs.length === 0}
    <p class="muted">No packs yet. Add one below.</p>
  {:else}
    <ul class="list">
      {#each packs as pack, i (pack.name)}
        {@const used = askedIn(i)}
        <li class="pack">
          {#if confirming === i}
            <div class="row">
              <p class="grow">Remove “{pack.name}” and its {pack.questions.length} questions?</p>
              <div class="actions">
                <button type="button" id="keep-pack-{i}" class="btn btn-secondary" onclick={() => keep(i)}>Keep it</button>
                <button type="button" class="btn btn-danger" disabled={busy} onclick={() => remove(i)}>Remove</button>
              </div>
            </div>
          {:else}
            <div class="row">
              <ToggleSwitch
                checked={pack.enabled}
                label="{pack.name} pack"
                {busy}
                onchange={(on) => change((c) => (c.features.poll.packs[i].enabled = on))}
              />
              <div class="label grow">
                <span class="name">{pack.name}</span>
                <span class="muted">
                  {pack.questions.length} {pack.questions.length === 1 ? 'question' : 'questions'}{used > 0 ? ` · ${used} asked this round` : ''}
                </span>
              </div>
              <div class="actions">
                <button type="button" class="btn btn-quiet" aria-expanded={open === i} aria-controls="pack-{i}" onclick={() => toggle(i)}>
                  {open === i ? 'Close' : 'Questions'}<span class="visually-hidden"> in {pack.name}</span>
                </button>
                <button type="button" id="remove-pack-{i}" class="btn btn-icon" aria-label="Remove {pack.name}" onclick={() => askRemove(i)}>
                  <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12 M18 6L6 18" /></svg>
                </button>
              </div>
            </div>
          {/if}

          {#if open === i}
            <div class="body" id="pack-{i}">
              <div class="pack-tools">
                <form class="rename" onsubmit={saveName}>
                  <label for="pack-name-{i}">Pack name</label>
                  <div class="inline">
                    <input id="pack-name-{i}" bind:value={rename} autocomplete="off" aria-describedby={renameProblem ? `pack-name-problem-${i}` : undefined} />
                    <button type="submit" class="btn btn-secondary" disabled={busy || !rename.trim() || !!renameProblem || rename.trim() === pack.name}>
                      Rename
                    </button>
                  </div>
                  {#if renameProblem}<p class="error" id="pack-name-problem-{i}">{renameProblem}</p>{/if}
                </form>
                {#if used > 0}
                  <button type="button" class="btn btn-quiet" disabled={busy} onclick={() => startOver(i)}>
                    Start over<span class="visually-hidden"> {pack.name}</span>
                  </button>
                {/if}
              </div>
              <PollQuestions {config} pack={i} asked={bot.poll_asked} {save} />
            </div>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}

  {#if error}<p class="error" role="alert">{error}</p>{/if}

  <form class="add" onsubmit={add}>
    <label for="new-pack">New pack</label>
    <div class="inline">
      <input
        id="new-pack"
        bind:value={newName}
        autocomplete="off"
        placeholder="Minecraft"
        maxlength={NAME_MAX + 1}
        aria-describedby={newProblem ? 'new-pack-problem' : undefined}
      />
      <button type="submit" class="btn btn-secondary" disabled={busy || !newName.trim() || !!newProblem}>Add pack</button>
    </div>
    {#if newProblem}<p class="error" id="new-pack-problem">{newProblem}</p>{/if}
  </form>
  <p class="muted hint">Questions: up to {POLL_LIMITS.question} characters, answers up to {POLL_LIMITS.answer}.</p>
</div>

<style>
  .packs {
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

  .count {
    font-weight: var(--weight-semibold);
  }

  .hint {
    font-size: var(--text-sm);
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
    background: var(--color-bg);
    list-style: none;
  }

  .pack + .pack {
    border-top: 1px solid var(--color-border);
  }

  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-3);
    min-height: calc(var(--touch-target) + var(--space-4));
    padding: var(--space-2) var(--space-2) var(--space-2) var(--space-4);
  }

  .grow {
    flex: 1 1 10rem;
    min-width: 0;
  }

  .label {
    display: flex;
    flex-direction: column;
  }

  .name {
    font-weight: var(--weight-semibold);
    overflow-wrap: anywhere;
  }

  .label .muted {
    font-size: var(--text-sm);
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
    margin-left: auto;
  }

  .body {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    padding: var(--space-2) var(--space-4) var(--space-4);
  }

  .pack-tools {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    justify-content: space-between;
    gap: var(--space-3);
  }

  .rename,
  .add {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .rename {
    flex: 1 1 16rem;
  }

  .inline {
    display: flex;
    gap: var(--space-2);
  }

  label {
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
  }

  input {
    flex: 1;
    min-width: 0;
    min-height: var(--touch-target);
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-bg);
  }
</style>
