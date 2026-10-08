// Remembers which profile this browser picked, until "Switch profile".
//
// A cookie, not localStorage: localStorage is separate per subdomain, but a
// cookie set on the parent domain (e.g. .example.com) is shared, so picking a
// profile in fitness also picks it in every other tool (decision #24).
// The server never reads it; it's just the browser's memory.

const NAME = 'home_tools_profile'
const ONE_YEAR_SEC = 60 * 60 * 24 * 365

/** The remembered profile ID, or null if none. */
export function rememberedProfileId(): string | null {
  for (const part of document.cookie.split('; ')) {
    const [key, value] = part.split('=')
    if (key === NAME && value) return decodeURIComponent(value)
  }
  return null
}

export function rememberProfile(id: string): void {
  document.cookie = `${NAME}=${encodeURIComponent(id)}; ${attributes(ONE_YEAR_SEC)}`
}

export function forgetProfile(): void {
  document.cookie = `${NAME}=; ${attributes(0)}`
}

/**
 * fitness.example.com → shared with all of .example.com. Hosts without
 * a subdomain (localhost, a bare IP) keep the cookie to themselves.
 */
function attributes(maxAgeSec: number): string {
  const host = location.hostname
  const parts = host.split('.')
  const isIP = /^[\d.]+$/.test(host)
  const domain = !isIP && parts.length >= 3 ? `; domain=.${parts.slice(1).join('.')}` : ''
  const secure = location.protocol === 'https:' ? '; secure' : ''
  return `path=/; max-age=${maxAgeSec}; samesite=lax${domain}${secure}`
}
