<!--
  Home Tools (tools.<domain>): one page with a shortcut to every tool, so
  there's one thing to bookmark (decision #46). No profile picker: nothing
  here is per person. The page's own accent is purple (index.html).
  The hero is the brand (wordmark + every tool's mascot); the cards below
  are the links.
-->
<script lang="ts">
  import { Hero } from '@/features/hero/hero'
  import { ToolCards } from '@/features/shortcuts/tool-cards'
</script>

<main class="home">
  <Hero />
  <ToolCards />
  <footer>Runs on your home server.</footer>
</main>

<style>
  .home {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
    min-height: 100dvh;
    padding-bottom: var(--space-7);
  }

  footer {
    margin-top: var(--space-5);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    text-align: center;
  }
</style>
