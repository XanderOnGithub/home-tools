//go:build webembed

// Package fitnessweb embeds the built fitness UI (dist/) into the Go binary.
// Only with `-tags webembed` (the Docker/`make build` path): dev and
// `make check` don't need a web build, since Vite serves the UI there.
package fitnessweb

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Dist is the built app (index.html at its root), or nil without webembed.
var Dist, _ = fs.Sub(dist, "dist")
