// Generates the app icons from the same blob code as the profile avatars
// (decision #22), so the icon is "one of us": the blob for the seed "gym".
//
//   pnpm --filter @home-tools/fitness favicon
//
// Writes public/favicon.svg (browser tabs) and public/apple-touch-icon.png
// (iPhone home screen; iOS ignores SVG icons). The PNG is rendered by
// headless Chrome: set CHROME to its path if it isn't in the macOS default.
// Rerun only when the seed, color or blob code changes; outputs are committed.

import { execFileSync } from 'node:child_process'
import { mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { blobFace, blobPath } from '../../src/features/profiles/blob/index.ts'

const SEED = 'gym'
const FILL = '#4ade80' // green accent, dark-theme shade: reads on light and dark tab bars
const EYE = '#1c1917' // same as the avatar's eyes
const EYE_SCALE = 1.35 // bigger than the avatar's, so they survive 16 px
const TOUCH_BG = '#f0fdf4' // iOS fills transparency with black; green-subtle instead
const CHROME = process.env.CHROME ?? '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'

const publicDir = fileURLToPath(new URL('../../public/', import.meta.url))

function blob(): string {
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
  return `<path d="${blobPath(SEED)}" fill="${FILL}"/><g transform="${face}">${eyes}</g>`
}

// Tab icon: cropped to the blob (it spans roughly 9..91 of the 100 box).
const favicon = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="7 7 86 86">${blob()}</svg>\n`
writeFileSync(join(publicDir, 'favicon.svg'), favicon)

// Home-screen icon: opaque square with padding (iOS rounds the corners).
const touch = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect width="100" height="100" fill="${TOUCH_BG}"/><g transform="translate(50 50) scale(0.82) translate(-50 -50)">${blob()}</g></svg>`
const tmp = mkdtempSync(join(tmpdir(), 'favicon-'))
const html = join(tmp, 'touch.html')
writeFileSync(html, `<body style="margin:0">${touch.replace('<svg ', '<svg width="180" height="180" ')}</body>`)
execFileSync(CHROME, [
  '--headless',
  '--disable-gpu',
  '--hide-scrollbars',
  '--window-size=180,180',
  `--screenshot=${join(publicDir, 'apple-touch-icon.png')}`,
  `file://${html}`,
], { stdio: 'ignore' })

console.log('wrote public/favicon.svg, public/apple-touch-icon.png')
