<!--
  Home Tools (tools.<domain>): one page with a shortcut to every tool, so
  there's one thing to bookmark (decision #46). No profile picker: nothing
  here is per person. The page's own accent is purple (index.html).
-->
<script lang="ts">
  import { ToolIcon } from '@home-tools/ui/tool-icon'
  import { HOME } from '@home-tools/ui/tools'
  import { ToolList } from '@/features/shortcuts/tool-list'
</script>

<main class="home">
  <header>
    <ToolIcon tool={HOME} size="lg" />
    <h1>Home Tools</h1>
    <p>Everything for the house, in one place.</p>
  </header>
  <ToolList />
</main>

<style>
  .home {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
    width: 100%;
    max-width: 30rem;
    min-height: 100dvh;
    margin: 0 auto;
    padding: clamp(var(--space-6), 12vh, var(--space-8)) var(--page-gutter) var(--space-7);
  }

  header {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
    text-align: center;
  }

  header :global(.tile) {
    margin-bottom: var(--space-2);
  }

  h1 {
    font-size: var(--text-2xl);
    font-weight: var(--weight-bold);
  }

  p {
    margin: 0;
    color: var(--color-text-muted);
  }
</style>
