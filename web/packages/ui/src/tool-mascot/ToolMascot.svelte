<!--
  A tool's mascot: a blob in the tool's color, seeded from the tool's name
  (the same blob code as profile avatars, so it blinks and glances), with
  one prop that says what the tool does: Fitness holds a dumbbell, Games a
  controller, Discord has a speech bubble, Home Tools wears a roof. The
  props are data in ./props.ts, shared with the favicon script.

  It sizes to its container (square). `bob` adds a slow idle float (the
  hero). Any link or button around it makes it hop on hover, tap or
  keyboard focus. All motion stops with reduced motion. Decorative: the
  tool's name is always next to it.
-->
<script lang="ts">
  import { ProfileAvatar } from '../profiles/profile-avatar'
  import type { Tool } from '../tools'
  import { PROPS, partAttrs } from './props'

  let { tool, bob = false, delay = 0 }: { tool: Tool; bob?: boolean; delay?: number } = $props()
</script>

<span class="mascot" class:bob data-accent={tool.color} style:--delay="-{delay}s" aria-hidden="true">
  <span class="hop">
    <ProfileAvatar name={tool.name} />
    <svg class="prop" viewBox="0 0 100 100">
      {#each PROPS[tool.id] ?? [] as part, i (i)}
        <svelte:element this={part.tag} xmlns="http://www.w3.org/2000/svg" class={part.paint} style:--i={i - 1} {...partAttrs(part)} />
      {/each}
    </svg>
  </span>
</span>

<style>
  .mascot {
    position: relative;
    display: block;
    width: 100%;
    aspect-ratio: 1;
  }

  .hop {
    position: absolute;
    inset: 0;
    transform-origin: 50% 90%;
  }

  .prop {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    overflow: visible;
  }

  /* Props are drawn like ink: character, not chrome. */
  .ink,
  .dot {
    fill: var(--color-mascot-prop);
  }

  .roof {
    stroke: var(--color-mascot-prop);
    stroke-width: 4;
    stroke-linejoin: round;
  }

  .skin {
    fill: var(--color-avatar);
  }

  .pad {
    fill: none;
    stroke: var(--color-avatar);
    stroke-width: 2.8;
    stroke-linecap: round;
  }

  .bubble {
    fill: var(--color-surface-raised);
    stroke: var(--color-mascot-prop);
    stroke-width: 2.5;
    stroke-linejoin: round;
  }

  @media (prefers-reduced-motion: no-preference) {
    .bob {
      animation: bob 3.6s var(--delay) infinite ease-in-out alternate;
    }

    .dot {
      animation: typing 1.4s calc(var(--i) * 0.18s) infinite ease-in-out;
    }

    /* Hop when the link or button around it is hovered, tapped or focused. */
    :global(:is(a, button):is(:hover, :focus-visible, :active)) .hop {
      animation: hop 0.6s var(--ease-out);
    }
  }

  @keyframes bob {
    from {
      translate: 0 0;
    }
    to {
      translate: 0 -6%;
    }
  }

  @keyframes typing {
    0%,
    60%,
    100% {
      translate: 0 0;
      opacity: 0.5;
    }
    30% {
      translate: 0 -2px;
      opacity: 1;
    }
  }

  /* Squash, jump, land: the same beat a cartoon character does. */
  @keyframes hop {
    0% {
      scale: 1 1;
      translate: 0 0;
    }
    18% {
      scale: 1.08 0.9;
      translate: 0 0;
    }
    45% {
      scale: 0.95 1.06;
      translate: 0 -14%;
    }
    72% {
      scale: 1.05 0.95;
      translate: 0 0;
    }
    100% {
      scale: 1 1;
      translate: 0 0;
    }
  }
</style>
