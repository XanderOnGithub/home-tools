package httpx

import "net/http"

// csp allows only this server's own files: scripts and styles from the
// built app (Svelte sets some inline style attributes, hence
// 'unsafe-inline' for styles only), images from here or generated in the
// page, no plugins, no <base> tricks, and no other site may frame a tool.
const csp = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data: blob:; connect-src 'self'; object-src 'none'; " +
	"base-uri 'none'; form-action 'self'; frame-ancestors 'none'"

// Secure wraps the whole server (decision #48). There are no logins (#10),
// so the browser is the only thing standing between a random website and
// the API:
//
//   - Cross-site writes are refused (CSRF). Without this, any page opened
//     on the LAN could POST a Minecraft console command: a text/plain body
//     needs no CORS preflight, and DecodeJSON doesn't look at the
//     Content-Type. Go's CrossOriginProtection checks the browser's
//     Sec-Fetch-Site / Origin headers; GET, HEAD and OPTIONS pass, and so
//     do requests without them (the Discord bot, curl), which no website
//     can make.
//   - Every response gets headers that stop MIME sniffing, leaking URLs in
//     Referer, framing (clickjacking) and loading anything from elsewhere.
func Secure(h http.Handler) http.Handler {
	csrf := http.NewCrossOriginProtection()
	csrf.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, http.StatusForbidden, "cross-site request refused")
	}))
	protected := csrf.Handler(h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hd := w.Header()
		hd.Set("Content-Security-Policy", csp)
		hd.Set("X-Content-Type-Options", "nosniff")
		hd.Set("Referrer-Policy", "same-origin")
		protected.ServeHTTP(w, r)
	})
}
