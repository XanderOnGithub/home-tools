package games

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

// Steam's A2S server query (UDP), as Valheim answers it on its query port
// (usually the game port + 1). Every packet starts with 0xFFFFFFFF; the
// server may first answer 'A' with a 4-byte challenge to send back.
var (
	a2sInfo   = append([]byte{0xFF, 0xFF, 0xFF, 0xFF, 'T'}, "Source Engine Query\x00"...)
	a2sPlayer = []byte{0xFF, 0xFF, 0xFF, 0xFF, 'U', 0xFF, 0xFF, 0xFF, 0xFF}
)

// a2sPlayers reads the count and max from A2S_INFO, then names from
// A2S_PLAYER (best effort: Valheim often sends empty names).
func a2sPlayers(ctx context.Context, addr string) (Players, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "udp", addr)
	if err != nil {
		return Players{}, fmt.Errorf("a2s: %w", err)
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	info, err := a2sAsk(conn, a2sInfo, 'I', func(challenge []byte) []byte {
		return append(append([]byte{}, a2sInfo...), challenge...)
	})
	if err != nil {
		return Players{}, err
	}
	p, err := parseA2SInfo(info)
	if err != nil {
		return Players{}, err
	}

	list, err := a2sAsk(conn, a2sPlayer, 'D', func(challenge []byte) []byte {
		return append([]byte{0xFF, 0xFF, 0xFF, 0xFF, 'U'}, challenge...)
	})
	if err == nil {
		p.Names = parseA2SPlayers(list)
	}
	return p, nil
}

// a2sAsk sends req and returns the reply's payload (after the 5-byte
// header) of type want, answering one challenge if the server asks.
func a2sAsk(conn net.Conn, req []byte, want byte, withChallenge func([]byte) []byte) ([]byte, error) {
	buf := make([]byte, 1400)
	for range 2 {
		if _, err := conn.Write(req); err != nil {
			return nil, fmt.Errorf("a2s: %w", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("a2s: %w", err)
		}
		if n < 5 || !bytes.Equal(buf[:4], []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
			return nil, errors.New("a2s: not an A2S reply")
		}
		switch buf[4] {
		case want:
			return append([]byte{}, buf[5:n]...), nil
		case 'A':
			if n < 9 {
				return nil, errors.New("a2s: short challenge")
			}
			req = withChallenge(buf[5:9])
		default:
			return nil, fmt.Errorf("a2s: unexpected reply %q", buf[4])
		}
	}
	return nil, errors.New("a2s: server kept asking for a challenge")
}

// parseA2SInfo skips protocol, name, map, folder, game and app id to reach
// the player count and max.
func parseA2SInfo(b []byte) (Players, error) {
	r := bytes.NewReader(b)
	if _, err := r.ReadByte(); err != nil { // protocol
		return Players{}, errors.New("a2s: short info")
	}
	for range 4 { // name, map, folder, game: null-terminated strings
		if _, err := readCString(r); err != nil {
			return Players{}, errors.New("a2s: short info")
		}
	}
	var appID uint16
	var online, maxPlayers byte
	if binary.Read(r, binary.LittleEndian, &appID) != nil ||
		binary.Read(r, binary.LittleEndian, &online) != nil ||
		binary.Read(r, binary.LittleEndian, &maxPlayers) != nil {
		return Players{}, errors.New("a2s: short info")
	}
	return Players{Online: int(online), Max: int(maxPlayers), Names: []string{}}, nil
}

// parseA2SPlayers returns the non-empty names: count byte, then per player
// index byte, name, int32 score, float32 seconds connected.
func parseA2SPlayers(b []byte) []string {
	names := []string{}
	r := bytes.NewReader(b)
	count, err := r.ReadByte()
	if err != nil {
		return names
	}
	for range count {
		if _, err := r.ReadByte(); err != nil {
			break
		}
		name, err := readCString(r)
		if err != nil {
			break
		}
		if _, err := r.Seek(8, 1); err != nil { // score + duration
			break
		}
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

func readCString(r *bytes.Reader) (string, error) {
	var b []byte
	for {
		c, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if c == 0 {
			return string(b), nil
		}
		b = append(b, c)
	}
}
