// Organic avatar blobs, computed from the profile ID: the same person always
// gets the same shape, and nothing is stored. Decision #22.
//
// The outline is a circle whose radius rises and falls with a few layered
// waves (3, 4 and 5 bumps around the circle, each at a random strength and
// angle). Waves can't make lone dents, and every blob is scaled to the same
// total bumpiness, so none come out plain round or broken-looking.

const WAVES = [3, 4, 5] // bumps per wave; lower = broader lobes
const SPREAD = 0.3 // biggest bump vs deepest dip, as a share of the radius
const RADIUS = 41 // in a 100×100 viewBox; the bumps stay inside it
const SAMPLES = 24 // points along the outline; enough to look smooth

/**
 * Returns an SVG path ("d" attribute) for a smooth blob in a 100×100 viewBox.
 * @param id - The profile's ID, e.g. "xander".
 */
export function blobPath(id: string): string {
  const next = seededRandom(id)
  const waves = WAVES.map((bumps) => ({
    bumps,
    strength: 0.4 + next() * 0.6, // never 0, so every wave shows a little
    angle: next() * 2 * Math.PI,
  }))

  // 1. How far each sample sticks out: the sum of all waves at that angle.
  const offsets: number[] = []
  for (let i = 0; i < SAMPLES; i++) {
    const t = (i / SAMPLES) * 2 * Math.PI
    offsets.push(waves.reduce((sum, w) => sum + w.strength * Math.cos(w.bumps * t + w.angle), 0))
  }

  // 2. Rescale so every blob has exactly SPREAD between its highest bump and
  // deepest dip: equally "blobby" for everyone.
  const min = Math.min(...offsets)
  const max = Math.max(...offsets)
  const mid = (min + max) / 2
  const points = offsets.map((o, i): Point => {
    const t = (i / SAMPLES) * 2 * Math.PI
    const r = RADIUS * (1 + ((o - mid) / (max - min)) * SPREAD)
    return [50 + r * Math.cos(t), 50 + r * Math.sin(t)]
  })

  return smoothPath(points)
}

/** Where and how a blob's eyes sit, and how fast they blink and glance. */
export type BlobFace = {
  x: number // center between the eyes, in the 100×100 viewBox
  y: number
  gap: number // distance from the center to each eye
  rx: number // eye, horizontal radius
  ry: number // eye, vertical radius (taller: oval eyes)
  tilt: number // degrees the pair is tilted
  blinkSec: number // one blink per this many seconds
  glanceSec: number // one look-around loop per this many seconds
  offsetSec: number // where in its loops this face starts, so no two sync up
}

/**
 * Returns the face for a profile's blob. Like the shape, it's computed
 * from the ID (with its own seed, so tweaking faces never changes shapes).
 */
export function blobFace(id: string): BlobFace {
  const next = seededRandom(id + ':face')
  const between = (min: number, max: number) => min + next() * (max - min)
  return {
    x: between(47, 53),
    y: between(43, 48),
    gap: between(11, 14),
    rx: between(5.5, 6.5),
    ry: between(8.5, 10),
    tilt: between(-6, 6),
    blinkSec: between(7, 11),
    glanceSec: between(16, 24),
    offsetSec: between(0, 24),
  }
}

type Point = [number, number]

/**
 * A smooth closed SVG path through points (Catmull-Rom → cubic Bézier).
 * For the segment p1 → p2, the handles follow the line from the previous
 * point to the next one, so the curve never has corners.
 */
function smoothPath(points: Point[]): string {
  const n = points.length
  const at = (i: number) => points[(i + n) % n] // wraps around
  let d = `M ${fmt(at(0))}`
  for (let i = 0; i < n; i++) {
    const [p0, p1, p2, p3] = [at(i - 1), at(i), at(i + 1), at(i + 2)]
    const c1: Point = [p1[0] + (p2[0] - p0[0]) / 6, p1[1] + (p2[1] - p0[1]) / 6]
    const c2: Point = [p2[0] - (p3[0] - p1[0]) / 6, p2[1] - (p3[1] - p1[1]) / 6]
    d += ` C ${fmt(c1)} ${fmt(c2)} ${fmt(p2)}`
  }
  return d + ' Z'
}

/** Returns a function giving repeatable "random" numbers in [0, 1) for seed. */
function seededRandom(seed: string): () => number {
  // FNV-1a hash: turns the seed into a number where every character matters.
  let h = 2166136261
  for (let i = 0; i < seed.length; i++) {
    h ^= seed.charCodeAt(i)
    h = Math.imul(h, 16777619)
  }
  // Each call steps the number forward (a linear congruential generator).
  return () => {
    h = (Math.imul(h, 1664525) + 1013904223) >>> 0
    return h / 2 ** 32
  }
}

const fmt = ([x, y]: Point) => `${x.toFixed(1)} ${y.toFixed(1)}`
