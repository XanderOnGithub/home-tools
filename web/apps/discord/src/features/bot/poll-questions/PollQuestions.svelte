<!--
  The poll's questions: a list, plus one form that adds a question or
  edits the one picked. Answers are typed one per line (quicker than a
  field per answer, and it works on phones). Limits are Discord's, checked
  here as you type and again by the server. Removing asks first.
-->
<script lang="ts">
  import { tick } from 'svelte'
  import { POLL_LIMITS, type Config, type PollQuestion } from '@/features/bot'

  let { config, save }: { config: Config; save: (change: (c: Config) => void) => Promise<string> } = $props()

  let editing = $state<number | null>(null) // index being edited; null = adding
  let question = $state('')
  let answersText = $state('')
  let multi = $state(false)
  let busy = $state(false)
  let error = $state('')
  let confirming = $state<number | null>(null)

  let questions = $derived(config.features.poll.questions)
  let answers = $derived(
    answersText
      .split('\n')
      .map((a) => a.trim())
      .filter(Boolean),
  )
  let problem = $derived.by(() => {
    const q = question.trim()
    if (q.length > POLL_LIMITS.question) return `The question is over ${POLL_LIMITS.question} characters.`
    const long = answers.find((a) => a.length > POLL_LIMITS.answer)
    if (long) return `“${long.slice(0, 20)}…” is over ${POLL_LIMITS.answer} characters.`
    if (answers.length > POLL_LIMITS.answers) return `At most ${POLL_LIMITS.answers} answers.`
    if (new Set(answers.map((a) => a.toLowerCase())).size !== answers.length) return 'An answer is there twice.'
    const dup = questions.findIndex((x, i) => i !== editing && x.question.toLowerCase() === q.toLowerCase())
    if (q && dup >= 0) return 'That question is already on the list.'
    return ''
  })
  let ready = $derived(question.trim() !== '' && answers.length >= 2 && !problem)

  async function submit(event: SubmitEvent) {
    event.preventDefault()
    if (!ready || busy) return
    const q: PollQuestion = { question: question.trim(), answers, multi }
    busy = true
    const at = editing
    error = await save((c) => {
      const list = c.features.poll.questions
      if (at === null) list.push(q)
      else list[at] = q
    })
    busy = false
    if (!error) reset()
  }

  async function edit(i: number) {
    const q = questions[i]
    editing = i
    question = q.question
    answersText = q.answers.join('\n')
    multi = q.multi ?? false
    await tick()
    document.getElementById('poll-question')?.focus()
  }

  function reset() {
    editing = null
    question = answersText = ''
    multi = false
  }

  async function askRemove(i: number) {
    confirming = i
    await tick()
    document.getElementById(`keep-question-${i}`)?.focus()
  }

  async function keep(i: number) {
    confirming = null
    await tick()
    document.getElementById(`remove-question-${i}`)?.focus()
  }

  async function remove(i: number) {
    busy = true
    error = await save((c) => c.features.poll.questions.splice(i, 1))
    busy = false
    confirming = null
    if (editing === i) reset()
  }
</script>

<div class="questions">
  <div class="title">
    <h3>Questions</h3>
    <span class="muted count">{questions.length}</span>
  </div>
  <p class="muted">Shuffled, so each one gets asked before any repeats.</p>

  {#if questions.length === 0}
    <p class="muted">None yet. Add the first one below.</p>
  {:else}
    <ol class="list">
      {#each questions as q, i (q.question)}
        <li class="row">
          {#if confirming === i}
            <p class="question-text">Remove “{q.question}”?</p>
            <div class="actions">
              <button type="button" id="keep-question-{i}" class="btn btn-secondary" onclick={() => keep(i)}>Keep it</button>
              <button type="button" class="btn btn-danger" disabled={busy} onclick={() => remove(i)}>Remove</button>
            </div>
          {:else}
            <div class="label">
              <span class="name">{q.question}</span>
              <span class="muted">{q.answers.join(' · ')}{q.multi ? ' · pick several' : ''}</span>
            </div>
            <div class="actions">
              <button type="button" class="btn btn-quiet" onclick={() => edit(i)} aria-label="Edit “{q.question}”">Edit</button>
              <button type="button" id="remove-question-{i}" class="btn btn-icon" aria-label="Remove “{q.question}”" onclick={() => askRemove(i)}>
                <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12 M18 6L6 18" /></svg>
              </button>
            </div>
          {/if}
        </li>
      {/each}
    </ol>
  {/if}

  <form class="form" onsubmit={submit}>
    <h4>{editing === null ? 'Add a question' : 'Edit question'}</h4>
    <div class="field">
      <label for="poll-question">Question</label>
      <input id="poll-question" bind:value={question} autocomplete="off" placeholder="Best pizza topping?" />
    </div>
    <div class="field">
      <label for="poll-answers">Answers <span class="muted hint">one per line, 2–{POLL_LIMITS.answers}</span></label>
      <textarea
        id="poll-answers"
        bind:value={answersText}
        rows="4"
        placeholder={'Pepperoni\nPineapple\nJust cheese'}
        aria-describedby={problem ? 'poll-problem' : undefined}
      ></textarea>
    </div>
    <label class="check">
      <input type="checkbox" bind:checked={multi} />
      People may pick more than one answer
    </label>
    {#if problem}<p class="error" id="poll-problem">{problem}</p>{/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <div class="actions">
      {#if editing !== null}
        <button type="button" class="btn btn-quiet" onclick={reset}>Cancel</button>
      {/if}
      <button type="submit" class="btn btn-primary" disabled={busy || !ready}>
        {busy ? 'Saving…' : editing === null ? 'Add question' : 'Save changes'}
      </button>
    </div>
  </form>
</div>

<style>
  .questions {
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

  h4 {
    margin: 0;
    font-size: var(--text-md);
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
    flex: 1 1 12rem;
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

  .question-text {
    flex: 1 1 12rem;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
  }

  .form {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  label {
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
  }

  .hint {
    font-weight: var(--weight-normal);
  }

  input:not([type='checkbox']),
  textarea {
    min-height: var(--touch-target);
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-bg);
  }

  textarea {
    resize: vertical;
    line-height: var(--leading-normal);
  }

  .check {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-height: var(--touch-target);
    font-size: var(--text-md);
    font-weight: var(--weight-medium);
    cursor: pointer;
  }

  .check input {
    width: 1.25rem;
    height: 1.25rem;
    accent-color: var(--color-accent);
  }
</style>
