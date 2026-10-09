package httpx

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// SPA serves a built single-page app from fsys (the folder holding
// index.html). Real files are served as-is; any other path gets index.html,
// so the app's own router handles URLs like /plans/abc on a reload.
//
// Vite puts content-hashed files under assets/, so they never change and
// may be cached forever; index.html must always be revalidated, or a
// browser keeps loading an old build.
func SPA(fsys fs.FS) http.Handler {
	files := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if info, err := fs.Stat(fsys, name); name == "" || err != nil || info.IsDir() {
			w.Header().Set("Cache-Control", "no-cache")
			http.ServeFileFS(w, r, fsys, "index.html")
			return
		}
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		files.ServeHTTP(w, r)
	})
}
