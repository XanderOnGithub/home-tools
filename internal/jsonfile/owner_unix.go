//go:build unix

package jsonfile

import (
	"os"
	"syscall"
)

// keepOwner gives f the owner of the file at path, if it exists, so
// replacing another program's file (a game's whitelist) doesn't take it
// away from that program. Best effort: only root may change owners, and
// a file we already own needs nothing.
func keepOwner(f *os.File, path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || int(st.Uid) == os.Getuid() && int(st.Gid) == os.Getgid() {
		return
	}
	// Not root: EPERM. Nothing more to do; the write itself still succeeds.
	_ = f.Chown(int(st.Uid), int(st.Gid))
}
