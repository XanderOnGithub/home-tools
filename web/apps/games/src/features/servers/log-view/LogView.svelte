<!--
  A server's live log (Server-Sent Events, decision #37): the last 200
  lines, then new ones as they're written.

  - The browser's EventSource reconnects by itself after a network blip;
    each (re)connect starts with the tail again, so lines are cleared on
    open instead of duplicated.
  - "end" = the container stopped: the stream is over until it starts
    again ("Reconnect"). "failure" = the server's explanation.
  - Lines arrive in bursts (200 at once on connect), so they're buffered
    and added once per animation frame; only the last MAX_LINES are kept.
  - Follows the newest line while you're at the bottom; scroll up to read
    and it stops following until you scroll back down.
-->
<script lang="ts">
  import { tick } from 'svelte'
  import { logsUrl } from '@/features/servers'

  let { id }: { id: string } = $props()

  const MAX_LINES = 1000

  let lines = $state<string[]>([])
  let status = $state<'connecting' | 'live' | 'ended' | 'failed'>('connecting')
  let failure = $state('')
  let attempt = $state(0) // bump to reconnect
  let box = $state<HTMLPreElement>()
  let following = true

  $effect(() => {
    void attempt // re-run (reconnect) when it changes
    const source = new EventSource(logsUrl(id))
    let buffer: string[] = []
    let frame = 0
    status = 'connecting'

    const flush = () => {
      frame = 0
      lines = [...lines, ...buffer].slice(-MAX_LINES)
      buffer = []
      if (following) tick().then(() => box && (box.scrollTop = box.scrollHeight))
    }

    source.onopen = () => {
      lines = []
      buffer = []
      status = 'live'
    }
    source.onmessage = (e) => {
      buffer.push(e.data)
      frame ||= requestAnimationFrame(flush)
    }
    source.addEventListener('end', () => {
      status = 'ended'
      source.close()
    })
    source.addEventListener('failure', (e) => {
      failure = (e as MessageEvent).data
      status = 'failed'
      source.close()
    })
    // Connection trouble: EventSource retries by itself unless the server
    // refused outright (then it's CLOSED).
    source.onerror = () => {
      if (source.readyState === EventSource.CLOSED && status !== 'ended') {
        failure = "Couldn't connect to the log."
        status = 'failed'
      } else if (status === 'live') {
        status = 'connecting'
      }
    }

    return () => {
      source.close()
      cancelAnimationFrame(frame)
    }
  })

  function onScroll() {
    if (!box) return
    following = box.scrollHeight - box.scrollTop - box.clientHeight < 24
  }

  /** Warnings and errors stand out (the word is in the line, so color
   * isn't the only signal). */
  function level(line: string): string {
    if (/\b(ERROR|SEVERE|FATAL|Exception)\b/.test(line)) return 'error'
    if (/\bWARN(ING)?\b/.test(line)) return 'warn'
    return ''
  }
</script>

<section class="log" aria-labelledby="log-title">
  <div class="head">
    <h2 id="log-title">Log</h2>
    <p class="status" role="status">
      {#if status === 'live'}
        <span class="dot live" aria-hidden="true"></span> Live
      {:else if status === 'connecting'}
        Connecting…
      {:else if status === 'ended'}
        Stopped; the log ends here.
      {:else}
        {failure}
      {/if}
    </p>
    {#if status === 'ended' || status === 'failed'}
      <button type="button" class="btn btn-quiet" onclick={() => attempt++}>Reconnect</button>
    {/if}
  </div>
  <!-- tabindex: a scrolling box must be reachable by keyboard to scroll it
       (WCAG 2.1.1); Safari doesn't make scrollers focusable by itself.
       aria-live off: a busy server would otherwise read out every line. -->
  <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
  <pre aria-label="Server log" bind:this={box} class="lines" tabindex="0" aria-live="off" onscroll={onScroll}>{#if lines.length}{#each lines as line, i (i)}<span class="line {level(line)}">{line}</span>{'\n'}{/each}{:else if status === 'live'}<span class="empty">Nothing logged yet.</span>{/if}</pre>
</section>

<style>
  .log {
    display: flex;
    flex-direction: column;
  }

  .head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2) var(--space-3);
    min-height: 3.25rem;
    padding: var(--space-2) var(--space-2) var(--space-2) var(--space-4);
    border-bottom: 1px solid var(--color-border);
  }

  h2 {
    font-size: var(--text-md);
  }

  .status {
    display: flex;
    flex: 1;
    align-items: center;
    gap: var(--space-2);
    margin: 0;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .dot {
    width: 0.4375rem;
    height: 0.4375rem;
    border-radius: var(--radius-full);
    background: var(--color-success-text);
  }

  @media (prefers-reduced-motion: no-preference) {
    .live {
      animation: pulse 2s ease-in-out infinite;
    }
  }

  @keyframes pulse {
    50% {
      opacity: 0.35;
    }
  }

  /* Long lines wrap (no sideways scrolling on a phone); the box scrolls. */
  .lines {
    height: min(30rem, 55dvh);
    margin: 0;
    padding: var(--space-3) var(--space-4);
    overflow-y: auto;
    background: var(--color-surface-sunken);
    font-family: var(--font-mono);
    font-size: 0.8125rem;
    line-height: 1.6;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  .lines:focus-visible {
    outline-offset: -2px;
  }

  .warn {
    color: var(--color-warning-text);
  }

  .error {
    color: var(--color-danger-text);
  }

  .empty {
    color: var(--color-text-muted);
  }
</style>
