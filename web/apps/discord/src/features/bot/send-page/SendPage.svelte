<!--
  Send (decision #59): post a message or a poll as the bot, in a channel
  picked here, e.g. a game night vote. Nothing is stored; the post lives
  in Discord. The form on the left, a preview of how it'll look on the
  right (under it on phones). The last channel used is remembered on this
  device.
-->
<script lang="ts">
  import { Card } from '@home-tools/ui/components/card'
  import { Field } from '@home-tools/ui/components/field'
  import { LoadError } from '@home-tools/ui/components/load-error'
  import { PageHeader } from '@home-tools/ui/components/page-header'
  import { toast } from '@home-tools/ui/toast'
  import { SettingsSkeleton } from '@/components/settings-skeleton'
  import { MESSAGE_MAX, POLL_LIMITS, sendPost, type Post } from '@/features/bot'
  import { SendPreview } from '@/features/bot/send-preview'
  import { settings } from '@/features/bot/settings'

  const s = settings()

  const LAST_CHANNEL = 'discord-send-channel'
  const OPEN_FOR = [
    { hours: 1, label: '1 hour' },
    { hours: 4, label: '4 hours' },
    { hours: 8, label: '8 hours' },
    { hours: 24, label: '1 day' },
    { hours: 72, label: '3 days' },
    { hours: 168, label: '1 week' },
  ]

  let kind = $state<'message' | 'poll'>('message')
  let channel = $state(remembered())
  let message = $state('')
  let everyone = $state(false)
  let question = $state('')
  let answers = $state(['', ''])
  let multi = $state(false)
  let hours = $state(24)
  let sending = $state(false)
  let error = $state('')

  /** Only channels the bot may post in; the remembered one if it still is. */
  let channels = $derived(
    (s.bot?.guilds ?? []).map((g) => ({ ...g, channels: g.channels.filter((c) => c.can_post) })).filter((g) => g.channels.length),
  )
  let channelName = $derived(channels.flatMap((g) => g.channels).find((c) => c.id === channel)?.name)

  let text = $derived((everyone ? `@everyone ${message.trim()}` : message.trim()).trim())
  let filled = $derived(answers.map((a) => a.trim()).filter(Boolean))
  let duplicate = $derived(new Set(filled.map((a) => a.toLowerCase())).size !== filled.length)

  /** Why it can't be sent yet, or '' when it can. Shown under the button. */
  let problem = $derived.by(() => {
    if (!s.bot?.connected) return "The bot isn't connected to Discord."
    if (!channelName) return 'Pick a channel.'
    if (text.length > MESSAGE_MAX) return `The message is ${text.length - MESSAGE_MAX} characters too long.`
    if (kind === 'message') return message.trim() ? '' : 'Write a message.'
    if (!question.trim()) return 'Write the question.'
    if (filled.length < 2) return 'Give it at least two answers.'
    if (duplicate) return 'Two answers are the same.'
    return ''
  })

  function remembered(): string {
    try {
      return localStorage.getItem(LAST_CHANNEL) ?? ''
    } catch {
      return ''
    }
  }

  function remember(id: string) {
    try {
      localStorage.setItem(LAST_CHANNEL, id)
    } catch {
      // private mode: it just isn't remembered
    }
  }

  function addAnswer(focus = true) {
    if (answers.length >= POLL_LIMITS.answers) return
    answers.push('')
    if (focus) queueMicrotask(() => document.getElementById(`answer-${answers.length - 1}`)?.focus())
  }

  function removeAnswer(i: number) {
    answers.splice(i, 1)
    document.getElementById(`answer-${Math.min(i, answers.length - 1)}`)?.focus()
  }

  /** Enter in an answer moves to the next one (adding it if it's the last). */
  function onAnswerKey(e: KeyboardEvent, i: number) {
    if (e.key !== 'Enter') return
    e.preventDefault()
    if (i < answers.length - 1) document.getElementById(`answer-${i + 1}`)?.focus()
    else if (answers[i].trim()) addAnswer()
  }

  async function send(e: SubmitEvent) {
    e.preventDefault()
    if (problem || sending) return
    const post: Post = { channel_id: channel, everyone }
    if (message.trim()) post.message = message.trim()
    if (kind === 'poll') post.poll = { question: question.trim(), answers: filled, multi, duration_hours: hours }
    sending = true
    error = ''
    try {
      const { url } = await sendPost(post)
      remember(channel)
      toast(`Posted in #${channelName}`, {
        action: { label: 'Open', run: () => window.open(url, '_blank', 'noopener') },
      })
      message = question = ''
      answers = ['', '']
      everyone = multi = false
    } catch (err) {
      error = (err as Error).message
    }
    sending = false
  }
</script>

<div class="page">
  <PageHeader title="Send" description="Post a message or a poll as the bot." />

  {#if s.status === 'error'}
    <LoadError what="the bot's channels" onretry={s.load} />
  {:else if s.bot}
    <div class="layout">
      <form onsubmit={send} novalidate>
        <Card padding="lg">
          <div class="fields">
            <Field label="Channel" hint={channels.length ? undefined : "The bot can't post anywhere yet. Add it to a server on the Bot page."}>
              {#snippet children(attrs)}
                <select {...attrs} bind:value={channel} disabled={!channels.length}>
                  <option value="" disabled>Pick a channel</option>
                  {#each channels as g (g.id)}
                    <optgroup label={g.name}>
                      {#each g.channels as c (c.id)}<option value={c.id}>#{c.name}</option>{/each}
                    </optgroup>
                  {/each}
                </select>
              {/snippet}
            </Field>

            <fieldset class="segmented">
              <legend class="visually-hidden">What to send</legend>
              <label><input type="radio" name="kind" value="message" bind:group={kind} />Message</label>
              <label><input type="radio" name="kind" value="poll" bind:group={kind} />Poll</label>
            </fieldset>

            {#if kind === 'poll'}
              <Field label="Question">
                {#snippet children(attrs)}
                  <input {...attrs} bind:value={question} maxlength={POLL_LIMITS.question} placeholder="Game night this week?" autocomplete="off" />
                {/snippet}
              </Field>

              <fieldset class="answers">
                <legend>Answers</legend>
                <ol>
                  {#each answers as _, i (i)}
                    <li>
                      <label class="visually-hidden" for="answer-{i}">Answer {i + 1}</label>
                      <input
                        id="answer-{i}"
                        bind:value={answers[i]}
                        maxlength={POLL_LIMITS.answer}
                        placeholder={i === 0 ? 'Friday' : i === 1 ? 'Saturday' : `Answer ${i + 1}`}
                        autocomplete="off"
                        enterkeyhint="next"
                        onkeydown={(e) => onAnswerKey(e, i)}
                      />
                      <button
                        type="button"
                        class="btn btn-icon"
                        aria-label="Remove answer {i + 1}"
                        disabled={answers.length <= 2}
                        onclick={() => removeAnswer(i)}
                      >
                        <svg viewBox="0 0 24 24"><path d="M6 6l12 12 M18 6L6 18" /></svg>
                      </button>
                    </li>
                  {/each}
                </ol>
                {#if answers.length < POLL_LIMITS.answers}
                  <button type="button" class="btn btn-quiet btn-sm add" onclick={() => addAnswer()}>
                    <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5v14 M5 12h14" /></svg>
                    Add answer
                  </button>
                {/if}
              </fieldset>

              <div class="row">
                <Field label="Open for">
                  {#snippet children(attrs)}
                    <select {...attrs} bind:value={hours}>
                      {#each OPEN_FOR as o (o.hours)}<option value={o.hours}>{o.label}</option>{/each}
                    </select>
                  {/snippet}
                </Field>
                <label class="check">
                  <input type="checkbox" bind:checked={multi} />
                  More than one answer
                </label>
              </div>
            {/if}

            <Field
              label={kind === 'poll' ? 'Message above the poll' : 'Message'}
              optional={kind === 'poll'}
              hint="{text.length} / {MESSAGE_MAX}"
            >
              {#snippet children(attrs)}
                <textarea
                  {...attrs}
                  bind:value={message}
                  rows={kind === 'poll' ? 2 : 5}
                  placeholder={kind === 'poll' ? 'Vote by Thursday!' : "Server's back up."}
                ></textarea>
              {/snippet}
            </Field>

            <label class="check">
              <input type="checkbox" bind:checked={everyone} />
              Ping @everyone
            </label>
          </div>

          <div class="footer">
            {#if error}<p class="error" role="alert">{error}</p>{:else if problem}<p class="problem">{problem}</p>{/if}
            <button type="submit" class="btn btn-primary" disabled={!!problem || sending} aria-busy={sending}>
              {#if sending}Posting…{:else}Post{channelName ? ` to #${channelName}` : ''}{/if}
            </button>
          </div>
        </Card>
      </form>

      <section class="preview" aria-label="Preview">
        <h2 class="preview-title">Preview</h2>
        <SendPreview
          bot={s.bot}
          {text}
          poll={kind === 'poll' ? { question: question.trim(), answers, multi, hours } : null}
        />
      </section>
    </div>
  {:else}
    <SettingsSkeleton sections={1} />
  {/if}
</div>

<style>
  .page > :global(.page-header) {
    margin-bottom: var(--space-5);
  }

  /* Phones: form, then the preview. Wide screens: side by side, the
   * preview staying in view while the form scrolls. */
  .layout {
    display: grid;
    gap: var(--space-6);
    align-items: start;
  }

  @media (min-width: 60rem) {
    .layout {
      grid-template-columns: minmax(0, 1fr) minmax(0, 26rem);
    }

    .preview {
      position: sticky;
      top: calc(var(--header-height) + var(--space-5));
    }
  }

  .fields {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  textarea {
    resize: vertical;
    line-height: var(--leading-normal);
  }

  /* Message | Poll: two radios drawn as one switch. */
  .segmented {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.125rem;
    margin: 0;
    padding: 0.1875rem;
    border: none;
    border-radius: var(--radius-md);
    background: var(--color-surface-sunken);
  }

  .segmented label {
    display: grid;
    place-items: center;
    min-height: 2.5rem;
    border-radius: calc(var(--radius-md) - 0.125rem);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
    cursor: pointer;
    transition:
      background var(--duration-fast) var(--ease-out),
      color var(--duration-fast) var(--ease-out);
  }

  .segmented input {
    position: absolute;
    opacity: 0;
    pointer-events: none;
  }

  .segmented label:has(:checked) {
    background: var(--color-surface);
    color: var(--color-text);
    box-shadow: var(--shadow-sm);
  }

  .segmented label:has(:focus-visible) {
    outline: var(--focus-ring);
    outline-offset: -2px;
  }

  .answers {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    border: none;
  }

  .answers legend {
    margin-bottom: var(--space-2);
    padding: 0;
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
  }

  ol {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: flex;
    gap: var(--space-1);
  }

  li input {
    flex: 1;
  }

  .add {
    align-self: flex-start;
    margin-left: calc(-1 * var(--space-3));
  }

  .add svg {
    width: 1rem;
    height: 1rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
  }

  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    gap: var(--space-3) var(--space-5);
  }

  .row > :global(.field) {
    min-width: 10rem;
  }

  .check {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    min-height: var(--touch-target);
    font-size: var(--text-sm);
    cursor: pointer;
  }

  .check input {
    width: 1.125rem;
    height: 1.125rem;
    margin: 0;
    accent-color: var(--color-accent);
  }

  .footer {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: flex-end;
    gap: var(--space-3);
    margin-top: var(--space-6);
    padding-top: var(--space-4);
    border-top: 1px solid var(--color-border);
  }

  .footer p {
    flex: 1 1 12rem;
    margin: 0;
    font-size: var(--text-sm);
  }

  .problem {
    color: var(--color-text-muted);
  }

  .error {
    color: var(--color-danger-text);
  }

  .preview {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .preview-title {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
  }
</style>
