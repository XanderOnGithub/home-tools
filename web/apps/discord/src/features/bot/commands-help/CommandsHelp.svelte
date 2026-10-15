<!--
  The bot's slash commands: what each does, who may use it, and whether
  it's on (a command whose feature is off is gone from Discord).
-->
<script lang="ts">
  import { Badge } from '@home-tools/ui/components/badge'
  import { Card } from '@home-tools/ui/components/card'
  import type { Features } from '@/features/bot'

  let { features }: { features: Features } = $props()

  let commands = $derived([
    { name: '/sens', what: 'Converts a mouse sensitivity between 17 games', who: 'Anyone', on: true },
    { name: '/restart', what: 'Restarts a game server, at most every 2 minutes', who: 'Verified', on: true },
    { name: '/whitelist', what: 'Adds, removes or lists players on a server', who: 'Verified', on: true },
    { name: '/blob', what: 'Makes a blob from a name and a color', who: 'Anyone', on: features.blob.enabled },
  ])
</script>

<Card padding="none">
  <ul class="card-list">
    {#each commands as c (c.name)}
      <li class:off={!c.on}>
        <code>{c.name}</code>
        <span class="what">{c.what}</span>
        <span class="badges">
          {#if c.on}
            <Badge tone={c.who === 'Verified' ? 'accent' : 'neutral'}>{c.who}</Badge>
          {:else}
            <Badge>Off</Badge>
          {/if}
        </span>
      </li>
    {/each}
  </ul>
</Card>

<style>
  li {
    display: grid;
    grid-template-columns: 6.5rem 1fr auto;
    align-items: center;
    gap: var(--space-1) var(--space-4);
    padding: var(--space-3) var(--space-4);
    font-size: var(--text-sm);
  }

  @media (max-width: 30rem) {
    li {
      grid-template-columns: 1fr auto;
      grid-template-areas: 'name badges' 'what what';
    }

    code {
      grid-area: name;
    }

    .what {
      grid-area: what;
    }

    .badges {
      grid-area: badges;
    }
  }

  code {
    font-family: var(--font-mono);
    font-weight: var(--weight-semibold);
  }

  .what {
    color: var(--color-text-muted);
  }

  .off code,
  .off .what {
    opacity: 0.55;
  }
</style>
