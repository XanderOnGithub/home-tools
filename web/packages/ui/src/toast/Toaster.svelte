<!--
  Where toasts appear: bottom center on phones (above the tab bar), bottom
  right on wide screens. A polite live region, so screen readers read each
  one without interrupting; errors use role="alert". Rendered once, by
  AppShell (pages without the shell can add their own).
-->
<script lang="ts">
  import { fly } from 'svelte/transition'
  import { dismiss, pause, resume, toasts } from './toast.svelte'

  const reduced = typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches
</script>

<div class="toaster" aria-live="polite">
  {#each toasts as t (t.id)}
    <div
      class="toast {t.tone}"
      role={t.tone === 'error' ? 'alert' : 'status'}
      transition:fly={{ y: reduced ? 0 : 16, duration: reduced ? 0 : 200 }}
      onmouseenter={() => pause(t.id)}
      onmouseleave={() => resume(t.id)}
      onfocusin={() => pause(t.id)}
      onfocusout={() => resume(t.id)}
    >
      <svg class="icon" viewBox="0 0 24 24" aria-hidden="true">
        {#if t.tone === 'success'}
          <path d="M5 12.5l4.5 4.5L19 7.5" />
        {:else if t.tone === 'error'}
          <path d="M12 8v5 M12 16.5h.01 M10.3 3.9L2.4 17.5A2 2 0 0 0 4.1 20.5h15.8a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0z" />
        {:else}
          <path d="M12 11v5 M12 7.5h.01 M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18z" />
        {/if}
      </svg>
      <p>{t.message}</p>
      {#if t.action}
        {@const action = t.action}
        <button
          type="button"
          class="action"
          onclick={() => {
            action.run()
            dismiss(t.id)
          }}>{action.label}</button
        >
      {/if}
      <button type="button" class="close" aria-label="Dismiss" onclick={() => dismiss(t.id)}>
        <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12 M18 6L6 18" /></svg>
      </button>
    </div>
  {/each}
</div>

<style>
  .toaster {
    position: fixed;
    z-index: 10;
    inset: auto var(--space-3) calc(4rem + env(safe-area-inset-bottom)) var(--space-3);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
    pointer-events: none;
  }

  @media (min-width: 40rem) {
    .toaster {
      inset: auto var(--space-5) var(--space-5) auto;
      align-items: flex-end;
    }
  }

  /* Inverted (dark on light, light on dark), like every OS's toasts: it
   * reads as "over" the page, not part of it. */
  .toast {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: max-content;
    max-width: min(26rem, 100%);
    min-height: 3rem;
    padding: var(--space-2) var(--space-2) var(--space-2) var(--space-4);
    border-radius: var(--radius-lg);
    background: var(--color-toast);
    color: var(--color-on-toast);
    box-shadow: var(--shadow-lg);
    font-size: var(--text-sm);
    pointer-events: auto;
  }

  p {
    flex: 1;
    margin: 0;
    font-weight: var(--weight-medium);
  }

  .icon {
    flex: none;
    width: 1.125rem;
    height: 1.125rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.25;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .success .icon {
    color: var(--color-toast-success);
  }

  .error .icon {
    color: var(--color-toast-danger);
  }

  .action {
    min-height: 2.25rem;
    padding: 0 var(--space-3);
    border: none;
    border-radius: var(--radius-md);
    background: none;
    color: var(--color-toast-action);
    font-weight: var(--weight-semibold);
    cursor: pointer;
  }

  .close {
    display: grid;
    flex: none;
    place-items: center;
    width: 2.25rem;
    height: 2.25rem;
    padding: 0;
    border: none;
    border-radius: var(--radius-md);
    background: none;
    color: var(--color-on-toast-muted);
    cursor: pointer;
  }

  @media (hover: hover) {
    .action:hover,
    .close:hover {
      background: var(--color-toast-hover);
    }
  }

  .close svg {
    width: 1rem;
    height: 1rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
  }
</style>
