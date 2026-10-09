//go:build !unix

package jsonfile

import "os"

// keepOwner does nothing where files have no Unix owner.
func keepOwner(*os.File, string) {}
