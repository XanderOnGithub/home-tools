// Keeps the screen on (Screen Wake Lock API) while a workout is open.
//
// Browsers only allow it on secure origins: HTTPS or localhost. Over plain
// HTTP on the LAN it silently does nothing (see decision #8, TLS). The
// browser also drops the lock whenever the tab is hidden, so it's
// re-requested each time the page becomes visible again.

/** Starts keeping the screen on; call the returned function to stop. */
export function keepScreenOn(): () => void {
  let lock: WakeLockSentinel | null = null
  let stopped = false

  async function request() {
    if (stopped || document.visibilityState !== 'visible' || !('wakeLock' in navigator)) return
    try {
      lock = await navigator.wakeLock.request('screen')
    } catch {
      // Not allowed (insecure origin, battery saver…): just carry on.
    }
  }

  const onVisible = () => void request()
  document.addEventListener('visibilitychange', onVisible)
  void request()

  return () => {
    stopped = true
    document.removeEventListener('visibilitychange', onVisible)
    void lock?.release()
  }
}
