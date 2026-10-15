<!--
  A Discord member's picture (decision #61), served by our own API (the
  browser never calls Discord). Until it loads, or when there's none, the
  first letter of their name on a gray circle. Decorative: the name is
  always next to it.
-->
<script lang="ts">
  import { memberAvatar } from '@/features/bot'

  let { userId, name, size = 'md' }: { userId: string; name: string; size?: 'sm' | 'md' } = $props()

  let failed = $state(false)
  let initial = $derived((name.trim()[0] ?? '?').toUpperCase())
</script>

<span class="avatar {size}" aria-hidden="true">
  <span class="initial">{initial}</span>
  {#if !failed}
    <img src={memberAvatar(userId)} alt="" loading="lazy" onerror={() => (failed = true)} />
  {/if}
</span>

<style>
  .avatar {
    position: relative;
    display: grid;
    flex: none;
    place-items: center;
    overflow: hidden;
    border-radius: var(--radius-full);
    background: var(--color-surface-sunken);
    color: var(--color-text-muted);
    font-weight: var(--weight-semibold);
  }

  .md {
    width: 2.25rem;
    height: 2.25rem;
    font-size: var(--text-sm);
  }

  .sm {
    width: 1.75rem;
    height: 1.75rem;
    font-size: var(--text-xs);
  }

  img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
</style>
