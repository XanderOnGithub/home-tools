<!--
  A gray placeholder shaped like the content that's coming, with a slow
  shimmer (none with reduced motion), so a page has its layout from the
  first frame instead of flashing blank (decision #57). Hidden from screen
  readers: wrap a page's skeleton in <SkeletonGroup label="…">, which
  announces "Loading …" once.

  width/height take any CSS length; `lines` draws that many text lines
  (the last one shorter), sized to the text around it.
-->
<script lang="ts">
  let {
    width = '100%',
    height = '1em',
    radius = 'md',
    lines = 0,
  }: {
    width?: string
    height?: string
    radius?: 'sm' | 'md' | 'lg' | 'full'
    lines?: number
  } = $props()
</script>

{#if lines > 0}
  <span class="lines" aria-hidden="true" style:width>
    {#each { length: lines } as _, i (i)}
      <span class="bone line" style:width={i === lines - 1 && lines > 1 ? '60%' : '100%'}></span>
    {/each}
  </span>
{:else}
  <span class="bone {radius}" aria-hidden="true" style:width style:height></span>
{/if}

<style>
  .bone {
    display: block;
    background: linear-gradient(
        90deg,
        transparent 0%,
        light-dark(rgb(255 255 255 / 0.55), rgb(255 255 255 / 0.04)) 50%,
        transparent 100%
      )
      var(--color-surface-sunken);
    background-size: 200% 100%;
    background-repeat: no-repeat;
    animation: shimmer var(--duration-slow) ease-in-out infinite;
  }

  .sm {
    border-radius: var(--radius-sm);
  }
  .md {
    border-radius: var(--radius-md);
  }
  .lg {
    border-radius: var(--radius-lg);
  }
  .full {
    border-radius: var(--radius-full);
  }

  .lines {
    display: flex;
    flex-direction: column;
    gap: 0.5em;
  }

  /* A text line: the x-height-ish band of a line of text. */
  .line {
    height: 0.75em;
    border-radius: var(--radius-sm);
  }

  @keyframes shimmer {
    from {
      background-position: 150% 0;
    }
    to {
      background-position: -50% 0;
    }
  }

  /* --duration-slow is 0 with reduced motion: a still gray block. */
  @media (prefers-reduced-motion: reduce) {
    .bone {
      animation: none;
    }
  }
</style>
