<!--
  The ⋯ button on a server and its small menu: Start (when stopped) or Stop
  (when running), and Restart (when running). Stop and Restart ask first in
  a small dialog, since both kick out anyone playing. Native Popover for
  the menu (Esc / outside click close it), native <dialog> for the confirm.

  `pending` is bindable so the card can say "Stopping…" while Docker works
  (up to a minute: the game saves its world first).
-->
<script lang="ts">
  import { tick } from 'svelte'
  import { runAction, type Action, type Server } from '@/features/servers'

  let {
    server,
    onchange,
    onerror,
    pending = $bindable(null),
  }: {
    server: Server
    onchange: (updated: Server) => void
    onerror: (message: string) => void
    pending?: Action | null
  } = $props()

  let trigger = $state<HTMLButtonElement>()
  let menu = $state<HTMLDivElement>()
  let dialog = $state<HTMLDialogElement>()
  let confirming = $state<'stop' | 'restart'>('stop')

  let running = $derived(server.state?.running ?? false)
  let menuId = $derived(`menu-${server.id}`)

  // Right under the button, right edges aligned (like the profile menu).
  function onBeforeToggle(event: ToggleEvent) {
    if (event.newState !== 'open' || !trigger || !menu) return
    const r = trigger.getBoundingClientRect()
    menu.style.top = `${r.bottom + 4}px`
    menu.style.right = `${document.documentElement.clientWidth - r.right}px`
  }

  async function choose(action: Action) {
    menu?.hidePopover()
    if (action === 'start') return run(action)
    confirming = action
    await tick()
    dialog?.showModal()
  }

  async function run(action: Action) {
    dialog?.close()
    pending = action
    try {
      onchange(await runAction(server.id, action))
    } catch (err) {
      onerror((err as Error).message)
    }
    pending = null
  }
</script>

<button
  bind:this={trigger}
  type="button"
  class="btn btn-icon trigger"
  popovertarget={menuId}
  aria-label="Actions for {server.name}"
  disabled={!!pending || !!server.busy || !server.state}
>
  <svg viewBox="0 0 24 24"><path d="M5 12h.01 M12 12h.01 M19 12h.01" /></svg>
</button>

<div bind:this={menu} id={menuId} class="menu" popover="auto" onbeforetoggle={onBeforeToggle}>
  {#if running}
    <button type="button" class="item" onclick={() => choose('stop')}>
      <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="6" y="6" width="12" height="12" rx="2" /></svg>
      Stop
    </button>
    <button type="button" class="item" onclick={() => choose('restart')}>
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M20 11a8 8 0 1 0-2.3 5.7 M20 4v7h-7" /></svg>
      Restart
    </button>
  {:else}
    <button type="button" class="item" onclick={() => choose('start')}>
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5l11 7-11 7z" /></svg>
      Start
    </button>
  {/if}
</div>

<dialog bind:this={dialog} class="dialog confirm" aria-labelledby="confirm-{server.id}" onclose={() => trigger?.focus()}>
  <h2 id="confirm-{server.id}">{confirming === 'stop' ? 'Stop' : 'Restart'} {server.name}?</h2>
  <p>Anyone playing is disconnected. The world is saved first, which can take up to a minute.</p>
  <div class="actions">
    <button type="button" class="btn btn-secondary" onclick={() => dialog?.close()}>Keep running</button>
    <button type="button" class="btn btn-danger" onclick={() => run(confirming)}>
      {confirming === 'stop' ? 'Stop' : 'Restart'}
    </button>
  </div>
</dialog>

<style>
  .trigger svg {
    stroke-width: 3.5;
  }

  /* Popovers default to centered on screen; pin it under the button. */
  .menu {
    position: fixed;
    inset: auto;
    min-width: 11rem;
    margin: 0;
    padding: var(--space-2);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface-raised);
    color: var(--color-text);
    box-shadow: var(--shadow-md);
  }

  .item {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: none;
    border-radius: var(--radius-sm);
    background: none;
    color: inherit;
    font-weight: var(--weight-medium);
    text-align: left;
    cursor: pointer;
  }

  @media (hover: hover) {
    .item:hover {
      background: var(--color-accent-subtle);
    }
  }

  .item svg {
    width: 1.25rem;
    height: 1.25rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .confirm {
    width: min(24rem, calc(100vw - 2 * var(--space-4)));
    padding: var(--space-6);
    gap: var(--space-3);
  }

  .confirm h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  .confirm p {
    margin: 0;
    color: var(--color-text-muted);
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: var(--space-2);
    margin-top: var(--space-3);
  }
</style>
