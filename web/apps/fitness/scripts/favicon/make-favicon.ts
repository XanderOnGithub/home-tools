// Generates every app's icons from the same blob code as the profile
// avatars (decision #22), so each icon is "one of us": each tool's seed
// and color come from TOOLS + HOME in @home-tools/ui (decision #46), the
// same list the tools home page draws its shortcuts from.
//   deploy/icons/https.png: the fitness blob in blue (Caddy's ZimaOS tile)
//
//   pnpm --filter @home-tools/fitness favicon
//
// Per app: public/favicon.svg (browser tabs) and public/apple-touch-icon.png
// (iPhone home screen; iOS ignores SVG icons). PNGs are rendered by headless
// Chrome: set CHROME to its path if it isn't in the macOS default.
// Rerun only when a seed, color or the blob code changes; outputs are committed.

import { execFileSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { blobFace, blobPath } from '../../../../packages/ui/src/profiles/blob/index.ts'
import type { ProfileColor } from '../../../../packages/ui/src/profiles/types.ts'
import { HOME, TOOLS } from '../../../../packages/ui/src/tools/index.ts'

// Accent (dark-theme shade, reads on light and dark tab bars) + its subtle bg.
const COLORS: Record<ProfileColor, { fill: string; bg: string }> = {
  green: { fill: '#4ade80', bg: '#f0fdf4' },
  blue: { fill: '#60a5fa', bg: '#eff6ff' },
  orange: { fill: '#fb923c', bg: '#fff7ed' },
  purple: { fill: '#c084fc', bg: '#faf5ff' },
}
const EYE = '#1c1917' // same as the avatar's eyes
const EYE_SCALE = 1.35 // bigger than the avatar's, so they survive 16 px
const CHROME = process.env.CHROME ?? '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'

const appsDir = fileURLToPath(new URL('../../../', import.meta.url))
const deployIcons = fileURLToPath(new URL('../../../../../deploy/icons/', import.meta.url))

function blob(seed: string, fill: string): string {
  const f = blobFace(seed)
  const eyes = [-1, 1]
    .map((side) => {
      const cx = (side * f.gap * 1.1).toFixed(2)
      const rx = (f.rx * EYE_SCALE).toFixed(2)
      const ry = (f.ry * EYE_SCALE).toFixed(2)
      return `<ellipse cx="${cx}" rx="${rx}" ry="${ry}" fill="${EYE}"/>`
    })
    .join('')
  const face = `translate(${f.x.toFixed(2)} ${f.y.toFixed(2)}) rotate(${f.tilt.toFixed(2)})`
  return `<path d="${blobPath(seed)}" fill="${fill}"/><g transform="${face}">${eyes}</g>`
}

/** Tab icon: cropped to the blob (it spans roughly 9..91 of the 100 box). */
function tabIcon(seed: string, fill: string, out: string): void {
  writeFileSync(out, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="7 7 86 86">${blob(seed, fill)}</svg>\n`)
}

/**
 * A 180 px square icon: opaque background (iOS fills transparency with
 * black) with padding (iOS and ZimaOS round the corners).
 */
function squareIcon(seed: string, { fill, bg }: { fill: string; bg: string }, out: string): void {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="180" height="180" viewBox="0 0 100 100"><rect width="100" height="100" fill="${bg}"/><g transform="translate(50 50) scale(0.82) translate(-50 -50)">${blob(seed, fill)}</g></svg>`
  const html = join(mkdtempSync(join(tmpdir(), 'favicon-')), 'icon.html')
  writeFileSync(html, `<body style="margin:0">${svg}</body>`)
  execFileSync(CHROME, [
    '--headless',
    '--disable-gpu',
    '--hide-scrollbars',
    '--window-size=180,180',
    `--screenshot=${out}`,
    `file://${html}`,
  ], { stdio: 'ignore' })
}

for (const { id: app, seed, color } of [...TOOLS, HOME]) {
  const colors = COLORS[color]
  const pub = join(appsDir, app, 'public')
  mkdirSync(pub, { recursive: true })
  tabIcon(seed, colors.fill, join(pub, 'favicon.svg'))
  squareIcon(seed, colors, join(pub, 'apple-touch-icon.png'))
  console.log(`wrote ${app}/public/favicon.svg, apple-touch-icon.png`)
}
mkdirSync(deployIcons, { recursive: true })
squareIcon('gym', COLORS.blue, join(deployIcons, 'https.png'))
console.log('wrote deploy/icons/https.png')
