<!--
  A problem or notice that concerns a whole page, said once at the top
  instead of on every card ("Docker isn't connected"). Tones: info,
  warning, danger; danger is announced right away (role="alert"), the
  others politely. One line of title, at most one line of description,
  and optional actions (a Try again button, a link).
-->
<script lang="ts">
  import type { Snippet } from 'svelte'

  let {
    tone = 'info',
    title,
    children,
    actions,
  }: {
    tone?: 'info' | 'warning' | 'danger'
    title: string
    children?: Snippet // the one line of description
    actions?: Snippet
  } = $props()

  // An "i" in a circle for info, a "!" in a triangle for warning and danger.
  const ICONS = {
    info: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18z M12 11v5 M12 7.5h.01',
    warning: 'M12 3.5 2.5 20h19L12 3.5z M12 10v4 M12 17h.01',
    danger: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18z M12 7.5v5 M12 16h.01',
  }
</script>

<div class="banner {tone}" role={tone === 'danger' ? 'alert' : 'status'}>
  <svg viewBox="0 0 24 24" aria-hidden="true"><path d={ICONS[tone]} /></svg>
  <div class="text">
    <p class="title">{title}</p>
    {#if children}<p class="description">{@render children()}</p>{/if}
  </div>
  {#if actions}<div class="actions">{@render actions()}</div>{/if}
</div>

<style>
  .banner {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
    color: var(--color-text);
  }

  svg {
    flex: none;
    width: 1.25rem;
    height: 1.25rem;
    margin-top: 0.0625rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .info svg {
    color: var(--color-accent-text);
  }

  .warning {
    background: var(--color-warning-subtle);
  }

  .warning svg {
    color: var(--color-warning-text);
  }

  .danger {
    background: var(--color-danger-subtle);
  }

  .danger svg {
    color: var(--color-danger-text);
  }

  .text {
    flex: 1 1 14rem;
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
    min-width: 0;
  }

  p {
    margin: 0;
  }

  .title {
    font-weight: var(--weight-semibold);
  }

  .description {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
  }
</style>
