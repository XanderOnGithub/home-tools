// Generates the app icons from the same blob code as the profile avatars
// (decision #22), so the icon is "one of us": the blob for the seed "gym".
//
//   pnpm --filter @home-tools/fitness favicon
//
// Writes public/favicon.svg (browser tabs), public/apple-touch-icon.png
// (iPhone home screen; iOS ignores SVG icons) and deploy/icons/https.png
// (the same blob in blue, for the Caddy app's tile in ZimaOS). The PNG is rendered by
// headless Chrome: set CHROME to its path if it isn't in the macOS default.
// Rerun only when the seed, color or blob code changes; outputs are committed.

import { execFileSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { blobFace, blobPath } from '../../src/features/profiles/blob/index.ts'

const SEED = 'gym'
const GREEN = { fill: '#4ade80', bg: '#f0fdf4' } // accent (dark-theme shade) + subtle
const BLUE = { fill: '#60a5fa', bg: '#eff6ff' }
const EYE = '#1c1917' // same as the avatar's eyes
const EYE_SCALE = 1.35 // bigger than the avatar's, so they survive 16 px
const CHROME = process.env.CHROME ?? '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'

const publicDir = fileURLToPath(new URL('../../public/', import.meta.url))
const deployIcons = fileURLToPath(new URL('../../../../../deploy/icons/', import.meta.url))

function blob(fill: string): string {
  const f = blobFace(SEED)
  const eyes = [-1, 1]
    .map((side) => {
      const cx = (side * f.gap * 1.1).toFixed(2)
      const rx = (f.rx * EYE_SCALE).toFixed(2)
      const ry = (f.ry * EYE_SCALE).toFixed(2)
      return `<ellipse cx="${cx}" rx="${rx}" ry="${ry}" fill="${EYE}"/>`
    })
    .join('')
  const face = `translate(${f.x.toFixed(2)} ${f.y.toFixed(2)}) rotate(${f.tilt.toFixed(2)})`
  return `<path d="${blobPath(SEED)}" fill="${fill}"/><g transform="${face}">${eyes}</g>`
}

// Tab icon: cropped to the blob (it spans roughly 9..91 of the 100 box).
// The light accent shade reads on both light and dark tab bars.
const favicon = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="7 7 86 86">${blob(GREEN.fill)}</svg>\n`
writeFileSync(join(publicDir, 'favicon.svg'), favicon)

/**
 * A 180 px square icon: opaque background (iOS fills transparency with
 * black) with padding (iOS and ZimaOS round the corners).
 */
function squareIcon({ fill, bg }: { fill: string; bg: string }, out: string): void {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="180" height="180" viewBox="0 0 100 100"><rect width="100" height="100" fill="${bg}"/><g transform="translate(50 50) scale(0.82) translate(-50 -50)">${blob(fill)}</g></svg>`
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

mkdirSync(deployIcons, { recursive: true })
squareIcon(GREEN, join(publicDir, 'apple-touch-icon.png'))
squareIcon(BLUE, join(deployIcons, 'https.png'))

console.log('wrote public/favicon.svg, public/apple-touch-icon.png, deploy/icons/https.png')
