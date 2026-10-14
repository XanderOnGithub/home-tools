// Generates every app's icons from the tool registry (TOOLS + HOME in
// @home-tools/ui, decisions #46 and #57): each tool's glyph in white on a
// rounded tile of its color, the same drawing as ToolIcon in the header's
// tool switcher and on the Home Tools page, so tab, home screen and app
// all match.
//   deploy/icons/https.png: the Home Tools tile (Caddy's ZimaOS tile)
//
//   pnpm --filter @home-tools/fitness favicon
//
// Per app: public/favicon.svg (browser tabs) and public/apple-touch-icon.png
// (iPhone home screen; iOS ignores SVG icons). PNGs are rendered by headless
// Chrome: set CHROME to its path if it isn't in the macOS default (and
// CHROME_ARGS="--no-sandbox" when running as root, e.g. in a container).
// Rerun only when a glyph or color changes; outputs are committed.

import { execFileSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import type { ProfileColor } from '../../../../packages/ui/src/profiles/types.ts'
import { HOME, TOOLS, type Tool } from '../../../../packages/ui/src/tools/index.ts'

// --color-tile in tokens.css: the same in both themes, white glyph ≥ 3:1.
const TILE: Record<ProfileColor, string> = {
  green: '#16a34a',
  blue: '#2563eb',
  orange: '#ea580c',
  purple: '#9333ea',
}
const CHROME = process.env.CHROME ?? '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'

const appsDir = fileURLToPath(new URL('../../../', import.meta.url))
const deployIcons = fileURLToPath(new URL('../../../../../deploy/icons/', import.meta.url))

/** The glyph, centered on a 24-unit tile (the same 62% as ToolIcon). */
function glyph(tool: Tool, stroke: number): string {
  const s = 0.62
  const o = (24 - 24 * s) / 2
  return `<g transform="translate(${o} ${o}) scale(${s})" fill="none" stroke="#fff" stroke-width="${stroke}" stroke-linecap="round" stroke-linejoin="round"><path d="${tool.icon}"/></g>`
}

/** Tab icon: a rounded tile (28% corners, like ToolIcon). */
function tabIcon(tool: Tool, out: string): void {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect width="24" height="24" rx="6.7" fill="${TILE[tool.color]}"/>${glyph(tool, 2.2)}</svg>\n`
  writeFileSync(out, svg)
}

/**
 * A 180 px square icon: opaque, edge to edge (iOS and ZimaOS round the
 * corners themselves), the glyph a little smaller for their margins.
 */
function squareIcon(tool: Tool, out: string): void {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="180" height="180" viewBox="0 0 24 24"><rect width="24" height="24" fill="${TILE[tool.color]}"/><g transform="translate(12 12) scale(0.9) translate(-12 -12)">${glyph(tool, 1.8)}</g></svg>`
  const html = join(mkdtempSync(join(tmpdir(), 'favicon-')), 'icon.html')
  writeFileSync(html, `<body style="margin:0">${svg}</body>`)
  execFileSync(CHROME, [
    ...(process.env.CHROME_ARGS?.split(' ').filter(Boolean) ?? []),
    '--headless',
    '--disable-gpu',
    '--hide-scrollbars',
    '--window-size=180,180',
    `--screenshot=${out}`,
    `file://${html}`,
  ], { stdio: 'ignore' })
}

for (const tool of [...TOOLS, HOME]) {
  const pub = join(appsDir, tool.id, 'public')
  mkdirSync(pub, { recursive: true })
  tabIcon(tool, join(pub, 'favicon.svg'))
  squareIcon(tool, join(pub, 'apple-touch-icon.png'))
  console.log(`wrote ${tool.id}/public/favicon.svg, apple-touch-icon.png`)
}

mkdirSync(deployIcons, { recursive: true })
squareIcon(HOME, join(deployIcons, 'https.png'))
console.log('wrote deploy/icons/https.png')
