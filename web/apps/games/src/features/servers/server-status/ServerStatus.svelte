<!--
  A server's state as a badge: the word says it, the dot backs it up
  (color is never the only signal). Running = green with a slow pulse
  (it's live), stopped = gray, unavailable = red with the reason on the
  line below; an action in flight ("Stopping…", from this page or anyone
  else's, e.g. the Discord bot) replaces it all. Uptime and players go on
  their own line (see `details`).
-->
<script lang="ts">
  import { Badge } from '@home-tools/ui/components/badge'
  import type { Action, Server } from '@/features/servers'

  let { server, pending = null }: { server: Server; pending?: Action | null } = $props()

  const DOING: Record<Action, string> = { start: 'Starting…', stop: 'Stopping…', restart: 'Restarting…' }
  let running = $derived(server.state?.running ?? false)
  let doing = $derived(pending ?? server.busy ?? null)
</script>

<span class="status">
  {#if doing}
    <Badge tone="warning" dot pulse>{DOING[doing]}</Badge>
  {:else if !server.state}
    <Badge tone="danger" dot>Unavailable</Badge>
  {:else if running}
    <Badge tone="success" dot pulse>Running</Badge>
  {:else}
    <Badge dot>Stopped</Badge>
  {/if}
</span>
{#if !doing && !server.state}
  <span class="reason">{server.error ?? 'Unknown error'}</span>
{/if}

<style>
  .status {
    display: flex;
  }

  .reason {
    color: var(--color-danger-text);
    font-size: var(--text-sm);
  }
</style>
