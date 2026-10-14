// Toasts: short confirmations ("Plan saved") and failures of actions that
// happen in the background, so nothing saves silently (decision #57).
// One queue for the whole app; <Toaster> (in AppShell) shows it.
//
// Not for errors a person must act on in a form: those stay next to the
// field. Not for loading failures: those are LoadError on the page.

export type ToastTone = 'success' | 'error' | 'info'

export type Toast = {
  id: number
  message: string
  tone: ToastTone
  action?: { label: string; run: () => void } // e.g. Undo
}

let next = 0
export const toasts = $state<Toast[]>([])
const timers = new Map<number, ReturnType<typeof setTimeout>>()

/** Shows a toast; it leaves by itself (errors stay longer). Returns its id. */
export function toast(message: string, options: { tone?: ToastTone; action?: Toast['action'] } = {}): number {
  const t: Toast = { id: ++next, message, tone: options.tone ?? 'success', action: options.action }
  toasts.push(t)
  if (toasts.length > 3) dismiss(toasts[0].id) // never a pile
  schedule(t)
  return t.id
}

toast.error = (message: string) => toast(message, { tone: 'error' })
toast.info = (message: string) => toast(message, { tone: 'info' })

export function dismiss(id: number): void {
  clearTimeout(timers.get(id))
  timers.delete(id)
  const i = toasts.findIndex((t) => t.id === id)
  if (i >= 0) toasts.splice(i, 1)
}

/** Hovering or focusing a toast pauses it, so it can be read and used. */
export function pause(id: number): void {
  clearTimeout(timers.get(id))
}

export function resume(id: number): void {
  const t = toasts.find((t) => t.id === id)
  if (t) schedule(t)
}

function schedule(t: Toast) {
  clearTimeout(timers.get(t.id))
  const ms = t.tone === 'error' || t.action ? 7000 : 3500
  timers.set(t.id, setTimeout(() => dismiss(t.id), ms))
}
