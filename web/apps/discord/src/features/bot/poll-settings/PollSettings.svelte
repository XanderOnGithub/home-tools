<!--
  The poll feature (decision #44): its switch, where and when it posts,
  and how long votes stay open. Every change saves at once, like the
  rest of the page. "Post one now" takes the next poll's turn, so trying
  it out never repeats a question.
-->
<script lang="ts">
  import { postPollNow, type BotView, type Config, type PollFeature } from '@/features/bot'
  import { PollQuestions } from '@/features/bot/poll-questions'
  import { ToggleSwitch } from '@/components/toggle-switch'

  let {
    config,
    bot,
    save,
    onposted,
  }: {
    config: Config
    bot: BotView
    save: (change: (c: Config) => void) => Promise<string>
    onposted: () => void
  } = $props()

  let busy = $state(false)
  let posting = $state(false)
  let error = $state('')
  let posted = $state('')

  let poll = $derived(config.features.poll)

  const EVERY = [1, 2, 3, 4, 5, 6, 7]
  const OPEN_FOR = [6, 12, 24, 48, 72]

  async function change(f: (p: PollFeature) => void) {
    busy = true
    posted = ''
    error = await save((c) => f(c.features.poll))
    busy = false
  }

  async function postNow() {
    posting = true
    posted = error = ''
    try {
      await postPollNow()
      posted = 'Posted. The next one goes out on schedule.'
      onposted()
    } catch (err) {
      error = (err as Error).message
    }
    posting = false
  }

  /** "Sat, Oct 10 at 6:00 PM", in this browser's locale. */
  function when(iso: string) {
    return new Date(iso).toLocaleString(undefined, { weekday: 'short', month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' })
  }
</script>

<section class="feature" aria-labelledby="poll-title">
  <div class="head">
    <div>
      <h2 id="poll-title">Poll</h2>
      <p class="muted">A question with answers to vote on, every few days, as a Discord poll.</p>
    </div>
    <ToggleSwitch checked={poll.enabled} label="Poll" {busy} onchange={(on) => change((p) => (p.enabled = on))} />
  </div>

  <div class="settings">
    <div class="field">
      <label for="poll-channel">Channel</label>
      <select id="poll-channel" value={poll.channel_id} disabled={busy} onchange={(e) => change((p) => (p.channel_id = e.currentTarget.value))}>
        <option value="">{bot.connected ? 'Pick a channel' : 'Connect the bot to pick one'}</option>
        {#each bot.guilds as g (g.id)}
          <optgroup label={g.name}>
            {#each g.channels as c (c.id)}
              <option value={c.id} disabled={!c.can_post}>#{c.name}{c.can_post ? '' : ' (no permission to post)'}</option>
            {/each}
          </optgroup>
        {/each}
      </select>
    </div>
    <div class="field">
      <label for="poll-every">How often</label>
      <select id="poll-every" value={poll.every_days} disabled={busy} onchange={(e) => change((p) => (p.every_days = Number(e.currentTarget.value)))}>
        {#each EVERY as n (n)}
          <option value={n}>{n === 1 ? 'Every day' : `Every ${n} days`}</option>
        {/each}
      </select>
    </div>
    <div class="field">
      <label for="poll-time">At</label>
      <input
        id="poll-time"
        type="time"
        value={poll.post_at}
        disabled={busy}
        onchange={(e) => e.currentTarget.value && change((p) => (p.post_at = e.currentTarget.value))}
      />
    </div>
    <div class="field">
      <label for="poll-open">Votes open for</label>
      <select id="poll-open" value={poll.duration_hours} disabled={busy} onchange={(e) => change((p) => (p.duration_hours = Number(e.currentTarget.value)))}>
        {#each OPEN_FOR as h (h)}
          <option value={h}>{h < 24 ? `${h} hours` : h === 24 ? '1 day' : `${h / 24} days`}</option>
        {/each}
      </select>
    </div>
  </div>

  {#if poll.enabled}
    <div class="now">
      <p class="muted">
        {#if bot.next_poll}Next poll: <time datetime={bot.next_poll}>{when(bot.next_poll)}</time>{/if}
      </p>
      <button type="button" class="btn btn-secondary" disabled={posting || !bot.connected} onclick={postNow}>
        {posting ? 'Posting…' : 'Post one now'}
      </button>
    </div>
  {/if}
  {#if posted}<p class="muted" role="status">{posted}</p>{/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}

  <PollQuestions {config} {save} />
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

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
  }

  .error {
    color: var(--color-danger-text);
  }

  .settings {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(11rem, 1fr));
    gap: var(--space-3);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0;
  }

  label {
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
  }

  select,
  input {
    min-width: 0;
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-bg);
    font-weight: var(--weight-medium);
  }

  select {
    cursor: pointer;
  }

  .now {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
  }
</style>
