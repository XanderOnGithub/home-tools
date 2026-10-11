//go:build !webembed

package toolsweb

import "io/fs"

// Dist is nil: built without -tags webembed, so the server serves no UI.
var Dist fs.FS
