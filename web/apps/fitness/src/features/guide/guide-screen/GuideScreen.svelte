<!--
  "How training works" (/guide, decision #54): the why behind what the
  app suggests, in plain words, one short section each. Other screens
  link to a section (/guide#reps). The numbers match what the app uses
  (features/training, tools/fitness recovery and strength); the research
  behind recovery is docs/research/muscle-recovery.md.
-->
<script lang="ts">
  import { PageHeader } from '@home-tools/ui/components/page-header'
  import { GOAL_KEYS, GOALS } from '@/features/training/guidance'

  const SECTIONS = [
    { id: 'effort', title: 'How hard' },
    { id: 'reps', title: 'How many reps' },
    { id: 'sets', title: 'How many sets' },
    { id: 'rest', title: 'How often, and rest' },
    { id: 'progress', title: 'Getting stronger' },
    { id: 'start', title: 'Every workout' },
  ]

  // Arriving from a link like /guide#reps: scroll to that section. After
  // the next frame, so it runs after the app shell focuses the page title.
  requestAnimationFrame(() => {
    const id = location.hash.slice(1)
    if (id) document.getElementById(id)?.scrollIntoView()
  })
</script>

<div class="guide">
  <PageHeader title="How training works" description="Why the app suggests what it does, in five minutes." />

  <nav aria-label="On this page">
    <ul class="toc">
      {#each SECTIONS as s (s.id)}
        <li><a href="#{s.id}">{s.title}</a></li>
      {/each}
    </ul>
  </nav>

  <section id="effort" aria-labelledby="effort-title">
    <h2 id="effort-title">How hard</h2>
    <p>
      <strong>Stop 1–2 reps before you couldn't do another one.</strong> That's close enough to build strength and muscle,
      and far enough to recover in two or three days. Going all the way to failure on every set adds a day or two of
      recovery for little extra.
    </p>
    <p class="muted">Not sure how many you had left? If the last rep was slow but clean, you're about right.</p>
  </section>

  <section id="reps" aria-labelledby="reps-title">
    <h2 id="reps-title">How many reps</h2>
    <p>It depends on what you're training for. Big lifts are the ones that move several muscles (squats, presses, rows).</p>
    <dl class="table">
      {#each GOAL_KEYS as key (key)}
        {@const g = GOALS[key]}
        <div>
          <dt>{g.label}</dt>
          <dd>
            {#if g.big[0] === g.small[0] && g.big[1] === g.small[1]}{g.big[0]}–{g.big[1]} reps{:else}{g.big[0]}–{g.big[1]} on big lifts, {g.small[0]}–{g.small[1]} on small ones{/if},
            {g.effort}.
            <span class="muted">{g.note}</span>
          </dd>
        </div>
      {/each}
    </dl>
    <p class="muted">
      Any of these build muscle if the sets are hard enough; the range mostly decides how heavy it feels. Pick your goal
      under Plans → Your routine; a plan can also have its own.
    </p>
  </section>

  <section id="sets" aria-labelledby="sets-title">
    <h2 id="sets-title">How many sets</h2>
    <p>
      Count <strong>hard sets per muscle per week</strong>. Under 5 builds very little, 5–9 builds muscle, and 10 or more
      builds it fastest. That's what "What your week works" on the Plans page counts for you.
    </p>
    <p class="muted">
      In one workout, 6–8 hard sets for a muscle is plenty: more takes 3–4 days to recover from. Two shorter days beat
      one long one.
    </p>
  </section>

  <section id="rest" aria-labelledby="rest-title">
    <h2 id="rest-title">How often, and rest</h2>
    <p>
      <strong>Work each muscle about twice a week</strong>, with two or three days between. Muscles get stronger while you
      rest, not while you train. The body map on Home shows what still needs rest and when it's ready.
    </p>
    <p class="muted">
      A new exercise (or one after a break) makes you sorer than usual the first time; the second time is much easier.
      Soreness alone doesn't tell you much: how your sets go does.
    </p>
  </section>

  <section id="progress" aria-labelledby="progress-title">
    <h2 id="progress-title">Getting stronger</h2>
    <ol class="steps">
      <li>Keep the weight and add a rep or two each time.</li>
      <li>When every set reaches the top of your range (12 for 8–12), add a little weight.</li>
      <li>You'll drop back to the bottom of the range. Build up again.</li>
    </ol>
    <p class="muted">
      Workout mode tells you which step you're on. For bodyweight exercises, "add weight" means a harder version or
      going slower.
    </p>
  </section>

  <section id="start" aria-labelledby="start-title">
    <h2 id="start-title">Every workout</h2>
    <ul class="steps">
      <li><strong>Warm up:</strong> 5 minutes of easy movement, then a light set or two of your first exercise.</li>
      <li><strong>Form first:</strong> a lighter weight done well beats a heavier one done badly. Each exercise has photos and steps.</li>
      <li><strong>Rest between sets:</strong> 1–2 minutes, longer for heavy lifts. The timer starts by itself.</li>
      <li><strong>Pain is a stop sign:</strong> burning muscles are normal, sharp or joint pain isn't.</li>
    </ul>
  </section>
</div>

<style>
  .guide {
    display: flex;
    flex-direction: column;
    gap: var(--space-6);
    max-width: 42rem;
  }

  section {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    scroll-margin-top: calc(var(--touch-target) + var(--space-7)); /* clears the sticky header */
    line-height: var(--leading-normal);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
  }

  .toc {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .toc a {
    display: inline-flex;
    align-items: center;
    min-height: var(--touch-target);
    padding: 0 var(--space-4);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-full);
    color: var(--color-text);
    font-weight: var(--weight-medium);
    text-decoration: none;
  }

  @media (hover: hover) {
    .toc a:hover {
      border-color: var(--color-accent);
      background: var(--color-accent-subtle);
    }
  }

  .table {
    display: flex;
    flex-direction: column;
    margin: 0;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
  }

  .table div {
    display: grid;
    grid-template-columns: 9rem 1fr;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
  }

  .table div + div {
    border-top: 1px solid var(--color-border);
  }

  dt {
    font-weight: var(--weight-semibold);
  }

  dd {
    margin: 0;
  }

  dd .muted {
    display: block;
    margin-top: var(--space-1);
    font-size: var(--text-sm);
  }

  .steps {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
    padding-left: var(--space-5);
  }
</style>
