// Shared dropdown menus (native Popover API: the browser closes them on Esc
// or an outside click and puts them in the top layer). Looks: .menu,
// .menu-item, .menu-label, .menu-separator in app.css. This places one
// under its trigger before it opens; call it from `onbeforetoggle`.

/**
 * Pins `menu` 8px under `trigger`, its left or right edge lined up with the
 * trigger's, and kept inside the window. Returns whether it's opening.
 */
export function placeUnder(
  event: ToggleEvent,
  trigger: HTMLElement | undefined,
  menu: HTMLElement | undefined,
  align: 'start' | 'end' = 'end',
): boolean {
  const open = event.newState === 'open'
  if (!open || !trigger || !menu) return open
  const r = trigger.getBoundingClientRect()
  const width = document.documentElement.clientWidth
  menu.style.top = `${r.bottom + 8}px`
  if (align === 'end') {
    menu.style.left = 'auto'
    menu.style.right = `${Math.max(8, width - r.right)}px`
  } else {
    menu.style.right = 'auto'
    menu.style.left = `${Math.max(8, r.left)}px`
  }
  return open
}
