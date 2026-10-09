//go:build webembed

// Package gamesweb embeds the built games UI (dist/) into the Go binary.
// Only with `-tags webembed`, like the fitness app (see its embed.go).
package gamesweb

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Dist is the built app (index.html at its root), or nil without webembed.
var Dist, _ = fs.Sub(dist, "dist")
