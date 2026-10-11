<!--
  Home's "Your body" card (decision #49): the muscle map, then the same
  thing in words (what needs rest, what's recovering), and a link to the
  Body page for details. The words are what screen readers get; the map
  is decorative for them.
-->
<script lang="ts">
  import { muscleList, regionStates, type MuscleRecovery } from '@/features/body'
  import { MuscleMap } from '@/features/body/muscle-map'

  let { recovery }: { recovery: MuscleRecovery[] } = $props()

  let rest = $derived(recovery.filter((m) => m.state === 'rest').map((m) => m.muscle))
  let recovering = $derived(recovery.filter((m) => m.state === 'recovering').map((m) => m.muscle))
  let trained = $derived(recovery.some((m) => m.last_worked))
</script>

<section class="body" aria-labelledby="body-title">
  <div class="head">
    <h2 id="body-title">Your body</h2>
    <a href="/body">Details<span class="visually-hidden"> about your muscles</span></a>
  </div>

  <div class="map"><MuscleMap states={regionStates(recovery)} /></div>

  {#if !trained}
    <p class="muted">Log a workout and you'll see here which muscles it worked and when they're ready again.</p>
  {:else if rest.length === 0 && recovering.length === 0}
    <p>Everything's ready to train.</p>
  {:else}
    <dl class="summary">
      {#if rest.length > 0}
        <div><dt>Needs rest</dt><dd>{muscleList(rest)}</dd></div>
      {/if}
      {#if recovering.length > 0}
        <div><dt>Recovering</dt><dd>{muscleList(recovering)}</dd></div>
      {/if}
    </dl>
  {/if}
</section>

<style>
  .body {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-3);
  }

  h2 {
    font-size: var(--text-lg);
    font-weight: var(--weight-bold);
  }

  .head a {
    display: inline-flex;
    align-items: center;
    min-height: var(--touch-target);
    color: var(--color-accent-text);
    font-weight: var(--weight-semibold);
  }

  .map {
    max-width: 20rem;
    width: 100%;
    margin: 0 auto;
  }

  .map :global(svg) {
    max-height: 13rem;
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
  }

  .summary {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
  }

  .summary div {
    display: flex;
    gap: var(--space-2);
  }

  dt {
    flex: none;
    font-weight: var(--weight-semibold);
  }

  dt::after {
    content: ':';
  }

  dd {
    margin: 0;
    color: var(--color-text-muted);
  }
</style>
