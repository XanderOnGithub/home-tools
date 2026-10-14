<!--
  Minecraft console (decision #41): type a command, see the game's answer.
  Anyone on the LAN may use it (#10). History lives only on this page;
  ↑/↓ in the box recall earlier commands, like a terminal. Answers are
  announced to screen readers (aria-live).
-->
<script lang="ts">
  import { tick } from 'svelte'
  import { runCommand } from '@/features/servers'

  let { id }: { id: string } = $props()

  type Entry = { command: string; output: string; failed: boolean }

  let entries = $state<Entry[]>([])
  let command = $state('')
  let busy = $state(false)
  let box = $state<HTMLOListElement>()
  let recall = -1 // index into sent commands while browsing with ↑/↓; -1 = new

  let sent = $derived(entries.map((e) => e.command))

  async function send(event: SubmitEvent) {
    event.preventDefault()
    const cmd = command.trim()
    if (!cmd || busy) return
    busy = true
    let entry: Entry
    try {
      const { output } = await runCommand(id, cmd)
      entry = { command: cmd, output: output || 'Done (no answer).', failed: false }
      command = ''
    } catch (err) {
      entry = { command: cmd, output: (err as Error).message, failed: true }
    } finally {
      busy = false
      recall = -1
    }
    entries = [...entries, entry]
    await tick()
    box?.lastElementChild?.scrollIntoView({ block: 'nearest' })
  }

  function onkeydown(event: KeyboardEvent) {
    if (event.key !== 'ArrowUp' && event.key !== 'ArrowDown') return
    if (sent.length === 0) return
    event.preventDefault()
    if (event.key === 'ArrowUp') recall = recall === -1 ? sent.length - 1 : Math.max(0, recall - 1)
    else recall = recall === -1 || recall === sent.length - 1 ? -1 : recall + 1
    command = recall === -1 ? '' : sent[recall]
  }
</script>

<section class="console" aria-labelledby="console-title">
  <h2 id="console-title" class="visually-hidden">Console</h2>

  {#if entries.length > 0}
    <ol class="entries" bind:this={box} aria-live="polite">
      {#each entries as e, i (i)}
        <li>
          <span class="command"><span aria-hidden="true">&gt;</span> {e.command}</span>
          <span class="output" class:failed={e.failed}>{e.output}</span>
        </li>
      {/each}
    </ol>
  {:else}
    <p class="muted">Run any server command, like <code>say Dinner's ready</code> or <code>whitelist add Steve</code>.</p>
  {/if}

  <form onsubmit={send}>
    <label class="visually-hidden" for="console-command">Command</label>
    <span class="prompt" aria-hidden="true">&gt;</span>
    <input
      id="console-command"
      type="text"
      bind:value={command}
      {onkeydown}
      placeholder="say hello"
      autocomplete="off"
      autocapitalize="off"
      spellcheck="false"
      enterkeyhint="send"
      maxlength="1000"
    />
    <button type="submit" class="btn btn-primary" disabled={busy || !command.trim()}>
      {busy ? 'Sending…' : 'Send'}
    </button>
  </form>
</section>

<style>
  .console {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4) var(--space-4);
    border-top: 1px solid var(--color-border);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  code,
  .entries,
  .prompt,
  input{
    font-family: var(--font-mono);
  }

  code {
    font-size: 0.9em;
  }

  .entries {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    max-height: min(16rem, 35dvh);
    margin: 0;
    padding: 0;
    overflow-y: auto;
    font-size: 0.8125rem;
    list-style: none;
  }

  .entries li {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .command {
    font-weight: var(--weight-semibold);
    overflow-wrap: anywhere;
  }

  .output {
    color: var(--color-text-muted);
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  .output.failed {
    color: var(--color-danger-text);
  }

  form {
    position: relative;
    display: flex;
    gap: var(--space-2);
  }

  .prompt {
    position: absolute;
    top: 50%;
    left: var(--space-3);
    translate: 0 -50%;
    color: var(--color-text-muted);
    pointer-events: none;
  }

  input{
    flex: 1;
    padding-left: var(--space-6);
  }
</style>
