<!--
  The top of the tools home page: the brand. A faint grid, the wordmark,
  one line, and every tool's mascot floating around it (the Home Tools
  cover). Mascots are decorative; the cards below are the links.
-->
<script lang="ts">
  import { ToolMascot } from '@home-tools/ui/tool-mascot'
  import { HOME, TOOLS } from '@home-tools/ui/tools'

  // Corners, in the order of the cover; delays keep them from bobbing in sync.
  const cast = [...TOOLS, HOME].map((tool, i) => ({ tool, spot: `spot-${i + 1}`, delay: i * 1.3 }))
</script>

<section class="hero" aria-labelledby="hero-title">
  <div class="grid" aria-hidden="true"></div>
  <div class="stage">
    {#each cast as { tool, spot, delay } (tool.id)}
      <span class="cast {spot}"><ToolMascot {tool} bob {delay} /></span>
    {/each}
    <div class="words">
      <h1 id="hero-title" class="wordmark">Home Tools</h1>
      <p class="tagline">Workouts, game servers and the Discord bot, for everyone in the house.</p>
    </div>
  </div>
</section>

<style>
  .hero {
    position: relative;
    overflow: hidden;
    padding-inline: var(--page-gutter);
  }

  /* The cover's graph paper: two sets of 1px lines, fading out toward the
   * edges so the hero blends into the page. */
  .grid {
    position: absolute;
    inset: 0;
    --line: color-mix(in oklab, var(--color-text) 9%, transparent);
    background-image: linear-gradient(var(--line) 1px, transparent 1px),
      linear-gradient(90deg, var(--line) 1px, transparent 1px);
    background-position: center top;
    background-size: var(--space-8) var(--space-8);
    mask-image: radial-gradient(ellipse 75% 85% at 50% 40%, black 35%, transparent 100%);
  }

  .stage {
    position: relative;
    display: grid;
    place-items: center;
    max-width: var(--page-width);
    min-height: clamp(20rem, 55vh, 34rem);
    margin: 0 auto;
    padding-block: calc(var(--cast-size) + var(--space-5)) calc(var(--cast-size) + var(--space-5));
    --cast-size: clamp(4.5rem, 3rem + 7vw, 9rem);
  }

  .words {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-4);
    text-align: center;
  }

  /* The brand, set like the cover: heavy, uppercase, tight. */
  .wordmark {
    font-size: clamp(2.75rem, 1.2rem + 8vw, 7rem);
    font-weight: 900;
    line-height: 0.95;
    letter-spacing: -0.04em;
    text-transform: uppercase;
  }

  .tagline {
    max-width: 28rem;
    margin: 0;
    color: var(--color-text-muted);
    font-size: clamp(var(--text-md), 0.9rem + 0.6vw, var(--text-lg));
    text-wrap: balance;
  }

  .cast {
    position: absolute;
    width: var(--cast-size);
  }

  .spot-1 {
    top: var(--space-5);
    left: 4%;
    rotate: -6deg;
  }
  .spot-2 {
    top: calc(var(--space-5) + 0.75rem);
    right: 5%;
    rotate: 5deg;
  }
  .spot-3 {
    bottom: var(--space-3);
    left: 7%;
    rotate: 4deg;
  }
  .spot-4 {
    bottom: calc(var(--space-3) + 0.5rem);
    right: 3%;
    rotate: -5deg;
  }

  /* Wide screens: the cast spreads toward the corners, like the cover. */
  @media (min-width: 64rem) {
    .spot-1 {
      left: 10%;
    }
    .spot-2 {
      right: 11%;
    }
    .spot-3 {
      left: 15%;
    }
    .spot-4 {
      right: 13%;
    }
  }
</style>
