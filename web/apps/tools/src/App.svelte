<!--
  Home Tools (tools.<domain>): one page with a shortcut to every tool, so
  there's one thing to bookmark (decision #46). No profile picker: nothing
  here is per person. The page's own accent is purple (index.html).
-->
<script lang="ts">
  import { ToolList } from '@/features/shortcuts/tool-list'
</script>

<main class="home">
  <h1>Home Tools</h1>
  <ToolList />
</main>

<style>
  /* Centered in the viewport, like the profile picker. */
  .home {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-6);
    min-height: 100dvh;
    padding: var(--space-5) var(--space-4);
  }

  h1 {
    font-size: var(--text-3xl);
    font-weight: var(--weight-extrabold);
  }
</style>
