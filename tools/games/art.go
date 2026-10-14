package games

import (
	"os"
	"path/filepath"
)

// Server art (decision #58): pictures the owner drops into
// data/games/art/, shown on the server cards. Nothing is configured and
// nothing is downloaded: a file with the right name is picked up on the
// next look. A server's own picture wins over its game's:
//
//	art/<server id>.jpg       this server's cover (e.g. a screenshot of its world)
//	art/<game>.jpg            every server of that game ("minecraft", "valheim")
//	art/<server id>-icon.png  a square icon, same order
//	art/<game>-icon.png
//
// Any of .webp, .jpg, .jpeg, .png. Lookups are a few os.Stat calls on a
// handful of servers; the snapshot (see snapshot.go) caches the result
// for the list, so it's not worth an index.

// ArtKind is which picture: the tall cover or the square icon.
type ArtKind string

const (
	ArtCover ArtKind = "cover"
	ArtIcon  ArtKind = "icon"
)

var allArtKinds = []ArtKind{ArtCover, ArtIcon}

// artTypes maps the extensions we look for to what we answer with, in
// the order they're tried.
var artTypes = []struct{ ext, mime string }{
	{".webp", "image/webp"},
	{".jpg", "image/jpeg"},
	{".jpeg", "image/jpeg"},
	{".png", "image/png"},
}

// ArtPath finds srv's picture of kind and its content type, or "" if
// there's none. Errors other than "not there" (permissions, a folder
// named like a picture) count as none too: art is decoration.
func (s *Store) ArtPath(srv Server, kind ArtKind) (path, mime string) {
	suffix := ""
	if kind == ArtIcon {
		suffix = "-icon"
	}
	dir := filepath.Join(s.dir, "art")
	// ID and game are validated names (no slashes or dots), so they're
	// safe as file names.
	for _, base := range []string{srv.ID, string(srv.Game)} {
		for _, t := range artTypes {
			p := filepath.Join(dir, base+suffix+t.ext)
			if info, err := os.Stat(p); err == nil && info.Mode().IsRegular() {
				return p, t.mime
			}
		}
	}
	return "", ""
}

// artOf lists which pictures srv has, for serverView.
func (s *Store) artOf(srv Server) []ArtKind {
	var out []ArtKind
	for _, k := range allArtKinds {
		if p, _ := s.ArtPath(srv, k); p != "" {
			out = append(out, k)
		}
	}
	return out
}
