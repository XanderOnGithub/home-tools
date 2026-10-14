<!--
  A small status label: "Running", "Verified", "Asked Oct 9". The dot
  carries the color; the word carries the meaning (never color alone).
  Tones: neutral (default), accent, success, warning, danger.
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  let {
    tone = 'neutral',
    dot = false,
    pulse = false,
    children,
  }: {
    tone?: 'neutral' | 'accent' | 'success' | 'warning' | 'danger'
    dot?: boolean
    pulse?: boolean // the dot breathes: something live (a running server)
    children: Snippet
  } = $props()
</script>

<span class="badge {tone}">
  {#if dot}<span class="dot" class:pulse aria-hidden="true"></span>{/if}
  {@render children()}
</span>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    min-height: 1.5rem;
    padding: 0 var(--space-2);
    border-radius: var(--radius-sm);
    background: var(--color-surface-sunken);
    color: var(--color-text-muted);
    font-size: var(--text-xs);
    font-weight: var(--weight-semibold);
    white-space: nowrap;
  }

  .dot {
    width: 0.4375rem;
    height: 0.4375rem;
    border-radius: var(--radius-full);
    background: currentColor;
  }

  .accent {
    background: var(--color-accent-subtle);
    color: var(--color-accent-text);
  }

  .success {
    background: var(--color-success-subtle);
    color: var(--color-success-text);
  }

  .warning {
    background: var(--color-warning-subtle);
    color: var(--color-warning-text);
  }

  .danger {
    background: var(--color-danger-subtle);
    color: var(--color-danger-text);
  }

  .pulse {
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    50% {
      opacity: 0.35;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .pulse {
      animation: none;
    }
  }
</style>
