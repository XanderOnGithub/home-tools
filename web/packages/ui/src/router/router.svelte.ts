// A tiny client-side router (History API), instead of a dependency.
//
// - `router.path` is reactive ($state): components re-render on navigation.
// - Internal <a href="/..."> links are intercepted globally, so pages use
//   plain links (real <a>: right-click, middle-click and screen readers
//   all behave normally) and nothing special is needed per link.
// - Back/forward buttons work via popstate.
// - Anchors: a link to #section on the same page is left to the browser
//   (it scrolls); /page#section keeps its hash, for the page to scroll to.
//
// The server must answer every app path with index.html (Vite does in dev;
// the Go server will when it serves the built app).

class Router {
  path = $state(location.pathname)

  constructor() {
    addEventListener('popstate', () => (this.path = location.pathname))
    document.addEventListener('click', (event) => this.#onClick(event))
  }

  /** Goes to path, which may end in #section (path stays the pathname). */
  navigate(path: string) {
    const pathname = path.split('#')[0]
    if (pathname === this.path && !path.includes('#')) return
    history.pushState(null, '', path)
    this.path = pathname
  }

  #onClick(event: MouseEvent) {
    // Leave anything that isn't a plain left click alone (new tab, etc.).
    if (event.defaultPrevented || event.button !== 0) return
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
    const a = (event.target as Element).closest('a')
    if (!a || a.target || a.hasAttribute('download')) return
    const url = new URL(a.href)
    if (url.origin !== location.origin) return
    if (url.hash && url.pathname === location.pathname) return // same page: the browser scrolls
    event.preventDefault()
    this.navigate(url.pathname + url.hash)
  }
}

export const router = new Router()
