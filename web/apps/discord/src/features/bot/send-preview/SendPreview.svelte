<!--
  Roughly how the post will look in Discord: today's persona as the
  author, the text, and the poll with its answers. A guide, not a pixel
  copy: Discord draws the real thing. Decorative for screen readers (the
  form above says everything), so it's aria-hidden.
-->
<script lang="ts">
  import { PERSONA_PNG, type BotView } from '@/features/bot'

  let {
    bot,
    text,
    poll,
  }: {
    bot: BotView
    text: string // as sent, with "@everyone" in front if pinging
    poll: { question: string; answers: string[]; multi: boolean; hours: number } | null
  } = $props()

  let version = $derived(`${bot.persona.date}-${bot.persona.color}-${encodeURIComponent(bot.persona.name)}`)

  /** "@everyone" is drawn as a mention, the rest as plain text. */
  let parts = $derived(text.split(/(@everyone)/).filter(Boolean))

  function lasts(h: number) {
    if (h % 24 === 0) return `${h / 24} ${h === 24 ? 'day' : 'days'}`
    return `${h} ${h === 1 ? 'hour' : 'hours'}`
  }
</script>

<div class="preview" aria-hidden="true">
  <img class="avatar" src="{PERSONA_PNG}?v={version}" alt="" width="40" height="40" />
  <div class="body">
    <p class="author">
      <span class="name">{bot.persona.name}</span>
      <span class="tag">APP</span>
      <span class="time">Today</span>
    </p>
    {#if text}
      <p class="text">
        {#each parts as part, i (i)}{#if part === '@everyone'}<span class="mention">@everyone</span>{:else}{part}{/if}{/each}
      </p>
    {/if}
    {#if poll}
      <div class="poll">
        <p class="question">{poll.question || 'Your question'}</p>
        <p class="hint">{poll.multi ? 'Select one or more answers' : 'Select one answer'}</p>
        <ul>
          {#each poll.answers as a, i (i)}
            <li class:placeholder={!a.trim()}>
              <span class="answer">{a.trim() || `Answer ${i + 1}`}</span>
              <span class="pick" class:square={poll.multi}></span>
            </li>
          {/each}
        </ul>
        <p class="hint">0 votes · {lasts(poll.hours)} left</p>
      </div>
    {/if}
    {#if !text && !poll}
      <p class="text empty">Your message shows here.</p>
    {/if}
  </div>
</div>

<style>
  /* Discord's own colors, fixed in both themes: this is a picture of
   * Discord, not part of our UI. */
  .preview {
    display: grid;
    grid-template-columns: 2.5rem 1fr;
    gap: var(--space-3);
    padding: var(--space-4);
    border-radius: var(--radius-lg);
    background: #313338;
    color: #dbdee1;
    font-size: 0.9375rem;
  }

  .avatar {
    width: 2.5rem;
    height: 2.5rem;
    border-radius: var(--radius-full);
    background: #2b2d31;
  }

  .body {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    min-width: 0;
  }

  p {
    margin: 0;
  }

  .author {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
  }

  .name {
    color: #f2f3f5;
    font-weight: var(--weight-medium);
  }

  .tag {
    padding: 0 0.25rem;
    border-radius: 0.1875rem;
    background: #5865f2;
    color: #fff;
    font-size: 0.625rem;
    font-weight: var(--weight-semibold);
    line-height: 1.4;
    align-self: center;
  }

  .time {
    color: #949ba4;
    font-size: var(--text-xs);
  }

  .text {
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  .empty {
    color: #949ba4;
  }

  .mention {
    padding: 0 0.125rem;
    border-radius: 0.1875rem;
    background: rgb(88 101 242 / 0.3);
    color: #c9cdfb;
    font-weight: var(--weight-medium);
  }

  .poll {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    max-width: 28rem;
    margin-top: var(--space-1);
    padding: var(--space-4);
    border: 1px solid #3f4147;
    border-radius: var(--radius-md);
    background: #2b2d31;
  }

  .question {
    color: #f2f3f5;
    font-weight: var(--weight-semibold);
    overflow-wrap: anywhere;
  }

  .hint {
    color: #949ba4;
    font-size: var(--text-xs);
  }

  ul {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    min-height: 2.5rem;
    padding: 0 var(--space-3);
    border-radius: var(--radius-md);
    background: #383a40;
  }

  .answer {
    overflow-wrap: anywhere;
  }

  .placeholder .answer {
    color: #80848e;
  }

  .pick {
    flex: none;
    width: 1.125rem;
    height: 1.125rem;
    border: 2px solid #80848e;
    border-radius: var(--radius-full);
  }

  .pick.square {
    border-radius: 0.25rem;
  }
</style>
