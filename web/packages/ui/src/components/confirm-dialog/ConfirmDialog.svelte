<!--
  "Are you sure?" for actions that can't be taken back or that affect other
  people (stop a server, delete a plan). Native <dialog> + showModal(): focus
  moves in, Esc closes, focus goes back to whatever opened it. The safe
  choice is on the left and gets focus first; the confirm button is red
  when `danger` (decision #57: one confirm, the same everywhere).

    <ConfirmDialog bind:this={confirm} title="Stop SummerCraft?" … onconfirm={stop} />
    confirm.open()
-->
<script lang="ts">
  let {
    title,
    message,
    confirmLabel,
    cancelLabel = 'Cancel',
    danger = false,
    onconfirm,
  }: {
    title: string
    message?: string
    confirmLabel: string
    cancelLabel?: string
    danger?: boolean
    onconfirm: () => void
  } = $props()

  let dialog = $state<HTMLDialogElement>()
  let opener: HTMLElement | null = null
  const id = `confirm-${Math.random().toString(36).slice(2, 9)}`

  export function open() {
    opener = document.activeElement as HTMLElement | null
    dialog?.showModal()
  }

  function confirm() {
    dialog?.close()
    onconfirm()
  }
</script>

<dialog
  bind:this={dialog}
  class="dialog confirm"
  aria-labelledby="{id}-title"
  aria-describedby={message ? `${id}-message` : undefined}
  onclose={() => opener?.focus()}
>
  <h2 id="{id}-title">{title}</h2>
  {#if message}<p id="{id}-message">{message}</p>{/if}
  <div class="actions">
    <!-- svelte-ignore a11y_autofocus -->
    <button type="button" class="btn btn-secondary" autofocus onclick={() => dialog?.close()}>{cancelLabel}</button>
    <button type="button" class="btn {danger ? 'btn-danger' : 'btn-primary'}" onclick={confirm}>{confirmLabel}</button>
  </div>
</dialog>

<style>
  .confirm {
    width: min(26rem, calc(100vw - 2 * var(--space-4)));
    padding: var(--space-5);
    gap: var(--space-2);
  }

  h2 {
    font-size: var(--text-lg);
  }

  p {
    margin: 0;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: var(--space-2);
    margin-top: var(--space-4);
  }
</style>
