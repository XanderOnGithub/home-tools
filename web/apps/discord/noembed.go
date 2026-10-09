//go:build !webembed

package discordweb

import "io/fs"

// Dist is nil: built without -tags webembed, so the server serves no UI.
var Dist fs.FS
