// Generates every app's icons from the tool registry (TOOLS + HOME in
// @home-tools/ui, decisions #46 and #62): each tool's mascot (blob, eyes and
// prop, the same drawing as ToolMascot in the header's tool switcher and on
// the Home Tools page), so tab, home screen and app all match.
//   deploy/icons/: the ZimaOS app tiles, each a PNG and an SVG:
//     home-tools  the Home Tools mascot (the app itself)
//     https       the same mascot holding a padlock (Caddy)
//     proxy       the same mascot holding a plug (the Docker socket proxy)
//
//   pnpm --filter @home-tools/fitness favicon
//
// Per app: public/favicon.svg (browser tabs: the mascot alone, its prop
// light or dark with the browser's theme) and public/apple-touch-icon.png
// (iPhone home screen; iOS ignores SVG icons: the mascot on its tinted grid,
// like the cards on the Home Tools page). PNGs are rendered by headless
// Chrome: set CHROME to its path if it isn't in the macOS default (and
// CHROME_ARGS="--no-sandbox" when running as root, e.g. in a container).
// Rerun when a mascot, prop or color changes; outputs are committed.

import { execFileSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { AVATAR, MASCOT_THEME_STYLE, mascotParts } from '../../../../packages/ui/src/tool-mascot/static.ts'
import { HOME, TOOLS, type Tool } from '../../../../packages/ui/src/tools/index.ts'

const CHROME = process.env.CHROME ?? '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'

const appsDir = fileURLToPath(new URL('../../../', import.meta.url))
const deployIcons = fileURLToPath(new URL('../../../../../deploy/icons/', import.meta.url))

// Each mascot's box in its 100×100 drawing, props included (Discord's
// bubble sticks out top right, the roof out the top), padded a little and
// square, so every tab icon fills its space the same way.
const BOX: Record<string, string> = {
  fitness: '2 2 96 96',
  games: '2 2 96 96',
  discord: '0 -18 112 112',
  tools: '-4 -6 108 108',
  https: '2 2 98 98',
  proxy: '2 2 100 100',
}
const box = (tool: Tool) => BOX[tool.id] ?? '0 0 100 100'

/** Tab icon: the mascot alone on a transparent background. */
function tabIcon(tool: Tool, out: string): void {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="${box(tool)}">${MASCOT_THEME_STYLE}${mascotParts(tool)}</svg>\n`
  writeFileSync(out, svg)
}

/**
 * A 180 px square icon: opaque, edge to edge (iOS and ZimaOS round the
 * corners themselves): the mascot on a light tint of its color with the
 * graph-paper grid of the Home Tools page.
 */
function squareIcon(tool: Tool, out: string): void {
  const color = AVATAR[tool.color]
  const tint = `color-mix(in oklab, ${color} 22%, #fff)`
  const line = `color-mix(in oklab, ${color} 30%, transparent)`
  const html = join(mkdtempSync(join(tmpdir(), 'favicon-')), 'icon.html')
  writeFileSync(
    html,
    `<body style="margin:0"><div style="width:180px;height:180px;display:grid;place-items:center;` +
      `background-color:${tint};background-image:linear-gradient(${line} 1px,transparent 1px),linear-gradient(90deg,${line} 1px,transparent 1px);` +
      `background-size:22.5px 22.5px;background-position:center">` +
      `<svg xmlns="http://www.w3.org/2000/svg" width="124" height="124" viewBox="${box(tool)}" style="overflow:visible">${mascotParts(tool, 'light')}</svg></div></body>`,
  )
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

// The server-side apps' ZimaOS tiles: the Home Tools mascot with each
// app's own prop (PROPS in tool-mascot/props.ts).
mkdirSync(deployIcons, { recursive: true })
const tiles: [string, Tool][] = [
  ['home-tools', HOME],
  ['https', { ...HOME, id: 'https' }],
  ['proxy', { ...HOME, id: 'proxy' }],
]
for (const [file, tool] of tiles) {
  squareIcon(tool, join(deployIcons, `${file}.png`))
  tabIcon(tool, join(deployIcons, `${file}.svg`))
  console.log(`wrote deploy/icons/${file}.png, ${file}.svg`)
}
