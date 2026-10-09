// The "rest over" sound: two short tones made with Web Audio (no audio
// file to load). Browsers only allow sound after a user gesture, and iOS
// starts audio suspended, so unlockChime() must run inside a tap (the ✓
// that starts the rest); playChime() then works when the timer ends.
// On iPhones the ring/silent switch may still mute it; vibration is
// Android-only, so this is the cue iOS gets.

let ctx: AudioContext | null = null

/** Call from a tap handler, before the first playChime(). */
export function unlockChime(): void {
  ctx ??= new AudioContext()
  if (ctx.state === 'suspended') void ctx.resume()
}

/** Two rising beeps, ~0.4 s. Silent if unlockChime() never ran. */
export function playChime(): void {
  if (!ctx) return
  const start = ctx.currentTime
  ;[880, 1175].forEach((freq, i) => {
    const t = start + i * 0.18
    const osc = ctx!.createOscillator()
    const gain = ctx!.createGain()
    osc.frequency.value = freq
    // Quick fade in/out: a hard start/stop clicks.
    gain.gain.setValueAtTime(0, t)
    gain.gain.linearRampToValueAtTime(0.3, t + 0.01)
    gain.gain.exponentialRampToValueAtTime(0.001, t + 0.15)
    osc.connect(gain).connect(ctx!.destination)
    osc.start(t)
    osc.stop(t + 0.16)
  })
}
