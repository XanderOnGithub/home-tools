// Small wrapper around fetch for the Go API. Every failure becomes an
// ApiError whose message is safe to show a person:
//   - 400: the server's own message (written for people, e.g. "unknown color")
//   - network down: "Couldn't reach the server…"
//   - a route the server doesn't have: "…may need a restart" (the Go server
//     doesn't reload code by itself, so a new endpoint 404s until restarted)
//   - anything else: a generic message (details stay in the server log)

export class ApiError extends Error {
  /** HTTP status, or 0 if the server couldn't be reached at all. */
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  let res: Response
  try {
    res = await fetch(path, {
      method,
      headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch {
    throw new ApiError(0, "Couldn't reach the server. Check that it's running.")
  }
  if (!res.ok) {
    // Our handlers always answer errors in JSON. Go's router answers an
    // unknown route (404) or method (405) in plain text, so plain text
    // means the server is older than this page.
    const isJSON = res.headers.get('Content-Type')?.includes('application/json')
    const data = isJSON ? await res.json().catch(() => null) : null
    let message = 'Something went wrong.'
    if (res.status === 400 && data?.error) message = data.error
    else if ((res.status === 404 || res.status === 405) && !isJSON)
      message = 'The server may be out of date. Try restarting it.'
    throw new ApiError(res.status, message)
  }
  return res.json()
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  put: <T>(path: string, body: unknown) => request<T>('PUT', path, body),
  post: <T>(path: string, body: unknown) => request<T>('POST', path, body),
}
