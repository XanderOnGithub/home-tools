<!--
  An exercise's photos as a looping "how it moves" demo. The catalog has
  two shots per exercise (start and end of the movement), so crossfading
  between them reads like the motion itself. Big: full width, 3:2 like
  the photos.

  - Auto-plays; the pause button stops it (WCAG 2.2.2: moving content
    must be pausable). With reduced motion it starts paused, no fade.
  - Tapping the photo shows the other shot and pauses.
  - Missing photos hide the whole block instead of a broken image.

  Remount per exercise ({#key}) so it starts at the first shot.
-->
<script lang="ts">
  import { prefersReducedMotion } from 'svelte/motion'

  let { images, name }: { images: string[]; name: string } = $props()

  const FRAME_MS = 1400 // per shot; slow enough to see each position

  let frame = $state(0)
  let paused = $state(prefersReducedMotion.current)
  let failed = $state(false)
  let animated = $derived(images.length > 1)
  let playing = $derived(animated && !paused)

  $effect(() => {
    if (!playing) return
    const timer = setInterval(() => (frame = (frame + 1) % images.length), FRAME_MS)
    return () => clearInterval(timer)
  })

  function step() {
    paused = true
    frame = (frame + 1) % images.length
  }
</script>

{#if images.length > 0 && !failed}
  <figure class="photos">
    <button
      type="button"
      class="frames"
      onclick={step}
      disabled={!animated}
      aria-label="{name}, photo {frame + 1} of {images.length}. Show the other photo"
    >
      {#each images as src, i (src)}
        <img src="/images/{src}" alt="" class:shown={i === frame} onerror={() => (failed = true)} />
      {/each}
    </button>

    {#if animated}
      <div class="controls">
        <span class="dots" aria-hidden="true">
          {#each images as src, i (src)}<span class:on={i === frame}></span>{/each}
        </span>
        <button
          type="button"
          class="play"
          aria-label={playing ? 'Pause photos' : 'Play photos'}
          onclick={() => (paused = !paused)}
        >
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d={playing ? 'M8 5v14 M16 5v14' : 'M8 5l11 7-11 7z'} />
          </svg>
        </button>
      </div>
    {/if}
  </figure>
{/if}

<style>
  .photos {
    position: relative;
    margin: 0;
  }

  /* Both shots stacked; only the current one is opaque. */
  .frames {
    display: grid;
    width: 100%;
    aspect-ratio: 3 / 2;
    padding: 0;
    overflow: hidden;
    border: none;
    border-radius: var(--radius-lg);
    background: #fff; /* the photos have white backgrounds; no flash in dark mode */
    cursor: pointer;
  }

  .frames:disabled {
    cursor: default;
  }

  .frames img {
    grid-area: 1 / 1;
    width: 100%;
    height: 100%;
    object-fit: contain;
    opacity: 0;
  }

  .frames img.shown {
    opacity: 1;
  }

  @media (prefers-reduced-motion: no-preference) {
    .frames img {
      transition: opacity 400ms var(--ease-out);
    }
  }

  .controls {
    position: absolute;
    right: var(--space-2);
    bottom: var(--space-2);
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding-left: var(--space-3);
    border-radius: var(--radius-full);
    background: rgb(0 0 0 / 0.55);
    color: #fff;
  }

  .dots {
    display: flex;
    gap: var(--space-1);
  }

  .dots span {
    width: 0.4rem;
    height: 0.4rem;
    border-radius: var(--radius-full);
    background: rgb(255 255 255 / 0.45);
  }

  .dots .on {
    background: #fff;
  }

  .play {
    display: grid;
    place-items: center;
    width: var(--touch-target);
    height: var(--touch-target);
    padding: 0;
    border: none;
    border-radius: var(--radius-full);
    background: none;
    color: inherit;
    cursor: pointer;
  }

  .play svg {
    width: 1.1rem;
    height: 1.1rem;
    fill: currentColor;
    stroke: currentColor;
    stroke-width: 2.5;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
</style>
