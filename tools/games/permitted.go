package games

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/XanderOnGithub/home-tools/internal/httpx"
	"github.com/XanderOnGithub/home-tools/internal/jsonfile"
)

// Valheim's whitelist is a file, permittedlist.txt: one SteamID64 per
// line, "//" lines are comments (decision #42). The game reads it when it
// starts, so changes take effect after the next restart; every answer
// says so. The file is edited in place (comments and order kept) and
// replaced atomically, so the game never reads half a list.

const permittedListName = "permittedlist.txt"

// steamID64 is an individual Steam account's 64-bit ID: 17 digits, always
// starting 7656119 (the base 76561197960265728 plus the account number).
var steamID64 = regexp.MustCompile(`^7656119[0-9]{10}$`)

// permittedList is the parsed file: its lines as they are (to write back
// unchanged) and the IDs on it, for O(1) lookups.
type permittedList struct {
	lines []string
	ids   map[string]struct{}
}

// entryID returns the Steam ID on a line, or "" for comments and blanks.
// Newer Valheim versions may write "Steam_<id>"; both count as the same ID.
func entryID(line string) string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "//") {
		return ""
	}
	return strings.TrimPrefix(line, "Steam_")
}

// readPermitted parses the file at path. A missing file is an empty list
// (Valheim creates it on its first start; we may get there first).
func readPermitted(path string) (permittedList, os.FileMode, error) {
	l := permittedList{ids: make(map[string]struct{})}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return l, 0o644, nil
	}
	if err != nil {
		return l, 0, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return l, 0, err
	}
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n")) // written on Windows
	for line := range strings.SplitSeq(strings.TrimRight(string(data), "\n"), "\n") {
		l.lines = append(l.lines, line)
		if id := entryID(line); id != "" {
			l.ids[id] = struct{}{}
		}
	}
	return l, info.Mode().Perm(), nil
}

// players returns the IDs on the list, in file order.
func (l permittedList) players() []string {
	out := []string{}
	for _, line := range l.lines {
		if id := entryID(line); id != "" {
			out = append(out, id)
		}
	}
	return out
}

// bytes is the file's new content: its lines, newline-terminated.
func (l permittedList) bytes() []byte {
	if len(l.lines) == 0 {
		return nil
	}
	return []byte(strings.Join(l.lines, "\n") + "\n")
}

// editPermitted adds or removes id on srv's list and says what happened.
// h.lists serializes edits (read-modify-write), so two at once can't
// lose one.
func (h *handlers) editPermitted(srv Server, id string, add bool) (commandResult, error) {
	h.lists.Lock()
	defer h.lists.Unlock()
	l, perm, err := readPermitted(srv.PermittedList)
	if err != nil {
		return commandResult{}, err
	}
	_, listed := l.ids[id]
	switch {
	case add && listed:
		return commandResult{Output: id + " is already on the permitted list."}, nil
	case !add && !listed:
		return commandResult{Output: id + " isn't on the permitted list."}, nil
	case add:
		if len(l.lines) == 0 {
			l.lines = append(l.lines, "// List permitted players ID  ONE per line") // Valheim's own header
		}
		l.lines = append(l.lines, id)
	default:
		kept := l.lines[:0]
		for _, line := range l.lines {
			if entryID(line) != id {
				kept = append(kept, line)
			}
		}
		l.lines = kept
	}
	if err := jsonfile.WriteFile(srv.PermittedList, l.bytes(), perm); err != nil {
		return commandResult{}, err
	}
	h.log.Info("valheim permitted list", "server", srv.ID, "player", id, "add", add)
	if add {
		return commandResult{Output: "Added " + id + " to the permitted list. They can join after the next server restart.", RestartNeeded: true}, nil
	}
	return commandResult{Output: "Removed " + id + " from the permitted list. That takes effect after the next server restart.", RestartNeeded: true}, nil
}

// permittedServer looks up {id} and checks it's a Valheim server with a
// permitted list set up, answering the error itself when it isn't.
func (h *handlers) permittedServer(w http.ResponseWriter, srv Server) bool {
	if srv.PermittedList == "" {
		httpx.WriteError(w, http.StatusConflict, "the whitelist isn't set up for this server (permitted_list)")
		return false
	}
	return true
}

// permittedFailed answers a file error: the details are ours (paths), so
// they go to the log and the person gets what to check.
func (h *handlers) permittedFailed(w http.ResponseWriter, srv Server, err error) {
	h.log.Error("valheim permitted list", "server", srv.ID, "err", err)
	httpx.WriteError(w, http.StatusBadGateway, fmt.Sprintf("couldn't open %s's permitted list; is its folder mounted into Home Tools?", srv.Name))
}
