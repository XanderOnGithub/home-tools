// What each tool's mascot holds (decision #63), as plain data, so the
// Svelte mascot (ToolMascot) and the favicon script (static.ts) draw the
// very same shapes. Coordinates are in the blob's 100×100 box; some props
// stick out of it (Discord's bubble, the roof), which is why the mascot's
// SVG has overflow: visible.
//
// `paint` says how a part is colored, so it reads in both themes:
//   ink     the prop's color (--color-mascot-prop: dark on light, light on dark)
//   skin    the mascot's own color (buttons on the controller)
//   pad     a stroke in the mascot's color (the d-pad)
//   bubble  a light speech bubble with an ink outline
//   roof    ink, with a rounded ink stroke so the corners aren't sharp
//   dot     ink; the typing dots (animated in ToolMascot)

export type PropPaint = 'ink' | 'skin' | 'pad' | 'bubble' | 'roof' | 'dot'

export type PropPart =
  | { tag: 'rect'; paint: PropPaint; x: number; y: number; width: number; height: number; rx: number }
  | { tag: 'circle'; paint: PropPaint; cx: number; cy: number; r: number }
  | { tag: 'path'; paint: PropPaint; d: string }

export const PROPS: Record<string, PropPart[]> = {
  // A dumbbell, held low in front.
  fitness: [
    { tag: 'rect', paint: 'ink', x: 24, y: 77.5, width: 52, height: 5, rx: 2.5 },
    { tag: 'rect', paint: 'ink', x: 25, y: 67, width: 9, height: 26, rx: 3 },
    { tag: 'rect', paint: 'ink', x: 66, y: 67, width: 9, height: 26, rx: 3 },
    { tag: 'rect', paint: 'ink', x: 17, y: 71.5, width: 7, height: 17, rx: 2.5 },
    { tag: 'rect', paint: 'ink', x: 76, y: 71.5, width: 7, height: 17, rx: 2.5 },
  ],
  // A controller: body, d-pad, two buttons.
  games: [
    {
      tag: 'path',
      paint: 'ink',
      d: 'M35 67h30a14 14 0 0 1 14 14v1a9 9 0 0 1-16 5.6L60 84H40l-3 3.6A9 9 0 0 1 21 82v-1a14 14 0 0 1 14-14z',
    },
    { tag: 'path', paint: 'pad', d: 'M34 74.5v9M29.5 79h9' },
    { tag: 'circle', paint: 'skin', cx: 64, cy: 76.5, r: 2.4 },
    { tag: 'circle', paint: 'skin', cx: 69.5, cy: 81, r: 2.4 },
  ],
  // A speech bubble, typing.
  discord: [
    { tag: 'path', paint: 'bubble', d: 'M74 -16h24a12 12 0 0 1 0 24H84l-9 8 1-8.4A12 12 0 0 1 74-16z' },
    { tag: 'circle', paint: 'dot', cx: 79, cy: -4, r: 2.6 },
    { tag: 'circle', paint: 'dot', cx: 86, cy: -4, r: 2.6 },
    { tag: 'circle', paint: 'dot', cx: 93, cy: -4, r: 2.6 },
  ],
  // A little roof, with a chimney.
  tools: [
    { tag: 'rect', paint: 'ink', x: 61, y: 1, width: 7, height: 14, rx: 1.5 },
    { tag: 'path', paint: 'roof', d: 'M50 -1L77 20H23z' },
  ],
}

/** The SVG attributes of a part, without its paint. */
export function partAttrs(part: PropPart): Record<string, string | number> {
  const { tag: _tag, paint: _paint, ...attrs } = part
  return attrs
}
