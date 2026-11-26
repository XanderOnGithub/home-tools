// A tool's mascot as a plain SVG string, for places that can't run Svelte:
// the favicon script writes tab icons and home-screen icons with it. Same
// blob, eyes and prop as ToolMascot, without the motion. Imports use .ts
// so Node can run this file directly.
import { blobFace, blobPath } from '../profiles/blob/index.ts'
import type { ProfileColor } from '../profiles/types.ts'
import type { Tool } from '../tools/index.ts'
import { PROPS, partAttrs, type PropPaint } from './props.ts'

// --color-avatar / --color-avatar-eye / --color-mascot-prop in tokens.css.
export const AVATAR: Record<ProfileColor, string> = {
  green: '#4ade80',
  blue: '#60a5fa',
  orange: '#fb923c',
  purple: '#c084fc',
}
const EYE = '#18181b'
const PROP_LIGHT = '#18181b' // on a light background
const PROP_DARK = '#e4e4e7' // on a dark one
const BUBBLE_LIGHT = '#ffffff'
const BUBBLE_DARK = '#27272a'

/**
 * The mascot's parts (blob, eyes, prop) in its 100×100 box, as SVG markup.
 * `theme: 'auto'` follows the viewer's light/dark setting with classes and
 * a media query (tab icons); 'light' bakes in the light colors (PNGs).
 */
export function mascotParts(tool: Tool, theme: 'auto' | 'light' = 'auto'): string {
  const face = blobFace(tool.name)
  const skin = AVATAR[tool.color]
  const eyes = [-1, 1]
    .map((side) => `<ellipse cx="${side * face.gap}" rx="${face.rx}" ry="${face.ry}" fill="${EYE}"/>`)
    .join('')
  const paint = (p: PropPaint): string => {
    const ink = theme === 'auto' ? 'class="ink"' : `fill="${PROP_LIGHT}"`
    switch (p) {
      case 'skin':
        return `fill="${skin}"`
      case 'pad':
        return `fill="none" stroke="${skin}" stroke-width="2.8" stroke-linecap="round"`
      case 'bubble':
        return theme === 'auto'
          ? 'class="bubble" stroke-width="2.5" stroke-linejoin="round"'
          : `fill="${BUBBLE_LIGHT}" stroke="${PROP_LIGHT}" stroke-width="2.5" stroke-linejoin="round"`
      case 'line':
        return theme === 'auto'
          ? 'class="ink-stroke" fill="none" stroke-width="5" stroke-linecap="round"'
          : `fill="none" stroke="${PROP_LIGHT}" stroke-width="5" stroke-linecap="round"`
      case 'roof':
        return theme === 'auto'
          ? 'class="ink ink-stroke" stroke-width="4" stroke-linejoin="round"'
          : `fill="${PROP_LIGHT}" stroke="${PROP_LIGHT}" stroke-width="4" stroke-linejoin="round"`
      default:
        return ink
    }
  }
  const prop = (PROPS[tool.id] ?? [])
    .map((part) => {
      const attrs = Object.entries(partAttrs(part))
        .map(([k, v]) => `${k}="${v}"`)
        .join(' ')
      return `<${part.tag} ${attrs} ${paint(part.paint)}/>`
    })
    .join('')
  return (
    `<path d="${blobPath(tool.name)}" fill="${skin}"/>` +
    `<g transform="translate(${face.x} ${face.y}) rotate(${face.tilt})">${eyes}</g>` +
    prop
  )
}

/** The <style> that colors an 'auto' mascot's prop for light and dark. */
export const MASCOT_THEME_STYLE =
  `<style>.ink{fill:${PROP_LIGHT}}.ink-stroke{stroke:${PROP_LIGHT}}.bubble{fill:${BUBBLE_LIGHT};stroke:${PROP_LIGHT}}` +
  `@media (prefers-color-scheme:dark){.ink{fill:${PROP_DARK}}.ink-stroke{stroke:${PROP_DARK}}.bubble{fill:${BUBBLE_DARK};stroke:${PROP_DARK}}}</style>`
