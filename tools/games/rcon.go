package games

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

// Minecraft RCON (the Source RCON protocol): TCP, little-endian packets of
//
//	int32 length (of everything after it) | int32 id | int32 type | body | 0 0
//
// Log in with a type-3 packet carrying the password (the reply's id is -1
// if it's wrong), then send a type-2 command and read the reply's body.
const (
	rconAuth    = 3
	rconExec    = 2
	rconMaxSize = 1 << 16 // no real reply comes close; guards against garbage
)

// rconRun logs in and runs one command, returning its output.
func rconRun(ctx context.Context, addr, password, command string) (string, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return "", fmt.Errorf("rcon: %w", err)
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	if err := rconWrite(conn, 1, rconAuth, password); err != nil {
		return "", err
	}
	id, _, err := rconRead(conn)
	if err != nil {
		return "", err
	}
	if id == -1 {
		return "", errors.New("rcon: wrong password")
	}

	if err := rconWrite(conn, 2, rconExec, command); err != nil {
		return "", err
	}
	for {
		id, body, err := rconRead(conn)
		if err != nil {
			return "", err
		}
		if id == 2 {
			return body, nil
		}
	}
}

func rconWrite(w io.Writer, id, kind int32, body string) error {
	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, int32(4+4+len(body)+2))
	binary.Write(&b, binary.LittleEndian, id)
	binary.Write(&b, binary.LittleEndian, kind)
	b.WriteString(body)
	b.Write([]byte{0, 0})
	if _, err := w.Write(b.Bytes()); err != nil {
		return fmt.Errorf("rcon: %w", err)
	}
	return nil
}

func rconRead(r io.Reader) (id int32, body string, err error) {
	var size int32
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return 0, "", fmt.Errorf("rcon: %w", err)
	}
	if size < 10 || size > rconMaxSize {
		return 0, "", fmt.Errorf("rcon: bad packet size %d", size)
	}
	buf := make([]byte, size)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, "", fmt.Errorf("rcon: %w", err)
	}
	id = int32(binary.LittleEndian.Uint32(buf[0:4]))
	return id, string(buf[8 : size-2]), nil
}
