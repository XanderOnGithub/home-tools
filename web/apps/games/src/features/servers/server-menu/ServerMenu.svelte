<!--
  The ⋯ button on a server and its menu: Start (when stopped) or Stop
  (when running), and Restart (when running). Stop and Restart ask first,
  since both kick out anyone playing. Shared menu look and ConfirmDialog
  from @home-tools/ui; the outcome is a toast ("SummerCraft 2 restarted"),
  so it's confirmed even if you've moved on.

  `pending` is bindable so the card can say "Stopping…" while Docker works
  (up to a minute: the game saves its world first).
-->
<script lang="ts">
  import { ConfirmDialog } from '@home-tools/ui/components/confirm-dialog'
  import { placeUnder } from '@home-tools/ui/menu'
  import { toast } from '@home-tools/ui/toast'
  import { runAction, type Action, type Server } from '@/features/servers'

  let {
    server,
    onchange,
    pending = $bindable(null),
  }: {
    server: Server
    onchange: (updated: Server) => void
    pending?: Action | null
  } = $props()

  let trigger = $state<HTMLButtonElement>()
  let menu = $state<HTMLDivElement>()
  let confirmStop = $state<ReturnType<typeof ConfirmDialog>>()
  let confirmRestart = $state<ReturnType<typeof ConfirmDialog>>()

  let running = $derived(server.state?.running ?? false)
  let menuId = $derived(`menu-${server.id}`)

  const DONE: Record<Action, string> = { start: 'started', stop: 'stopped', restart: 'restarted' }

  function choose(action: Action) {
    menu?.hidePopover()
    if (action === 'start') run('start')
    else if (action === 'stop') confirmStop?.open()
    else confirmRestart?.open()
  }

  async function run(action: Action) {
    pending = action
    try {
      onchange(await runAction(server.id, action))
      toast(`${server.name} ${DONE[action]}`)
    } catch (err) {
      toast.error(`Couldn't ${action} ${server.name}: ${(err as Error).message}`)
    }
    pending = null
  }

  const WARNING = 'Anyone playing is disconnected. The world is saved first, which can take up to a minute.'
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

<div bind:this={menu} id={menuId} class="menu" popover="auto" onbeforetoggle={(e) => placeUnder(e, trigger, menu, 'end')}>
  {#if running}
    <button type="button" class="menu-item" onclick={() => choose('restart')}>
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M20 11a8 8 0 1 0-2.3 5.7 M20 4v7h-7" /></svg>
      Restart
    </button>
    <button type="button" class="menu-item menu-item-danger" onclick={() => choose('stop')}>
      <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="6" y="6" width="12" height="12" rx="2" /></svg>
      Stop
    </button>
  {:else}
    <button type="button" class="menu-item" onclick={() => choose('start')}>
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5l11 7-11 7z" /></svg>
      Start
    </button>
  {/if}
</div>

<ConfirmDialog
  bind:this={confirmStop}
  title="Stop {server.name}?"
  message={WARNING}
  cancelLabel="Keep running"
  confirmLabel="Stop server"
  danger
  onconfirm={() => run('stop')}
/>
<ConfirmDialog
  bind:this={confirmRestart}
  title="Restart {server.name}?"
  message={WARNING}
  cancelLabel="Keep running"
  confirmLabel="Restart"
  onconfirm={() => run('restart')}
/>

<style>
  .trigger svg {
    stroke-width: 3.5;
  }
</style>
