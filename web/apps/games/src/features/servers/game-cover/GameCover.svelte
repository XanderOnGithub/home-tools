<!--
  The tall cover on a server card (decision #58): the picture dropped
  into data/games/art when there is one (`src`), else a scene drawn here
  for the game: original art, not the games' own. Minecraft = a blocky
  landscape at midday; Valheim = mountains and pines at dusk.
  Decorative: the server's name is right under it.
-->
<script lang="ts">
  import type { Game } from '@/features/servers'

  let { game, src = null }: { game: Game; src?: string | null } = $props()
  let failed = $state(false)

  /** A stepped skyline: one column per `w` px at each height, closed to the bottom. */
  function steps(heights: number[], w: number, bottom = 400): string {
    let d = `M0 ${bottom} V${heights[0]}`
    heights.forEach((h, i) => {
      d += ` H${i * w} V${h} H${(i + 1) * w}`
    })
    return `${d} V${bottom} Z`
  }

  /** A jagged mountain range through the given (x, y) peaks and valleys. */
  function range(points: [number, number][], bottom = 400): string {
    return `M0 ${bottom} ` + points.map(([x, y]) => `L${x} ${y}`).join(' ') + ` L300 ${bottom} Z`
  }

  /** A pine: three stacked triangles on a short trunk, foot at (x, y). */
  function pine(x: number, y: number, h: number): string {
    const w = h * 0.42
    const tiers = [0, 0.3, 0.55].map((t, i) => {
      const top = y - h + t * h
      const half = w * (0.55 + i * 0.22)
      const base = top + h * 0.42
      return `M${x} ${top} L${x + half} ${base} L${x - half} ${base} Z`
    })
    return tiers.join(' ') + ` M${x - 1.5} ${y - h * 0.08} h3 v${h * 0.08} h-3 Z`
  }

  // Block world: 15 columns of 20 px.
  const FAR = [214, 214, 200, 200, 186, 186, 200, 214, 214, 200, 192, 192, 206, 214, 214]
  const MID = [262, 248, 248, 234, 234, 248, 262, 262, 248, 234, 220, 220, 234, 248, 262]
  const NEAR = [300, 300, 286, 286, 286, 300, 314, 314, 300, 300, 286, 300, 300, 314, 314]
</script>

<div class="cover {game}" aria-hidden="true">
  {#if src && !failed}
    <img {src} alt="" onerror={() => (failed = true)} />
  {:else if game === 'minecraft'}
    <svg viewBox="0 0 300 400" preserveAspectRatio="xMidYMid slice">
      <defs>
        <linearGradient id="mc-sky" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stop-color="#5aa9e6" />
          <stop offset="1" stop-color="#bfe3f7" />
        </linearGradient>
      </defs>
      <rect width="300" height="400" fill="url(#mc-sky)" />
      <rect x="214" y="54" width="36" height="36" fill="#fff6c2" />
      <rect x="220" y="60" width="24" height="24" fill="#ffe680" />
      <!-- Square clouds. -->
      <path d="M30 80h60v12H30z M42 68h24v12H42z M120 110h50v10h-50z M132 100h20v10h-20z" fill="#fff" opacity="0.85" />
      <path d={steps(FAR, 20)} fill="#9cc7a4" />
      <path d={steps(MID, 20)} fill="#5e9e4b" />
      <!-- Blocky trees on the middle hills. -->
      <path d="M58 194h28v28H58z M66 222h12v26H66z M218 176h28v28h-28z M226 204h12v30h-12z" fill="#3f7a33" />
      <path d="M68 222h8v26h-8z M228 204h8v30h-8z" fill="#7a5532" />
      <path d={steps(NEAR, 20)} fill="#7cbf4f" />
      <!-- Dirt under the grass, with a few darker blocks. -->
      <path d={steps(NEAR.map((h) => h + 12), 20)} fill="#8b6239" />
      <path d="M20 340h20v20H20z M100 360h20v20h-20z M180 334h20v20h-20z M240 372h20v20h-20z M140 392h20v8h-20z" fill="#74502d" />
      <path d="M60 380h20v20H60z M200 388h20v12h-20z" fill="#8f8f8f" />
    </svg>
  {:else}
    <svg viewBox="0 0 300 400" preserveAspectRatio="xMidYMid slice">
      <defs>
        <linearGradient id="vh-sky" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stop-color="#14263a" />
          <stop offset="0.55" stop-color="#3d5a6e" />
          <stop offset="0.8" stop-color="#c98a52" />
          <stop offset="1" stop-color="#e8b170" />
        </linearGradient>
        <linearGradient id="vh-mist" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stop-color="#e8d2b0" stop-opacity="0" />
          <stop offset="1" stop-color="#e8d2b0" stop-opacity="0.55" />
        </linearGradient>
      </defs>
      <rect width="300" height="400" fill="url(#vh-sky)" />
      <!-- Stars. -->
      <g fill="#fff">
        <circle cx="40" cy="40" r="1.2" opacity="0.8" />
        <circle cx="96" cy="72" r="0.9" opacity="0.6" />
        <circle cx="150" cy="30" r="1.1" opacity="0.7" />
        <circle cx="210" cy="58" r="0.8" opacity="0.5" />
        <circle cx="262" cy="24" r="1.3" opacity="0.8" />
        <circle cx="244" cy="96" r="0.8" opacity="0.5" />
      </g>
      <!-- A low sun behind the ridges. -->
      <circle cx="196" cy="268" r="34" fill="#f6cf8f" opacity="0.9" />
      <path d={range([[0, 230], [40, 196], [72, 214], [118, 150], [150, 186], [182, 170], [228, 120], [262, 176], [300, 160]])} fill="#5b6f82" />
      <!-- Snow caps on the far range. -->
      <path d="M118 150l-12 18 8-3 4 6 5-8 7 4z M228 120l-14 22 9-4 5 7 6-9 9 5z" fill="#dfe7ee" />
      <path d={range([[0, 286], [36, 244], [70, 262], [110, 226], [160, 270], [200, 236], [244, 262], [300, 228]])} fill="#34495c" />
      <rect y="250" width="300" height="80" fill="url(#vh-mist)" />
      <path d={range([[0, 330], [50, 304], [104, 322], [160, 300], [214, 318], [270, 296], [300, 306]])} fill="#1b2a38" />
      <path
        d={[pine(24, 352, 70), pine(54, 344, 52), pine(250, 350, 76), pine(282, 340, 58), pine(214, 356, 44)].join(' ')}
        fill="#0f1a24"
      />
      <rect y="350" width="300" height="50" fill="#0f1a24" />
    </svg>
  {/if}
</div>

<style>
  .cover {
    position: relative;
    aspect-ratio: 3 / 4;
    overflow: hidden;
    background: var(--color-surface-sunken);
  }

  img,
  svg {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
</style>
