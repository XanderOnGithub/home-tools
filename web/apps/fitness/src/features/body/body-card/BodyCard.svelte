<!--
  Home's "Your body" card (decision #49): the muscle map, then the same
  thing in words (what needs rest, what's recovering, what's getting
  stronger, #50), and a link to the
  Body page for details. The words are what screen readers get; the map
  is decorative for them.
-->
<script lang="ts">
  import { isStronger, muscleList, recoveryShades, RECOVERY_LEGEND, type MuscleRecovery, type MuscleStrength } from '@/features/body'
  import { MuscleMap } from '@/features/body/muscle-map'

  let { recovery, strength }: { recovery: MuscleRecovery[]; strength: MuscleStrength[] } = $props()

  // Most improved first.
  let stronger = $derived(
    strength
      .filter((m) => isStronger(m.change))
      .toSorted((a, b) => b.change - a.change)
      .map((m) => m.muscle),
  )

  let rest = $derived(recovery.filter((m) => m.state === 'rest').map((m) => m.muscle))
  let recovering = $derived(recovery.filter((m) => m.state === 'recovering').map((m) => m.muscle))
  let trained = $derived(recovery.some((m) => m.last_worked))
</script>

<section class="body" aria-labelledby="body-title">
  <div class="head">
    <h2 id="body-title">Your body</h2>
    <a href="/body">Details<span class="visually-hidden"> about your muscles</span></a>
  </div>

  <div class="map"><MuscleMap shades={recoveryShades(recovery)} legend={RECOVERY_LEGEND} /></div>

  {#if !trained && stronger.length === 0}
    <p class="muted">Log a workout and you'll see here which muscles it worked and when they're ready again.</p>
  {:else}
    {#if trained && rest.length === 0 && recovering.length === 0}
      <p>Everything's ready to train.</p>
    {/if}
    <dl class="summary">
      {#if stronger.length > 0}
        <div class="up"><dt>Getting stronger</dt><dd>{muscleList(stronger)}</dd></div>
      {/if}
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

  /* A full touch target that doesn't push the card's top down. */
  .head a {
    display: inline-flex;
    align-items: center;
    min-height: var(--touch-target);
    margin-block: calc(-1 * var(--space-3));
    color: var(--color-accent-text);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
    text-decoration: none;
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

  .up dt {
    color: var(--color-accent-text);
  }
</style>
